package t32

import (
	"errors"
	"fmt"
	"image"
	"image/color"
)

func (c *Container) AddImage(img *image.Paletted) (int, error) {
	if c == nil {
		return -1, ErrContainerIsNil
	}
	width, height, indices, colors, err := palettedData(img)
	if err != nil {
		return -1, err
	}
	if len(c.Textures) == int(^uint32(0)) {
		return -1, errors.New("texture total exceeds uint32")
	}
	if c.totalPaletteCount() == int(^uint16(0)) {
		return -1, errors.New("palette total exceeds uint16")
	}

	texture := Texture{
		Index:    uint32(len(c.Textures)),
		Width:    width,
		Height:   height,
		Indices:  indices,
		Palettes: []Palette{{Colors: colors}},
	}
	if err := configureTexture(&texture); err != nil {
		return -1, err
	}
	c.Textures = append(c.Textures, texture)
	c.syncCounts()
	return len(c.Textures) - 1, nil
}

func (c *Container) AddPaletteAtImage(textureIndex int, img *image.Paletted) (int, error) {
	texture, err := c.textureAt(textureIndex)
	if err != nil {
		return -1, err
	}
	if img == nil {
		return -1, errors.New("palette image is nil")
	}
	bounds := img.Bounds()
	if bounds.Dx() != texture.Width || bounds.Dy() != texture.Height {
		return -1, fmt.Errorf(
			"palette image size is %dx%d, expected %dx%d",
			bounds.Dx(),
			bounds.Dy(),
			texture.Width,
			texture.Height,
		)
	}
	colors, err := paletteColors(img.Palette)
	if err != nil {
		return -1, err
	}
	if maxIndex(texture.Indices) >= len(colors) {
		return -1, fmt.Errorf(
			"palette has %d colors but texture uses index %d",
			len(colors),
			maxIndex(texture.Indices),
		)
	}
	if c.totalPaletteCount() == int(^uint16(0)) {
		return -1, errors.New("palette total exceeds uint16")
	}

	texture.Palettes = append(texture.Palettes, Palette{Colors: colors})
	if err := configureTexture(texture); err != nil {
		texture.Palettes = texture.Palettes[:len(texture.Palettes)-1]
		return -1, err
	}
	c.syncCounts()
	return len(texture.Palettes) - 1, nil
}

func (c *Container) RemoveImageAt(textureIndex int) error {
	if _, err := c.textureAt(textureIndex); err != nil {
		return err
	}
	copy(c.Textures[textureIndex:], c.Textures[textureIndex+1:])
	c.Textures = c.Textures[:len(c.Textures)-1]
	for i := range c.Textures {
		c.Textures[i].Index = uint32(i)
	}
	c.syncCounts()
	return nil
}

func (c *Container) RemovePaletteAtImage(textureIndex, paletteIndex int) error {
	texture, _, err := c.textureAndPalette(textureIndex, paletteIndex)
	if err != nil {
		return err
	}
	if len(texture.Palettes) == 1 {
		return errors.New("an image must retain at least one palette")
	}
	copy(texture.Palettes[paletteIndex:], texture.Palettes[paletteIndex+1:])
	texture.Palettes = texture.Palettes[:len(texture.Palettes)-1]
	if err := configureTexture(texture); err != nil {
		return err
	}
	c.syncCounts()
	return nil
}

func (c *Container) ImageAt(textureIndex int) (*image.Paletted, error) {
	return c.PaletteAtImage(textureIndex, 0)
}

func (c *Container) PaletteAtImage(textureIndex, paletteIndex int) (*image.Paletted, error) {
	texture, palette, err := c.textureAndPalette(textureIndex, paletteIndex)
	if err != nil {
		return nil, err
	}
	colors := make(color.Palette, len(palette.Colors))
	for i, entry := range palette.Colors {
		colors[i] = color.NRGBA{R: entry.R, G: entry.G, B: entry.B, A: entry.A}
	}
	img := image.NewPaletted(image.Rect(0, 0, texture.Width, texture.Height), colors)
	copy(img.Pix, texture.Indices)
	return img, nil
}

func palettedData(img *image.Paletted) (int, int, []byte, []RGBA, error) {
	if img == nil {
		return 0, 0, nil, nil, errors.New("image is nil")
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if _, err := powerOfTwoExponent(width); err != nil {
		return 0, 0, nil, nil, fmt.Errorf("image width: %w", err)
	}
	if _, err := powerOfTwoExponent(height); err != nil {
		return 0, 0, nil, nil, fmt.Errorf("image height: %w", err)
	}
	colors, err := paletteColors(img.Palette)
	if err != nil {
		return 0, 0, nil, nil, err
	}
	indices := make([]byte, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := img.ColorIndexAt(bounds.Min.X+x, bounds.Min.Y+y)
			if int(index) >= len(colors) {
				return 0, 0, nil, nil, fmt.Errorf(
					"pixel (%d,%d) uses palette index %d outside %d colors",
					x,
					y,
					index,
					len(colors),
				)
			}
			indices[y*width+x] = index
		}
	}
	return width, height, indices, colors, nil
}

func paletteColors(palette color.Palette) ([]RGBA, error) {
	if len(palette) == 0 || len(palette) > 256 {
		return nil, fmt.Errorf("palette has %d colors, expected 1..256", len(palette))
	}
	colors := make([]RGBA, len(palette))
	for i := range palette {
		converted := color.NRGBAModel.Convert(palette[i]).(color.NRGBA)
		colors[i] = RGBA{R: converted.R, G: converted.G, B: converted.B, A: converted.A}
	}
	return colors, nil
}

func configureTexture(texture *Texture) error {
	if texture == nil {
		return errors.New("texture is nil")
	}
	tw, err := powerOfTwoExponent(texture.Width)
	if err != nil {
		return fmt.Errorf("texture width: %w", err)
	}
	th, err := powerOfTwoExponent(texture.Height)
	if err != nil {
		return fmt.Errorf("texture height: %w", err)
	}
	if len(texture.Indices) != texture.Width*texture.Height {
		return fmt.Errorf(
			"texture has %d indices, expected %d",
			len(texture.Indices),
			texture.Width*texture.Height,
		)
	}
	if len(texture.Palettes) == 0 {
		return errors.New("texture has no palettes")
	}

	limit := 16
	if maxIndex(texture.Indices) >= 16 {
		limit = 256
	}
	for i := range texture.Palettes {
		colorTotal := len(texture.Palettes[i].Colors)
		if colorTotal == 0 || colorTotal > 256 {
			return fmt.Errorf("palette %d has %d colors, expected 1..256", i, colorTotal)
		}
		if maxIndex(texture.Indices) >= colorTotal {
			return fmt.Errorf(
				"palette %d has %d colors but texture uses index %d",
				i,
				colorTotal,
				maxIndex(texture.Indices),
			)
		}
		if colorTotal > 16 {
			limit = 256
		}
	}

	psm := PSMT4
	if limit == 256 {
		psm = PSMT8
	}
	tbw := texture.Width / 64
	pageHeight := 128
	if psm == PSMT8 {
		pageHeight = 64
	}
	if texture.Height > pageHeight && tbw < 2 {
		tbw = 2
	}
	if tbw > 0x3f {
		return fmt.Errorf("texture buffer width %d exceeds GS TEX0", tbw)
	}
	texture.GSTex0.PSM = psm
	texture.GSTex0.TW = tw
	texture.GSTex0.TH = th
	texture.GSTex0.TBW = uint8(tbw)
	texture.GSTex0.CPSM = PSMCT32
	texture.GSTex0.CSM = 0
	texture.GSTex0.CSA = 0
	texture.GSTex0.CLD = 1
	return nil
}

func powerOfTwoExponent(value int) (uint8, error) {
	if value <= 0 || value > 1024 || value&(value-1) != 0 {
		return 0, fmt.Errorf("%d is not a power of two in the range 1..1024", value)
	}
	exponent := uint8(0)
	for 1<<exponent < value {
		exponent++
	}
	return exponent, nil
}

func maxIndex(indices []byte) int {
	maximum := 0
	for _, index := range indices {
		if int(index) > maximum {
			maximum = int(index)
		}
	}
	return maximum
}

func (c *Container) totalPaletteCount() int {
	total := 0
	if c == nil {
		return total
	}
	for i := range c.Textures {
		total += len(c.Textures[i].Palettes)
	}
	return total
}

func (c *Container) syncCounts() {
	if c == nil {
		return
	}
	c.TextureTotal = uint32(len(c.Textures))
	c.PalettePacketTotal = uint16(c.totalPaletteCount())
	c.PaletteTotal = 0
	if len(c.Textures) > 0 {
		c.PaletteTotal = uint32(len(c.Textures[0].Palettes))
	}
}
