package utils

import (
	"path"
	"strings"
	"unicode"
)

func FilterUnprintableString(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, str)
}

func Basename(p string) string {
	return path.Base(p)
}

func BasenameWithoutExt(p string) string {
	return strings.TrimSuffix(Basename(p), path.Ext(p))
}

func ParentDirectory(p string) string {
	return path.Dir(p)
}
