package main

import (
	"fmt"
	"log"

	rayguistyle "github.com/anasrar/chihuahua/internal/raygui_style"
	"github.com/anasrar/chihuahua/pkg/dat"
	"github.com/anasrar/chihuahua/pkg/ems"
	"github.com/anasrar/chihuahua/pkg/oms"
	"github.com/anasrar/chihuahua/pkg/utils"
	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	CameraSpeed float32 = 20
)

func drop(filePath string) error {
	dat0 := dat.New()
	if err := dat.FromPath(dat0, filePath); err != nil {
		return err
	}

	var datScp *dat.Entry
	datOms := []*dat.Entry{}
	datEms := []*dat.Entry{}

	for _, entry := range dat0.Entries {
		t := utils.FilterUnprintableString(entry.Type)

		switch t {
		case "SCP":
			datScp = entry
		case "OMS":
			datOms = append(datOms, entry)
		case "EMS":
			datEms = append(datEms, entry)
		}
	}

	if datScp == nil {
		return fmt.Errorf("SCP not found")
	}

	scp = datScp

	dat1 := dat.New()
	if err := dat.FromPathWithOffsetSize(dat1, filePath, datScp.Offset, datScp.Size); err != nil {
		return err
	}

	for _, model := range models {
		rl.UnloadMesh(model.Model.Meshes)
	}

	models = []*Model{}

	for _, entry := range dat1.Entries {
		t := utils.FilterUnprintableString(entry.Type)
		switch t {
		case "TM3":
			if err := LoadTextures(filePath, entry.Offset, entry.Size); err != nil {
				return err
			}
		case "SCR":
			if err := LoadModels(filePath, entry.Offset); err != nil {
				return err
			}
		default:
		}
	}

	omsEntries = []*Object{}

	for _, entry := range datOms {
		om := oms.New()
		if err := oms.FromPathWithOffset(om, filePath, entry.Offset); err != nil {
			return err
		}

		for _, omEntry := range om.Entries {
			omsEntries = append(omsEntries, &Object{
				RenderLabel: false,
				Entry:       omEntry,
			})
		}

	}

	emsEntries = []*ems.Entry{}

	for _, entry := range datEms {
		em := ems.New()
		if err := ems.FromPathWithOffset(em, filePath, entry.Offset); err != nil {
			return err
		}

		emsEntries = append(emsEntries, em.Entries...)
	}

	datPath = filePath
	modelContentRectangle.Height = float32(24*len(models)) + 4

	return nil
}

func main() {
	rl.InitWindow(int32(width), int32(height), "Room Viewer")
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	rayguistyle.Load()
	defer rayguistyle.Unload()

	checked := rl.GenImageChecked(20, 20, 1, 1, rl.White, rl.Gray)
	textureDefault = rl.LoadTextureFromImage(checked)
	rl.UnloadImage(checked)
	defer rl.UnloadTexture(textureDefault)

	rl.EnableDepthTest()
	rl.EnableColorBlend()
	rl.EnableDepthMask()

	for !rl.WindowShouldClose() {
		if rl.IsWindowResized() {
			width = float32(rl.GetScreenWidth())
			height = float32(rl.GetScreenHeight())

			tm3PreviewRectangle = rl.NewRectangle(width-74, 58, 64, height-108)
		}

		if rl.IsFileDropped() {
			filePath := rl.LoadDroppedFiles()[0]
			defer rl.UnloadDroppedFiles()

			if err := drop(filePath); err != nil {
				log.Println(err)
			}
		}

		if rl.IsKeyDown(rl.KeyW) {
			rl.CameraMoveForward(&camera, CameraSpeed*rl.GetFrameTime(), 0)
		}
		if rl.IsKeyDown(rl.KeyS) {
			rl.CameraMoveForward(&camera, -CameraSpeed*rl.GetFrameTime(), 0)
		}

		if rl.IsKeyDown(rl.KeyA) {
			rl.CameraMoveRight(&camera, -CameraSpeed*rl.GetFrameTime(), 0)
		}
		if rl.IsKeyDown(rl.KeyD) {
			rl.CameraMoveRight(&camera, CameraSpeed*rl.GetFrameTime(), 0)
		}

		if rl.IsKeyDown(rl.KeyQ) {
			rl.CameraMoveUp(&camera, -CameraSpeed*rl.GetFrameTime())
		}
		if rl.IsKeyDown(rl.KeyE) {
			rl.CameraMoveUp(&camera, CameraSpeed*rl.GetFrameTime())
		}

		if rl.IsKeyDown(rl.KeyLeft) {
			rl.CameraYaw(&camera, 1*rl.GetFrameTime(), 0)
		}
		if rl.IsKeyDown(rl.KeyRight) {
			rl.CameraYaw(&camera, -1*rl.GetFrameTime(), 0)
		}

		if rl.IsKeyDown(rl.KeyUp) {
			rl.CameraPitch(&camera, 0.5*rl.GetFrameTime(), 0, 0, 0)
		}
		if rl.IsKeyDown(rl.KeyDown) {
			rl.CameraPitch(&camera, -0.5*rl.GetFrameTime(), 0, 0, 0)
		}

		rl.BeginDrawing()
		rl.ClearBackground(background)

		rl.BeginMode3D(camera)

		for _, model := range models {
			if model.Render {
				rl.PushMatrix()

				rl.Translatef(model.Translation.X, model.Translation.Y, model.Translation.Z)
				rl.Rotatef(model.Rotation.X, 1, 0, 0)
				rl.Rotatef(model.Rotation.Y, 0, 1, 0)
				rl.Rotatef(model.Rotation.Z, 0, 0, 1)
				rl.Scalef(model.Scale.X, model.Scale.Y, model.Scale.Z)

				rl.DrawModel(*model.Model, rl.NewVector3(0, 0, 0), 1, rl.White)

				rl.PopMatrix()
			}
		}

		if showOms {
			ray := rl.GetMouseRay(rl.GetMousePosition(), camera)
			for _, obj := range omsEntries {
				obj.RenderLabel = rl.GetRayCollisionBox(
					ray,
					rl.NewBoundingBox(
						rl.NewVector3(obj.Translation[0]-0.2, obj.Translation[1], obj.Translation[2]-0.2),
						rl.NewVector3(obj.Translation[0]+0.2, obj.Translation[1]+0.4, obj.Translation[2]+0.2),
					),
				).Hit

				rl.DrawCube(rl.NewVector3(obj.Translation[0], obj.Translation[1]+0.2, obj.Translation[2]), 0.4, 0.4, 0.4, rl.Blue)
			}
		}

		if showEms {
			for _, enemy := range emsEntries {
				rl.DrawCube(rl.NewVector3(enemy.Translation[0], enemy.Translation[1]+0.2, enemy.Translation[2]), 0.4, 0.4, 0.4, rl.Red)
			}
		}

		rl.DrawGrid(4, 0.5)

		rl.EndMode3D()

		if showOms {
			for _, obj := range omsEntries {
				if obj.RenderLabel {
					screenPosition := rl.GetWorldToScreen(rl.NewVector3(obj.Translation[0], obj.Translation[1]+.4, obj.Translation[2]), camera)
					raygui.Label(rl.NewRectangle(screenPosition.X, screenPosition.Y, 120, 18), obj.Name)
				}
			}
		}

		raygui.ScrollPanel(
			modelRectangle,
			"",
			modelContentRectangle,
			&modelScroll,
			&modelView,
		)

		// rl.DrawRectangle(
		// 	int32(modelRectangle.X+modelScroll.X),
		// 	int32(modelRectangle.Y+modelScroll.Y),
		// 	int32(modelContentRectangle.Width),
		// 	int32(modelContentRectangle.Height),
		// 	rl.Fade(rl.Red, 0.1),
		// )

		rl.BeginScissorMode(
			int32(modelView.X),
			int32(modelView.Y),
			int32(modelView.Width),
			int32(modelView.Height),
		)

		{
			y := modelRectangle.Y + modelScroll.Y
			for i, model := range models {
				rect := rl.NewRectangle(12, (24*float32(i))+4+y, 24, 24)
				inside := rl.CheckCollisionRecs(rect, modelRectangle)

				if inside {
					check := rl.NewRectangle(12, (24*float32(i))+4+y, 14, 14)
					r := rl.GetCollisionRec(check, modelRectangle)
					model.Render = raygui.CheckBox(r, model.Name, model.Render)
				}
			}
		}

		rl.EndScissorMode()

		showOms = raygui.CheckBox(rl.NewRectangle(8, 218, 14, 14), "Show OMS", showOms)
		showEms = raygui.CheckBox(rl.NewRectangle(8, 240, 14, 14), "Show EMS", showEms)

		if scp == nil {
			raygui.Disable()
		}
		if raygui.Button(rl.NewRectangle(8, height-40, 132, 32), "Convert To GLTF") {
			go func() {
				log.Println("Convert to GLTF")
				if err := ConvertToGlft(); err != nil {
					log.Println(err)
				} else {
					log.Println("Convert done")
				}
			}()
		}
		if scp == nil {
			raygui.Enable()
		}

		background = raygui.ColorPicker(rl.NewRectangle(width-74, 8, 42, 42), "", background)

		raygui.ScrollPanel(
			tm3PreviewRectangle,
			"",
			tm3PreviewContentRectangle,
			&tm3PreviewScroll,
			&tm3PreviewView,
		)

		// rl.DrawRectangle(
		// 	int32(previewRectangle.X+previewScroll.X),
		// 	int32(previewRectangle.Y+previewScroll.Y),
		// 	int32(previewContentRectangle.Width),
		// 	int32(previewContentRectangle.Height),
		// 	rl.Fade(rl.Red, 0.1),
		// )

		rl.BeginScissorMode(
			int32(tm3PreviewView.X),
			int32(tm3PreviewView.Y),
			int32(tm3PreviewView.Width),
			int32(tm3PreviewView.Height),
		)

		{
			y := tm3PreviewRectangle.Y + tm3PreviewScroll.Y

			for i, index := range textureIndices {
				rect := rl.NewRectangle(width-73, float32(i)*42+y+1, 42, 42)
				isInside := rl.CheckCollisionRecs(tm3PreviewRectangle, rect)
				if !isInside {
					continue
				}

				texture := textures[index]

				rl.DrawTexturePro(
					texture.Texture,
					rl.NewRectangle(0, 0, float32(texture.Texture.Width), float32(texture.Texture.Height)),
					rl.NewRectangle(width-73, float32(i)*42+y+1, 42, 42),
					rl.Vector2Zero(),
					0,
					rl.White,
				)

			}
		}

		rl.EndScissorMode()

		if raygui.Button(rl.NewRectangle(width-116, height-40, 108, 32), "Reset View") {
			camera.Position = rl.NewVector3(0, 2.8, 2.8)
			camera.Target = rl.NewVector3(0, 1.2, 0)
		}

		rl.EndDrawing()

	}

	for _, model := range models {
		rl.UnloadMesh(model.Model.Meshes)
	}

	for _, index := range textureIndices {
		rl.UnloadTexture(textures[index].Texture)
		delete(textures, index)
	}
}
