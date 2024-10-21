package main

import (
	"bytes"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var timPath = ""

var (
	width  float32 = 600
	height float32 = 500
)

var background = rl.NewColor(0x12, 0x12, 0x12, 0xFF)

var buf = bytes.NewBuffer([]byte{})
var textures = []rl.Texture2D{}
var position = rl.NewVector2(0, 0)
var scale = float32(1)

var canConvert = false
