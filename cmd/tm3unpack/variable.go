package main

import (
	"context"
	"fmt"

	"github.com/anasrar/chihuahua/pkg/tm3"
)

var GitCommitHash = "Dev Mode"

var tm3Path = ""
var tm3Data *tm3.Tm3 = nil

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
var logs = fmt.Sprintf("Build %s\nDrag and Drop TM3 File\n", GitCommitHash)
