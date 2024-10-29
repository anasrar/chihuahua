package scr

import (
	"bytes"
	"fmt"
	"image/png"
	"os"

	"github.com/anasrar/chihuahua/pkg/tim3"
	"github.com/anasrar/chihuahua/pkg/tm3"
	"github.com/anasrar/chihuahua/pkg/utils"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

func ConvertToGlft(
	scrPath string,
	tm3Path string,
	textureShift int,
) error {
	doc := gltf.NewDocument()
	materials := map[uint16]int{}
	zero := float64(0)
	one := float64(1)

	if tm3Path != "" {
		tm := tm3.New()
		if err := tm3.FromPath(tm, tm3Path); err != nil {
			return err
		}

		for i, entry := range tm.Entries {
			tim := tim3.New()
			if err := tim3.FromPathWithOffset(tim, tm3Path, entry.Offset); err != nil {
				return err
			}

			var buf bytes.Buffer
			picture := tim.Pictures[0]
			if err := png.Encode(&buf, tim3.PictureToImage(picture)); err != nil {
				return err
			}

			index, _ := modeler.WriteImage(doc, fmt.Sprintf("%s_%03d", entry.Name, i), "image/png", &buf)
			doc.Textures = append(doc.Textures, &gltf.Texture{
				Source: gltf.Index(index),
			})

			k := len(doc.Materials)
			materials[uint16(i)] = k
			doc.Materials = append(doc.Materials,
				&gltf.Material{
					PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
						BaseColorTexture: &gltf.TextureInfo{
							Index: index,
						},
						MetallicFactor:  &zero,
						RoughnessFactor: &one,
					},
					AlphaMode: gltf.AlphaMask,
				},
			)
		}
	}

	s := New()
	if err := FromPath(s, scrPath); err != nil {
		return err
	}

	output := fmt.Sprintf(
		"%s/GLTF_%s",
		utils.ParentDirectory(scrPath),
		utils.Basename(scrPath),
	)

	if err := os.MkdirAll(output, os.ModePerm); err != nil {
		return err
	}

	nodes := 0

	doc.Skins = []*gltf.Skin{{
		Skeleton: gltf.Index(nodes),
		Joints:   []int{nodes},
	}}

	rootBoneIndexInNodes := nodes
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name: "root",
	})
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, nodes)
	nodes++

	bonesWorldPosition := map[int16][3]float32{
		0: {0, 0, 0},
	}
	inverse := [][4][4]float32{{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}}
	bones := 1
	for _, bone := range s.Nodes[0].Mdb.Bones {

		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name:        fmt.Sprintf("%d", bones),
			Translation: [3]float64{float64(bone.X), float64(bone.Y), float64(bone.Z)},
		})

		parentNode := doc.Nodes[rootBoneIndexInNodes+int(bone.Parent)]

		parentNode.Children = append(
			parentNode.Children,
			nodes,
		)
		doc.Skins[0].Joints = append(
			doc.Skins[0].Joints,
			nodes,
		)

		parentWorldPosition := bonesWorldPosition[bone.Parent]
		parentWorldPositionX := parentWorldPosition[0]
		parentWorldPositionY := parentWorldPosition[1]
		parentWorldPositionZ := parentWorldPosition[2]

		worldPositionX := parentWorldPositionX + bone.X
		worldPositionY := parentWorldPositionY + bone.Y
		worldPositionZ := parentWorldPositionZ + bone.Z

		bonesWorldPosition[int16(bones)] = [3]float32{
			worldPositionX,
			worldPositionY,
			worldPositionZ,
		}

		bones++
		nodes++

		inverse = append(
			inverse, [4][4]float32{
				{1, 0, 0, -worldPositionX},
				{0, 1, 0, -worldPositionY},
				{0, 0, 1, -worldPositionZ},
				{0, 0, 0, 1},
			},
		)
	}
	doc.Skins[0].InverseBindMatrices = gltf.Index(modeler.WriteAccessor(doc, gltf.TargetArrayBuffer, inverse))

	for _, node := range s.Nodes {
		name := utils.FilterUnprintableString(node.Name)
		primitives := []*gltf.Primitive{}

		for _, vb := range node.Mdb.VertexBuffers {
			materialIndex, ok := materials[vb.Material+uint16(textureShift)]
			if !ok {
				k := len(doc.Materials)
				materials[vb.Material+uint16(textureShift)] = k
				materialIndex = k
				doc.Materials = append(doc.Materials,
					&gltf.Material{
						PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
							MetallicFactor:  &zero,
							RoughnessFactor: &one,
						},
						AlphaMode: gltf.AlphaMask,
					},
				)
			}

			for _, index := range vb.Indices {
				p1 := vb.Vertices[index[0]]
				p2 := vb.Vertices[index[1]]
				p3 := vb.Vertices[index[2]]

				uv1 := vb.Uvs[index[0]]
				uv2 := vb.Uvs[index[1]]
				uv3 := vb.Uvs[index[2]]

				attributes := gltf.PrimitiveAttributes{
					gltf.POSITION:   modeler.WritePosition(doc, [][3]float32{p1, p2, p3}),
					gltf.TEXCOORD_0: modeler.WriteTextureCoord(doc, [][2]float32{uv1, uv2, uv3}),
				}

				if len(vb.Joints) != 0 {
					joint1 := vb.Joints[index[0]]
					joint2 := vb.Joints[index[1]]
					joint3 := vb.Joints[index[2]]

					attributes[gltf.JOINTS_0] = modeler.WriteJoints(doc, [][4]uint8{joint1, joint2, joint3})
				}

				if len(vb.Weights) != 0 {
					weight1 := vb.Weights[index[0]]
					weight2 := vb.Weights[index[1]]
					weight3 := vb.Weights[index[2]]

					attributes[gltf.WEIGHTS_0] = modeler.WriteWeights(doc, [][4]float32{weight1, weight2, weight3})
				}

				if len(vb.Normals) != 0 {
					normal1 := vb.Normals[index[0]]
					normal2 := vb.Normals[index[1]]
					normal3 := vb.Normals[index[2]]

					attributes[gltf.NORMAL] = modeler.WriteNormal(doc, [][3]float32{normal1, normal2, normal3})
				}

				primitives = append(
					primitives,
					&gltf.Primitive{
						Indices:    gltf.Index(modeler.WriteIndices(doc, []uint16{0, 1, 2})),
						Attributes: attributes,
						Material:   gltf.Index(materialIndex),
					},
				)
			}
		}

		doc.Meshes = append(doc.Meshes, &gltf.Mesh{
			Name:       name,
			Primitives: primitives,
		})
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name: name,
			Mesh: gltf.Index(len(doc.Meshes) - 1),
			Skin: gltf.Index(0),
		})
		doc.Nodes[0].Children = append(doc.Nodes[0].Children, nodes)
		nodes++
	}

	if err := gltf.Save(doc, fmt.Sprintf("%s/%s.gltf", output, utils.BasenameWithoutExt(scrPath))); err != nil {
		return err
	}

	return nil
}
