package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var pngPath = ""
var bpp = uint(8)
var format = "TIM3"

var (
	width  float32 = 600
	height float32 = 500
)

var background = [3]float32{0.071, 0.071, 0.071}
var showInfo = true

var camera = rl.NewCamera2D(rl.Vector2Zero(), rl.Vector2Zero(), 0, 1)
var matrix = rl.MatrixIdentity()
var entries = []rl.Texture2D{}
var colors = []color.RGBA{}

var bpps = [2]string{"8BitPerPixel", "4BitPerPixel"}
var bppIndex = 0

var formats = [2]string{"TIM3", "TIM2"}
var formatIndex = 0

var canConvert = false
