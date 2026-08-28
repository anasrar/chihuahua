package main

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/anasrar/chihuahua/pkg/tim2"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Texture struct {
	Picture *tim2.Picture
	Texture rl.Texture2D
	Ref     imgui.TextureRef
}

func NewTexture(picture *tim2.Picture, texture rl.Texture2D) *Texture {
	return &Texture{
		Picture: picture,
		Texture: texture,
		Ref:     *imgui.NewTextureRefTextureID(imgui.TextureID(texture.ID)),
	}
}
