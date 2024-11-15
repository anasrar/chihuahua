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
	"github.com/anasrar/chihuahua/pkg/scr"
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
	case scr.Signature:
		var s scr.Scr
		if err := scr.FromPath(&s, filePath); err != nil {
			return err
		}

		for _, model := range models {
			rl.UnloadMesh(model.Model.Meshes)
		}

		models = []*Model{}

		for i, node := range s.Nodes {
			for j, vb := range node.Mdb.VertexBuffers {
				name := fmt.Sprintf("%s_%03d_%03d", utils.FilterUnprintableString(node.Name), i, j)

				var mesh rl.Mesh
				mesh.TriangleCount = int32(len(vb.Indices))
				mesh.VertexCount = int32(len(vb.Indices) * 3)

				vertices := []float32{}
				uvs := []float32{}

				for _, index := range vb.Indices {
					p1 := vb.Vertices[index[0]]
					p2 := vb.Vertices[index[1]]
					p3 := vb.Vertices[index[2]]

					vertices = append(vertices, p1[:]...)
					vertices = append(vertices, p2[:]...)
					vertices = append(vertices, p3[:]...)

					uv1 := vb.Uvs[index[0]]
					uv2 := vb.Uvs[index[1]]
					uv3 := vb.Uvs[index[2]]

					uvs = append(uvs, uv1[:]...)
					uvs = append(uvs, uv2[:]...)
					uvs = append(uvs, uv3[:]...)
				}

				mesh.Vertices = &vertices[0]
				mesh.Texcoords = &uvs[0]

				rl.UploadMesh(&mesh, false)

				model := rl.LoadModelFromMesh(mesh)

				if texture, found := textures[int(vb.Material)]; found {
					rl.SetMaterialTexture(model.Materials, rl.MapDiffuse, texture.Texture)
				} else {
					rl.SetMaterialTexture(model.Materials, rl.MapDiffuse, textureDefault)
				}

				models = append(
					models,
					NewModel(
						name,
						&model,
						int(vb.Material),
						rl.NewVector3(node.Translation[0], node.Translation[1], node.Translation[2]),
						rl.NewVector3(node.Rotation[0], node.Rotation[1], node.Rotation[2]),
						rl.NewVector3(node.Scale[0], node.Scale[1], node.Scale[2]),
					),
				)
			}
		}

		boneTree = NewBoneNode(0, 0, 0)
		bones = []*BoneNode{
			boneTree,
		}

		for _, bone := range s.Nodes[0].Mdb.Bones {
			node := NewBoneNode(bone.Translation[0], bone.Translation[1], bone.Translation[2])
			bones = append(bones, node)

			bones[bone.Parent].Children = append(
				bones[bone.Parent].Children,
				node,
			)
		}

		scrPath = filePath
		textureShift = 0
	case tm3.Signature:
		tm := tm3.New()
		if err := tm3.FromPath(tm, filePath); err != nil {
			return err
		}

		for _, index := range textureIndices {
			rl.UnloadTexture(textures[index].Texture)
			delete(textures, index)
		}

		textureIndices = []int{}
		textureTotal = int(tm.EntryTotal)

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
			defer rl.UnloadImage(img)

			texture := rl.LoadTextureFromImage(img)
			textures[i] = NewTexture(texture)

			textureIndices = append(textureIndices, i)
		}

		for _, model := range models {
			if texture, found := textures[model.Texture]; found {
				rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, texture.Texture)
			} else {
				rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, textureDefault)
			}
		}

		tm3Path = filePath
		textureShift = 0
	default:
		return fmt.Errorf("Format not supported")
	}

	return nil
}

func main() {
	rl.InitWindow(int32(width), int32(height), "SCR Viewer")
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

	boneRender = rl.LoadRenderTexture(int32(width), int32(height))
	defer rl.UnloadRenderTexture(boneRender)

	for !rl.WindowShouldClose() {
		rlig.Update()

		if rl.IsWindowResized() {
			width = float32(rl.GetScreenWidth())
			height = float32(rl.GetScreenHeight())

			rl.UnloadRenderTexture(boneRender)
			boneRender = rl.LoadRenderTexture(int32(width), int32(height))
		}

		if rl.IsFileDropped() {
			filePath := rl.LoadDroppedFiles()[0]
			defer rl.UnloadDroppedFiles()

			if err := drop(filePath); err != nil {
				log.Println(err)
			}
		}

		if rl.IsKeyDown(rl.KeyW) {
			rl.CameraMoveForward(&camera, 1*rl.GetFrameTime(), 0)
		}
		if rl.IsKeyDown(rl.KeyS) {
			rl.CameraMoveForward(&camera, -1*rl.GetFrameTime(), 0)
		}

		if rl.IsKeyDown(rl.KeyA) {
			rl.CameraMoveRight(&camera, -1*rl.GetFrameTime(), 0)
		}
		if rl.IsKeyDown(rl.KeyD) {
			rl.CameraMoveRight(&camera, 1*rl.GetFrameTime(), 0)
		}

		if rl.IsKeyDown(rl.KeyQ) {
			rl.CameraMoveUp(&camera, -1*rl.GetFrameTime())
		}
		if rl.IsKeyDown(rl.KeyE) {
			rl.CameraMoveUp(&camera, 1*rl.GetFrameTime())
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
		imgui.BeginDisabledV(scrPath == "")
		if imgui.Button("Convert To GLTF") {
			go func() {
				log.Println("Convert to GLTF")
				if err := scr.ConvertToGlft(scrPath, tm3Path, textureShift); err != nil {
					log.Println(err)
				} else {
					log.Println("Convert done")
				}
			}()
		}
		imgui.EndDisabled()
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(12, 12), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
		imgui.SetNextWindowSizeV(imgui.NewVec2(200, 300), imgui.CondFirstUseEver)
		imgui.BeginV("Inspector", nil, imgui.WindowFlagsNone)
		imgui.BeginDisabledV(textureTotal == 0)
		if imgui.Button("Shift -1") {
			textureShift = (textureShift - 1) % textureTotal
			for _, model := range models {
				if texture, found := textures[model.Texture+textureShift]; found {
					rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, texture.Texture)
				} else {
					rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, textureDefault)
				}
			}
		}
		imgui.SameLineV(0, 4)
		if imgui.Button("Shift +1") {
			textureShift = (textureShift + 1) % textureTotal
			for _, model := range models {
				if texture, found := textures[model.Texture+textureShift]; found {
					rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, texture.Texture)
				} else {
					rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, textureDefault)
				}
			}
		}
		imgui.EndDisabled()
		imgui.Checkbox("Show Bones", &showBones)
		imgui.Checkbox("Apply SCR Transform", &applyScrTransform)
		imgui.Separator()
		imgui.BeginChildStrV("MdbRegion", imgui.NewVec2(0, 0), imgui.ChildFlagsNavFlattened, imgui.WindowFlagsHorizontalScrollbar)
		for _, model := range models {
			imgui.Checkbox(model.Name, &model.Render)
		}
		imgui.EndChild()
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

		rl.BeginMode3D(camera)

		for _, model := range models {
			if model.Render {
				rl.PushMatrix()

				if applyScrTransform {
					rl.Translatef(model.Translation.X, model.Translation.Y, model.Translation.Z)
					rl.Rotatef(model.Rotation.X, 1, 0, 0)
					rl.Rotatef(model.Rotation.Y, 0, 1, 0)
					rl.Rotatef(model.Rotation.Z, 0, 0, 1)
					rl.Scalef(model.Scale.X, model.Scale.Y, model.Scale.Z)
				}

				rl.DrawModel(*model.Model, rl.NewVector3(0, 0, 0), 1, rl.White)

				rl.PopMatrix()
			}
		}
		rl.DrawGrid(4, 0.5)

		rl.EndMode3D()

		// TODO: refactor using raylib model bones, https://www.raylib.com/examples/models/loader.html?name=models_loading_m3d
		if showBones {
			rl.BeginTextureMode(boneRender)
			rl.ClearBackground(rl.NewColor(0, 0, 0, 0))
			rl.BeginMode3D(camera)
			DrawBoneTree(boneTree)
			rl.EndMode3D()
			rl.EndTextureMode()

			rl.DrawTextureRec(boneRender.Texture, rl.NewRectangle(0, 0, width, -height), rl.Vector2Zero(), rl.White)
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
