package main

import (
	"bytes"
	"image/png"

	"github.com/anasrar/chihuahua/pkg/tim3"
	"github.com/anasrar/chihuahua/pkg/tm3"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Texture struct {
	Texture rl.Texture2D
}

func NewTexture(texture rl.Texture2D) *Texture {
	return &Texture{
		Texture: texture,
	}
}

func LoadTextures(datPath string, offset uint32, size uint32) error {
	tm := tm3.New()
	if err := tm3.FromPathWithOffsetSize(tm, datPath, offset, size); err != nil {
		return err
	}

	for _, index := range textureIndices {
		rl.UnloadTexture(textures[index].Texture)
		delete(textures, index)
	}

	textureIndices = []int{}

	for i, entry := range tm.Entries {
		tim := tim3.New()
		if err := tim3.FromPathWithOffset(tim, datPath, entry.Offset); err != nil {
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

	for _, model := range models {
		if texture, found := textures[model.Texture]; found {
			rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, texture.Texture)
		} else {
			rl.SetMaterialTexture(model.Model.Materials, rl.MapDiffuse, textureDefault)
		}
	}
	return nil
}
