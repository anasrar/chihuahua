package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var t32Path = ""

var (
	width  float32 = 600
	height float32 = 500
)

var background = [3]float32{0.071, 0.071, 0.071}

var camera = rl.NewCamera2D(rl.Vector2Zero(), rl.Vector2Zero(), 0, 1)
var matrix = rl.MatrixIdentity()
var mode = ModeSingle
var entries = []*Entry{}
var currentEntry = -1
var stride = int32(0)
var strideTotal = int32(0)

var canConvert = false
