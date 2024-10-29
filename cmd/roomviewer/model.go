package main

import (
	"fmt"

	"github.com/anasrar/chihuahua/pkg/scr"
	"github.com/anasrar/chihuahua/pkg/utils"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Model struct {
	Name        string
	Model       *rl.Model
	Texture     int
	Render      bool
	Translation rl.Vector3
	Rotation    rl.Vector3
	Scale       rl.Vector3
}

func NewModel(
	name string,
	model *rl.Model,
	texture int,
	translation rl.Vector3,
	rotation rl.Vector3,
	scale rl.Vector3,
) *Model {
	return &Model{
		Name:        name,
		Model:       model,
		Texture:     texture,
		Render:      true,
		Translation: translation,
		Rotation:    rotation,
		Scale:       scale,
	}
}

func LoadModels(datPath string, offset uint32) error {
	var s scr.Scr
	if err := scr.FromPathWithOffset(&s, datPath, offset); err != nil {
		return err
	}

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
					node.Translation,
					node.Rotation,
					node.Scale,
				),
			)
		}
	}
	return nil
}
