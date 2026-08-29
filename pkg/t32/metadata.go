package t32

type MetadataPalette struct {
	Index uint64 `json:"index"`
	Path  string `json:"path"`
}

type MetadataTexture struct {
	Index    uint64            `json:"index"`
	Palettes []MetadataPalette `json:"palettes"`
}

type Metadata struct {
	Textures []MetadataTexture `json:"textures"`
}
