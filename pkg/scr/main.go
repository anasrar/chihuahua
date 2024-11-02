package scr

import (
	"fmt"
	"io"
	"os"

	"github.com/anasrar/chihuahua/pkg/buffer"
	"github.com/anasrar/chihuahua/pkg/mdb"
)

const (
	Signature uint32 = 0x00726373
)

type Scr struct {
	Offset    uint32  `json:"offset"`
	NodeTotal uint32  `json:"node_total"`
	Nodes     []*Node `json:"nodes"`
}

func (self *Scr) unmarshal(stream io.ReadWriteSeeker) error {
	if _, err := buffer.Seek(stream, int64(self.Offset), buffer.SeekStart); err != nil {
		return err
	}

	signature := uint32(0)
	if _, err := buffer.ReadUint32LE(stream, &signature); err != nil {
		return err
	}

	if signature != Signature {
		return fmt.Errorf("SCR signature not match")
	}

	if _, err := buffer.Seek(stream, 4, buffer.SeekCurrent); err != nil {
		return err
	}

	if _, err := buffer.ReadUint32LE(stream, &self.NodeTotal); err != nil {
		return err
	}

	if _, err := buffer.Seek(stream, 4, buffer.SeekCurrent); err != nil {
		return err
	}

	nodeOffsets := []int64{}
	offset := uint32(0)
	for range self.NodeTotal {
		if _, err := buffer.ReadUint32LE(stream, &offset); err != nil {
			return err
		}
		nodeOffsets = append(nodeOffsets, int64(self.Offset+offset))
	}

	mdbOffset := int32(0)
	for _, offset := range nodeOffsets {
		if _, err := buffer.Seek(stream, offset, buffer.SeekStart); err != nil {
			return err
		}

		if _, err := buffer.ReadInt32LE(stream, &mdbOffset); err != nil {
			return err
		}

		if _, err := buffer.Seek(stream, 4, buffer.SeekCurrent); err != nil {
			return err
		}

		name := ""
		if _, err := buffer.ReadString(stream, &name, 8); err != nil {
			return err
		}

		scaleX := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &scaleX); err != nil {
			return err
		}

		scaleY := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &scaleY); err != nil {
			return err
		}

		scaleZ := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &scaleZ); err != nil {
			return err
		}

		rotationX := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &rotationX); err != nil {
			return err
		}

		rotationY := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &rotationY); err != nil {
			return err
		}

		rotationZ := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &rotationZ); err != nil {
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

		var m mdb.Mdb
		if err := mdb.FromStreamWithOffset(&m, stream, uint32(int32(offset)+mdbOffset)); err != nil {
			return err
		}

		self.Nodes = append(
			self.Nodes,
			NewNode(
				&m,
				name,
				[3]float32{scaleX, scaleY, scaleZ},
				[3]float32{rotationX, rotationY, rotationZ},
				[3]float32{translationX, translationY, translationZ},
			),
		)
	}

	return nil
}

func New() *Scr {
	return &Scr{
		Offset:    0,
		NodeTotal: 0,
		Nodes:     []*Node{},
	}
}

func FromStreamWithOffset(scr *Scr, stream io.ReadWriteSeeker, offset uint32) error {
	scr.Offset = offset
	return scr.unmarshal(stream)
}

func FromStream(scr *Scr, stream io.ReadWriteSeeker) error {
	return FromStreamWithOffset(scr, stream, 0)
}

func FromPathWithOffset(scr *Scr, filePath string, offset uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scr.Offset = offset
	return scr.unmarshal(file)
}

func FromPath(scr *Scr, filePath string) error {
	return FromPathWithOffset(scr, filePath, 0)
}
