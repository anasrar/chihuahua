package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
	rlig "github.com/anasrar/chihuahua/pkg/raylib_imgui"
	"github.com/anasrar/chihuahua/pkg/utils"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func drop(filePath string) error {
	pngFile, err := os.Open(filePath)
	if err != nil {
		return err
	}

	img, err := png.Decode(pngFile)
	if err != nil {
		return err
	}

	imgPaletted, ok := img.(*image.Paletted)
	if !ok {
		return fmt.Errorf("PNG is not in indexed mode")
	}
	colorTotal := len(imgPaletted.Palette)

	if colorTotal > 256 {
		return fmt.Errorf("PNG colors exceeds the maximum allowable limit of 256")
	}

	if colorTotal == 16 {
		bpp = 4
		bppIndex = 1
	} else {
		bpp = 8
		bppIndex = 0
	}

	rlImg := rl.NewImageFromImage(img)
	defer rl.UnloadImage(rlImg)
	entries = []rl.Texture2D{rl.LoadTextureFromImage(rlImg)}

	pngPath = filePath
	canConvert = true

	colors = []color.RGBA{}
	for _, v := range imgPaletted.Palette {
		c, _ := v.(color.RGBA)
		colors = append(colors, color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A})
	}

	matrix = rl.MatrixTranslate(
		(width/2)-(float32(entries[0].Width)/2),
		(height/2)-(float32(entries[0].Height)/2),
		0,
	)
	return nil
}

func zoom(wheel float32) {
	scale := float32(0)
	switch wheel {
	case 1:
		scale = 6.0 / 5.0
	case -1:
		scale = 5.0 / 6.0
	}
	positionX := rl.GetMousePosition().X
	positionY := rl.GetMousePosition().Y
	matrix = rl.MatrixMultiply(
		matrix,
		rl.MatrixTranslate(-positionX, -positionY, 0),
	)
	matrix = rl.MatrixMultiply(
		matrix,
		rl.MatrixScale(scale, scale, 1),
	)
	matrix = rl.MatrixMultiply(
		matrix,
		rl.MatrixTranslate(positionX, positionY, 0),
	)
}

func gui() {
	rl.InitWindow(int32(width), int32(height), "DAT Unpacker")
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	rlig.Load()
	defer rlig.Unload()

	for !rl.WindowShouldClose() {
		rlig.Update()

		if rl.IsWindowResized() {
			width = float32(rl.GetScreenWidth())
			height = float32(rl.GetScreenHeight())
		}

		if rl.IsFileDropped() {
			filePath := rl.LoadDroppedFiles()[0]
			defer rl.UnloadDroppedFiles()
			if err := drop(filePath); err != nil {
				log.Println(err)
			}
		}

		if rl.IsMouseButtonDown(1) {
			positionX := rl.GetMouseDelta().X
			positionY := rl.GetMouseDelta().Y
			matrix = rl.MatrixMultiply(
				matrix,
				rl.MatrixTranslate(positionX, positionY, 0),
			)
		}

		wheel := rl.GetMouseWheelMoveV().Y
		if wheel != 0 {
			zoom(wheel)
		}

		imgui.NewFrame()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, 12), imgui.CondAlways, imgui.NewVec2(1, 0))
		imgui.BeginV("View", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.ColorEdit3V("Background", &(background), imgui.ColorEditFlagsNoInputs)
		imgui.Checkbox("Info", &showInfo)
		if imgui.Button("Reset View") {
			if len(entries) != 0 {
				matrix = rl.MatrixTranslate(
					(width/2)-(float32(entries[0].Width)/2),
					(height/2)-(float32(entries[0].Height)/2),
					0,
				)
			}
		}
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, height-12), imgui.CondAlways, imgui.NewVec2(1, 1))
		imgui.BeginV("ToTim", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.BeginDisabledV(!canConvert)
		imgui.PushIDStr("BitPerPixel")
		if imgui.BeginComboV("", bpps[bppIndex], imgui.ComboFlagsWidthFitPreview) {

			for i := range bpps {
				flags := imgui.SelectableFlagsNone
				if i == 1 && len(colors) > 16 {
					flags = imgui.SelectableFlagsDisabled
				}

				selected := i == bppIndex
				if imgui.SelectableBoolV(bpps[i], selected, flags, imgui.NewVec2(0, 0)) {
					bppIndex = i
					switch i {
					case 0:
						bpp = 8
					case 1:
						bpp = 4
					}
				}

				if selected {
					imgui.SetItemDefaultFocus()
				}
			}

			imgui.EndCombo()
		}
		imgui.PopID()
		imgui.SameLineV(0, 4)
		imgui.PushIDStr("Format")
		if imgui.BeginComboV("", formats[formatIndex], imgui.ComboFlagsWidthFitPreview) {

			for i := range formats {
				flags := imgui.SelectableFlagsNone

				selected := i == formatIndex
				if imgui.SelectableBoolV(formats[i], selected, flags, imgui.NewVec2(0, 0)) {
					formatIndex = i
					format = formats[i]
				}

				if selected {
					imgui.SetItemDefaultFocus()
				}
			}

			imgui.EndCombo()
		}
		imgui.PopID()
		imgui.SameLineV(0, 4)
		if imgui.Button("Convert To TIM") {
			go func() {
				log.Println("Convert PNG to TIM")
				if err := convert(pngPath, bpp, format); err != nil {
					log.Println(err)
				} else {
					log.Println("Convert done")
				}
			}()
		}
		imgui.EndDisabled()
		imgui.End()

		if len(entries) != 0 && showInfo {
			imgui.SetNextWindowPosV(imgui.NewVec2(12, height-12), imgui.CondAlways, imgui.NewVec2(0, 1))
			imgui.BeginV("Graphic Synthesizer", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
			entry := entries[0]

			TH := uint8(math.Log2(float64(entry.Height)))
			TW := uint8(math.Log2(float64(entry.Width)))
			TBW := entry.Width / 64

			imgui.Text(
				fmt.Sprintf(
					"Texture Height: %d\nTexture Width: %d\nTexture Buffer Width: %d",
					TH,
					TW,
					TBW,
				),
			)
			imgui.End()
		}

		rl.BeginDrawing()
		rl.ClearBackground(
			rl.NewColor(
				uint8(background[0]*0xFF),
				uint8(background[1]*0xFF),
				uint8(background[2]*0xFF),
				0xFF,
			),
		)
		for _, entry := range entries {
			rl.BeginMode2D(camera)

			translate := rl.NewVector3(0, 0, 0)
			rotation := rl.NewQuaternion(0, 0, 0, 1)
			scale := rl.NewVector3(1, 1, 1)

			utils.MatrixDecompose(matrix, &translate, &rotation, &scale)

			rl.DrawRectangleLinesEx(rl.NewRectangle(translate.X, translate.Y, float32(entry.Width)*scale.X, float32(entry.Height)*scale.Y), 1, rl.Gray)

			rl.PushMatrix()
			rl.Translatef(translate.X, translate.Y, 0)
			rl.Scalef(scale.X, scale.Y, 1)

			rl.DrawTexture(entry, 0, 0, rl.White)

			rl.PopMatrix()

			for i, c := range colors {
				rl.DrawRectangle(int32(i%8*8)+12, int32(i/8*8)+12, 8, 8, c)
			}
			rl.EndMode2D()
		}
		rlig.Render()
		rl.EndDrawing()

	}
}
