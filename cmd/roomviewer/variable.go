package main

import (
	"github.com/anasrar/chihuahua/pkg/dat"
	"github.com/anasrar/chihuahua/pkg/ems"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var datPath = ""
var scp *dat.Entry

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

var showEms = true
var emsEntries = []*ems.Entry{}

var showOms = true
var omsEntries = []*Object{}

var background = [3]float32{0.071, 0.071, 0.071}
