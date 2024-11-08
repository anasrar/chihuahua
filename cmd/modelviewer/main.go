package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"

	rayguistyle "github.com/anasrar/chihuahua/internal/raygui_style"
	"github.com/anasrar/chihuahua/pkg/dat"
	"github.com/anasrar/chihuahua/pkg/scr"
	"github.com/anasrar/chihuahua/pkg/tim3"
	"github.com/anasrar/chihuahua/pkg/tm3"
	"github.com/anasrar/chihuahua/pkg/utils"
	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func loadModel(index int) error {
	tm3Entry := tm3Entries[index]
	mdEntry := mdEntries[index]

	{
		tm := tm3.New()
		if err := tm3.FromPathWithOffsetSize(tm, tm3Entry.Source, tm3Entry.Offset, tm3Entry.Size); err != nil {
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
			if err := tim3.FromPathWithOffset(tim, tm3Entry.Source, entry.Offset); err != nil {
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
		textureShift = 0
	}

	{
		s := scr.New()
		if err := scr.FromPathWithOffset(s, mdEntry.Source, mdEntry.Offset); err != nil {
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
				joints := []int32{}
				weights := []float32{}

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

					for _, joint := range vb.Joints[index[0]] {
						joints = append(joints, int32(joint))
					}
					for _, joint := range vb.Joints[index[1]] {
						joints = append(joints, int32(joint))
					}
					for _, joint := range vb.Joints[index[2]] {
						joints = append(joints, int32(joint))
					}

					weight1 := vb.Weights[index[0]]
					weight2 := vb.Weights[index[1]]
					weight3 := vb.Weights[index[2]]

					weights = append(weights, weight1[:]...)
					weights = append(weights, weight2[:]...)
					weights = append(weights, weight3[:]...)
				}

				mesh.Vertices = &vertices[0]
				mesh.Texcoords = &uvs[0]
				mesh.BoneIds = &joints[0]
				mesh.BoneWeights = &weights[0]

				rl.UploadMesh(&mesh, false)

				bones := []rl.BoneInfo{{
					Name:   [32]int8{0x72, 0x6f, 0x6f, 0x74},
					Parent: -1,
				}}
				pose := []rl.Transform{{
					Translation: rl.Vector3Zero(),
					Rotation:    rl.QuaternionIdentity(),
					Scale:       rl.Vector3One(),
				}}
				bonesWorldPosition := map[int16]rl.Vector3{
					0: rl.Vector3Zero(),
				}

				for i, bone := range node.Mdb.Bones {
					boneName := [32]int8{}
					for i, c := range []uint8(bone.Name) {
						boneName[i] = int8(c)
					}

					bones = append(
						bones,
						rl.BoneInfo{
							Name:   boneName,
							Parent: int32(bone.Parent),
						},
					)

					parentWorldPosition := bonesWorldPosition[bone.Parent]
					worldPosition := rl.Vector3Add(parentWorldPosition, rl.NewVector3(bone.X, bone.Y, bone.Z))
					bonesWorldPosition[int16(i+1)] = worldPosition

					pose = append(
						pose,
						rl.Transform{
							Translation: worldPosition,
							Rotation:    rl.QuaternionIdentity(),
							Scale:       rl.Vector3One(),
						},
					)
				}

				model := rl.LoadModelFromMesh(mesh)
				model.BoneCount = int32(len(bones))
				model.Bones = &bones[0]
				model.BindPose = &pose[0]

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
	}

	modelIndex = index
	motionIndex = -1

	return nil
}

func drop(filePath string) error {
	dat0 := dat.New()
	if err := dat.FromPath(dat0, filePath); err != nil {
		return err
	}

	tmpTm3Entries := []*Entry{}
	tmpMdEntries := []*Entry{}
	tmpMotEntries := []*Entry{}

	for i, entry := range dat0.Entries {
		t := utils.FilterUnprintableString(entry.Type)
		name := fmt.Sprintf("%s_%03d", t, i)
		switch t {
		case "MD":
			tmpMdEntries = append(tmpMdEntries, NewEntry(name, entry))
		case "TM3":
			tmpTm3Entries = append(tmpTm3Entries, NewEntry(name, entry))
		case "MOT":
			tmpMotEntries = append(tmpMotEntries, NewEntry(name, entry))
		default:
			continue
		}
	}

	if len(tmpMdEntries) == 0 {
		return fmt.Errorf("MD not found")
	}

	tm3Entries = tmpTm3Entries
	mdEntries = tmpMdEntries
	motEntries = tmpMotEntries

	mdContentRectangle.Height = float32(len(mdEntries))*24 + 8
	motContentRectangle.Height = float32(len(motEntries))*24 + 8

	if err := loadModel(0); err != nil {
		return err
	}

	datPath = filePath
	return nil
}

func loadMotion(index int) error {
	log.Println("TODO: implement load motion")
	motionIndex = index
	return nil
}

func main() {
	rl.InitWindow(int32(width), int32(height), "Model Viewer")
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

		hasModel := false
		for _, model := range models {
			hasModel = true
			if model.Render {
				rl.DrawModel(*model.Model, rl.NewVector3(0, 0, 0), 1, rl.White)
			}
		}
		rl.DrawGrid(4, 0.5)

		rl.EndMode3D()

		if hasModel && showBones {
			rl.BeginTextureMode(boneRender)
			rl.ClearBackground(rl.NewColor(0, 0, 0, 0))
			rl.BeginMode3D(camera)

			model := models[0]

			rl.DrawCube(rl.Vector3Zero(), .02, .02, .02, rl.Green)
			bones := model.Model.GetBones()
			pose := model.Model.GetBindPose()
			for j := int32(1); j < model.Model.BoneCount; j++ {
				rl.DrawCube(pose[j].Translation, .02, .02, .02, rl.Green)
				rl.DrawLine3D(pose[j].Translation, pose[bones[j].Parent].Translation, rl.Green)
			}

			rl.EndMode3D()
			rl.EndTextureMode()

			rl.DrawTextureRec(boneRender.Texture, rl.NewRectangle(0, 0, width, -height), rl.Vector2Zero(), rl.White)
		}

		raygui.ScrollPanel(
			mdRectangle,
			"",
			mdContentRectangle,
			&mdScroll,
			&mdView,
		)

		// rl.DrawRectangle(
		// 	int32(mdRectangle.X+mdScroll.X),
		// 	int32(mdRectangle.Y+mdScroll.Y),
		// 	int32(mdContentRectangle.Width),
		// 	int32(mdContentRectangle.Height),
		// 	rl.Fade(rl.Red, 0.1),
		// )

		rl.BeginScissorMode(
			int32(mdView.X),
			int32(mdView.Y),
			int32(mdView.Width),
			int32(mdView.Height),
		)

		{
			y := mdRectangle.Y + mdScroll.Y
			mousePosition := rl.GetMousePosition()
			for i, entry := range mdEntries {
				rect := rl.NewRectangle(12, (24*float32(i))+4+y, mdContentRectangle.Width, 24)
				inside := rl.CheckCollisionRecs(rect, mdRectangle)

				if inside {
					r := rl.GetCollisionRec(rect, mdRectangle)
					hover := rl.CheckCollisionPointRec(mousePosition, r)

					if hover {
						rl.DrawRectangleRec(r, rl.NewColor(0x2A, 0x2A, 0x2A, 0xFF))

						if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && i != modelIndex {
							if err := loadModel(i); err != nil {
								log.Println(err)
							}
						}
					} else if i == modelIndex {
						rl.DrawRectangleRec(r, rl.NewColor(0x1F, 0x1F, 0x1F, 0xFF))
					}

					raygui.Label(rect, entry.Name)
				}
			}
		}

		rl.EndScissorMode()

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

			if raygui.Button(rl.NewRectangle(8, 428, 87, 32), "Shift -1") {
				textureShift = (textureShift - 1) % textureTotal
				for _, model := range models {
					if texture, found := textures[model.Texture+textureShift]; found {
						rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, texture.Texture)
					} else {
						rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, textureDefault)
					}
				}
			}

			if raygui.Button(rl.NewRectangle(103, 428, 87, 32), "Shift +1") {
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

		showBones = raygui.CheckBox(rl.NewRectangle(8, 468, 14, 14), "Show Bones", showBones)

		if !hasModel {
			raygui.Disable()
		}

		if raygui.Button(rl.NewRectangle(8, 492, 132, 32), "Convert To GLTF") {
			go func() {
				log.Println("Convert Model to GLTF")
				if err := ConvertModelToGlft(datPath, tm3Entries[modelIndex], mdEntries[modelIndex], textureShift); err != nil {
					log.Println(err)
				} else {
					log.Println("Convert done")
				}
			}()
		}

		if !hasModel {
			raygui.Enable()
		}

		raygui.ScrollPanel(
			motRectangle,
			"",
			motContentRectangle,
			&motScroll,
			&motView,
		)

		// rl.DrawRectangle(
		//      int32(motRectangle.X+motScroll.X),
		//      int32(motRectangle.Y+motScroll.Y),
		//      int32(motContentRectangle.Width),
		//      int32(motContentRectangle.Height),
		//      rl.Fade(rl.Red, 0.1),
		// )

		rl.BeginScissorMode(
			int32(motView.X),
			int32(motView.Y),
			int32(motView.Width),
			int32(motView.Height),
		)

		{
			y := motRectangle.Y + motScroll.Y
			mousePosition := rl.GetMousePosition()
			for i, entry := range motEntries {
				rect := rl.NewRectangle(12, (24*float32(i))+4+y, motContentRectangle.Width, 24)
				inside := rl.CheckCollisionRecs(rect, motRectangle)

				if inside {
					r := rl.GetCollisionRec(rect, motRectangle)
					hover := rl.CheckCollisionPointRec(mousePosition, r)

					if hover {
						rl.DrawRectangleRec(r, rl.NewColor(0x2A, 0x2A, 0x2A, 0xFF))

						if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
							go func() {
								loadMotion(i)
							}()
						}
					} else if i == motionIndex {
						rl.DrawRectangleRec(r, rl.NewColor(0x1F, 0x1F, 0x1F, 0xFF))
					}

					raygui.Label(rect, entry.Name)
				}
			}
		}

		rl.EndScissorMode()

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
