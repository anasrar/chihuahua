package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var timPath = ""

var (
	width  float32 = 600
	height float32 = 500
)

var background = rl.NewColor(0x12, 0x12, 0x12, 0xFF)

var camera = rl.NewCamera2D(rl.Vector2Zero(), rl.Vector2Zero(), 0, 1)
var matrix = rl.MatrixIdentity()
var mode = ModeSingle
var entries = []*Entry{}
var currentEntry = -1

var canConvert = false

var previewRectangle = rl.NewRectangle(width-74, 58, 64, height-108)
var previewContentRectangle = rl.NewRectangle(0, 0, 42, 0)
var previewScroll = rl.NewVector2(0, 0)
var previewView = rl.NewRectangle(0, 0, 0, 0)
