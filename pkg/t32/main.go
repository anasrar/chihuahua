package t32

import (
	"image/color"
	"io"
	"os"

	"github.com/anasrar/chihuahua/pkg/buffer"
)

type T32 struct {
	Offset      uint32 `json:"offset"`
	ImageWidth  uint16 `json:"image_width"`
	ImageHeight uint16 `json:"image_height"`
	ImageData   []byte
	ClutData    []*color.RGBA
}

func New() *T32 {
	return &T32{
		Offset:      0,
		ImageWidth:  0,
		ImageHeight: 0,
		ImageData:   []byte{},
		ClutData:    []*color.RGBA{},
	}
}

func (self *T32) unmarshal(stream io.ReadWriteSeeker) error {
	if _, err := buffer.Seek(stream, int64(self.Offset), buffer.SeekStart); err != nil {
		return err
	}

	if _, err := buffer.Seek(stream, 12, buffer.SeekCurrent); err != nil {
		return err
	}

	clutOffset := uint32(0)
	if _, err := buffer.ReadUint32LE(stream, &clutOffset); err != nil {
		return err
	}

	if _, err := buffer.Seek(stream, 208, buffer.SeekCurrent); err != nil {
		return err
	}

	imageDataSize := clutOffset - 256
	imageData := make([]byte, imageDataSize)
	if _, err := buffer.ReadBytes(stream, imageData); err != nil {
		return err
	}
	self.ImageData = imageData

	self.ImageWidth = 128
	self.ImageHeight = uint16(imageDataSize / 128)

	if _, err := buffer.Seek(stream, 256, buffer.SeekCurrent); err != nil {
		return err
	}

	rgba := make([]byte, 4)
	for range 256 {
		if _, err := buffer.ReadBytes(stream, rgba); err != nil {
			return err
		}

		self.ClutData = append(
			self.ClutData,
			&color.RGBA{
				R: rgba[0],
				G: rgba[1],
				B: rgba[2],
				A: uint8(float64(rgba[3]) / 0x80 * 0xFF),
			},
		)
	}

	twiddle := []*color.RGBA{}
	for i := 0; i < 256; i += 32 {
		twiddle = append(twiddle, self.ClutData[i+0:i+8]...)
		twiddle = append(twiddle, self.ClutData[i+16:i+24]...)
		twiddle = append(twiddle, self.ClutData[i+8:i+16]...)
		twiddle = append(twiddle, self.ClutData[i+24:i+32]...)
	}

	self.ClutData = twiddle

	return nil
}

func FromPathWithOffset(t32 *T32, filePath string, offset uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	t32.Offset = offset
	return t32.unmarshal(file)
}

func FromPath(t32 *T32, filePath string) error {
	return FromPathWithOffset(t32, filePath, 0)
}
