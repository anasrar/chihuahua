package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/anasrar/chihuahua/pkg/tim3"
	"github.com/anasrar/chihuahua/pkg/utils"
)

func convert(pngPath string) error {
	pngFile, err := os.Open(pngPath)
	if err != nil {
		return err
	}

	img, err := png.Decode(pngFile)
	if err != nil {
		return err
	}

	imgPaletted, ok := img.(*image.Paletted)
	if !ok {
		return fmt.Errorf("PNG is not in indexed mode")
	}

	if len(imgPaletted.Palette) > 256 {
		return fmt.Errorf("PNG colors exceeds the maximum allowable limit of 256")
	}

	output := filepath.Join(
		utils.ParentDirectory(pngPath),
		fmt.Sprintf("TM3_%s.tm3", utils.BasenameWithoutExt(pngPath)),
	)

	timFile, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer timFile.Close()

	if err := tim3.ImagePalettedToFile(imgPaletted, timFile); err != nil {
		return err
	}
	return nil
}
