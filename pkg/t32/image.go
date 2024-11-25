package t32

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/anasrar/chihuahua/pkg/buffer"
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

func ImagePalettedToFile(t32Path string, img *image.Paletted, output *os.File) error {
	colorTotal := len(img.Palette)

	if colorTotal > 256 {
		return fmt.Errorf("PNG colors exceeds the maximum allowable limit of 256")
	}

	t32File, err := os.Open(t32Path)
	if err != nil {
		return err
	}
	defer t32File.Close()

	if _, err := buffer.Seek(t32File, 12, buffer.SeekCurrent); err != nil {
		return err
	}

	clutOffset := uint32(0)
	if _, err := buffer.ReadUint32LE(t32File, &clutOffset); err != nil {
		return err
	}

	imageDataSize := clutOffset - 256
	targetDataSIze := uint32(img.Rect.Max.X * img.Rect.Max.Y)

	if imageDataSize != targetDataSIze {
		return fmt.Errorf("PNG size is not match, expected %d, got %d", imageDataSize, targetDataSIze)
	}

	if _, err := buffer.Seek(t32File, 0, buffer.SeekStart); err != nil {
		return err
	}

	headerImage := make([]byte, 224)
	if _, err := buffer.ReadBytes(t32File, headerImage); err != nil {
		return err
	}

	if _, err := buffer.WriteBytes(output, headerImage); err != nil {
		return err
	}

	imageData := []byte{}

	imgWidth := img.Rect.Max.X
	imgHeight := img.Rect.Max.Y

	for y := 0; y < imgHeight; y += 64 {
		for x := 0; x < imgWidth; x += 128 {
			indices := []uint8{}

			for ty := 0; ty < 64; ty++ {
				for tx := 0; tx < 128; tx++ {
					indices = append(indices, img.ColorIndexAt(x+tx, y+ty))
				}
			}

			imageData = append(imageData, graphicsynthesizer.Swizzle8(indices, 128, 64)...)
		}
	}

	if _, err := buffer.WriteBytes(output, imageData); err != nil {
		return err
	}

	pad := make([]byte, 32)
	if _, err := buffer.WriteBytes(output, pad); err != nil {
		return err
	}

	headerPalette := []byte{
		0x46, 0x00, 0x00, 0x30, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x70, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x0E, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x51, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x10, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x52, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x53, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x40, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	if _, err := buffer.WriteBytes(output, headerPalette); err != nil {
		return err
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

	for _, c := range twiddle {
		a := uint8(float32(c.A) / 255 * 0x80)
		if _, err := buffer.WriteBytes(output, []byte{c.R, c.G, c.B, a}); err != nil {
			return err
		}
	}

	return nil
}
