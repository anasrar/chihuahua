package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/anasrar/chihuahua/pkg/t32"
	"github.com/anasrar/chihuahua/pkg/utils"
)

func unpack(
	ctx context.Context,
	t32Path string,
	onStart,
	onDone func(total uint32, current uint32, name string),
) error {
	file, err := os.Open(t32Path)
	if err != nil {
		return err
	}

	container := t32.New()
	if err := t32.FromStreamWithOffset(container, file, 0); err != nil {
		return err
	}

	outputFilesDirPath := filepath.Join(utils.ParentDirectory(t32Path), fmt.Sprintf("UNPACK_%s", utils.Basename(t32Path)), "FILES")
	if err := os.MkdirAll(outputFilesDirPath, os.ModePerm); err != nil {
		return err
	}

	total := 0
	for _, texture := range container.Textures {
		if !texture.IsIndexed() {
			continue
		}

		total += len(texture.Palettes)
	}

	outputMetadataPath := filepath.Join(utils.ParentDirectory(t32Path), fmt.Sprintf("UNPACK_%s", utils.Basename(t32Path)), "METADATA.json")

	md := t32.Metadata{
		Textures: make([]t32.MetadataTexture, 0),
	}

	current := 0
	for textureIndex, texture := range container.Textures {
		if !texture.IsIndexed() {
			log.Printf(
				"skip texture %d: pixel format %#x is not indexed",
				textureIndex,
				uint8(texture.GSTex0.PSM),
			)
			continue
		}

		tx := t32.MetadataTexture{
			Index:    uint64(texture.Index),
			Palettes: make([]t32.MetadataPalette, 0),
		}

		for paletteIndex := range texture.Palettes {
			filename := fmt.Sprintf("texture_%03d_palette_%03d.png", textureIndex, paletteIndex)
			outputPath := filepath.Join(outputFilesDirPath, filename)

			current += 1
			onStart(uint32(total), uint32(current), filename)

			decoded, err := container.DecodeTexture(textureIndex, paletteIndex)
			if err != nil {
				return err
			}

			if err := utils.WritePNG(outputPath, decoded.Image()); err != nil {
				log.Fatal(err)
			}
			onDone(uint32(total), uint32(current), filename)

			tx.Palettes = append(tx.Palettes, t32.MetadataPalette{
				Index: uint64(paletteIndex),
				Path:  outputPath,
			})
		}

		md.Textures = append(md.Textures, tx)
	}

	buf, err := json.MarshalIndent(md, "", "\t")
	if err != nil {
		return err
	}

	metadataFile, err := os.OpenFile(outputMetadataPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer metadataFile.Close()

	if _, err := metadataFile.Write(buf); err != nil {
		return err
	}

	return nil
}
