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

func DrawModelSkeleton(model *Model, animated bool) {
	pose := model.BindPose
	color := rl.Green
	if animated {
		pose = model.CurrentPose
		color = rl.Blue
	}
	if len(pose) == 0 {
		return
	}

	for boneIndex, transform := range pose {
		rl.DrawCube(transform.Translation, .02, .02, .02, color)
		parent := int(model.Bones[boneIndex].Parent)
		if parent >= 0 && parent < len(pose) {
			rl.DrawLine3D(transform.Translation, pose[parent].Translation, color)
		}
	}
}
