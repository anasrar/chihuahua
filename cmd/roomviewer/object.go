package main

import "github.com/anasrar/chihuahua/pkg/oms"

type Object struct {
	*oms.Entry
	RenderLabel bool
}
