package main

import (
	"fmt"
	"log"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/anasrar/chihuahua/pkg/dat"
	"github.com/anasrar/chihuahua/pkg/ems"
	"github.com/anasrar/chihuahua/pkg/oms"
	rlig "github.com/anasrar/chihuahua/pkg/raylib_imgui"
	"github.com/anasrar/chihuahua/pkg/utils"
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

	return nil
}

func main() {
	rl.InitWindow(int32(width), int32(height), "Room Viewer")
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	rlig.Load()
	defer rlig.Unload()
	imgui.StyleColorsDark()

	checked := rl.GenImageChecked(20, 20, 1, 1, rl.White, rl.Gray)
	textureDefault = rl.LoadTextureFromImage(checked)
	rl.UnloadImage(checked)
	defer rl.UnloadTexture(textureDefault)

	rl.EnableDepthTest()
	rl.EnableColorBlend()
	rl.EnableDepthMask()

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

		imgui.NewFrame()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, 12), imgui.CondAlways, imgui.NewVec2(1, 0))
		imgui.BeginV("View", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.ColorEdit3V("Background", &(background), imgui.ColorEditFlagsNoInputs)
		if imgui.Button("Reset View") {
			camera.Position = rl.NewVector3(0, 2.8, 2.8)
			camera.Target = rl.NewVector3(0, 1.2, 0)
		}
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, 82), imgui.CondAlways, imgui.NewVec2(1, 0))
		imgui.SetNextWindowSizeConstraints(imgui.NewVec2(72, 108), imgui.NewVec2(72, 216))
		imgui.BeginV("TM3", nil, imgui.WindowFlagsNoMove)
		for _, index := range textureIndices {
			texture := textures[index]

			imgui.Image(imgui.TextureID(texture.Texture.ID), imgui.NewVec2(42, 42))
		}
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(12, height-12), imgui.CondAlways, imgui.NewVec2(0, 1))
		imgui.BeginV("ToGltf", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.BeginDisabledV(scp == nil)
		if imgui.Button("Convert To GLTF") {
			go func() {
				log.Println("Convert to GLTF")
				if err := ConvertToGlft(); err != nil {
					log.Println(err)
				} else {
					log.Println("Convert done")
				}
			}()
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

		imgui.SetNextWindowPosV(imgui.NewVec2(12, 12), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
		imgui.SetNextWindowSizeV(imgui.NewVec2(200, 300), imgui.CondFirstUseEver)
		imgui.BeginV("Inspector", nil, imgui.WindowFlagsNone)
		imgui.Checkbox("Show OMS", &showOms)
		imgui.Checkbox("Show EMS", &showEms)
		imgui.Separator()
		imgui.BeginChildStrV("MdbRegion", imgui.NewVec2(0, 0), imgui.ChildFlagsNavFlattened, imgui.WindowFlagsHorizontalScrollbar)
		for _, model := range models {
			imgui.Checkbox(model.Name, &model.Render)
		}
		imgui.EndChild()
		imgui.End()

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
					position := imgui.MousePos()
					imgui.SetNextWindowPosV(imgui.NewVec2(position.X, position.Y), imgui.CondAlways, imgui.NewVec2(0, 1))
					imgui.BeginV("Information", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoFocusOnAppearing)
					imgui.Text(obj.Name)
					imgui.End()
				}
			}
		}

		rlig.Render()
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
