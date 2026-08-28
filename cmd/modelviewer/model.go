package main

import (
	"fmt"
	"unsafe"

	"github.com/anasrar/chihuahua/pkg/mot"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	meshBufferPosition = 0
	meshBufferNormal   = 2
)

type MeshBuffers struct {
	Vertices     []float32
	Normals      []float32
	Texcoords    []float32
	BoneIndices  []uint8
	BoneWeights  []float32
	AnimVertices []float32
	AnimNormals  []float32
}

func NewMeshBuffers(vertices, normals, texcoords []float32, boneIndices []uint8, boneWeights []float32) MeshBuffers {
	buffers := MeshBuffers{
		Vertices:     vertices,
		Normals:      normals,
		Texcoords:    texcoords,
		BoneIndices:  boneIndices,
		BoneWeights:  boneWeights,
		AnimVertices: append([]float32(nil), vertices...),
	}
	if len(normals) != 0 {
		buffers.AnimNormals = append([]float32(nil), normals...)
	}
	return buffers
}

func (buffers *MeshBuffers) Validate(vertexCount, boneCount int) error {
	if len(buffers.Vertices) != vertexCount*3 {
		return fmt.Errorf("mesh has %d position values, expected %d", len(buffers.Vertices), vertexCount*3)
	}
	if len(buffers.Texcoords) != 0 && len(buffers.Texcoords) != vertexCount*2 {
		return fmt.Errorf("mesh has %d texture coordinate values, expected %d", len(buffers.Texcoords), vertexCount*2)
	}
	if len(buffers.Normals) != 0 && len(buffers.Normals) != vertexCount*3 {
		return fmt.Errorf("mesh has %d normal values, expected %d", len(buffers.Normals), vertexCount*3)
	}
	if len(buffers.BoneIndices) != vertexCount*4 || len(buffers.BoneWeights) != vertexCount*4 {
		return fmt.Errorf("mesh skin data does not contain four influences for each of its %d vertices", vertexCount)
	}

	for i, boneIndex := range buffers.BoneIndices {
		if buffers.BoneWeights[i] != 0 && int(boneIndex) >= boneCount {
			return fmt.Errorf("mesh vertex influence refers to bone %d, but skeleton has %d bones", boneIndex, boneCount)
		}
	}
	return nil
}

func (buffers *MeshBuffers) Attach(mesh *rl.Mesh) {
	mesh.Vertices = unsafe.SliceData(buffers.Vertices)
	mesh.Normals = unsafe.SliceData(buffers.Normals)
	mesh.Texcoords = unsafe.SliceData(buffers.Texcoords)
	mesh.BoneIndices = unsafe.SliceData(buffers.BoneIndices)
	mesh.BoneWeights = unsafe.SliceData(buffers.BoneWeights)
	mesh.AnimVertices = unsafe.SliceData(buffers.AnimVertices)
	mesh.AnimNormals = unsafe.SliceData(buffers.AnimNormals)
}

type Model struct {
	Name        string
	Model       *rl.Model
	Texture     int
	Render      bool
	Translation rl.Vector3
	Rotation    rl.Vector3
	Scale       rl.Vector3

	Buffers        MeshBuffers
	Bones          []rl.BoneInfo
	BindLocal      []rl.Transform
	BindPose       []rl.Transform
	CurrentPose    []rl.Transform
	BoneMatrices   []rl.Matrix
	NormalMatrices []rl.Matrix
}

func NewModel(
	name string,
	model *rl.Model,
	texture int,
	translation rl.Vector3,
	rotation rl.Vector3,
	scale rl.Vector3,
	buffers MeshBuffers,
	bones []rl.BoneInfo,
	bindLocal []rl.Transform,
	bindPose []rl.Transform,
) *Model {
	result := &Model{
		Name:           name,
		Model:          model,
		Texture:        texture,
		Render:         true,
		Translation:    translation,
		Rotation:       rotation,
		Scale:          scale,
		Buffers:        buffers,
		Bones:          bones,
		BindLocal:      bindLocal,
		BindPose:       bindPose,
		CurrentPose:    make([]rl.Transform, len(bones)),
		BoneMatrices:   make([]rl.Matrix, len(bones)),
		NormalMatrices: make([]rl.Matrix, len(bones)),
	}

	model.Skeleton.BoneCount = int32(len(bones))
	model.Skeleton.Bones = unsafe.SliceData(result.Bones)
	model.Skeleton.BindPose = unsafe.SliceData(result.BindPose)
	return result
}

func buildWorldPose(bones []rl.BoneInfo, local []rl.Transform) ([]rl.Transform, error) {
	if len(local) != len(bones) {
		return nil, fmt.Errorf("skeleton has %d bones but %d local transforms", len(bones), len(local))
	}

	world := append([]rl.Transform(nil), local...)
	for i := range world {
		parent := int(bones[i].Parent)
		if parent < 0 {
			continue
		}
		if parent >= i {
			return nil, fmt.Errorf("bone %d has non-topological parent %d", i, parent)
		}

		world[i].Rotation = rl.QuaternionMultiply(world[parent].Rotation, world[i].Rotation)
		world[i].Scale = rl.Vector3Multiply(world[i].Scale, world[parent].Scale)
		world[i].Translation = rl.Vector3Multiply(world[i].Translation, world[parent].Scale)
		world[i].Translation = rl.Vector3RotateByQuaternion(world[i].Translation, world[parent].Rotation)
		world[i].Translation = rl.Vector3Add(world[i].Translation, world[parent].Translation)
	}
	return world, nil
}

func transformMatrix(transform rl.Transform) rl.Matrix {
	return rl.MatrixMultiply(
		rl.MatrixMultiply(
			rl.MatrixScale(transform.Scale.X, transform.Scale.Y, transform.Scale.Z),
			rl.QuaternionToMatrix(transform.Rotation),
		),
		rl.MatrixTranslate(transform.Translation.X, transform.Translation.Y, transform.Translation.Z),
	)
}

func (model *Model) ApplyAnimation(animation []mot.Transform) error {
	if len(animation) < len(model.Bones) {
		return fmt.Errorf("MOT frame has %d bones, model %q requires %d", len(animation), model.Name, len(model.Bones))
	}

	local := make([]rl.Transform, len(model.BindLocal))
	copy(local, model.BindLocal)
	for boneIndex, transform := range animation[:len(local)] {
		translation := rl.NewVector3(transform.Translation[0], transform.Translation[1], transform.Translation[2])
		if boneIndex == 1 {
			local[boneIndex].Translation = translation
		} else {
			local[boneIndex].Translation = rl.Vector3Add(model.BindLocal[boneIndex].Translation, translation)
		}
		local[boneIndex].Rotation = rl.QuaternionMultiply(
			model.BindLocal[boneIndex].Rotation,
			rl.QuaternionFromEuler(transform.Rotation[0], transform.Rotation[1], transform.Rotation[2]),
		)
	}

	currentPose, err := buildWorldPose(model.Bones, local)
	if err != nil {
		return fmt.Errorf("build pose for model %q: %w", model.Name, err)
	}
	copy(model.CurrentPose, currentPose)

	for boneIndex := range model.Bones {
		bindMatrix := transformMatrix(model.BindPose[boneIndex])
		currentMatrix := transformMatrix(model.CurrentPose[boneIndex])
		model.BoneMatrices[boneIndex] = rl.MatrixMultiply(rl.MatrixInvert(bindMatrix), currentMatrix)
		model.NormalMatrices[boneIndex] = rl.MatrixTranspose(rl.MatrixInvert(model.BoneMatrices[boneIndex]))
	}

	vertexCount := len(model.Buffers.Vertices) / 3
	for vertexIndex := 0; vertexIndex < vertexCount; vertexIndex++ {
		positionOffset := vertexIndex * 3
		influenceOffset := vertexIndex * 4
		basePosition := rl.NewVector3(
			model.Buffers.Vertices[positionOffset],
			model.Buffers.Vertices[positionOffset+1],
			model.Buffers.Vertices[positionOffset+2],
		)
		animatedPosition := rl.Vector3Zero()
		animatedNormal := rl.Vector3Zero()
		hasInfluence := false

		for influence := 0; influence < 4; influence++ {
			weight := model.Buffers.BoneWeights[influenceOffset+influence]
			if weight == 0 {
				continue
			}
			boneIndex := int(model.Buffers.BoneIndices[influenceOffset+influence])
			if boneIndex >= len(model.BoneMatrices) {
				return fmt.Errorf("model %q vertex %d refers to missing bone %d", model.Name, vertexIndex, boneIndex)
			}

			position := rl.Vector3Transform(basePosition, model.BoneMatrices[boneIndex])
			animatedPosition = rl.Vector3Add(animatedPosition, rl.Vector3Scale(position, weight))
			hasInfluence = true

			if len(model.Buffers.Normals) != 0 {
				baseNormal := rl.NewVector3(
					model.Buffers.Normals[positionOffset],
					model.Buffers.Normals[positionOffset+1],
					model.Buffers.Normals[positionOffset+2],
				)
				normal := rl.Vector3Transform(baseNormal, model.NormalMatrices[boneIndex])
				animatedNormal = rl.Vector3Add(animatedNormal, rl.Vector3Scale(normal, weight))
			}
		}

		if !hasInfluence {
			animatedPosition = basePosition
		}
		model.Buffers.AnimVertices[positionOffset] = animatedPosition.X
		model.Buffers.AnimVertices[positionOffset+1] = animatedPosition.Y
		model.Buffers.AnimVertices[positionOffset+2] = animatedPosition.Z

		if len(model.Buffers.AnimNormals) != 0 {
			if !hasInfluence {
				animatedNormal = rl.NewVector3(
					model.Buffers.Normals[positionOffset],
					model.Buffers.Normals[positionOffset+1],
					model.Buffers.Normals[positionOffset+2],
				)
			}
			model.Buffers.AnimNormals[positionOffset] = animatedNormal.X
			model.Buffers.AnimNormals[positionOffset+1] = animatedNormal.Y
			model.Buffers.AnimNormals[positionOffset+2] = animatedNormal.Z
		}
	}

	mesh := *model.Model.Meshes
	rl.UpdateMeshBuffer(mesh, meshBufferPosition, float32Bytes(model.Buffers.AnimVertices), 0)
	if len(model.Buffers.AnimNormals) != 0 {
		rl.UpdateMeshBuffer(mesh, meshBufferNormal, float32Bytes(model.Buffers.AnimNormals), 0)
	}
	return nil
}

func float32Bytes(values []float32) []byte {
	if len(values) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(values))), len(values)*int(unsafe.Sizeof(float32(0))))
}

func (model *Model) Unload() {
	if model == nil || model.Model == nil {
		return
	}
	if model.Model.Meshes != nil {
		mesh := model.Model.Meshes
		mesh.Vertices = nil
		mesh.Normals = nil
		mesh.Texcoords = nil
		mesh.BoneIndices = nil
		mesh.BoneWeights = nil
		mesh.AnimVertices = nil
		mesh.AnimNormals = nil
	}
	model.Model.Skeleton.BoneCount = 0
	model.Model.Skeleton.Bones = nil
	model.Model.Skeleton.BindPose = nil
	model.Model.CurrentPose = nil
	model.Model.BoneMatrices = nil
	rl.UnloadModel(*model.Model)
	model.Model = nil
}
