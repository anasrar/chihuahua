package main

import (
	"bytes"

	"github.com/anasrar/chihuahua/pkg/tim2"
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
var picture *tim2.Picture
var camera = rl.NewCamera2D(rl.Vector2Zero(), rl.Vector2Zero(), 0, 1)
var matrix = rl.MatrixIdentity()

var canConvert = false
