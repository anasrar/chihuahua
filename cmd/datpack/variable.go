package main

import (
	"context"
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var GitCommitHash = "Dev Mode"

var metadataPath = ""

var (
	width  float32 = 600
	height float32 = 500
)

var ctx, cancel = context.WithCancel(context.Background())

var progress = float32(0)
var canUnpack = false
var canCancel = false

var logAutoScroll = true
var logRectangle = rl.NewRectangle(0, 0, width, height-48)
var logContentRectangle = rl.NewRectangle(0, 0, width-20, 48)
var logScroll = rl.NewVector2(0, 0)
var logView = rl.NewRectangle(0, 0, 0, 0)
var logs = fmt.Sprintf("Build %s\nDrag and Drop METADATA.json File\n", GitCommitHash)
