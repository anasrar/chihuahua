package tim3

import (
	"image"

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

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := indices[y*width+x]
			img.Set(x, y, picture.ClutData[index])
		}
	}

	return img
}
