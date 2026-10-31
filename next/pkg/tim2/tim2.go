package tim2

import "github.com/anasrar/chihuahua/next/pkg/baseline"

type Header struct {
	Signature     uint32 `json:"signature"`
	FormatVersion uint8  `json:"format_version"`
	FormatID      uint8  `json:"format_id"`
	PictureTotal  uint16 `json:"pictures"`
	Reserved      uint64 `json:"reserved"`
}

type Container struct {
	File    *baseline.File `json:"from"`
	Header  *Header        `json:"header"`
	Picture *Picture       `json:"picture"`
}
