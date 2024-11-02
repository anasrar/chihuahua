package ems

import (
	"fmt"
	"io"
	"os"

	"github.com/anasrar/chihuahua/pkg/buffer"
)

const (
	Signature uint32 = 0x00534D45
)

type Ems struct {
	Offset     uint32   `json:"offset"`
	EntryTotal uint32   `json:"entry_total"`
	Entries    []*Entry `json:"entries"`
}

func (self *Ems) unmarshal(stream io.ReadWriteSeeker) error {
	if _, err := buffer.Seek(stream, int64(self.Offset), buffer.SeekStart); err != nil {
		return err
	}

	signature := uint32(0)
	if _, err := buffer.ReadUint32LE(stream, &signature); err != nil {
		return err
	}

	if signature != Signature {
		return fmt.Errorf("EMS signature not match")
	}

	if _, err := buffer.ReadUint32LE(stream, &self.EntryTotal); err != nil {
		return err
	}

	if self.EntryTotal == 0 {
		return nil
	}

	for range self.EntryTotal {
		// TODO: research this padding
		if _, err := buffer.Seek(stream, 4, buffer.SeekCurrent); err != nil {
			return err
		}

		translationX := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &translationX); err != nil {
			return err
		}

		translationY := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &translationY); err != nil {
			return err
		}

		translationZ := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &translationZ); err != nil {
			return err
		}

		self.Entries = append(
			self.Entries,
			NewEntry(
				[3]float32{translationX, translationY, translationZ},
			),
		)

		// TODO: research this padding
		if _, err := buffer.Seek(stream, 48, buffer.SeekCurrent); err != nil {
			return err
		}

	}

	return nil
}

func New() *Ems {
	return &Ems{
		Offset:     0,
		EntryTotal: 0,
		Entries:    []*Entry{},
	}
}

func FromPathWithOffset(ems *Ems, filePath string, offset uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	ems.Offset = offset
	return ems.unmarshal(file)
}

func FromPath(ems *Ems, filePath string) error {
	return FromPathWithOffset(ems, filePath, 0)
}
