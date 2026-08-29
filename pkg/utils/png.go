package utils

import (
	"image"
	"image/png"
	"os"
)

func WritePNG(outputPath string, imageData image.Image) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := png.Encode(file, imageData); err != nil {
		return err
	}
	return nil
}
