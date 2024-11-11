package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"os"

	rayguistyle "github.com/anasrar/chihuahua/internal/raygui_style"
	"github.com/anasrar/chihuahua/pkg/buffer"
	"github.com/anasrar/chihuahua/pkg/scr"
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

		modelContentRectangle.Height = float32(24*len(models)) + 4

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

		tm3PreviewContentRectangle.Height = float32(tm.EntryTotal*42) + 1

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

	rayguistyle.Load()
	defer rayguistyle.Unload()

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
		if rl.IsWindowResized() {
			width = float32(rl.GetScreenWidth())
			height = float32(rl.GetScreenHeight())

			tm3PreviewRectangle = rl.NewRectangle(width-74, 58, 64, height-108)

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

		rl.BeginDrawing()
		rl.ClearBackground(background)

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

		{
			if textureTotal == 0 {
				raygui.Disable()
			}

			if raygui.Button(rl.NewRectangle(8, 218, 87, 32), "Shift -1") {
				textureShift = (textureShift - 1) % textureTotal
				for _, model := range models {
					if texture, found := textures[model.Texture+textureShift]; found {
						rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, texture.Texture)
					} else {
						rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, textureDefault)
					}
				}
			}

			if raygui.Button(rl.NewRectangle(103, 218, 87, 32), "Shift +1") {
				textureShift = (textureShift + 1) % textureTotal
				for _, model := range models {
					if texture, found := textures[model.Texture+textureShift]; found {
						rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, texture.Texture)
					} else {
						rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, textureDefault)
					}
				}
			}

			if textureTotal == 0 {
				raygui.Enable()
			}
		}

		showBones = raygui.CheckBox(rl.NewRectangle(8, 258, 14, 14), "Show Bones", showBones)
		applyScrTransform = raygui.CheckBox(rl.NewRectangle(8, 282, 14, 14), "Apply SCR Transform", applyScrTransform)

		if scrPath == "" {
			raygui.Disable()
		}
		if raygui.Button(rl.NewRectangle(8, height-40, 132, 32), "Convert To GLTF") {
			go func() {
				log.Println("Convert to GLTF")
				if err := scr.ConvertToGlft(scrPath, tm3Path, textureShift); err != nil {
					log.Println(err)
				} else {
					log.Println("Convert done")
				}
			}()
		}
		if scrPath == "" {
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
