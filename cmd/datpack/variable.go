package main

import (
	"context"
	"fmt"

	"github.com/anasrar/chihuahua/pkg/dat"
)

var GitCommitHash = "Dev Mode"

var metadataPath = ""
var datMetadata *dat.Metadata = nil

var (
	width  float32 = 600
	height float32 = 500
)

var ctx, cancel = context.WithCancel(context.Background())

var progress = float32(0)
var canPack = false
var canCancel = false

var logAutoScroll = true
var logUpdate = false
var logs = fmt.Sprintf("Build %s\nDrag and Drop METADATA.json File\n", GitCommitHash)
