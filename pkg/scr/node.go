package scr

import (
	"github.com/anasrar/chihuahua/pkg/mdb"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Node struct {
	Mdb         *mdb.Mdb   `json:"mdb"`
	Name        string     `json:"name"`
	Translation rl.Vector3 `json:"translation"`
	Rotation    rl.Vector3 `json:"rotation"`
	Scale       rl.Vector3 `json:"scale"`
}

func NewNode(
	m *mdb.Mdb,
	name string,
	scale rl.Vector3,
	rotation rl.Vector3,
	translation rl.Vector3,
) *Node {
	return &Node{
		Mdb:         m,
		Name:        name,
		Translation: translation,
		Rotation:    rotation,
		Scale:       scale,
	}
}
