package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Model struct {
	Name        string
	Model       *rl.Model
	Texture     int
	Render      bool
	Translation rl.Vector3
	Rotation    rl.Vector3
	Scale       rl.Vector3
}

func NewModel(
	name string,
	model *rl.Model,
	texture int,
	translation rl.Vector3,
	rotation rl.Vector3,
	scale rl.Vector3,
) *Model {
	return &Model{
		Name:        name,
		Model:       model,
		Texture:     texture,
		Render:      true,
		Translation: translation,
		Rotation:    rotation,
		Scale:       scale,
	}
}
