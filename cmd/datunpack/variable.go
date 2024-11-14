package main

import (
	"context"
	"fmt"

	"github.com/anasrar/chihuahua/pkg/dat"
)

var GitCommitHash = "Dev Mode"

var datPath = ""
var datData *dat.Dat = nil

type OffsetUnit int

const (
	OffsetUnitDecimal OffsetUnit = iota
	OffsetUnitHex
)

var offsetUnit = OffsetUnitHex

var (
	width  float32 = 600
	height float32 = 500
)

var ctx, cancel = context.WithCancel(context.Background())

var progress = float32(0)
var canUnpack = false
var canCancel = false

var logAutoScroll = true
var logUpdate = false
var logs = fmt.Sprintf("Build %s\nDrag and Drop DAT File\n", GitCommitHash)
