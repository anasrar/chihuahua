package main

import (
	"github.com/anasrar/chihuahua/pkg/t32"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Entry struct {
	Source  string
	Name    string
	Png     []byte
	Texture rl.Texture2D
	Picture *t32.T32
}
