package t32

import (
	"image"

	graphicsynthesizer "github.com/anasrar/chihuahua/pkg/graphic_synthesizer"
)

func T32ToImage(t32 *T32) *image.NRGBA {
	width := int(t32.ImageWidth)
	height := int(t32.ImageHeight)
	dataSize := len(t32.ImageData)
	data := make([]byte, dataSize)
	copy(data, t32.ImageData)

	indices := []byte{}
	for i := 0; i < dataSize; i += 128 * 64 {
		chunk := data[i : i+8192]
		chunk = graphicsynthesizer.Unswizzle8(chunk, 128, 64)
		indices = append(indices, chunk...)
	}

	raw := []uint8{}
	for _, index := range indices {
		c := t32.ClutData[index]
		raw = append(raw, c.R)
		raw = append(raw, c.G)
		raw = append(raw, c.B)
		raw = append(raw, c.A)
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	copy(img.Pix, raw)

	return img
}
