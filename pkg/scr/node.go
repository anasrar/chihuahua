package scr

import (
	"github.com/anasrar/chihuahua/pkg/mdb"
)

type Node struct {
	Mdb         *mdb.Mdb   `json:"mdb"`
	Name        string     `json:"name"`
	Translation [3]float32 `json:"translation"`
	Rotation    [3]float32 `json:"rotation"`
	Scale       [3]float32 `json:"scale"`
}

func NewNode(
	m *mdb.Mdb,
	name string,
	scale [3]float32,
	rotation [3]float32,
	translation [3]float32,
) *Node {
	return &Node{
		Mdb:         m,
		Name:        name,
		Translation: translation,
		Rotation:    rotation,
		Scale:       scale,
	}
}
