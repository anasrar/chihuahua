package main

import "github.com/anasrar/chihuahua/pkg/dat"

type Entry struct {
	Name string
	*dat.Entry
}

func NewEntry(name string, entry *dat.Entry) *Entry {
	return &Entry{
		Name:  name,
		Entry: entry,
	}
}
