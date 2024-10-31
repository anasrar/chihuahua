package mot

import (
	"fmt"
	"io"
	"os"

	"github.com/anasrar/chihuahua/pkg/buffer"
)

const (
	Signature uint32 = 0x3362746D
)

type Mot struct {
	Offset              uint32    `json:"offset"`
	Size                uint32    `json:"size"`
	FrameTotal          uint16    `json:"frame_total"`
	RecordTotal         uint8     `json:"record_total"`
	UseInverseKinematic uint8     `json:"use_inverse_kinematic"`
	Records             []*Record `json:"record"`
}

func (self *Mot) unmarshal(stream io.ReadWriteSeeker) error {
	if self.Size == 0 {
		size, _ := buffer.Seek(stream, 0, buffer.SeekEnd)
		self.Size = uint32(size) - self.Offset
	}

	if _, err := buffer.Seek(stream, int64(self.Offset), buffer.SeekStart); err != nil {
		return err
	}

	signature := uint32(0)
	if _, err := buffer.ReadUint32LE(stream, &signature); err != nil {
		return err
	}

	if signature != Signature {
		return fmt.Errorf("MOT signature not match")
	}

	if _, err := buffer.ReadUint16LE(stream, &self.FrameTotal); err != nil {
		return err
	}

	if _, err := buffer.ReadUint8(stream, &self.RecordTotal); err != nil {
		return err
	}

	if _, err := buffer.ReadUint8(stream, &self.UseInverseKinematic); err != nil {
		return err
	}

	keyframeOffsets := []int64{}
	for range self.RecordTotal {
		record := NewRecord()

		if _, err := buffer.ReadUint8(stream, &record.Target); err != nil {
			return err
		}

		record.Target += 1

		if _, err := buffer.ReadUint8(stream, &record.Channel); err != nil {
			return err
		}

		if _, err := buffer.ReadUint16LE(stream, &record.CurveTotal); err != nil {
			return err
		}

		if _, err := buffer.ReadUint32LE(stream, &record.UseGlobalTransform); err != nil {
			return err
		}

		offset := uint32(0)
		if _, err := buffer.ReadUint32LE(stream, &offset); err != nil {
			return err
		}

		record.IsNull = offset == 0 || offset > self.Size

		keyframeOffsets = append(
			keyframeOffsets,
			int64(self.Offset+offset),
		)

		self.Records = append(
			self.Records,
			record,
		)
	}

	for i, offset := range keyframeOffsets {
		record := self.Records[i]

		if record.IsNull {
			continue
		}

		if _, err := buffer.Seek(stream, offset, buffer.SeekStart); err != nil {
			return err
		}

		if _, err := buffer.ReadUint16LE(stream, &record.Position); err != nil {
			return err
		}

		if _, err := buffer.ReadUint16LE(stream, &record.PositionDelta); err != nil {
			return err
		}

		if _, err := buffer.ReadUint16LE(stream, &record.Tangent0); err != nil {
			return err
		}

		if _, err := buffer.ReadUint16LE(stream, &record.TangentDelta0); err != nil {
			return err
		}

		if _, err := buffer.ReadUint16LE(stream, &record.Tangent1); err != nil {
			return err
		}

		if _, err := buffer.ReadUint16LE(stream, &record.TangentDelta1); err != nil {
			return err
		}

		for range record.CurveTotal {
			curve := NewCurve()

			if _, err := buffer.ReadUint8(stream, &curve.FrameDelta); err != nil {
				return err
			}

			if _, err := buffer.ReadUint8(stream, &curve.ControlPoint); err != nil {
				return err
			}

			if _, err := buffer.ReadUint8(stream, &curve.ControlTangent0); err != nil {
				return err
			}

			if _, err := buffer.ReadUint8(stream, &curve.ControlTangent1); err != nil {
				return err
			}

			record.Curves = append(record.Curves, curve)
		}
	}

	return nil
}

func New() *Mot {
	return &Mot{
		Offset:              0,
		FrameTotal:          0,
		RecordTotal:         0,
		UseInverseKinematic: 0,
		Records:             []*Record{},
	}
}

func FromPathWithOffsetSize(mot *Mot, filePath string, offset uint32, size uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	mot.Offset = offset
	mot.Size = size
	return mot.unmarshal(file)
}

func FromPath(mot *Mot, filePath string) error {
	return FromPathWithOffsetSize(mot, filePath, 0, 0)
}
