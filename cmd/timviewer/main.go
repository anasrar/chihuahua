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
	case tim2.Signature:
		tim := tim2.New()
		if err := tim2.FromPath(tim, filePath); err != nil {
			return err
		}

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

		position = rl.NewVector2((width/2)-(float32(textures[0].Width)/2), (height/2)-(float32(textures[0].Height)/2))
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
			position = rl.Vector2Add(rl.GetMouseDelta(), position)
		}

		wheel := rl.GetMouseWheelMoveV().Y * 0.25
		if wheel != 0 {
			scale += wheel
			scale = max(scale, 0.2)
		}

		rl.BeginDrawing()
		rl.ClearBackground(background)

		for _, tex := range textures {
			rl.DrawRectangleLinesEx(rl.NewRectangle(position.X, position.Y, float32(tex.Width)*scale, float32(tex.Height)*scale), 1, rl.Gray)
			rl.DrawTextureEx(tex, position, 0, scale, rl.White)
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
				position = rl.NewVector2((width/2)-(float32(tex.Width)/2), (height/2)-(float32(tex.Height)/2))
			}
			scale = 1
		}

		rl.EndDrawing()

	}
	rayguistyle.Unload()

	for _, tex := range textures {
		rl.UnloadTexture(tex)
	}

	rl.CloseWindow()
}
