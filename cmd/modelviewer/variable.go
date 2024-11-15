package main

import (
	"github.com/anasrar/chihuahua/pkg/bone"
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
var boneTree = NewBoneNode(bone.New(0, "root", 0, 0, 0, 0, 0, 0, -1))
var boneNodes = []*BoneNode{boneTree}

var motEntries = []*Entry{}
var motionIndex = -1

var frames = [][]*bone.Bone{} // NOTE: frames -> bone
var frameTotal = int32(0)
var frameIndex = int32(0)
var framePlay = false

var background = [3]float32{0.071, 0.071, 0.071}
