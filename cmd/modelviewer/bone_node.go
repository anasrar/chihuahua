package main

import (
	"github.com/anasrar/chihuahua/pkg/bone"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type BoneNode struct {
	Bone     *bone.Bone
	Children []*BoneNode
}

func NewBoneNode(bone *bone.Bone) *BoneNode {
	return &BoneNode{
		Bone:     bone,
		Children: []*BoneNode{},
	}
}

func DrawBoneTree(node *BoneNode, frame int32) {
	rl.PushMatrix()

	if node.Bone.Index != 1 {
		rl.Translatef(node.Bone.Translation[0], node.Bone.Translation[1], node.Bone.Translation[2])
	}

	if motionIndex != -1 {
		bone := frames[frame][node.Bone.Index]
		rl.Translatef(bone.Translation[0], bone.Translation[1], bone.Translation[2])
		rl.Rotatef(bone.Rotation[0]*rl.Rad2deg, 1, 0, 0)
		rl.Rotatef(bone.Rotation[1]*rl.Rad2deg, 0, 1, 0)
		rl.Rotatef(bone.Rotation[2]*rl.Rad2deg, 0, 0, 1)
	}

	rl.DrawCube(rl.Vector3Zero(), .02, .02, .02, rl.Blue)

	for _, child := range node.Children {
		rl.DrawLine3D(
			rl.Vector3Zero(),
			rl.NewVector3(
				child.Bone.Translation[0],
				child.Bone.Translation[1],
				child.Bone.Translation[2],
			),
			rl.Blue,
		)

		DrawBoneTree(child, frame)
	}

	rl.PopMatrix()
}
