package tim2

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
)

type Picture struct {
	TotalSize      uint32  `json:"total_size"`
	ClutSize       uint32  `json:"clut_size"`
	ImageSize      uint32  `json:"image_size"`
	HeaderSize     uint16  `json:"header_size"`
	ClutColors     uint16  `json:"clut_colors"`
	PictFormat     uint8   `json:"pict_format"`
	MipMapTextures uint8   `json:"mipmap_textures"`
	ClutType       uint8   `json:"clut_type"`
	ImageType      uint8   `json:"image_type"`
	ImageWidth     uint16  `json:"image_width"`
	ImageHeight    uint16  `json:"image_height"`
	GSTex0         uint64  `json:"gs_tex0"`
	GSTex1         uint64  `json:"gs_tex1"`
	GSRegs         uint32  `json:"gs_regs"`
	GSTexClut      uint32  `json:"gs_tex_clut"`
	ImageData      []uint8 `json:"image_data" length:"ImageSize"`
	ClutData       []uint8 `json:"clut_data" length:"ClutSize"`
}

func ToImageRBGA(container *Container) (*image.RGBA, error) {
	picture := container.Picture
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
		for _, v := range data {
			low := v & 0xF
			high := (v >> 4) & 0xF
			indices = append(indices, low)
			indices = append(indices, high)
		}
	case 0x05:
		indices = append(indices, data...)
	default:
		return nil, fmt.Errorf("TIM2: image type 0x%x is not supported", picture.ImageType)
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
