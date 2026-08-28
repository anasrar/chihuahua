package main

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/anasrar/chihuahua/pkg/tim2"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Entry struct {
	Source     string
	Name       string
	Png        []byte
	Texture    rl.Texture2D
	TextureRef imgui.TextureRef
	Picture    *tim2.Picture
}
