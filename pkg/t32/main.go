package t32

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"

	graphicsynthesizer "github.com/anasrar/chihuahua/pkg/graphic_synthesizer"
	"github.com/anasrar/chihuahua/pkg/utils"
)

type Header struct {
	TextureTotal       uint32 `json:"texture_total"`
	ImageUploadWidth   uint16 `json:"image_upload_width"`
	ImageUploadHeight  uint16 `json:"image_upload_height"`
	ImageDMAOffset     uint32 `json:"image_dma_offset"`
	PaletteDMAOffset   uint32 `json:"palette_dma_offset"`
	ImagePacketTotal   uint16 `json:"image_packet_total"`
	PalettePacketTotal uint16 `json:"palette_packet_total"`
	Reserved0          uint32 `json:"reserved_0"`
	Reserved1          uint32 `json:"reserved_1"`
	// PaletteTotal is the number of palettes owned by the first texture.
	PaletteTotal uint32 `json:"palette_total"`
}

// TextureLink precedes every texture descriptor except the first one.
type TextureLink struct {
	Index        uint32 `json:"index"`
	PaletteTotal uint32 `json:"palette_total"`
}

type Palette struct {
	Colors []RGBA `json:"colors" skip:""`
}

type Texture struct {
	GSTex0Raw uint64 `json:"gs_tex_0_raw"`
	GSTex1Raw uint32 `json:"gs_tex_1_raw"`
	Reserved  uint32 `json:"reserved"`

	GSTex0   GSTex0    `json:"gs_tex_0" skip:""`
	GSTex1   GSTex1    `json:"gs_tex_1" skip:""`
	Index    uint32    `json:"index" skip:""`
	Width    int       `json:"width" skip:""`
	Height   int       `json:"height" skip:""`
	Indices  []byte    `json:"indices" skip:""`
	Palettes []Palette `json:"palettes" skip:""`
}

func (t *Texture) IsIndexed() bool {
	return t != nil && (t.GSTex0.PSM == PSMT8 || t.GSTex0.PSM == PSMT4)
}

type Container struct {
	Offset uint32 `json:"offset" skip:""`
	Header
	Textures []Texture `json:"textures" skip:""`
	ImageDMA DMAChain  `json:"image_dma" skip:""`
}

type DecodedTexture struct {
	Width   int              `json:"width"`
	Height  int              `json:"height"`
	Format  PixelStorageMode `json:"format"`
	Indices []byte           `json:"indices"`
	Palette []RGBA           `json:"palette"`
	Pixels  []RGBA           `json:"pixels"`
}

func New() *Container {
	return &Container{
		Textures: []Texture{},
		ImageDMA: DMAChain{Transfers: []DMATransfer{}},
	}
}

func (c *Container) unmarshal(stream io.ReadSeeker) error {
	if c == nil {
		return ErrContainerIsNil
	}
	if stream == nil {
		return errors.New("stream is nil")
	}

	streamSize, err := utils.SeekerSize(stream)
	if err != nil {
		return err
	}
	baseOffset := uint64(c.Offset)
	if err := ValidateRegion(streamSize, baseOffset, HeaderSize, "T32 header"); err != nil {
		return err
	}

	parsed := New()
	parsed.Offset = c.Offset
	if err := unmarshalAt(stream, baseOffset, &parsed.Header); err != nil {
		return fmt.Errorf("T32 header: %w", err)
	}
	if err := ValidateHeader(parsed.Header, streamSize-baseOffset); err != nil {
		return err
	}

	descriptorOffset := baseOffset + HeaderSize
	paletteTotal := uint64(0)
	paletteCounts := make([]int, 0, parsed.TextureTotal)
	for i := 0; i < int(parsed.TextureTotal); i++ {
		link := TextureLink{Index: uint32(i), PaletteTotal: parsed.PaletteTotal}
		if i > 0 {
			if err := unmarshalAt(stream, descriptorOffset, &link); err != nil {
				return fmt.Errorf("texture link %d: %w", i, err)
			}
			descriptorOffset += TextureLinkSize
		}
		if link.PaletteTotal == 0 {
			return fmt.Errorf("texture %d has no palettes", i)
		}
		if descriptorOffset+DescriptorSize > baseOffset+uint64(parsed.ImageDMAOffset) {
			return fmt.Errorf(
				"texture descriptor %d ends at %#x, after image DMA offset %#x",
				i,
				descriptorOffset+DescriptorSize-baseOffset,
				parsed.ImageDMAOffset,
			)
		}

		var texture Texture
		if err := unmarshalAt(stream, descriptorOffset, &texture); err != nil {
			return fmt.Errorf("texture descriptor %d: %w", i, err)
		}
		descriptorOffset += DescriptorSize
		if texture.Reserved != 0 {
			return fmt.Errorf(
				"%w: texture %d reserved field is %#x",
				ErrUnsupportedT32,
				i,
				texture.Reserved,
			)
		}
		texture.GSTex0 = DecodeGSTex0(texture.GSTex0Raw)
		texture.GSTex1 = DecodeGSTex1(texture.GSTex1Raw)
		if err := ValidateTexture(texture); err != nil {
			return fmt.Errorf("texture %d: %w", i, err)
		}
		texture.Index = link.Index
		texture.Width = texture.GSTex0.Width()
		texture.Height = texture.GSTex0.Height()
		texture.Palettes = make([]Palette, 0, link.PaletteTotal)
		parsed.Textures = append(parsed.Textures, texture)
		paletteCounts = append(paletteCounts, int(link.PaletteTotal))
		paletteTotal += uint64(link.PaletteTotal)
	}
	if paletteTotal != uint64(parsed.PalettePacketTotal) {
		return fmt.Errorf(
			"texture palette counts total %d, but header contains %d palette packets",
			paletteTotal,
			parsed.PalettePacketTotal,
		)
	}

	parsed.ImageDMA, err = ReadDMAChain(
		stream,
		streamSize,
		baseOffset,
		parsed.ImageDMAOffset,
		parsed.ImagePacketTotal,
		"image",
	)
	if err != nil {
		return err
	}
	paletteDMA, err := ReadDMAChain(
		stream,
		streamSize,
		baseOffset,
		parsed.PaletteDMAOffset,
		parsed.PalettePacketTotal,
		"palette",
	)
	if err != nil {
		return err
	}

	for i := range parsed.ImageDMA.Transfers {
		if err := ValidateUpload(parsed.ImageDMA.Transfers[i], "image", i); err != nil {
			return err
		}
	}
	for i := range paletteDMA.Transfers {
		if err := ValidateUpload(paletteDMA.Transfers[i], "palette", i); err != nil {
			return err
		}
		if len(paletteDMA.Transfers[i].Data) < 256*4 {
			return fmt.Errorf("palette DMA transfer %d has %#x bytes, need %#x", i, len(paletteDMA.Transfers[i].Data), 256*4)
		}
	}

	firstImage := parsed.ImageDMA.Transfers[0].Upload
	if firstImage.Width != parsed.ImageUploadWidth || firstImage.Height != parsed.ImageUploadHeight {
		return fmt.Errorf(
			"%w: header image upload is %dx%d but first GIF upload is %dx%d",
			ErrUnsupportedT32,
			parsed.ImageUploadWidth,
			parsed.ImageUploadHeight,
			firstImage.Width,
			firstImage.Height,
		)
	}

	packetIndex := 0
	for textureIndex := range parsed.Textures {
		texture := &parsed.Textures[textureIndex]
		paletteCount := paletteCounts[textureIndex]
		for paletteIndex := 0; paletteIndex < paletteCount; paletteIndex++ {
			transfer := &paletteDMA.Transfers[packetIndex]
			if paletteIndex == 0 && transfer.Upload.DestinationBase != texture.GSTex0.CBP {
				return fmt.Errorf(
					"texture %d CLUT base is %#x but its first palette uploads to %#x",
					textureIndex,
					texture.GSTex0.CBP,
					transfer.Upload.DestinationBase,
				)
			}
			fullPalette := DecodeCSM1RGBA32(transfer.Data)
			colors := fullPalette
			if texture.GSTex0.PSM == PSMT4 {
				base := int(texture.GSTex0.CSA&0x0f) * 16
				colors = append([]RGBA(nil), fullPalette[base:base+16]...)
			} else {
				colors = append([]RGBA(nil), fullPalette...)
			}
			texture.Palettes = append(texture.Palettes, Palette{Colors: colors})
			packetIndex++
		}
	}

	if err := parsed.decodeAllIndices(); err != nil {
		return err
	}
	*c = *parsed
	return nil
}

func (c *Container) decodeAllIndices() (err error) {
	gsMutex.Lock()
	defer gsMutex.Unlock()
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("graphicsynthesizer memory access: %v", recovered)
		}
	}()
	clear(graphicsynthesizer.GsMem)

	for i := range c.ImageDMA.Transfers {
		transfer := &c.ImageDMA.Transfers[i]
		upload := transfer.Upload
		graphicsynthesizer.WriteTexPSMCT32(
			int(upload.DestinationBase),
			int(upload.DestinationBufferWidth),
			int(upload.DestinationX),
			int(upload.DestinationY),
			int(upload.Width),
			int(upload.Height),
			transfer.Data,
		)
	}

	for i := range c.Textures {
		texture := &c.Textures[i]
		pixelTotal := texture.Width * texture.Height
		switch texture.GSTex0.PSM {
		case PSMT8:
			texture.Indices = make([]byte, pixelTotal)
			graphicsynthesizer.ReadTexPSMT8(
				int(texture.GSTex0.TBP0),
				int(texture.GSTex0.TBW),
				0,
				0,
				texture.Width,
				texture.Height,
				texture.Indices,
			)
		case PSMT4:
			packed := make([]byte, (pixelTotal+1)/2)
			graphicsynthesizer.ReadTexPSMT4(
				int(texture.GSTex0.TBP0),
				int(texture.GSTex0.TBW),
				0,
				0,
				texture.Width,
				texture.Height,
				packed,
			)
			texture.Indices = unpack4(packed, pixelTotal)
		default:
			return fmt.Errorf("texture %d: %w", i, ErrNotIndexed)
		}
	}
	return nil
}

func (c *Container) DecodeTexture(textureIndex, paletteIndex int) (*DecodedTexture, error) {
	texture, palette, err := c.textureAndPalette(textureIndex, paletteIndex)
	if err != nil {
		return nil, err
	}
	pixels := make([]RGBA, len(texture.Indices))
	for i, index := range texture.Indices {
		if int(index) >= len(palette.Colors) {
			return nil, fmt.Errorf(
				"texture %d pixel %d uses palette index %d outside %d entries",
				textureIndex,
				i,
				index,
				len(palette.Colors),
			)
		}
		pixels[i] = palette.Colors[index]
	}
	return &DecodedTexture{
		Width:   texture.Width,
		Height:  texture.Height,
		Format:  texture.GSTex0.PSM,
		Indices: append([]byte(nil), texture.Indices...),
		Palette: append([]RGBA(nil), palette.Colors...),
		Pixels:  pixels,
	}, nil
}

func (c *Container) PixelIndices(textureIndex int) ([]byte, error) {
	texture, err := c.textureAt(textureIndex)
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), texture.Indices...), nil
}

func (c *Container) ColorPalette(textureIndex, paletteIndex int) ([]RGBA, error) {
	_, palette, err := c.textureAndPalette(textureIndex, paletteIndex)
	if err != nil {
		return nil, err
	}
	return append([]RGBA(nil), palette.Colors...), nil
}

func (c *Container) textureAt(textureIndex int) (*Texture, error) {
	if c == nil {
		return nil, ErrContainerIsNil
	}
	if textureIndex < 0 || textureIndex >= len(c.Textures) {
		return nil, fmt.Errorf("texture index %d is outside %d textures", textureIndex, len(c.Textures))
	}
	return &c.Textures[textureIndex], nil
}

func (c *Container) textureAndPalette(textureIndex, paletteIndex int) (*Texture, *Palette, error) {
	texture, err := c.textureAt(textureIndex)
	if err != nil {
		return nil, nil, err
	}
	if paletteIndex < 0 || paletteIndex >= len(texture.Palettes) {
		return nil, nil, fmt.Errorf(
			"palette index %d is outside %d palettes at texture %d",
			paletteIndex,
			len(texture.Palettes),
			textureIndex,
		)
	}
	return texture, &texture.Palettes[paletteIndex], nil
}

func FromStreamWithOffset(t32 *Container, stream io.ReadSeeker, offset uint32) error {
	if t32 == nil {
		return ErrContainerIsNil
	}
	t32.Offset = offset
	return t32.unmarshal(stream)
}

func DecodeImage(
	stream io.ReadSeeker,
	offset uint32,
	textureIndex int,
	paletteIndex int,
) (image.Image, error) {
	t32 := New()
	if err := FromStreamWithOffset(t32, stream, offset); err != nil {
		return nil, err
	}
	return T32ToImage(t32, textureIndex, paletteIndex)
}

func DecodePalettedImage(
	stream io.ReadSeeker,
	offset uint32,
	textureIndex int,
	paletteIndex int,
) (*image.Paletted, error) {
	t32 := New()
	if err := FromStreamWithOffset(t32, stream, offset); err != nil {
		return nil, err
	}
	return t32.PaletteAtImage(textureIndex, paletteIndex)
}

func T32ToImage(t32 *Container, textureIndex, paletteIndex int) (image.Image, error) {
	if t32 == nil {
		return nil, ErrContainerIsNil
	}
	decoded, err := t32.DecodeTexture(textureIndex, paletteIndex)
	if err != nil {
		return nil, err
	}
	return decoded.Image(), nil
}

func ValidateHeader(header Header, dataSize uint64) error {
	if header.TextureTotal == 0 {
		return fmt.Errorf("%w: texture total is zero", ErrUnsupportedT32)
	}
	descriptorSize := uint64(DescriptorSize)
	if header.TextureTotal > 1 {
		descriptorSize += uint64(header.TextureTotal-1) * (TextureLinkSize + DescriptorSize)
	}
	minimumImageDMAOffset := uint64(HeaderSize) + descriptorSize
	if uint64(header.ImageDMAOffset) < minimumImageDMAOffset {
		return fmt.Errorf(
			"%w: image DMA offset %#x overlaps texture descriptors ending at %#x",
			ErrUnsupportedT32,
			header.ImageDMAOffset,
			minimumImageDMAOffset,
		)
	}
	if header.ImagePacketTotal == 0 {
		return fmt.Errorf("%w: image packet total is zero", ErrUnsupportedT32)
	}
	if header.PalettePacketTotal == 0 || header.PaletteTotal == 0 {
		return fmt.Errorf("%w: palette total is zero", ErrUnsupportedT32)
	}
	if header.ImageUploadWidth == 0 || header.ImageUploadHeight == 0 {
		return fmt.Errorf("%w: image upload dimensions are zero", ErrUnsupportedT32)
	}
	if header.Reserved0 != 0 || header.Reserved1 != 0 {
		return fmt.Errorf(
			"%w: non-zero header reserved fields %#x, %#x",
			ErrUnsupportedT32,
			header.Reserved0,
			header.Reserved1,
		)
	}
	if uint64(header.ImageDMAOffset) >= dataSize {
		return fmt.Errorf("image DMA offset %#x is outside T32 data size %#x", header.ImageDMAOffset, dataSize)
	}
	if uint64(header.PaletteDMAOffset) >= dataSize {
		return fmt.Errorf("palette DMA offset %#x is outside T32 data size %#x", header.PaletteDMAOffset, dataSize)
	}
	if header.PaletteDMAOffset <= header.ImageDMAOffset {
		return fmt.Errorf(
			"palette DMA offset %#x does not follow image DMA offset %#x",
			header.PaletteDMAOffset,
			header.ImageDMAOffset,
		)
	}
	return nil
}

func ValidateTexture(texture Texture) error {
	if !texture.IsIndexed() {
		return fmt.Errorf("%w: %s", ErrNotIndexed, texture.GSTex0.PSM)
	}
	if texture.GSTex0.CPSM != PSMCT32 {
		return fmt.Errorf("unsupported CLUT format %s", texture.GSTex0.CPSM)
	}
	if texture.GSTex0.CSM != 0 {
		return errors.New("unsupported CSM2 palette storage")
	}
	if texture.GSTex0.TW > 10 || texture.GSTex0.TH > 10 {
		return fmt.Errorf(
			"texture dimensions 2^%d x 2^%d exceed the PS2 GS limit",
			texture.GSTex0.TW,
			texture.GSTex0.TH,
		)
	}
	if texture.GSTex0.TBW == 0 {
		return errors.New("texture buffer width is zero")
	}
	return nil
}

func (d *DecodedTexture) Image() image.Image {
	if d == nil {
		return nil
	}
	img := image.NewNRGBA(image.Rect(0, 0, d.Width, d.Height))
	for i, pixel := range d.Pixels {
		position := i * 4
		img.Pix[position+0] = pixel.R
		img.Pix[position+1] = pixel.G
		img.Pix[position+2] = pixel.B
		img.Pix[position+3] = pixel.A
	}
	return img
}

func (d *DecodedTexture) PalettedImage() *image.Paletted {
	if d == nil {
		return nil
	}
	palette := make(color.Palette, len(d.Palette))
	for i, entry := range d.Palette {
		palette[i] = color.NRGBA{R: entry.R, G: entry.G, B: entry.B, A: entry.A}
	}
	img := image.NewPaletted(image.Rect(0, 0, d.Width, d.Height), palette)
	copy(img.Pix, d.Indices)
	return img
}

func ValidateUpload(transfer DMATransfer, name string, index int) error {
	upload := transfer.Upload
	if upload.Direction != TransferHostToLocal {
		return fmt.Errorf("%s DMA transfer %d is not host-to-local", name, index)
	}
	if upload.DestinationPSM != PSMCT32 {
		return fmt.Errorf(
			"%s DMA transfer %d has destination format %s, expected PSMCT32",
			name,
			index,
			upload.DestinationPSM,
		)
	}
	if upload.DestinationBufferWidth == 0 || upload.Width == 0 || upload.Height == 0 {
		return fmt.Errorf("%s DMA transfer %d has zero upload dimensions", name, index)
	}
	required := uint64(upload.Width) * uint64(upload.Height) * 4
	if upload.DataSize != required || uint64(len(transfer.Data)) != required {
		return fmt.Errorf(
			"%s DMA transfer %d has %#x bytes, expected %#x for %dx%d PSMCT32",
			name,
			index,
			len(transfer.Data),
			required,
			upload.Width,
			upload.Height,
		)
	}
	return nil
}

func DecodeCSM1RGBA32(payload []byte) []RGBA {
	palette := make([]RGBA, 256)
	for logicalIndex := range palette {
		storedIndex := ClutPermutation(logicalIndex)
		position := storedIndex * 4
		palette[logicalIndex] = RGBA{
			R: payload[position+0],
			G: payload[position+1],
			B: payload[position+2],
			A: ExpandPS2Alpha(payload[position+3]),
		}
	}
	return palette
}

func unpack4(packed []byte, pixelTotal int) []byte {
	indices := make([]byte, pixelTotal)
	for i := range indices {
		if i&1 == 0 {
			indices[i] = packed[i/2] & 0x0f
		} else {
			indices[i] = packed[i/2] >> 4
		}
	}
	return indices
}
