package oms

import (
	"fmt"
	"io"
	"os"

	"github.com/anasrar/chihuahua/pkg/buffer"
)

const (
	Signature uint32 = 0x00534D4F
)

type Oms struct {
	Offset     uint32   `json:"offset"`
	EntryTotal uint32   `json:"entry_total"`
	Entries    []*Entry `json:"entries"`
}

func (self *Oms) unmarshal(stream io.ReadWriteSeeker) error {
	if _, err := buffer.Seek(stream, int64(self.Offset), buffer.SeekStart); err != nil {
		return err
	}

	signature := uint32(0)
	if _, err := buffer.ReadUint32LE(stream, &signature); err != nil {
		return err
	}

	if signature != Signature {
		return fmt.Errorf("OMS signature not match")
	}

	if _, err := buffer.ReadUint32LE(stream, &self.EntryTotal); err != nil {
		return err
	}

	if self.EntryTotal == 0 {
		return nil
	}

	if _, err := buffer.Seek(stream, 56, buffer.SeekCurrent); err != nil {
		return err
	}

	for range self.EntryTotal {
		name := ""
		if _, err := buffer.ReadString(stream, &name, 8); err != nil {
			return err
		}

		// TODO: add this 8 char as fx string
		if _, err := buffer.Seek(stream, 8, buffer.SeekCurrent); err != nil {
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
				name,
				[3]float32{translationX, translationY, translationZ},
			),
		)

		if _, err := buffer.Seek(stream, 36, buffer.SeekCurrent); err != nil {
			return err
		}

	}

	return nil
}

func New() *Oms {
	return &Oms{
		Offset:     0,
		EntryTotal: 0,
		Entries:    []*Entry{},
	}
}

func FromPathWithOffset(oms *Oms, filePath string, offset uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	oms.Offset = offset
	return oms.unmarshal(file)
}

func FromPath(oms *Oms, filePath string) error {
	return FromPathWithOffset(oms, filePath, 0)
}
