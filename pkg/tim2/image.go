package tim2

import (
	"image"
)

func PictureToImage(picture *Picture) *image.NRGBA {
	width := int(picture.ImageWidth)
	height := int(picture.ImageHeight)
	data := make([]byte, len(picture.ImageData))
	copy(data, picture.ImageData)
	indices := []byte{}

	switch picture.ImageType {
	case ImageType4BitTexture:
		for _, v := range data {
			low := v & 0xF
			high := (v >> 4) & 0xF
			indices = append(indices, low)
			indices = append(indices, high)
		}
	case ImageType8BitTexture:
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
