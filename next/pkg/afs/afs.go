package afs

import "github.com/anasrar/chihuahua/next/pkg/baseline"

type Header struct {
	SignatureVersion uint32 `json:"signature_version"`
}

func (header *Header) Signature() uint32 {
	return header.SignatureVersion & 0x00FFFFFF
}

func (header *Header) Version() uint32 {
	return (header.SignatureVersion & 0xFF000000) >> (3 * 8)
}

type EntryDescription struct {
	Name       [32]byte `json:"name"`
	Year       uint16   `json:"year"`
	Month      uint16   `json:"month"`
	Day        uint16   `json:"day"`
	Hour       uint16   `json:"hour"`
	Minute     uint16   `json:"minute"`
	Second     uint16   `json:"second"`
	CustomData uint32   `json:"custom_data"`
}

type Entry struct {
	Source         string            `json:"source" skip:""`
	AbsoluteOffset uint32            `json:"absolute_offset" skip:""`
	Offset         uint32            `json:"offset"`
	Length         uint32            `json:"length"`
	Description    *EntryDescription `json:"description" skip:""`
	IsNull         bool              `json:"is_null" skip:""`
}

type AttributeData struct {
	Offset uint32 `json:"offset"`
	Length uint32 `json:"length"`
}

type Container struct {
	File          *baseline.File `json:"from" skip:""`
	Header        *Header        `json:"header"`
	EntryTotal    uint32         `json:"entry_total"`
	Entries       []*Entry       `json:"entries" skip:""`
	AttributeData *AttributeData `json:"attribute_data" skip:""`
}
