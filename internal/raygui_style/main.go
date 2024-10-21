package rayguistyle

import (
	_ "embed"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed PixelOperator.ttf
var defaultFontSource []byte
var DefaultFont rl.Font

func Load() {
	chars := []int32{0}
	for i := int32(32); i < 256; i++ {
		chars = append(chars, i)
	}

	DefaultFont = rl.LoadFontFromMemory(".ttf", defaultFontSource, int32(len(defaultFontSource)), 16, &chars[0], int32(len(chars)))
	raygui.SetFont(DefaultFont)

	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_SIZE, 16)

	raygui.SetStyle(raygui.DEFAULT, raygui.BORDER_COLOR_NORMAL, 0x3A3A3AFF)
	raygui.SetStyle(raygui.DEFAULT, raygui.BORDER_COLOR_FOCUSED, 0x4A4A4AFF)
	raygui.SetStyle(raygui.DEFAULT, raygui.BORDER_COLOR_PRESSED, 0x3F3F3FFF)

	raygui.SetStyle(raygui.DEFAULT, raygui.BACKGROUND_COLOR, 0x121212FF)
	raygui.SetStyle(raygui.DEFAULT, raygui.BACKGROUND_COLOR, 0x121212FF)

	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_COLOR_NORMAL, 0xDADADAFF)
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_COLOR_FOCUSED, 0xDADADAFF)
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_COLOR_PRESSED, 0xDADADAFF)

	raygui.SetStyle(raygui.BUTTON, raygui.BORDER_WIDTH, 1)
	raygui.SetStyle(raygui.BUTTON, raygui.BASE_COLOR_NORMAL, 0x202020FF)
	raygui.SetStyle(raygui.BUTTON, raygui.BORDER_COLOR_NORMAL, 0x404040FF)
	raygui.SetStyle(raygui.BUTTON, raygui.TEXT_COLOR_NORMAL, 0xDADADAFF)

	raygui.SetStyle(raygui.BUTTON, raygui.BASE_COLOR_FOCUSED, 0x303030FF)
	raygui.SetStyle(raygui.BUTTON, raygui.BORDER_COLOR_FOCUSED, 0x404040FF)
	raygui.SetStyle(raygui.BUTTON, raygui.TEXT_COLOR_FOCUSED, 0xDADADAFF)

	raygui.SetStyle(raygui.BUTTON, raygui.BASE_COLOR_PRESSED, 0x3A3A3AFF)
	raygui.SetStyle(raygui.BUTTON, raygui.BORDER_COLOR_PRESSED, 0x404040FF)
	raygui.SetStyle(raygui.BUTTON, raygui.TEXT_COLOR_PRESSED, 0xDADADAFF)

	raygui.SetStyle(raygui.BUTTON, raygui.BASE_COLOR_DISABLED, 0x3A3A3A60)
	raygui.SetStyle(raygui.BUTTON, raygui.BORDER_COLOR_DISABLED, 0x40404060)
	raygui.SetStyle(raygui.BUTTON, raygui.TEXT_COLOR_DISABLED, 0xDADADA60)

	raygui.SetStyle(raygui.CHECKBOX, raygui.BASE_COLOR_NORMAL, 0x202020FF)
	raygui.SetStyle(raygui.CHECKBOX, raygui.BORDER_COLOR_NORMAL, 0x404040FF)
	raygui.SetStyle(raygui.CHECKBOX, raygui.TEXT_COLOR_NORMAL, 0xDADADAFF)

	raygui.SetStyle(raygui.CHECKBOX, raygui.BASE_COLOR_FOCUSED, 0x303030FF)
	raygui.SetStyle(raygui.CHECKBOX, raygui.BORDER_COLOR_FOCUSED, 0x404040FF)
	raygui.SetStyle(raygui.CHECKBOX, raygui.TEXT_COLOR_FOCUSED, 0xDADADAFF)

	raygui.SetStyle(raygui.CHECKBOX, raygui.BASE_COLOR_PRESSED, 0x3A3A3AFF)
	raygui.SetStyle(raygui.CHECKBOX, raygui.BORDER_COLOR_PRESSED, 0x404040FF)
	raygui.SetStyle(raygui.CHECKBOX, raygui.TEXT_COLOR_PRESSED, 0xDADADAFF)

	raygui.SetStyle(raygui.PROGRESSBAR, raygui.BORDER_COLOR_FOCUSED, 0xDADADAFF)
	raygui.SetStyle(raygui.PROGRESSBAR, raygui.BASE_COLOR_PRESSED, 0xDADADAFF)
}

func Unload() {
	rl.UnloadFont(DefaultFont)
}
