package t32

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/anasrar/binarium"
	graphicsynthesizer "github.com/anasrar/chihuahua/pkg/graphic_synthesizer"
)

type paletteLayout struct {
	PacketOffset  uint64
	PayloadOffset uint64
	Destination   uint16
	Payload       []byte
}

func (c *Container) Write(output io.Writer) error {
	if c == nil {
		return ErrContainerIsNil
	}
	if output == nil {
		return errors.New("output writer is nil")
	}
	if len(c.Textures) == 0 {
		return errors.New("T32 has no images")
	}

	imageHeight, imagePayload, palettePayloads, err := c.preparePayloads()
	if err != nil {
		return err
	}
	c.syncCounts()

	descriptorEnd := uint64(HeaderSize + DescriptorSize)
	if len(c.Textures) > 1 {
		descriptorEnd += uint64(len(c.Textures)-1) * (TextureLinkSize + DescriptorSize)
	}
	imageDMAOffset := alignUp(descriptorEnd, 0x10)
	imagePacketOffset := alignUp(imageDMAOffset+2*DMATagSize, 0x80)
	imagePayloadOffset := imagePacketOffset + UploadPacketSize
	paletteDMAOffset := alignUp(imagePayloadOffset+uint64(len(imagePayload)), 0x80)
	palettePacketOffset := alignUp(
		paletteDMAOffset+uint64(len(palettePayloads)+1)*DMATagSize,
		0x80,
	)

	palettes := make([]paletteLayout, len(palettePayloads))
	packetOffset := palettePacketOffset
	for i := range palettePayloads {
		palettes[i] = paletteLayout{
			PacketOffset:  packetOffset,
			PayloadOffset: packetOffset + UploadPacketSize,
			Destination:   uint16(i),
			Payload:       palettePayloads[i],
		}
		packetOffset = alignUp(palettes[i].PayloadOffset+uint64(len(palettes[i].Payload)), 0x80)
	}
	totalSize := palettes[len(palettes)-1].PayloadOffset + uint64(len(palettes[len(palettes)-1].Payload))
	if totalSize > uint64(int(^uint(0)>>1)) {
		return fmt.Errorf("T32 size %#x exceeds addressable memory", totalSize)
	}

	header := Header{
		TextureTotal:       uint32(len(c.Textures)),
		ImageUploadWidth:   64,
		ImageUploadHeight:  uint16(imageHeight),
		ImageDMAOffset:     uint32(imageDMAOffset),
		PaletteDMAOffset:   uint32(paletteDMAOffset),
		ImagePacketTotal:   1,
		PalettePacketTotal: uint16(len(palettePayloads)),
		PaletteTotal:       uint32(len(c.Textures[0].Palettes)),
	}
	c.Header = header

	data := make([]byte, int(totalSize))
	if err := marshalAt(data, 0, header); err != nil {
		return fmt.Errorf("marshal T32 header: %w", err)
	}
	descriptorOffset := uint64(HeaderSize)
	for i := range c.Textures {
		if i > 0 {
			link := TextureLink{Index: uint32(i), PaletteTotal: uint32(len(c.Textures[i].Palettes))}
			if err := marshalAt(data, descriptorOffset, link); err != nil {
				return fmt.Errorf("marshal texture link %d: %w", i, err)
			}
			descriptorOffset += TextureLinkSize
		}
		if err := marshalAt(data, descriptorOffset, c.Textures[i]); err != nil {
			return fmt.Errorf("marshal texture descriptor %d: %w", i, err)
		}
		descriptorOffset += DescriptorSize
	}

	imageQWC := uint16(len(imagePayload) / 16)
	imageTag := newDMATag(imageQWC+6, DMAIDReference, uint32(imagePacketOffset-imageDMAOffset))
	if err := marshalAt(data, imageDMAOffset, imageTag); err != nil {
		return fmt.Errorf("marshal image DMA tag: %w", err)
	}
	if err := marshalAt(data, imageDMAOffset+DMATagSize, newDMATag(0, DMAIDEnd, 0)); err != nil {
		return fmt.Errorf("marshal image DMA end tag: %w", err)
	}
	imagePacket := newGIFUploadPacket(0, 1, 0, 0, 64, uint16(imageHeight), imageQWC)
	if err := marshalAt(data, imagePacketOffset, imagePacket); err != nil {
		return fmt.Errorf("marshal image GIF packet: %w", err)
	}
	copy(data[imagePayloadOffset:], imagePayload)

	for i := range palettes {
		address := uint32(palettes[i].PacketOffset - paletteDMAOffset)
		tag := newDMATag(64+6, DMAIDReference, address)
		if err := marshalAt(data, paletteDMAOffset+uint64(i)*DMATagSize, tag); err != nil {
			return fmt.Errorf("marshal palette DMA tag %d: %w", i, err)
		}
		packet := newGIFUploadPacket(palettes[i].Destination, 1, 0, 0, 16, 16, 64)
		if err := marshalAt(data, palettes[i].PacketOffset, packet); err != nil {
			return fmt.Errorf("marshal palette GIF packet %d: %w", i, err)
		}
		copy(data[palettes[i].PayloadOffset:], palettes[i].Payload)
	}
	paletteEndOffset := paletteDMAOffset + uint64(len(palettes))*DMATagSize
	if err := marshalAt(data, paletteEndOffset, newDMATag(0, DMAIDEnd, 0)); err != nil {
		return fmt.Errorf("marshal palette DMA end tag: %w", err)
	}

	return writeFull(output, data)
}

func (c *Container) preparePayloads() (height int, imagePayload []byte, palettePayloads [][]byte, err error) {
	if c.totalPaletteCount() > int(^uint16(0)) {
		return 0, nil, nil, errors.New("palette total exceeds uint16")
	}
	if c.totalPaletteCount() > 0x4000 {
		return 0, nil, nil, errors.New("palette total exceeds GS base-pointer range")
	}

	textureBase := 0
	paletteBase := 0
	for i := range c.Textures {
		texture := &c.Textures[i]
		texture.Index = uint32(i)
		if err := configureTexture(texture); err != nil {
			return 0, nil, nil, fmt.Errorf("texture %d: %w", i, err)
		}
		allocation := textureAllocationUnits(texture)
		if textureBase+allocation > 0x4000 {
			return 0, nil, nil, errors.New("textures exceed the 4 MiB GS memory")
		}
		texture.GSTex0.TBP0 = uint16(textureBase)
		texture.GSTex0.CBP = uint16(paletteBase)
		texture.GSTex0Raw = EncodeGSTex0(texture.GSTex0)
		texture.GSTex0.Raw = texture.GSTex0Raw
		texture.GSTex1Raw = 0x260
		texture.GSTex1 = DecodeGSTex1(texture.GSTex1Raw)
		texture.Reserved = 0
		textureBase += allocation
		paletteBase += len(texture.Palettes)
	}
	if textureBase == 0 || textureBase*16 > 0x7fff {
		return 0, nil, nil, fmt.Errorf(
			"combined GS image upload height %d exceeds one T32 GIF packet",
			textureBase,
		)
	}

	imagePayload, err = c.buildImagePayload(textureBase)
	if err != nil {
		return 0, nil, nil, err
	}
	palettePayloads = make([][]byte, 0, c.totalPaletteCount())
	for textureIndex := range c.Textures {
		texture := &c.Textures[textureIndex]
		for paletteIndex := range texture.Palettes {
			fullPalette := make([]RGBA, 256)
			copy(fullPalette, texture.Palettes[paletteIndex].Colors)
			payload, err := EncodeCSM1RGBA32(fullPalette)
			if err != nil {
				return 0, nil, nil, fmt.Errorf(
					"texture %d palette %d: %w",
					textureIndex,
					paletteIndex,
					err,
				)
			}
			palettePayloads = append(palettePayloads, payload)
		}
	}
	return textureBase, imagePayload, palettePayloads, nil
}

func (c *Container) buildImagePayload(height int) (payload []byte, err error) {
	gsMutex.Lock()
	defer gsMutex.Unlock()
	defer func() {
		if recovered := recover(); recovered != nil {
			payload = nil
			err = fmt.Errorf("graphicsynthesizer memory access: %v", recovered)
		}
	}()
	clear(graphicsynthesizer.GsMem)

	for i := range c.Textures {
		texture := &c.Textures[i]
		switch texture.GSTex0.PSM {
		case PSMT8:
			graphicsynthesizer.WriteTexPSMT8(
				int(texture.GSTex0.TBP0),
				int(texture.GSTex0.TBW),
				0,
				0,
				texture.Width,
				texture.Height,
				texture.Indices,
			)
		case PSMT4:
			graphicsynthesizer.WriteTexPSMT4(
				int(texture.GSTex0.TBP0),
				int(texture.GSTex0.TBW),
				0,
				0,
				texture.Width,
				texture.Height,
				pack4(texture.Indices),
			)
		default:
			return nil, fmt.Errorf("texture %d: %w", i, ErrNotIndexed)
		}
	}

	payload = make([]byte, 64*height*4)
	graphicsynthesizer.ReadTexPSMCT32(0, 1, 0, 0, 64, height, payload)
	return payload, nil
}

func textureAllocationUnits(texture *Texture) int {
	pagesX := (texture.Width + 127) / 128
	pagesY := (texture.Height + 127) / 128
	if texture.GSTex0.PSM == PSMT8 {
		pagesY = (texture.Height + 63) / 64
	}
	return pagesX * pagesY * 32
}

func pack4(indices []byte) []byte {
	packed := make([]byte, (len(indices)+1)/2)
	for i, index := range indices {
		if i&1 == 0 {
			packed[i/2] = index & 0x0f
		} else {
			packed[i/2] |= (index & 0x0f) << 4
		}
	}
	return packed
}

func newDMATag(qwc uint16, id DMAID, address uint32) DMATag {
	return DMATag{
		Word0: uint64(qwc) | uint64(id&0x07)<<28 | uint64(address&0x7fffffff)<<32,
	}
}

func newGIFUploadPacket(
	destinationBase uint16,
	destinationBufferWidth uint8,
	destinationX uint16,
	destinationY uint16,
	width uint16,
	height uint16,
	payloadQWC uint16,
) GIFUploadPacket {
	return GIFUploadPacket{
		SetupTag: GIFTag{Low: uint64(1)<<60 | 4, High: GIFRegisterAD},
		BITBLTBUF: GIFAD{
			Value: uint64(destinationBase&0x3fff)<<32 |
				uint64(destinationBufferWidth&0x3f)<<48 |
				uint64(PSMCT32)<<56,
			Register: GIFRegisterBITBLTBUF,
		},
		TRXPOS: GIFAD{
			Value:    uint64(destinationX&0x7ff)<<32 | uint64(destinationY&0x7ff)<<48,
			Register: GIFRegisterTRXPOS,
		},
		TRXREG: GIFAD{
			Value:    uint64(width&0xfff) | uint64(height&0xfff)<<32,
			Register: GIFRegisterTRXREG,
		},
		TRXDIR: GIFAD{Register: GIFRegisterTRXDIR},
		ImageTag: GIFTag{
			Low: uint64(payloadQWC&0x7fff) | uint64(1)<<15 | uint64(2)<<58,
		},
	}
}

func marshalAt(data []byte, offset uint64, value any) error {
	encoded, err := binarium.Marshal(binary.LittleEndian, value)
	if err != nil {
		return err
	}
	if err := ValidateRegion(uint64(len(data)), offset, uint64(len(encoded)), "marshal target"); err != nil {
		return err
	}
	copy(data[offset:offset+uint64(len(encoded))], encoded)
	return nil
}

func alignUp(value, alignment uint64) uint64 {
	return (value + alignment - 1) &^ (alignment - 1)
}

func writeFull(output io.Writer, data []byte) error {
	for len(data) > 0 {
		written, err := output.Write(data)
		if err != nil {
			return fmt.Errorf("write T32: %w", err)
		}
		if written == 0 {
			return io.ErrShortWrite
		}
		data = data[written:]
	}
	return nil
}
