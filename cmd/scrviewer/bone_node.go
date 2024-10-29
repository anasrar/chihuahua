package main

import rl "github.com/gen2brain/raylib-go/raylib"

type BoneNode struct {
	X        float32
	Y        float32
	Z        float32
	Children []*BoneNode
}

func NewBoneNode(x, y, z float32) *BoneNode {
	return &BoneNode{
		X:        x,
		Y:        y,
		Z:        z,
		Children: []*BoneNode{},
	}
}

func DrawBoneTree(node *BoneNode) {
	rl.PushMatrix()

	rl.Translatef(node.X, node.Y, node.Z)

	rl.DrawCube(rl.Vector3Zero(), .02, .02, .02, rl.Green)

	for _, child := range node.Children {
		rl.DrawLine3D(rl.Vector3Zero(), rl.NewVector3(child.X, child.Y, child.Z), rl.Green)

		DrawBoneTree(child)
	}

	rl.PopMatrix()
}
