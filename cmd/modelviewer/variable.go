package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var datPath = ""

var (
	width  float32 = 1000
	height float32 = 900
)

var camera = rl.NewCamera3D(
	rl.NewVector3(0, 2.8, 2.8),
	rl.NewVector3(0, 1.2, 0),
	rl.NewVector3(0, 1, 0),
	45,
	rl.CameraPerspective,
)

var tm3Entries = []*Entry{}
var textureDefault rl.Texture2D
var textures = map[int]*Texture{}
var textureIndices = []int{}
var textureTotal = 0
var textureShift = 0

var mdEntries = []*Entry{}
var models = []*Model{}
var modelIndex = 0

var showBones = false
var boneRender rl.RenderTexture2D

var motEntries = []*Entry{}
var motionIndex = -1

var background = rl.NewColor(0x12, 0x12, 0x12, 0xFF)

var mdRectangle = rl.NewRectangle(8, 8, 182, 202)
var mdContentRectangle = rl.NewRectangle(0, 0, 162, 0)
var mdScroll = rl.NewVector2(0, 0)
var mdView = rl.NewRectangle(0, 0, 0, 0)

var tm3PreviewRectangle = rl.NewRectangle(width-74, 58, 64, height-108)
var tm3PreviewContentRectangle = rl.NewRectangle(0, 0, 42, 0)
var tm3PreviewScroll = rl.NewVector2(0, 0)
var tm3PreviewView = rl.NewRectangle(0, 0, 0, 0)

var modelRectangle = rl.NewRectangle(8, 218, 182, 202)
var modelContentRectangle = rl.NewRectangle(0, 0, 162, 0)
var modelScroll = rl.NewVector2(0, 0)
var modelView = rl.NewRectangle(0, 0, 0, 0)

var motRectangle = rl.NewRectangle(8, 532, 182, 202)
var motContentRectangle = rl.NewRectangle(0, 0, 162, 0)
var motScroll = rl.NewVector2(0, 0)
var motView = rl.NewRectangle(0, 0, 0, 0)
