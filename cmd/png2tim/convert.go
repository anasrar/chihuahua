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

func convert(pngPath string, bpp uint) error {
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
	colorTotal := len(imgPaletted.Palette)

	if colorTotal > 256 {
		return fmt.Errorf("PNG colors exceeds the maximum allowable limit of 256")
	}

	if bpp == 4 && colorTotal > 16 {
		return fmt.Errorf("PNG colors greater than 16 can not use 4 bit perpixel")
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

	if err := tim3.ImagePalettedToFile(imgPaletted, bpp, timFile); err != nil {
		return err
	}
	return nil
}
