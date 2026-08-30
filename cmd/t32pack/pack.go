package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/anasrar/chihuahua/pkg/t32"
	"github.com/anasrar/chihuahua/pkg/utils"
)

func pack(
	ctx context.Context,
	metadataPath string,
	onStart,
	onDone func(total uint32, current uint32, name string),
) error {
	metadataBuf, err := os.ReadFile(metadataPath)
	if err != nil {
		return err
	}

	var m t32.Metadata
	if err := json.Unmarshal(metadataBuf, &m); err != nil {
		return err
	}

	t := t32.New()

	parentDir := utils.ParentDirectory(metadataPath)

	total := 0
	current := 0
	for _, textures := range m.Textures {
		for range textures.Palettes {
			total += 1
		}
	}

	for textureIndex, textures := range m.Textures {
		for palettedIndex, palette := range textures.Palettes {
			current += 1
			filename := filepath.Base(palette.Path)
			onStart(uint32(total), uint32(current), filename)

			pngFile, err := os.Open(palette.Path)
			if err != nil {
				return err
			}
			defer pngFile.Close()

			img, err := png.Decode(pngFile)
			if err != nil {
				return err
			}

			palettedImg, isPalleted := img.(*image.Paletted)
			if !isPalleted {
				return fmt.Errorf("PNG not a palleted: %s", palette.Path)
			}

			if palettedIndex == 0 {
				t.AddImage(palettedImg)
			} else {
				t.AddPaletteAtImage(textureIndex, palettedImg)
			}

			onDone(uint32(total), uint32(current), filename)
		}
	}

	packPath := filepath.Join(parentDir, "OUTPUT.t32")
	packFile, err := os.OpenFile(packPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer packFile.Close()

	if err := t.Write(packFile); err != nil {
		return err
	}

	return nil
}
