package utils

import (
	"path/filepath"
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
	return filepath.Base(p)
}

func BasenameWithoutExt(p string) string {
	return strings.TrimSuffix(Basename(p), filepath.Ext(p))
}

func ParentDirectory(p string) string {
	return filepath.Dir(p)
}
