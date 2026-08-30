package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/AllenDang/cimgui-go/imgui"
	rlig "github.com/anasrar/chihuahua/pkg/raylib_imgui"
	"github.com/anasrar/chihuahua/pkg/t32"
	"github.com/anasrar/chihuahua/pkg/utils"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func drop(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var m t32.Metadata
	if err := json.Unmarshal(file, &m); err != nil {
		return err
	}

	for _, entry := range entries {
		rl.UnloadTexture(entry.Texture)
		entry.TextureRef.Destroy()
	}

	entries = []*Entry{}

	for _, textures := range m.Textures {
		for _, palette := range textures.Palettes {
			filename := filepath.Base(palette.Path)

			pngFile, err := os.Open(palette.Path)
			if err != nil {
				return err
			}
			defer pngFile.Close()

			img, err := png.Decode(pngFile)
			if err != nil {
				return err
			}

			palettedImg, isPalleted := img.(*image.Paletted)
			if !isPalleted {
				return fmt.Errorf("PNG not a palleted: %s", palette.Path)
			}

			buf := bytes.NewBuffer([]byte{})
			if err := png.Encode(buf, palettedImg); err != nil {
				return err
			}

			imgMem := rl.LoadImageFromMemory(".png", buf.Bytes(), int32(buf.Len()))
			texture := rl.LoadTextureFromImage(imgMem)
			rl.UnloadImage(imgMem)

			entries = append(
				entries,
				&Entry{
					Name:       filename,
					Texture:    texture,
					TextureRef: *imgui.NewTextureRefTextureID(imgui.TextureID(texture.ID)),
				},
			)
		}
	}

	entry := entries[0]
	currentEntry = 0

	matrix = rl.MatrixTranslate(
		(width/2)-(float32(entry.Texture.Width)/2),
		(height/2)-(float32(entry.Texture.Height)/2),
		0,
	)

	canConvert = true

	return nil
}

func zoom(wheel float32) {
	isInsidePreview := rl.CheckCollisionPointRec(rl.GetMousePosition(), zoomDeadZone)

	if isInsidePreview {
		return
	}

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
	rl.InitWindow(int32(width), int32(height), "T32 Unpack")
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
			} else {
				metadataPath = filePath
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
				matrix = rl.MatrixTranslate(
					(width/2)-(float32(entry.Texture.Width)/2),
					(height/2)-(float32(entry.Texture.Height)/2),
					0,
				)
			}
		}
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, 104), imgui.CondAlways, imgui.NewVec2(1, 0))
		imgui.SetNextWindowSizeConstraints(imgui.NewVec2(84, 108), imgui.NewVec2(84, 216))
		imgui.BeginV("Entries", nil, imgui.WindowFlagsNoMove)
		for i, entry := range entries {
			if imgui.ImageButton(entry.Name, entry.TextureRef, imgui.NewVec2(42, 42)) {
				currentEntry = i
				matrix = rl.MatrixTranslate(
					(width/2)-(float32(entry.Texture.Width)/2),
					(height/2)-(float32(entry.Texture.Height)/2),
					0,
				)
			}
		}

		{
			windowEntriesRect := imgui.InternalCurrentWindow().Size()
			zoomDeadZoneMin := imgui.InternalCurrentWindow().Pos()
			zoomDeadZone.X = zoomDeadZoneMin.X
			zoomDeadZone.Y = zoomDeadZoneMin.Y
			zoomDeadZone.Width = windowEntriesRect.X
			zoomDeadZone.Height = windowEntriesRect.Y
		}

		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, height-12), imgui.CondAlways, imgui.NewVec2(1, 1))
		imgui.BeginV("ToPng", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.BeginDisabledV(!canConvert)
		if imgui.Button("Pack") {
			go func() {
				if err := pack(
					context.Background(),
					metadataPath,
					func(total, current uint32, name string) {
						log.Printf("% 8d/%d(%s): start\n", current, total, name)
					},
					func(total, current uint32, name string) {
						log.Printf("% 8d/%d(%s): done\n", current, total, name)
					},
				); err != nil {
					log.Print(err)
				}
			}()
		}
		imgui.EndDisabled()
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(12, height-12), imgui.CondAlways, imgui.NewVec2(0, 1))
		imgui.BeginV("Credit", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		if currentEntry == -1 {
			imgui.Text("Drag and drop METADATA.json")
		}
		imgui.Text("https://github.com/anasrar/chihuahua")
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

			rl.DrawRectangleLinesEx(rl.NewRectangle(translate.X, translate.Y, float32(entry.Texture.Width)*scale.X, float32(entry.Texture.Height)*scale.Y), 1, rl.Gray)

			rl.PushMatrix()
			rl.Translatef(translate.X, translate.Y, 0)
			rl.Scalef(scale.X, scale.Y, 1)

			rl.DrawTexture(entry.Texture, 0, 0, rl.White)

			rl.PopMatrix()

			rl.EndMode2D()
		}

		rlig.Render()
		rl.EndDrawing()
	}

	for _, entry := range entries {
		rl.UnloadTexture(entry.Texture)
		entry.TextureRef.Destroy()
	}
}
