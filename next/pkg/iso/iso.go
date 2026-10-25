package iso

import (
	"encoding/binary"
	"time"

	"github.com/anasrar/chihuahua/next/pkg/baseline"
)

const SECTOR_SIZE = 2048

var (
	le = binary.LittleEndian
	be = binary.BigEndian
)

type Extent struct {
	LBA    int64 `json:"lba"`
	Offset int64 `json:"offset"`
	Length int64 `json:"length"`
	Sparse bool  `json:"sparse,omitempty"`
}

type Node struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Offset      int64     `json:"offset"`
	LBA         int64     `json:"lba"`
	Length      int64     `json:"size"`
	Date        time.Time `json:"date"`
	IsDirectory bool      `json:"is_directory"`
	IsFile      bool      `json:"is_file"`
	Source      string    `json:"source,omitempty"`
	Extents     []Extent  `json:"extents,omitempty"`
	Children    []*Node   `json:"children,omitempty"`
	Error       string    `json:"error,omitempty"`
}

type Volume struct {
	Name            string `json:"name"`
	SystemID        string `json:"system_id,omitempty"`
	LogicalVolumeID string `json:"logical_volume_id,omitempty"`
	VolumeID        string `json:"volume_id,omitempty"`
	FileSetID       string `json:"file_set_id,omitempty"`
}

type Image struct {
	Volume  Volume `json:"volume"`
	Sectors int64  `json:"sectors"`
	Root    *Node  `json:"root"`
}

type Container struct {
	File  *baseline.File `json:"from"`
	Image *Image         `json:"image"`
}
