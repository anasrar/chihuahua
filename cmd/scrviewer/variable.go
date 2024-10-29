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
var textureTotal = 0
var textureShift = 0

var applyScrTransform = false
var models = []*Model{}

var showBones = false
var boneTree = NewBoneNode(0, 0, 0)
var bones = []*BoneNode{}
var boneRender rl.RenderTexture2D

var background = rl.NewColor(0x12, 0x12, 0x12, 0xFF)

var tm3PreviewRectangle = rl.NewRectangle(width-74, 58, 64, height-108)
var tm3PreviewContentRectangle = rl.NewRectangle(0, 0, 42, 0)
var tm3PreviewScroll = rl.NewVector2(0, 0)
var tm3PreviewView = rl.NewRectangle(0, 0, 0, 0)

var modelRectangle = rl.NewRectangle(8, 8, 182, 202)
var modelContentRectangle = rl.NewRectangle(0, 0, 162, 0)
var modelScroll = rl.NewVector2(0, 0)
var modelView = rl.NewRectangle(0, 0, 0, 0)
