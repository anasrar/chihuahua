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
	"github.com/anasrar/chihuahua/pkg/utils"
	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func drop(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

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

		picture = tim.Pictures[0]

		buf = bytes.NewBuffer([]byte{})
		if err := png.Encode(buf, tim3.PictureToImage(tim.Pictures[0])); err != nil {
			return err
		}

		img := rl.LoadImageFromMemory(".png", buf.Bytes(), int32(buf.Len()))
		for _, tex := range textures {
			rl.UnloadTexture(tex)
		}
		textures = []rl.Texture2D{rl.LoadTextureFromImage(img)}
		rl.UnloadImage(img)

		matrix = rl.MatrixTranslate(
			(width/2)-(float32(textures[0].Width)/2),
			(height/2)-(float32(textures[0].Height)/2),
			0,
		)

		canConvert = true
	case tim2.Signature:
		tim := tim2.New()
		if err := tim2.FromPath(tim, filePath); err != nil {
			return err
		}

		picture = tim.Pictures[0]

		buf = bytes.NewBuffer([]byte{})
		if err := png.Encode(buf, tim2.PictureToImage(tim.Pictures[0])); err != nil {
			return err
		}

		img := rl.LoadImageFromMemory(".png", buf.Bytes(), int32(buf.Len()))
		for _, tex := range textures {
			rl.UnloadTexture(tex)
		}
		textures = []rl.Texture2D{rl.LoadTextureFromImage(img)}
		rl.UnloadImage(img)

		matrix = rl.MatrixTranslate(
			(width/2)-(float32(textures[0].Width)/2),
			(height/2)-(float32(textures[0].Height)/2),
			0,
		)

		canConvert = true
	default:
		return fmt.Errorf("Format not supported")
	}

	return nil
}

func main() {
	rl.InitWindow(int32(width), int32(height), "TIM Viewer")
	rl.SetTargetFPS(30)

	rayguistyle.Load()

	for !rl.WindowShouldClose() {
		if rl.IsWindowResized() {
			width = float32(rl.GetScreenWidth())
			height = float32(rl.GetScreenHeight())
		}

		if rl.IsFileDropped() {
			filePath := rl.LoadDroppedFiles()[0]

			if err := drop(filePath); err != nil {
				log.Println(err)
			} else {
				timPath = filePath
			}

			rl.UnloadDroppedFiles()
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

		rl.BeginDrawing()
		rl.ClearBackground(background)

		rl.BeginMode2D(camera)

		for _, tex := range textures {
			translate := rl.NewVector3(0, 0, 0)
			rotation := rl.NewQuaternion(0, 0, 0, 1)
			scale := rl.NewVector3(1, 1, 1)

			utils.MatrixDecompose(matrix, &translate, &rotation, &scale)

			rl.DrawRectangleLinesEx(rl.NewRectangle(translate.X, translate.Y, float32(tex.Width)*scale.X, float32(tex.Height)*scale.Y), 1, rl.Gray)

			rl.PushMatrix()
			rl.Translatef(translate.X, translate.Y, 0)
			rl.Scalef(scale.X, scale.Y, 1)

			rl.DrawTexture(tex, 0, 0, rl.White)

			rl.PopMatrix()
		}

		rl.EndMode2D()

		if picture != nil {
			rl.DrawTextEx(
				rayguistyle.DefaultFont,
				fmt.Sprintf(
					"%dx%d\n%s\n%s\nClut Colors %d",
					picture.ImageWidth,
					picture.ImageHeight,
					picture.ClutType,
					picture.ImageType,
					picture.ClutColors,
				),
				rl.NewVector2(8, 8),
				16,
				0,
				rl.White,
			)

			for i, c := range picture.ClutData {
				rl.DrawRectangle(int32(i%8*8)+8, int32(i/8*8)+78, 8, 8, *c)
			}
		}

		background = raygui.ColorPicker(rl.NewRectangle(width-74, 8, 42, 42), "", background)

		if !canConvert {
			raygui.Disable()
		}

		if raygui.Button(rl.NewRectangle(8, height-40, 132, 32), "Convert To PNG") {
			for range textures {
				packFile, err := os.OpenFile(fmt.Sprintf("%s/%s.png", utils.ParentDirectory(timPath), utils.BasenameWithoutExt(timPath)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					log.Println(err)
					continue
				}
				defer packFile.Close()

				if _, err := packFile.Write(buf.Bytes()); err != nil {
					log.Println(err)
					continue
				}

				log.Println("Exported")
			}
		}

		if !canConvert {
			raygui.Enable()
		}

		if raygui.Button(rl.NewRectangle(width-116, height-40, 108, 32), "Reset View") {
			for _, tex := range textures {
				matrix = rl.MatrixTranslate(
					(width/2)-(float32(tex.Width)/2),
					(height/2)-(float32(tex.Height)/2),
					0,
				)
			}
		}

		rl.EndDrawing()

	}
	rayguistyle.Unload()

	for _, tex := range textures {
		rl.UnloadTexture(tex)
	}

	rl.CloseWindow()
}
