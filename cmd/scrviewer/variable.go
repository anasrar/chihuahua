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

var background = [3]float32{0.071, 0.071, 0.071}
