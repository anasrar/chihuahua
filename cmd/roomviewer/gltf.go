package main

import (
	"bytes"
	"fmt"
	"image/png"
	"os"

	"github.com/anasrar/chihuahua/pkg/dat"
	"github.com/anasrar/chihuahua/pkg/scr"
	"github.com/anasrar/chihuahua/pkg/tim3"
	"github.com/anasrar/chihuahua/pkg/tm3"
	"github.com/anasrar/chihuahua/pkg/utils"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

func ConvertToGlft() error {
	if scp == nil {
		return fmt.Errorf("SCP not found")
	}

	if datPath == "" {
		return fmt.Errorf("DAT not found")
	}

	dat0 := dat.New()
	if err := dat.FromPathWithOffsetSize(dat0, datPath, scp.Offset, scp.Size); err != nil {
		return err
	}

	doc := gltf.NewDocument()
	materials := map[uint16]int{}
	zero := float64(0)
	one := float64(1)
	nodes := 0

	for i, entry := range dat0.Entries {
		t := utils.FilterUnprintableString(entry.Type)
		switch t {
		case "TM3":
			tm := tm3.New()
			if err := tm3.FromPathWithOffsetSize(tm, datPath, entry.Offset, entry.Size); err != nil {
				return err
			}

			for i, entry := range tm.Entries {
				tim := tim3.New()
				if err := tim3.FromPathWithOffset(tim, datPath, entry.Offset); err != nil {
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
		case "SCR":
			doc.Nodes = append(doc.Nodes, &gltf.Node{
				Name: fmt.Sprintf("SCR_%03d", i),
			})
			doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, nodes)
			parentNodeIndex := nodes
			nodes++

			s := scr.New()
			if err := scr.FromPathWithOffset(s, datPath, entry.Offset); err != nil {
				return err
			}

			for _, node := range s.Nodes {
				name := utils.FilterUnprintableString(node.Name)
				primitives := []*gltf.Primitive{}

				for _, vb := range node.Mdb.VertexBuffers {
					materialIndex, ok := materials[vb.Material]
					if !ok {
						k := len(doc.Materials)
						materials[vb.Material] = k
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

				orientation := rl.QuaternionFromEuler(node.Rotation[2], node.Rotation[1], node.Rotation[1])

				doc.Meshes = append(doc.Meshes, &gltf.Mesh{
					Name:       name,
					Primitives: primitives,
				})
				doc.Nodes = append(doc.Nodes, &gltf.Node{
					Name:        name,
					Mesh:        gltf.Index(len(doc.Meshes) - 1),
					Translation: [3]float64{float64(node.Translation[0]), float64(node.Translation[1]), float64(node.Translation[2])},
					Rotation:    [4]float64{float64(orientation.X), float64(orientation.Y), float64(orientation.Z), float64(orientation.W)},
					Scale:       [3]float64{float64(node.Scale[0]), float64(node.Scale[1]), float64(node.Scale[2])},
				})
				doc.Nodes[parentNodeIndex].Children = append(doc.Nodes[parentNodeIndex].Children, nodes)
				nodes++
			}
		default:
		}
	}

	output := fmt.Sprintf(
		"%s/GLTF_%s",
		utils.ParentDirectory(datPath),
		utils.Basename(datPath),
	)

	if err := os.MkdirAll(output, os.ModePerm); err != nil {
		return err
	}

	if err := gltf.Save(doc, fmt.Sprintf("%s/%s.gltf", output, utils.BasenameWithoutExt(datPath))); err != nil {
		return err
	}

	return nil
}
