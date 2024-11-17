package tim3

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/anasrar/chihuahua/pkg/buffer"
	graphicsynthesizer "github.com/anasrar/chihuahua/pkg/graphic_synthesizer"
	"github.com/anasrar/chihuahua/pkg/tim2"
)

func PictureToImage(picture *tim2.Picture) *image.NRGBA {
	width := int(picture.ImageWidth)
	height := int(picture.ImageHeight)
	swizzle := width >= 128 && height >= 128
	data := make([]byte, len(picture.ImageData))
	copy(data, picture.ImageData)
	indices := []byte{}

	switch picture.ImageType {
	case tim2.ImageType4BitTexture:
		if swizzle {
			data = graphicsynthesizer.Unswizzle4(data, width, height)
		}

		for _, v := range data {
			low := v & 0xF
			high := (v >> 4) & 0xF
			indices = append(indices, low)
			indices = append(indices, high)
		}
	case tim2.ImageType8BitTexture:
		if swizzle {
			data = graphicsynthesizer.Unswizzle8(data, width, height)
		}

		indices = append(indices, data...)
	}

	raw := []uint8{}
	for _, index := range indices {
		c := picture.ClutData[index]
		raw = append(raw, c.R)
		raw = append(raw, c.G)
		raw = append(raw, c.B)
		raw = append(raw, c.A)
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	copy(img.Pix, raw)

	return img
}

func ImagePalettedToFile(img *image.Paletted, output *os.File) error {
	colorTotal := len(img.Palette)

	if colorTotal > 256 {
		return fmt.Errorf("PNG colors exceeds the maximum allowable limit of 256")
	}

	width := img.Rect.Max.X
	height := img.Rect.Max.Y
	indices := img.Pix

	swizzle := width >= 128 && height >= 128
	if swizzle {
		indices = graphicsynthesizer.Swizzle8(indices, width, height)
	}

	colors := []*color.RGBA{}
	for _, c := range img.Palette {
		c32, _ := c.(color.RGBA)
		colors = append(colors, &c32)
	}

	// NOTE: fill colors to 256
	{
		diff := 256 - colorTotal
		for range diff {
			colors = append(colors, &color.RGBA{R: 0, G: 0, B: 0, A: 0})
		}
	}

	twiddle := []*color.RGBA{}
	for i := 0; i < 256; i += 32 {
		twiddle = append(twiddle, colors[i+0:i+8]...)
		twiddle = append(twiddle, colors[i+16:i+24]...)
		twiddle = append(twiddle, colors[i+8:i+16]...)
		twiddle = append(twiddle, colors[i+24:i+32]...)
	}

	if _, err := buffer.WriteUint32LE(output, Signature); err != nil {
		return err
	}

	// NOTE: FileHeader.format_version
	if _, err := buffer.WriteUint8(output, 4); err != nil {
		return err
	}

	// NOTE: FileHeader.format_id
	if _, err := buffer.WriteUint8(output, 6); err != nil {
		return err
	}

	// NOTE: FileHeader.picturees
	if _, err := buffer.WriteUint16LE(output, 1); err != nil {
		return err
	}

	// NOTE: FileHeader.reserved
	if _, err := buffer.WriteBytes(output, []byte{0, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}

	// NOTE: Picture.total_size = clut_size + image_size +  header_size
	if _, err := buffer.WriteUint32LE(output, uint32(1024+(width*height)+48)); err != nil {
		return err
	}

	// NOTE: Picture.clut_size
	if _, err := buffer.WriteUint32LE(output, 1024); err != nil {
		return err
	}

	// NOTE: Picture.image_size
	if _, err := buffer.WriteUint32LE(output, uint32(width*height)); err != nil {
		return err
	}

	// NOTE: Picture.header_size
	if _, err := buffer.WriteUint16LE(output, 48); err != nil {
		return err
	}

	// NOTE: Picture.clut_colors
	if _, err := buffer.WriteUint16LE(output, 256); err != nil {
		return err
	}

	// NOTE: Picture.pict_format
	if _, err := buffer.WriteUint8(output, 0); err != nil {
		return err
	}

	// NOTE: Picture.mipmap_textures
	if _, err := buffer.WriteUint8(output, 1); err != nil {
		return err
	}

	// NOTE: Picture.clut_type = RGBA32|0x80
	if _, err := buffer.WriteUint8(output, 3); err != nil {
		return err
	}

	// NOTE: Picture.image_type = 8bpp
	if _, err := buffer.WriteUint8(output, 5); err != nil {
		return err
	}

	// NOTE: Picture.image_width
	if _, err := buffer.WriteUint16LE(output, uint16(width)); err != nil {
		return err
	}

	// NOTE: Picture.image_height
	if _, err := buffer.WriteUint16LE(output, uint16(height)); err != nil {
		return err
	}

	// TODO: construct GsTex0
	// NOTE: Picture.gs_tex0
	if _, err := buffer.WriteUint64LE(output, 0); err != nil {
		return err
	}

	// NOTE: Picture.gs_tex1
	if _, err := buffer.WriteUint64LE(output, 608); err != nil {
		return err
	}

	// NOTE: Picture.gs_regs
	if _, err := buffer.WriteUint32LE(output, 0); err != nil {
		return err
	}

	// NOTE: Picture.gs_tex_clut
	if _, err := buffer.WriteUint32LE(output, 0); err != nil {
		return err
	}

	// NOTE: Picture.image_data
	if _, err := buffer.WriteBytes(output, indices); err != nil {
		return err
	}

	// NOTE: Picture.clut_data
	for _, c := range twiddle {
		a := uint8(float32(c.A) / 255 * 0x80)
		if _, err := buffer.WriteBytes(output, []byte{c.R, c.G, c.B, a}); err != nil {
			return err
		}
	}

	return nil
}
