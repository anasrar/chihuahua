package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"os"

	rayguistyle "github.com/anasrar/chihuahua/internal/raygui_style"
	"github.com/anasrar/chihuahua/pkg/buffer"
	"github.com/anasrar/chihuahua/pkg/tim2"
	"github.com/anasrar/chihuahua/pkg/tim3"
	"github.com/anasrar/chihuahua/pkg/tm3"
	"github.com/anasrar/chihuahua/pkg/utils"
	"github.com/gen2brain/raylib-go/raygui"
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
			if err := tim3.FromPathWithOffsetSize(tim, filePath, entry.Offset); err != nil {
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

		previewContentRectangle.Height = float32(tm.EntryTotal*42) + 1

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
	isInsidePreview := rl.CheckCollisionPointRec(rl.GetMousePosition(), previewRectangle)

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

	rayguistyle.Load()
	defer rayguistyle.Unload()

	for !rl.WindowShouldClose() {
		if rl.IsWindowResized() {
			width = float32(rl.GetScreenWidth())
			height = float32(rl.GetScreenHeight())

			previewRectangle = rl.NewRectangle(width-74, 58, 64, height-108)
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

		rl.BeginDrawing()
		rl.ClearBackground(background)

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

			rl.DrawTextEx(
				rayguistyle.DefaultFont,
				fmt.Sprintf(
					"%dx%d\n%s\n%s\nClut Colors %d",
					entry.Picture.ImageWidth,
					entry.Picture.ImageHeight,
					entry.Picture.ClutType,
					entry.Picture.ImageType,
					entry.Picture.ClutColors,
				),
				rl.NewVector2(8, 8),
				16,
				0,
				rl.White,
			)

			for i, c := range entry.Picture.ClutData {
				rl.DrawRectangle(int32(i%8*8)+8, int32(i/8*8)+78, 8, 8, *c)
			}
		}

		background = raygui.ColorPicker(rl.NewRectangle(width-74, 8, 42, 42), "", background)

		if mode == ModeMultiple {
			raygui.ScrollPanel(
				previewRectangle,
				"",
				previewContentRectangle,
				&previewScroll,
				&previewView,
			)

			// rl.DrawRectangle(
			// 	int32(previewRectangle.X+previewScroll.X),
			// 	int32(previewRectangle.Y+previewScroll.Y),
			// 	int32(previewContentRectangle.Width),
			// 	int32(previewContentRectangle.Height),
			// 	rl.Fade(rl.Red, 0.1),
			// )

			rl.BeginScissorMode(
				int32(previewView.X),
				int32(previewView.Y),
				int32(previewView.Width),
				int32(previewView.Height),
			)

			{
				y := previewRectangle.Y + previewScroll.Y

				for i, entry := range entries {
					rect := rl.NewRectangle(width-73, float32(i)*42+y+1, 42, 42)
					isInside := rl.CheckCollisionRecs(previewRectangle, rect)
					if !isInside {
						continue
					}

					bound := rl.GetCollisionRec(previewRectangle, rect)
					tint := rl.Gray

					isHover := rl.CheckCollisionPointRec(rl.GetMousePosition(), bound)
					if isHover {
						tint = rl.White

						if rl.IsMouseButtonPressed(0) {
							currentEntry = i
							matrix = rl.MatrixTranslate(
								(width/2)-(float32(entry.Texture.Width)/2),
								(height/2)-(float32(entry.Texture.Height)/2),
								0,
							)
						}
					}

					rl.DrawTexturePro(
						entry.Texture,
						rl.NewRectangle(0, 0, float32(entry.Texture.Width), float32(entry.Texture.Height)),
						rl.NewRectangle(width-73, float32(i)*42+y+1, 42, 42),
						rl.Vector2Zero(),
						0,
						tint,
					)

				}
			}

			rl.EndScissorMode()
		}

		if !canConvert {
			raygui.Disable()
		}

		if raygui.Button(rl.NewRectangle(8, height-40, 132, 32), "Convert To PNG") {
			go func() {
				convert2png()
			}()
		}

		if !canConvert {
			raygui.Enable()
		}

		if raygui.Button(rl.NewRectangle(width-116, height-40, 108, 32), "Reset View") {
			if currentEntry != -1 {
				entry := entries[currentEntry]
				matrix = rl.MatrixTranslate(
					(width/2)-(float32(entry.Texture.Width)/2),
					(height/2)-(float32(entry.Texture.Height)/2),
					0,
				)
			}
		}

		rl.EndDrawing()

	}
	for _, entry := range entries {
		rl.UnloadTexture(entry.Texture)
	}
}
