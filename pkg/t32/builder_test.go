package t32

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func testPaletted(width, height, colorTotal int, seed uint8) *image.Paletted {
	palette := make(color.Palette, colorTotal)
	for i := range palette {
		palette[i] = color.NRGBA{
			R: uint8(i*7) + seed,
			G: uint8(i*3) + seed,
			B: uint8(i*11) + seed,
			A: 0xff,
		}
	}
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	for i := range img.Pix {
		img.Pix[i] = uint8(i % colorTotal)
	}
	return img
}

func TestNestedPaletteBuilderRoundTrip(t *testing.T) {
	container := New()
	textureIndex, err := container.AddImage(testPaletted(128, 64, 16, 1))
	if err != nil {
		t.Fatal(err)
	}
	if textureIndex != 0 {
		t.Fatalf("first texture index is %d", textureIndex)
	}
	paletteIndex, err := container.AddPaletteAtImage(0, testPaletted(128, 64, 16, 9))
	if err != nil {
		t.Fatal(err)
	}
	if paletteIndex != 1 {
		t.Fatalf("second palette index is %d", paletteIndex)
	}
	if _, err := container.AddImage(testPaletted(128, 64, 32, 17)); err != nil {
		t.Fatal(err)
	}

	var encoded bytes.Buffer
	if err := container.Write(&encoded); err != nil {
		t.Fatal(err)
	}
	prefix := bytes.Repeat([]byte{0xa5}, 37)
	stream := bytes.NewReader(append(prefix, encoded.Bytes()...))
	decoded := New()
	if err := FromStreamWithOffset(decoded, stream, uint32(len(prefix))); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Textures) != 2 {
		t.Fatalf("decoded %d textures", len(decoded.Textures))
	}
	if len(decoded.Textures[0].Palettes) != 2 || len(decoded.Textures[1].Palettes) != 1 {
		t.Fatalf(
			"decoded nested palette totals %d and %d",
			len(decoded.Textures[0].Palettes),
			len(decoded.Textures[1].Palettes),
		)
	}
	if decoded.Textures[0].GSTex0.PSM != PSMT4 {
		t.Fatalf("first texture format is %s", decoded.Textures[0].GSTex0.PSM)
	}
	if decoded.Textures[1].GSTex0.PSM != PSMT8 {
		t.Fatalf("second texture format is %s", decoded.Textures[1].GSTex0.PSM)
	}

	paletteImage, err := decoded.PaletteAtImage(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if paletteImage.Bounds() != image.Rect(0, 0, 128, 64) {
		t.Fatalf("palette image bounds are %v", paletteImage.Bounds())
	}
	fromStream, err := DecodePalettedImage(bytes.NewReader(encoded.Bytes()), 0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(paletteImage.Pix, fromStream.Pix) {
		t.Fatal("indexed pixels differ")
	}
}

func TestRemoveImageAndPalette(t *testing.T) {
	container := New()
	if _, err := container.AddImage(testPaletted(128, 64, 16, 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := container.AddPaletteAtImage(0, testPaletted(128, 64, 16, 2)); err != nil {
		t.Fatal(err)
	}
	if _, err := container.AddImage(testPaletted(128, 64, 16, 3)); err != nil {
		t.Fatal(err)
	}
	if err := container.RemovePaletteAtImage(0, 1); err != nil {
		t.Fatal(err)
	}
	if err := container.RemovePaletteAtImage(0, 0); err == nil {
		t.Fatal("removed the last palette")
	}
	if err := container.RemoveImageAt(0); err != nil {
		t.Fatal(err)
	}
	if len(container.Textures) != 1 || container.Textures[0].Index != 0 {
		t.Fatal("image removal did not reindex textures")
	}
}

func TestAddPaletteDimensionMismatch(t *testing.T) {
	container := New()
	if _, err := container.AddImage(testPaletted(128, 64, 16, 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := container.AddPaletteAtImage(0, testPaletted(64, 64, 16, 2)); err == nil {
		t.Fatal("accepted a palette image with different dimensions")
	}
}

func TestCompactPSMT4TexturesRoundTrip(t *testing.T) {
	dimensions := [][2]int{
		{128, 128},
		{64, 64},
		{128, 64},
		{64, 32},
		{128, 256},
		{128, 64},
		{128, 128},
		{64, 64},
		{128, 256},
		{128, 64},
		{64, 64},
		{32, 32},
		{64, 64},
		{32, 32},
		{128, 64},
		{64, 32},
	}
	expectedBases := []uint16{0, 32, 40, 56, 60, 124, 140, 172, 180, 244, 260, 268, 270, 278, 280, 296}
	expectedWidths := []uint8{2, 1, 2, 1, 2, 2, 2, 1, 2, 2, 1, 0, 1, 0, 2, 1}

	container := New()
	for i, size := range dimensions {
		if _, err := container.AddImage(testPaletted(size[0], size[1], 16, uint8(i))); err != nil {
			t.Fatalf("add image %d: %v", i, err)
		}
	}

	var encoded bytes.Buffer
	if err := container.Write(&encoded); err != nil {
		t.Fatal(err)
	}
	if container.ImageUploadWidth != 64 || container.ImageUploadHeight != 320 {
		t.Fatalf("image upload is %dx%d, expected 64x320", container.ImageUploadWidth, container.ImageUploadHeight)
	}
	for i := range container.Textures {
		if container.Textures[i].GSTex0.TBP0 != expectedBases[i] {
			t.Fatalf("texture %d base is %d, expected %d", i, container.Textures[i].GSTex0.TBP0, expectedBases[i])
		}
		if container.Textures[i].GSTex0.TBW != expectedWidths[i] {
			t.Fatalf("texture %d buffer width is %d, expected %d", i, container.Textures[i].GSTex0.TBW, expectedWidths[i])
		}
	}

	decoded := New()
	if err := FromStreamWithOffset(decoded, bytes.NewReader(encoded.Bytes()), 0); err != nil {
		t.Fatal(err)
	}
	for i := range decoded.Textures {
		if !bytes.Equal(decoded.Textures[i].Indices, container.Textures[i].Indices) {
			t.Fatalf("texture %d indices differ after round trip", i)
		}
	}
}
