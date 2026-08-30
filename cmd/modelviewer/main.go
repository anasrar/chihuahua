package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/anasrar/chihuahua/pkg/bone"
	"github.com/anasrar/chihuahua/pkg/dat"
	"github.com/anasrar/chihuahua/pkg/mot"
	rlig "github.com/anasrar/chihuahua/pkg/raylib_imgui"
	"github.com/anasrar/chihuahua/pkg/scr"
	"github.com/anasrar/chihuahua/pkg/t32"
	"github.com/anasrar/chihuahua/pkg/tim2"
	"github.com/anasrar/chihuahua/pkg/tim3"
	"github.com/anasrar/chihuahua/pkg/tm3"
	"github.com/anasrar/chihuahua/pkg/utils"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func loadModel(index int) error {

	if len(tm3Entries) != 0 {
		tm3Entry := tm3Entries[index]

		tm := tm3.New()
		if err := tm3.FromPathWithOffsetSize(tm, tm3Entry.Source, tm3Entry.Offset, tm3Entry.Size); err != nil {
			return err
		}

		for _, index := range textureIndices {
			rl.UnloadTexture(textures[index].Texture)
			textures[index].Ref.Destroy()
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
			textures[i] = NewTexture(tim.Pictures[0], texture)

			textureIndices = append(textureIndices, i)
		}

		textureShift = 0
	}

	if len(t32Entries) != 0 {
		t32Entry := t32Entries[index]

		file, err := os.Open(t32Entry.Source)
		if err != nil {
			return err
		}
		defer file.Close()

		t := t32.New()
		log.Print(t32Entry.Source, t32Entry.Offset)
		if err := t32.FromStreamWithOffset(t, file, t32Entry.Offset); err != nil {
			return err
		}

		for _, index := range textureIndices {
			rl.UnloadTexture(textures[index].Texture)
			textures[index].Ref.Destroy()
			delete(textures, index)
		}

		textureIndices = []int{}
		textureTotal = 0

		for textureIndex, texture := range t.Textures {
			if !texture.IsIndexed() {
				continue
			}

			for paletteIndex := range texture.Palettes {
				decoded, err := t.DecodeTexture(textureIndex, paletteIndex)
				if err != nil {
					return err
				}

				buf := bytes.NewBuffer([]byte{})
				if err := png.Encode(buf, decoded.Image()); err != nil {
					return err
				}

				img := rl.LoadImageFromMemory(".png", buf.Bytes(), int32(buf.Len()))
				defer rl.UnloadImage(img)

				texture := rl.LoadTextureFromImage(img)
				textures[textureTotal] = NewTexture(&tim2.Picture{}, texture)

				textureIndices = append(textureIndices, textureTotal)

				textureTotal += 1
			}
		}

		textureShift = 0
	}

	mdEntry := mdEntries[index]
	{
		s := scr.New()
		if err := scr.FromPathWithOffset(s, mdEntry.Source, mdEntry.Offset); err != nil {
			return err
		}

		for _, model := range models {
			model.Unload()
		}

		models = []*Model{}

		for i, node := range s.Nodes {
			bones := []rl.BoneInfo{{
				Name:   [32]int8{0x72, 0x6f, 0x6f, 0x74},
				Parent: -1,
			}}
			bindLocal := []rl.Transform{{
				Translation: rl.Vector3Zero(),
				Rotation:    rl.QuaternionIdentity(),
				Scale:       rl.Vector3One(),
			}}

			for _, mdbBone := range node.Mdb.Bones {
				boneName := [32]int8{}
				for nameIndex, c := range []uint8(mdbBone.Name) {
					if nameIndex == len(boneName) {
						break
					}
					boneName[nameIndex] = int8(c)
				}

				bones = append(bones, rl.BoneInfo{
					Name:   boneName,
					Parent: int32(mdbBone.Parent),
				})
				bindLocal = append(bindLocal, rl.Transform{
					Translation: rl.NewVector3(mdbBone.Translation[0], mdbBone.Translation[1], mdbBone.Translation[2]),
					Rotation:    rl.QuaternionIdentity(),
					Scale:       rl.Vector3One(),
				})
			}

			bindPose, err := buildWorldPose(bones, bindLocal)
			if err != nil {
				return fmt.Errorf("build skeleton for SCR node %q: %w", utils.FilterUnprintableString(node.Name), err)
			}

			for j, vb := range node.Mdb.VertexBuffers {
				if len(vb.Indices) == 0 {
					continue
				}
				name := fmt.Sprintf("%s_%03d_%03d", utils.FilterUnprintableString(node.Name), i, j)

				vertices := []float32{}
				normals := []float32{}
				uvs := []float32{}
				joints := []uint8{}
				weights := []float32{}

				hasNormals := len(vb.Normals) != 0
				if hasNormals && len(vb.Normals) != len(vb.Vertices) {
					return fmt.Errorf("mesh %q has %d vertices but %d normals", name, len(vb.Vertices), len(vb.Normals))
				}
				if len(vb.Uvs) != len(vb.Vertices) {
					return fmt.Errorf("mesh %q has %d vertices but %d texture coordinates", name, len(vb.Vertices), len(vb.Uvs))
				}
				hasSkin := len(vb.Joints) != 0 || len(vb.Weights) != 0
				if hasSkin && (len(vb.Joints) != len(vb.Vertices) || len(vb.Weights) != len(vb.Vertices)) {
					return fmt.Errorf("mesh %q has incomplete joint/weight data", name)
				}

				for _, triangle := range vb.Indices {
					for _, rawVertexIndex := range triangle {
						vertexIndex := int(rawVertexIndex)
						if vertexIndex < 0 || vertexIndex >= len(vb.Vertices) {
							return fmt.Errorf("mesh %q triangle refers to invalid vertex %d", name, vertexIndex)
						}

						position := vb.Vertices[vertexIndex]
						vertices = append(vertices, position[:]...)
						uv := vb.Uvs[vertexIndex]
						uvs = append(uvs, uv[:]...)
						if hasNormals {
							normal := vb.Normals[vertexIndex]
							normals = append(normals, normal[:]...)
						}

						if hasSkin {
							joint := vb.Joints[vertexIndex]
							weight := vb.Weights[vertexIndex]
							joints = append(joints, joint[:]...)
							weights = append(weights, weight[:]...)
						} else {
							joints = append(joints, 0, 0, 0, 0)
							weights = append(weights, 1, 0, 0, 0)
						}
					}
				}

				buffers := NewMeshBuffers(vertices, normals, uvs, joints, weights)
				vertexCount := len(vertices) / 3
				if err := buffers.Validate(vertexCount, len(bones)); err != nil {
					return fmt.Errorf("prepare mesh %q: %w", name, err)
				}

				mesh := rl.Mesh{
					TriangleCount: int32(len(vb.Indices)),
					VertexCount:   int32(vertexCount),
					BoneCount:     int32(len(bones)),
				}
				buffers.Attach(&mesh)
				rl.UploadMesh(&mesh, true)

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
						buffers,
						bones,
						bindLocal,
						bindPose,
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

	frames = [][]mot.Transform{}
	frameTotal = int32(0)
	frameIndex = int32(0)
	framePlay = false
	frameDirty = false

	return nil
}

func drop(filePath string) error {
	dat0 := dat.New()
	if err := dat.FromPath(dat0, filePath); err != nil {
		return err
	}

	tmpTm3Entries := []*Entry{}
	tmpT32Entries := []*Entry{}
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
		case "T32":
			tmpT32Entries = append(tmpT32Entries, NewEntry(name, entry))
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
	t32Entries = tmpT32Entries
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
	if m.FrameTotal == 0 {
		return fmt.Errorf("MOT contains no frames")
	}

	quantizedFrames, err := m.QuantizeTransforms(len(boneNodes))
	if err != nil {
		return err
	}
	for _, model := range models {
		if len(quantizedFrames[0]) < len(model.Bones) {
			return fmt.Errorf("MOT has %d bone transforms, model %q requires %d", len(quantizedFrames[0]), model.Name, len(model.Bones))
		}
	}
	frames = quantizedFrames

	motionIndex = index
	frameIndex = 0
	frameTotal = int32(m.FrameTotal - 1)
	framePlay = true
	frameDirty = true
	return nil
}

func applyMotionFrame() error {
	if motionIndex == -1 || !frameDirty {
		return nil
	}
	if frameIndex < 0 || int(frameIndex) >= len(frames) {
		return fmt.Errorf("MOT frame %d is outside the decoded frame range", frameIndex)
	}

	for _, model := range models {
		if err := model.ApplyAnimation(frames[frameIndex]); err != nil {
			return err
		}
	}
	frameDirty = false
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

		if framePlay && motionIndex != -1 {
			frameIndex = (frameIndex + 1) % (frameTotal + 1)
			frameDirty = true
		}

		imgui.NewFrame()

		imgui.SetNextWindowPosV(imgui.NewVec2(12, height-12), imgui.CondAlways, imgui.NewVec2(0, 1))
		imgui.BeginV("Credit", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.Text("https://github.com/anasrar/chihuahua")
		imgui.End()

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
			imgui.Image(texture.Ref, imgui.NewVec2(42, 42))
			if imgui.BeginItemTooltip() {
				imgui.Text(
					fmt.Sprintf(
						"%dx%d\n%s\n%s\nClut Colors %d",
						texture.Picture.ImageWidth,
						texture.Picture.ImageHeight,
						texture.Picture.ClutType,
						texture.Picture.ImageType,
						texture.Picture.ClutColors,
					),
				)
				imgui.Image(texture.Ref, imgui.NewVec2(float32(texture.Texture.Width), float32(texture.Texture.Height)))
				imgui.EndTooltip()
			}
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
		previousFrame := frameIndex
		imgui.SliderInt("Frame", &(frameIndex), 0, int32(frameTotal))
		if frameIndex != previousFrame {
			frameDirty = true
		}
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

		if err := applyMotionFrame(); err != nil {
			log.Println(err)
			framePlay = false
			frameDirty = false
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

			DrawModelSkeleton(models[0], motionIndex != -1)

			rl.EndMode3D()
			rl.EndTextureMode()

			rl.DrawTextureRec(boneRender.Texture, rl.NewRectangle(0, 0, width, -height), rl.Vector2Zero(), rl.White)
		}

		rlig.Render()
		rl.EndDrawing()
	}

	for _, model := range models {
		model.Unload()
	}

	for _, index := range textureIndices {
		rl.UnloadTexture(textures[index].Texture)
		textures[index].Ref.Destroy()
		delete(textures, index)
	}
}
