package tm3

import (
	"bytes"

	"github.com/anasrar/chihuahua/next/pkg/baseline"
)

type Header struct {
	Signature uint32 `json:"signature"`
	Total     uint32 `json:"total"`
	Padding   uint64 `json:"-"`
}

type OffsetContainer struct {
	Total   uint32   `json:"total" skip:""`
	Entries []uint32 `json:"value" length:"Total"`
}

type NameContainer struct {
	Total   uint32    `json:"total" skip:""`
	Entries [][8]byte `json:"value" length:"Total"`
}

type Container struct {
	File   *baseline.File   `json:"from" skip:""`
	Header *Header          `json:"header" skip:""`
	Offset *OffsetContainer `json:"offset" skip:""`
	Name   *NameContainer   `json:"name" skip:""`
}

type Entry struct {
	Source         string `json:"source"`
	AbsoluteOffset uint32 `json:"absolute_offset"`
	Offset         uint32 `json:"offset"`
	Length         uint32 `json:"length"`
	Name           string `json:"name"`
}

func (container *Container) ToEntries() []*Entry {
	entries := []*Entry{}
	total := len(container.Name.Entries)

	for i, name := range container.Name.Entries {
		isLast := (total - 1) == i

		offset := container.Offset.Entries[i]
		size := offset - uint32(container.File.Length)
		if !isLast {
			size = container.Offset.Entries[i+1] - offset
		}

		entry := &Entry{
			Source:         container.File.Source,
			AbsoluteOffset: uint32(container.File.Offset) + offset,
			Offset:         offset,
			Length:         size,
			Name:           string(bytes.Trim(name[:], "\x00")),
		}

		entries = append(entries, entry)
	}

	return entries
}
