package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/anasrar/chihuahua/pkg/bone"
	"github.com/anasrar/chihuahua/pkg/dat"
	"github.com/anasrar/chihuahua/pkg/mot"
	rlig "github.com/anasrar/chihuahua/pkg/raylib_imgui"
	"github.com/anasrar/chihuahua/pkg/scr"
	"github.com/anasrar/chihuahua/pkg/tim3"
	"github.com/anasrar/chihuahua/pkg/tm3"
	"github.com/anasrar/chihuahua/pkg/utils"
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
					worldPosition := rl.Vector3Add(parentWorldPosition, rl.NewVector3(bone.Translation[0], bone.Translation[1], bone.Translation[2]))
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

		boneTree = NewBoneNode(bone.New(0, "root", 0, 0, 0, 0, 0, 0, -1))
		boneNodes = []*BoneNode{boneTree}

		for _, bone := range s.Nodes[0].Mdb.Bones {
			node := NewBoneNode(bone)
			boneNodes = append(boneNodes, node)

			boneNodes[bone.Parent].Children = append(
				boneNodes[bone.Parent].Children,
				node,
			)
		}
	}

	modelIndex = index
	motionIndex = -1

	frames = [][]*bone.Bone{}
	frameTotal = int32(0)
	frameIndex = int32(0)
	framePlay = false

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

	if err := loadModel(0); err != nil {
		return err
	}

	datPath = filePath
	return nil
}

func loadMotion(index int) error {
	motEntry := motEntries[index]

	m := mot.New()
	if err := mot.FromPathWithOffsetSize(m, motEntry.Source, motEntry.Offset, motEntry.Size); err != nil {
		return err
	}

	frames = [][]*bone.Bone{}

	for range m.FrameTotal {
		bones := []*bone.Bone{}
		for _, node := range boneNodes {
			bones = append(bones, bone.New(node.Bone.Index, node.Bone.Name, 0, 0, 0, 0, 0, 0, node.Bone.Parent))
		}
		frames = append(frames, bones)
	}

	for _, record := range m.Records {
		if record.IsNull {
			continue
		}

		values := record.QuantizeHermite(m.FrameTotal)

		for frame, value := range values {
			switch record.Channel {
			case 16:
				frames[frame][record.Target].Translation[0] = value
			case 17:
				frames[frame][record.Target].Translation[1] = value
			case 18:
				frames[frame][record.Target].Translation[2] = value
			case 19:
				frames[frame][record.Target].Rotation[0] = value
			case 20:
				frames[frame][record.Target].Rotation[1] = value
			case 21:
				frames[frame][record.Target].Rotation[2] = value
			}
		}
	}

	motionIndex = index
	frameIndex = 0
	frameTotal = int32(m.FrameTotal - 1)
	framePlay = true
	return nil
}

func main() {
	rl.InitWindow(int32(width), int32(height), "Model Viewer")
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

		if framePlay {
			frameIndex = (frameIndex + 1) % frameTotal
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

		imgui.SetNextWindowPosV(imgui.NewVec2(12, 12), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
		imgui.SetNextWindowSizeV(imgui.NewVec2(200, 200), imgui.CondFirstUseEver)
		imgui.BeginV("MD", nil, imgui.WindowFlagsNone)
		for i, entry := range mdEntries {
			imgui.PushIDStr(entry.Name)
			imgui.BeginDisabledV(i == modelIndex)
			if imgui.Button("View") {
				if err := loadModel(i); err != nil {
					log.Println(err)
				}
			}
			imgui.EndDisabled()
			imgui.PopID()
			imgui.SameLineV(0, 4)
			imgui.Text(entry.Name)
		}
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(12, 224), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
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
		imgui.BeginDisabledV(len(models) == 0)
		if imgui.Button("Convert To GLTF") {
			go func() {
				log.Println("Convert Model to GLTF")
				if err := ConvertModelToGlft(datPath, tm3Entries[modelIndex], mdEntries[modelIndex], textureShift); err != nil {
					log.Println(err)
				} else {
					log.Println("Convert done")
				}
			}()
		}
		imgui.EndDisabled()
		imgui.Separator()
		imgui.BeginChildStrV("MdbRegion", imgui.NewVec2(0, 0), imgui.ChildFlagsNavFlattened, imgui.WindowFlagsHorizontalScrollbar)
		for _, model := range models {
			imgui.Checkbox(model.Name, &model.Render)
		}
		imgui.EndChild()
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(12, 536), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
		imgui.SetNextWindowSizeV(imgui.NewVec2(200, 240), imgui.CondFirstUseEver)
		imgui.BeginV("MOT", nil, imgui.WindowFlagsNoFocusOnAppearing)
		imgui.BeginDisabledV(motionIndex == -1)
		imgui.Checkbox("Play", &framePlay)
		imgui.SliderInt("Frame", &(frameIndex), 0, int32(frameTotal))
		imgui.EndDisabled()
		imgui.Separator()
		imgui.BeginChildStrV("MotRegion", imgui.NewVec2(0, 0), imgui.ChildFlagsNavFlattened, imgui.WindowFlagsHorizontalScrollbar)
		for i, entry := range motEntries {
			imgui.PushIDStr(entry.Name)
			imgui.BeginDisabledV(i == motionIndex)
			if imgui.Button("Play") {
				if err := loadMotion(i); err != nil {
					log.Println(err)
				}
			}
			imgui.EndDisabled()
			imgui.PopID()
			imgui.SameLineV(0, 4)
			imgui.Text(entry.Name)
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

			if motionIndex == -1 {
				model := models[0]

				rl.DrawCube(rl.Vector3Zero(), .02, .02, .02, rl.Green)
				bones := model.Model.GetBones()
				pose := model.Model.GetBindPose()
				for j := int32(1); j < model.Model.BoneCount; j++ {
					rl.DrawCube(pose[j].Translation, .02, .02, .02, rl.Green)
					rl.DrawLine3D(pose[j].Translation, pose[bones[j].Parent].Translation, rl.Green)
				}
			} else {
				DrawBoneTree(boneTree, frameIndex)
			}

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
