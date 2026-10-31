package tim3

import (
	"bytes"
	"fmt"
	"image"
	"image/color"

	"github.com/anasrar/chihuahua/next/pkg/tim2"
	graphicsynthesizer "github.com/anasrar/chihuahua/pkg/graphic_synthesizer"
)

func ToImageRBGA(container *tim2.Container) (*image.RGBA, error) {
	picture := container.Picture
	width := int(picture.ImageWidth)
	height := int(picture.ImageHeight)
	swizzle := width >= 128 && height >= 128

	data := make([]byte, len(picture.ImageData))
	copy(data, picture.ImageData)
	clut := make([]byte, len(picture.ClutData))
	copy(clut, picture.ClutData)
	indices := []byte{}

	clutColors := make([]*color.RGBA, 0)
	clutReader := bytes.NewReader(clut)
	rgba := make([]byte, 4)
	for range picture.ClutColors {
		if _, err := clutReader.Read(rgba); err != nil {
			return nil, err
		}

		clutColors = append(
			clutColors,
			&color.RGBA{
				R: rgba[0],
				G: rgba[1],
				B: rgba[2],
				A: uint8(float64(rgba[3]) / 0x80 * 0xFF),
			},
		)
	}

	if picture.ClutColors >= 32 {
		twiddle := []*color.RGBA{}

		for i := 0; i < int(picture.ClutColors); i += 32 {
			twiddle = append(twiddle, clutColors[i+0:i+8]...)
			twiddle = append(twiddle, clutColors[i+16:i+24]...)
			twiddle = append(twiddle, clutColors[i+8:i+16]...)
			twiddle = append(twiddle, clutColors[i+24:i+32]...)
		}

		clutColors = twiddle
	}

	switch picture.ImageType {
	case 0x04:
		if swizzle {
			data = graphicsynthesizer.Unswizzle4(data, width, height)
		}

		for _, v := range data {
			low := v & 0xF
			high := (v >> 4) & 0xF
			indices = append(indices, low)
			indices = append(indices, high)
		}
	case 0x05:
		if swizzle {
			data = graphicsynthesizer.Unswizzle8(data, width, height)
		}

		indices = append(indices, data...)
	default:
		return nil, fmt.Errorf("TIM3: image type 0x%x is not supported", picture.ImageType)
	}

	raw := []uint8{}
	for _, index := range indices {
		c := clutColors[index]
		raw = append(raw, c.R)
		raw = append(raw, c.G)
		raw = append(raw, c.B)
		raw = append(raw, c.A)
	}

	img := image.NewRGBA(image.Rect(0, 0, int(picture.ImageWidth), int(picture.ImageHeight)))
	copy(img.Pix, raw)

	return img, nil
}
