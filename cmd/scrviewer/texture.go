package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Texture struct {
	Texture rl.Texture2D
}

func NewTexture(texture rl.Texture2D) *Texture {
	return &Texture{
		Texture: texture,
	}
}
