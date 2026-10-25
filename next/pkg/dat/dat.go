package dat

import "github.com/anasrar/chihuahua/next/pkg/baseline"

type Header struct {
	EntryTotal uint32    `json:"entry_total"`
	Offsets    []uint32  `json:"offsets" length:"EntryTotal"`
	Types      [][4]byte `json:"types" length:"EntryTotal"`
}

type Entry struct {
	Source         string `json:"source" skip:""`
	AbsoluteOffset uint32 `json:"absolute_offset" skip:""`
	Offset         uint32 `json:"offset" skip:""`
	Length         uint32 `json:"length" skip:""`
	Type           string `json:"type" skip:""`
	IsNull         bool   `json:"is_null" skip:""`
}

type Container struct {
	File    *baseline.File `json:"from"`
	Header  *Header        `json:"header"`
	Entries []*Entry       `json:"entries"`
}
