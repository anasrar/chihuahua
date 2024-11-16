package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/anasrar/chihuahua/pkg/buffer"
	rlig "github.com/anasrar/chihuahua/pkg/raylib_imgui"
	"github.com/anasrar/chihuahua/pkg/tim2"
	"github.com/anasrar/chihuahua/pkg/tim3"
	"github.com/anasrar/chihuahua/pkg/tm3"
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

	switch signature {
	case tim3.Signature:
		tim := tim3.New()
		if err := tim3.FromPath(tim, filePath); err != nil {
			return err
		}

		for _, entry := range entries {
			rl.UnloadTexture(entry.Texture)
		}

		buf := bytes.NewBuffer([]byte{})
		if err := png.Encode(buf, tim3.PictureToImage(tim.Pictures[0])); err != nil {
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
				Picture: tim.Pictures[0],
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
	case tim2.Signature:
		tim := tim2.New()
		if err := tim2.FromPath(tim, filePath); err != nil {
			return err
		}

		for _, entry := range entries {
			rl.UnloadTexture(entry.Texture)
		}

		buf := bytes.NewBuffer([]byte{})
		if err := png.Encode(buf, tim2.PictureToImage(tim.Pictures[0])); err != nil {
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
				Picture: tim.Pictures[0],
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
	case tm3.Signature:
		tm := tm3.New()
		if err := tm3.FromPath(tm, filePath); err != nil {
			return err
		}

		for _, entry := range entries {
			rl.UnloadTexture(entry.Texture)
		}

		entries = []*Entry{}

		for i, entry := range tm.Entries {
			tim := tim3.New()
			if err := tim3.FromPathWithOffset(tim, filePath, entry.Offset); err != nil {
				return err
			}

			buf := bytes.NewBuffer([]byte{})
			if err := png.Encode(buf, tim3.PictureToImage(tim.Pictures[0])); err != nil {
				return err
			}

			img := rl.LoadImageFromMemory(".png", buf.Bytes(), int32(buf.Len()))
			texture := rl.LoadTextureFromImage(img)
			rl.UnloadImage(img)

			entries = append(
				entries,
				&Entry{
					Source:  filePath,
					Name:    fmt.Sprintf("%s_%03d", utils.FilterUnprintableString(entry.Name), i),
					Png:     buf.Bytes(),
					Texture: texture,
					Picture: tim.Pictures[0],
				},
			)
		}

		entry := entries[0]
		currentEntry = 0

		matrix = rl.MatrixTranslate(
			(width/2)-(float32(entry.Texture.Width)/2),
			(height/2)-(float32(entry.Texture.Height)/2),
			0,
		)

		mode = ModeMultiple
		canConvert = true
	default:
		return fmt.Errorf("Format not supported")
	}

	return nil
}

func zoom(wheel float32) {
	isInsidePreview := rl.CheckCollisionPointRec(rl.GetMousePosition(), zoomDeadZone)

	if mode == ModeMultiple && isInsidePreview {
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

func convert2png() {
	if currentEntry == -1 {
		return
	}
	entry := entries[currentEntry]

	pngFile, err := os.OpenFile(fmt.Sprintf("%s/%s.png", utils.ParentDirectory(timPath), entry.Name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer pngFile.Close()

	if _, err := pngFile.Write(entry.Png); err != nil {
		log.Println(err)
		return
	}
	log.Println("Converted")
}

func main() {
	rl.InitWindow(int32(width), int32(height), "TIM Viewer")
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	rlig.Load()
	defer rlig.Unload()
	imgui.StyleColorsDark()

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
				timPath = filePath
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
		imgui.Checkbox("GS Info", &showGsInfo)
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

		if currentEntry != -1 {
			imgui.SetNextWindowPosV(imgui.NewVec2(12, 12), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
			imgui.BeginV("Information", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
			entry := entries[currentEntry]
			imgui.Text(
				fmt.Sprintf(
					"%dx%d\n%s\n%s\nClut Colors %d",
					entry.Picture.ImageWidth,
					entry.Picture.ImageHeight,
					entry.Picture.ClutType,
					entry.Picture.ImageType,
					entry.Picture.ClutColors,
				),
			)
			imgui.End()
		}

		if mode == ModeMultiple {
			imgui.SetNextWindowPosV(imgui.NewVec2(width-12, 104), imgui.CondAlways, imgui.NewVec2(1, 0))
			imgui.SetNextWindowSizeConstraints(imgui.NewVec2(84, 108), imgui.NewVec2(84, 216))
			imgui.BeginV("Entries", nil, imgui.WindowFlagsNoMove)
			for i, entry := range entries {
				if imgui.ImageButton(entry.Name, imgui.TextureID(entry.Texture.ID), imgui.NewVec2(42, 42)) {
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
		}

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, height-12), imgui.CondAlways, imgui.NewVec2(1, 1))
		imgui.BeginV("ToPng", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.BeginDisabledV(!canConvert)
		if imgui.Button("Convert To PNG") {
			go func() {
				convert2png()
			}()
		}
		imgui.EndDisabled()
		imgui.End()

		if currentEntry != -1 && showGsInfo {
			imgui.SetNextWindowPosV(imgui.NewVec2(12, height-12), imgui.CondAlways, imgui.NewVec2(0, 1))
			imgui.BeginV("Graphic Synthesizer", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
			entry := entries[currentEntry]

			CLD := (entry.Picture.GsTex0 >> 61) & 0x7
			CSA := (entry.Picture.GsTex0 >> 56) & 0x1F
			CSM := (entry.Picture.GsTex0 >> 55) & 0x1
			CPSM := (entry.Picture.GsTex0 >> 51) & 0xF
			CBP := (entry.Picture.GsTex0 >> 37) & 0x3FFF
			TFX := (entry.Picture.GsTex0 >> 35) & 0x3
			TCC := (entry.Picture.GsTex0 >> 34) & 0x1
			TH := (entry.Picture.GsTex0 >> 30) & 0xF
			TW := (entry.Picture.GsTex0 >> 26) & 0xF
			PSM := (entry.Picture.GsTex0 >> 20) & 0x3F
			TBW := (entry.Picture.GsTex0 >> 14) & 0x3F
			TBP0 := entry.Picture.GsTex0 & 0x3FFF

			imgui.Text(
				fmt.Sprintf(
					"CLUT Buffer Load Control: %d\nCLUT Entry Offset: %d\nCLUT Storage Mode: %d\nCLUT Pixel Storage Format: %d\nCLUT Buffer Base Point: %d\nTexture Function: %d\nTexture Color Component: %d\nTexture Height: %d\nTexture Width: %d\nTexture Pixel Storage Format: %d\nTexture Buffer Width: %d\nTexture Base Point: %d",
					CLD,
					CSA,
					CSM,
					CPSM,
					CBP,
					TFX,
					TCC,
					TH,
					TW,
					PSM,
					TBW,
					TBP0,
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

			for i, c := range entry.Picture.ClutData {
				rl.DrawRectangle(int32(i%8*8)+16, int32(i/8*8)+90, 8, 8, *c)
			}
		}

		rlig.Render()
		rl.EndDrawing()
	}

	for _, entry := range entries {
		rl.UnloadTexture(entry.Texture)
	}
}
