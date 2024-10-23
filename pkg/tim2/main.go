package tim2

import (
	"fmt"
	"image/color"
	"io"
	"os"

	"github.com/anasrar/chihuahua/pkg/buffer"
)

const (
	Signature uint32 = 0x324D4954
)

type Tim2 struct {
	Offset        uint32        `json:"offset"`
	FormatVersion FormatVersion `json:"format_version"`
	FormatId      FormatId      `json:"format_id"`
	PictureTotal  uint16        `json:"picture_total"`
	Pictures      []*Picture    `json:"pictures"`
}

func New() *Tim2 {
	return &Tim2{
		Offset:        0,
		FormatVersion: FormatVersionReserved,
		FormatId:      FormatId16Alignment,
		PictureTotal:  0,
		Pictures:      []*Picture{},
	}
}

func (self *Tim2) unmarshal(stream io.ReadWriteSeeker) error {
	if _, err := buffer.Seek(stream, int64(self.Offset), buffer.SeekStart); err != nil {
		return err
	}

	signature := uint32(0)
	if _, err := buffer.ReadUint32LE(stream, &signature); err != nil {
		return err
	}

	if signature != Signature {
		return fmt.Errorf("TIM2 signature not match")
	}

	version := uint8(0)
	if _, err := buffer.ReadUint8(stream, &version); err != nil {
		return err
	}
	self.FormatVersion = FormatVersion(version)

	id := uint8(0)
	if _, err := buffer.ReadUint8(stream, &id); err != nil {
		return err
	}
	self.FormatId = FormatId(id)

	pictureTotal := uint16(0)
	if _, err := buffer.ReadUint16LE(stream, &pictureTotal); err != nil {
		return err
	}
	self.PictureTotal = pictureTotal

	if _, err := buffer.Seek(stream, 8, buffer.SeekCurrent); err != nil {
		return err
	}

	for range self.PictureTotal {
		picture := Picture{
			ClutData: []*color.RGBA{},
		}

		if _, err := buffer.ReadUint32LE(stream, &picture.TotalSize); err != nil {
			return err
		}

		if _, err := buffer.ReadUint32LE(stream, &picture.ClutSize); err != nil {
			return err
		}

		if _, err := buffer.ReadUint32LE(stream, &picture.ImageSize); err != nil {
			return err
		}

		if _, err := buffer.ReadUint16LE(stream, &picture.HeaderSize); err != nil {
			return err
		}

		if _, err := buffer.ReadUint16LE(stream, &picture.ClutColors); err != nil {
			return err
		}

		if _, err := buffer.ReadUint8(stream, &picture.PictureFormat); err != nil {
			return err
		}

		if _, err := buffer.ReadUint8(stream, &picture.MipMapTextures); err != nil {
			return err
		}

		clut := uint8(0)
		if _, err := buffer.ReadUint8(stream, &clut); err != nil {
			return err
		}
		picture.ClutType = ClutType(clut)

		imageType := uint8(0)
		if _, err := buffer.ReadUint8(stream, &imageType); err != nil {
			return err
		}
		picture.ImageType = ImageType(imageType)

		if _, err := buffer.ReadUint16LE(stream, &picture.ImageWidth); err != nil {
			return err
		}

		if _, err := buffer.ReadUint16LE(stream, &picture.ImageHeight); err != nil {
			return err
		}

		if _, err := buffer.ReadUint64LE(stream, &picture.GsTex0); err != nil {
			return err
		}

		if _, err := buffer.ReadUint64LE(stream, &picture.GsTex1); err != nil {
			return err
		}

		if _, err := buffer.ReadUint32LE(stream, &picture.GsRegs); err != nil {
			return err
		}

		if _, err := buffer.ReadUint32LE(stream, &picture.GsTexClut); err != nil {
			return err
		}

		buf := make([]byte, picture.ImageSize)
		if _, err := buffer.ReadBytes(stream, buf); err != nil {
			return err
		}
		picture.ImageData = buf

		rgba := make([]byte, 4)
		for range picture.ClutColors {
			if _, err := buffer.ReadBytes(stream, rgba); err != nil {
				return err
			}

			picture.ClutData = append(
				picture.ClutData,
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
				twiddle = append(twiddle, picture.ClutData[i+0:i+8]...)
				twiddle = append(twiddle, picture.ClutData[i+16:i+24]...)
				twiddle = append(twiddle, picture.ClutData[i+8:i+16]...)
				twiddle = append(twiddle, picture.ClutData[i+24:i+32]...)
			}

			picture.ClutData = twiddle
		}

		self.Pictures = append(self.Pictures, &picture)
	}

	return nil
}

func FromPathWithOffsetSize(tim2 *Tim2, filePath string, offset uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tim2.Offset = offset
	return tim2.unmarshal(file)
}

func FromPath(dat *Tim2, filePath string) error {
	return FromPathWithOffsetSize(dat, filePath, 0)
}
