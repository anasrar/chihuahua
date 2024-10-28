package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var scrPath = ""
var tm3Path = ""

var (
	width  float32 = 800
	height float32 = 500
)

var camera = rl.NewCamera3D(
	rl.NewVector3(0, 2.8, 2.8),
	rl.NewVector3(0, 1.2, 0),
	rl.NewVector3(0, 1, 0),
	45,
	rl.CameraPerspective,
)

var textureDefault rl.Texture2D
var textures = map[int]*Texture{}
var textureIndices = []int{}

var models = []*Model{}

var background = rl.NewColor(0x12, 0x12, 0x12, 0xFF)

var tm3PreviewRectangle = rl.NewRectangle(width-74, 58, 64, height-108)
var tm3PreviewContentRectangle = rl.NewRectangle(0, 0, 42, 0)
var tm3PreviewScroll = rl.NewVector2(0, 0)
var tm3PreviewView = rl.NewRectangle(0, 0, 0, 0)
