package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var timPath = ""

var (
	width  float32 = 600
	height float32 = 500
)

var background = [3]float32{0.071, 0.071, 0.071}
var showGsInfo = true

var camera = rl.NewCamera2D(rl.Vector2Zero(), rl.Vector2Zero(), 0, 1)
var matrix = rl.MatrixIdentity()
var mode = ModeSingle
var entries = []*Entry{}
var currentEntry = -1

var canConvert = false

var zoomDeadZone = rl.NewRectangle(width-74, 58, 64, height-108)
