package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/anasrar/chihuahua/pkg/buffer"
	rlig "github.com/anasrar/chihuahua/pkg/raylib_imgui"
	"github.com/anasrar/chihuahua/pkg/t32"
	"github.com/anasrar/chihuahua/pkg/utils"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func drop(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	signature := uint32(0)
	if _, err := buffer.ReadUint32LE(file, &signature); err != nil {
		return err
	}

	t := t32.New()
	if err := t32.FromPath(t, filePath); err != nil {
		return err
	}

	for _, entry := range entries {
		rl.UnloadTexture(entry.Texture)
	}

	buf := bytes.NewBuffer([]byte{})
	if err := png.Encode(buf, t32.T32ToImage(t)); err != nil {
		return err
	}

	img := rl.LoadImageFromMemory(".png", buf.Bytes(), int32(buf.Len()))
	texture := rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)

	entries = []*Entry{
		{
			Source:  filePath,
			Name:    utils.BasenameWithoutExt(filePath),
			Png:     buf.Bytes(),
			Texture: texture,
			Picture: t,
		},
	}

	currentEntry = 0

	matrix = rl.MatrixTranslate(
		(width/2)-(float32(texture.Width)/2),
		(height/2)-(float32(texture.Height)/2),
		0,
	)

	mode = ModeSingle
	canConvert = true
	stride = 0
	strideTotal = int32(texture.Height) / 64

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

func convert2png(stride, strideTotal int32) error {
	if currentEntry == -1 {
		return fmt.Errorf("T32 not found")
	}
	entry := entries[currentEntry]

	pngFile, err := os.OpenFile(filepath.Join(utils.ParentDirectory(t32Path), fmt.Sprintf("%s.png", entry.Name)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err)
		return err
	}
	defer pngFile.Close()

	if stride <= 1 {
		if _, err := pngFile.Write(entry.Png); err != nil {
			return err
		}
	} else {
		t32Img := t32.T32ToImage(entry.Picture)
		pngWidth := 128 * stride
		pngHeight := (strideTotal / stride) * 64
		pngImg := image.NewNRGBA(image.Rect(0, 0, int(pngWidth), int(pngHeight)))

		for i := range strideTotal {
			scrY := 64 * i
			dstX := 128 * (i % stride)
			dstY := 64 * (i / stride)

			if dstY < pngHeight {
				// NOTE: copy 128 x 64 pixels
				for y := range 64 {
					for x := range 128 {
						c := t32Img.At(x, int(scrY)+y)
						pngImg.Set(int(dstX)+x, int(dstY)+y, c)
					}
				}
			}
		}

		buf := bytes.NewBuffer([]byte{})
		if err := png.Encode(buf, pngImg); err != nil {
			return err
		}

		if _, err := pngFile.Write(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func png2t32(filePath string) error {
	if currentEntry == -1 {
		return fmt.Errorf("T32 not found")
	}
	entry := entries[currentEntry]

	pngFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer pngFile.Close()

	img, err := png.Decode(pngFile)
	if err != nil {
		return err
	}

	imgPaletted, ok := img.(*image.Paletted)
	if !ok {
		return fmt.Errorf("PNG is not in indexed mode")
	}

	output := filepath.Join(
		utils.ParentDirectory(filePath),
		fmt.Sprintf("T32_%s.t32", utils.BasenameWithoutExt(filePath)),
	)

	t32File, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer t32File.Close()

	return t32.ImagePalettedToFile(entry.Source, imgPaletted, t32File)
}

func main() {
	rl.InitWindow(int32(width), int32(height), "T32 Viewer")
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

			if canConvertPng2t32 {
				log.Println("Convert PNG to T32")
				if err := png2t32(filePath); err != nil {
					log.Println(err)
				} else {
					log.Println("Converted")
				}
			} else {
				if err := drop(filePath); err != nil {
					log.Println(err)
				} else {
					t32Path = filePath
				}
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
		if imgui.Button("Reset View") {
			if currentEntry != -1 {
				entry := entries[currentEntry]
				if stride <= 1 {
					matrix = rl.MatrixTranslate(
						(width/2)-(float32(entry.Texture.Width)/2),
						(height/2)-(float32(entry.Texture.Height)/2),
						0,
					)
				} else {
					matrix = rl.MatrixTranslate(
						(width/2)-(float32((128*stride)/2)),
						(height/2)-(float32(((strideTotal/stride)*64)/2)),
						0,
					)
				}
			}
		}
		imgui.End()

		if currentEntry != -1 {
			imgui.SetNextWindowPosV(imgui.NewVec2(12, 12), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
			imgui.BeginV("Information", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
			entry := entries[currentEntry]
			if stride <= 1 {
				imgui.Text(
					fmt.Sprintf(
						"%dx%d",
						entry.Picture.ImageWidth,
						entry.Picture.ImageHeight,
					),
				)
			} else {
				imgui.Text(
					fmt.Sprintf(
						"%dx%d",
						128*stride,
						(strideTotal/stride)*64,
					),
				)
			}
			imgui.End()
		}

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, height-12), imgui.CondAlways, imgui.NewVec2(1, 1))
		imgui.BeginV("ToPng", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.BeginDisabledV(!canConvert)
		imgui.PushIDStr("Stride")
		if imgui.SliderInt("", &(stride), 0, strideTotal) {
			if currentEntry != -1 {
				entry := entries[currentEntry]
				if stride <= 1 {
					matrix = rl.MatrixTranslate(
						(width/2)-(float32(entry.Texture.Width)/2),
						(height/2)-(float32(entry.Texture.Height)/2),
						0,
					)
				} else {
					matrix = rl.MatrixTranslate(
						(width/2)-(float32((128*stride)/2)),
						(height/2)-(float32(((strideTotal/stride)*64)/2)),
						0,
					)
				}
			}
		}
		imgui.PopID()
		imgui.SameLineV(0, 4)
		if imgui.Button("Convert To PNG") {
			go func() {
				log.Println("Convert to PNG")
				if err := convert2png(stride, strideTotal); err != nil {
					log.Println(err)
				} else {
					log.Println("Converted")
				}
			}()
		}
		imgui.EndDisabled()
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(12, height-12), imgui.CondAlways, imgui.NewVec2(0, 1))
		imgui.BeginV("WindowPng2T32", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.BeginDisabledV(!canConvert)

		if imgui.Button("PNG2T32") {
			imgui.OpenPopupStr("PNG2T32")
			canConvertPng2t32 = true
		}

		center := imgui.MainViewport().Center()
		imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.NewVec2(0.5, 0.5))
		if imgui.BeginPopupModalV("PNG2T32", nil, imgui.WindowFlagsAlwaysAutoResize) {
			imgui.Text("Drag and drop PNG")
			if imgui.Button("Cancel") {
				imgui.CloseCurrentPopup()
				canConvertPng2t32 = false
			}
			imgui.EndPopup()
		}

		imgui.EndDisabled()
		imgui.End()

		rl.BeginDrawing()
		rl.ClearBackground(
			rl.NewColor(
				uint8(background[0]*0xFF),
				uint8(background[1]*0xFF),
				uint8(background[2]*0xFF),
				0xFF,
			),
		)

		if currentEntry != -1 {
			entry := entries[currentEntry]

			rl.BeginMode2D(camera)

			translate := rl.NewVector3(0, 0, 0)
			rotation := rl.NewQuaternion(0, 0, 0, 1)
			scale := rl.NewVector3(1, 1, 1)

			utils.MatrixDecompose(matrix, &translate, &rotation, &scale)

			if stride <= 1 {
				rl.DrawRectangleLinesEx(rl.NewRectangle(translate.X, translate.Y, float32(entry.Texture.Width)*scale.X, float32(entry.Texture.Height)*scale.Y), 1, rl.Gray)
			} else {
				rl.DrawRectangleLinesEx(rl.NewRectangle(translate.X, translate.Y, float32(128*stride)*scale.X, float32((strideTotal/stride)*64)*scale.Y), 1, rl.Gray)
			}

			rl.PushMatrix()
			rl.Translatef(translate.X, translate.Y, 0)
			rl.Scalef(scale.X, scale.Y, 1)

			if stride == 0 {
				rl.DrawTexture(entry.Texture, 0, 0, rl.White)
			} else {
				for i := range strideTotal {
					x := float32(128 * (i % stride))
					y := float32(64 * (i / stride))
					rl.DrawTextureRec(entry.Texture, rl.NewRectangle(0, float32(64*i), 128, 64), rl.NewVector2(x, y), rl.White)
				}
			}

			rl.PopMatrix()

			rl.EndMode2D()

			for i, c := range entry.Picture.ClutData {
				rl.DrawRectangle(int32(i%8*8)+16, int32(i/8*8)+54, 8, 8, *c)
			}
		}

		rlig.Render()
		rl.EndDrawing()
	}

	for _, entry := range entries {
		rl.UnloadTexture(entry.Texture)
	}
}
