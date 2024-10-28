package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Model struct {
	Name    string
	Model   *rl.Model
	Texture int
}

func NewModel(name string, model *rl.Model, texture int) *Model {
	return &Model{
		Name:    name,
		Model:   model,
		Texture: texture,
	}
}
