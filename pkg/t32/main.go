// Package t32 decodes the compact, single-texture T32 variant.
//
// The input must already be decompressed. Fixed binary structures are read
// from io.ReadSeeker with binarium, while PS2 GS addressing is handled by
// github.com/anasrar/graphicsynthesizer.
package t32

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"

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
	PaletteTotal       uint32 `json:"palette_total"`
}

type Texture struct {
	GSTex0Raw uint64 `json:"gs_tex_0_raw"`
	GSTex1Raw uint32 `json:"gs_tex_1_raw"`
	Reserved  uint32 `json:"reserved"`

	GSTex0 GSTex0 `json:"gs_tex_0" skip:""`
	GSTex1 GSTex1 `json:"gs_tex_1" skip:""`
}

func (t *Texture) IsIndexed() bool {
	return t != nil && (t.GSTex0.PSM == PSMT8 || t.GSTex0.PSM == PSMT4)
}

type Container struct {
	Offset uint32 `json:"offset" skip:""`
	Header
	Texture    Texture  `json:"texture" skip:""`
	ImageDMA   DMAChain `json:"image_dma" skip:""`
	PaletteDMA DMAChain `json:"palette_dma" skip:""`
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
		ImageDMA:   DMAChain{Transfers: []DMATransfer{}},
		PaletteDMA: DMAChain{Transfers: []DMATransfer{}},
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
	if err := unmarshalAt(stream, baseOffset+HeaderSize, &parsed.Texture); err != nil {
		return fmt.Errorf("texture descriptor: %w", err)
	}
	if parsed.Texture.Reserved != 0 {
		return fmt.Errorf("%w: texture reserved field is %#x", ErrUnsupportedT32, parsed.Texture.Reserved)
	}
	parsed.Texture.GSTex0 = DecodeGSTex0(parsed.Texture.GSTex0Raw)
	parsed.Texture.GSTex1 = DecodeGSTex1(parsed.Texture.GSTex1Raw)
	if err := ValidateTexture(parsed.Texture); err != nil {
		return err
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
	parsed.PaletteDMA, err = ReadDMAChain(
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
	for i := range parsed.PaletteDMA.Transfers {
		if err := ValidateUpload(parsed.PaletteDMA.Transfers[i], "palette", i); err != nil {
			return err
		}
	}

	firstImage := parsed.ImageDMA.Transfers[0].Upload
	if firstImage.Width != parsed.ImageUploadWidth || firstImage.Height != parsed.ImageUploadHeight {
		return fmt.Errorf(
			"%w: header image upload is %dx%d but GIF upload is %dx%d",
			ErrUnsupportedT32,
			parsed.ImageUploadWidth,
			parsed.ImageUploadHeight,
			firstImage.Width,
			firstImage.Height,
		)
	}

	paletteIndex := int(parsed.Texture.GSTex0.CBP)
	if paletteIndex >= len(parsed.PaletteDMA.Transfers) {
		return fmt.Errorf(
			"texture CLUT packet %d is outside %d palette transfers",
			paletteIndex,
			len(parsed.PaletteDMA.Transfers),
		)
	}

	*c = *parsed
	return nil
}

func (c *Container) DecodeTexture() (*DecodedTexture, error) {
	if c == nil {
		return nil, ErrContainerIsNil
	}
	indices, err := c.PixelIndices()
	if err != nil {
		return nil, err
	}
	palette, err := c.ColorPalette()
	if err != nil {
		return nil, err
	}

	pixels := make([]RGBA, len(indices))
	for i, index := range indices {
		if int(index) >= len(palette) {
			return nil, fmt.Errorf(
				"pixel %d uses palette index %d outside %d entries",
				i,
				index,
				len(palette),
			)
		}
		pixels[i] = palette[index]
	}

	return &DecodedTexture{
		Width:   c.Texture.GSTex0.Width(),
		Height:  c.Texture.GSTex0.Height(),
		Format:  c.Texture.GSTex0.PSM,
		Indices: indices,
		Palette: palette,
		Pixels:  pixels,
	}, nil
}

func (c *Container) PixelIndices() (indices []byte, err error) {
	if c == nil {
		return nil, ErrContainerIsNil
	}
	if !c.Texture.IsIndexed() {
		return nil, ErrNotIndexed
	}

	gsMutex.Lock()
	defer gsMutex.Unlock()
	defer func() {
		if recovered := recover(); recovered != nil {
			indices = nil
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

	width := c.Texture.GSTex0.Width()
	height := c.Texture.GSTex0.Height()
	pixelTotal := width * height
	tbw := int(c.Texture.GSTex0.TBW)

	switch c.Texture.GSTex0.PSM {
	case PSMT8:
		indices = make([]byte, pixelTotal)
		graphicsynthesizer.ReadTexPSMT8(
			int(c.Texture.GSTex0.TBP0), tbw, 0, 0, width, height, indices,
		)
		return indices, nil

	case PSMT4:
		packed := make([]byte, (pixelTotal+1)/2)
		graphicsynthesizer.ReadTexPSMT4(
			int(c.Texture.GSTex0.TBP0), tbw, 0, 0, width, height, packed,
		)
		indices = make([]byte, pixelTotal)
		for i := range indices {
			if i&1 == 0 {
				indices[i] = packed[i/2] & 0x0f
			} else {
				indices[i] = packed[i/2] >> 4
			}
		}
		return indices, nil

	default:
		return nil, ErrNotIndexed
	}
}

func (c *Container) ColorPalette() ([]RGBA, error) {
	if c == nil {
		return nil, ErrContainerIsNil
	}
	if !c.Texture.IsIndexed() {
		return nil, ErrNotIndexed
	}

	packetIndex := int(c.Texture.GSTex0.CBP)
	if packetIndex < 0 || packetIndex >= len(c.PaletteDMA.Transfers) {
		return nil, fmt.Errorf(
			"CLUT packet %d is outside %d palette transfers",
			packetIndex,
			len(c.PaletteDMA.Transfers),
		)
	}
	payload := c.PaletteDMA.Transfers[packetIndex].Data
	if len(payload) < 256*4 {
		return nil, fmt.Errorf("palette payload has %#x bytes, need %#x", len(payload), 256*4)
	}

	fullPalette := DecodeCSM1RGBA32(payload)
	if c.Texture.GSTex0.PSM == PSMT8 {
		return fullPalette, nil
	}

	base := int(c.Texture.GSTex0.CSA&0x0f) * 16
	palette := make([]RGBA, 16)
	copy(palette, fullPalette[base:base+16])
	return palette, nil
}

func FromStreamWithOffset(t32 *Container, stream io.ReadSeeker, offset uint32) error {
	if t32 == nil {
		return ErrContainerIsNil
	}
	t32.Offset = offset
	return t32.unmarshal(stream)
}

func FromStream(t32 *Container, stream io.ReadSeeker) error {
	return FromStreamWithOffset(t32, stream, 0)
}

func FromPathWithOffset(t32 *Container, filePath string, offset uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return FromStreamWithOffset(t32, file, 0)
}

func FromPath(c *Container, filePath string) error {
	return FromPathWithOffset(c, filePath, 0)
}

func DecodeImage(stream io.ReadSeeker) (image.Image, error) {
	t32 := New()
	if err := FromStream(t32, stream); err != nil {
		return nil, err
	}
	return T32ToImage(t32)
}

func DecodePalettedImage(stream io.ReadSeeker) (*image.Paletted, error) {
	t32 := New()
	if err := FromStream(t32, stream); err != nil {
		return nil, err
	}
	return T32ToPalettedImage(t32)
}

func T32ToImage(t32 *Container) (image.Image, error) {
	if t32 == nil {
		return nil, ErrContainerIsNil
	}
	decoded, err := t32.DecodeTexture()
	if err != nil {
		return nil, err
	}
	return decoded.Image(), nil
}

func T32ToPalettedImage(t32 *Container) (*image.Paletted, error) {
	if t32 == nil {
		return nil, ErrContainerIsNil
	}
	decoded, err := t32.DecodeTexture()
	if err != nil {
		return nil, err
	}
	return decoded.PalettedImage(), nil
}

func ValidateHeader(header Header, dataSize uint64) error {
	if header.TextureTotal != 1 {
		return fmt.Errorf("%w: texture total is %d, expected 1", ErrUnsupportedT32, header.TextureTotal)
	}
	if header.ImageDMAOffset != ImageDMAOffset {
		return fmt.Errorf(
			"%w: image DMA offset is %#x, expected compact descriptor end %#x",
			ErrUnsupportedT32,
			header.ImageDMAOffset,
			ImageDMAOffset,
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

func (self *DecodedTexture) Image() image.Image {
	if self == nil {
		return nil
	}
	img := image.NewNRGBA(image.Rect(0, 0, self.Width, self.Height))
	for i, pixel := range self.Pixels {
		position := i * 4
		img.Pix[position+0] = pixel.R
		img.Pix[position+1] = pixel.G
		img.Pix[position+2] = pixel.B
		img.Pix[position+3] = pixel.A
	}
	return img
}

func (self *DecodedTexture) PalettedImage() *image.Paletted {
	if self == nil {
		return nil
	}
	palette := make(color.Palette, len(self.Palette))
	for i, entry := range self.Palette {
		palette[i] = color.NRGBA{R: entry.R, G: entry.G, B: entry.B, A: entry.A}
	}
	img := image.NewPaletted(image.Rect(0, 0, self.Width, self.Height), palette)
	copy(img.Pix, self.Indices)
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
