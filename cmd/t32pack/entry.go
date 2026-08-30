package main

import (
	"github.com/AllenDang/cimgui-go/imgui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Entry struct {
	Name       string
	Texture    rl.Texture2D
	TextureRef imgui.TextureRef
}
