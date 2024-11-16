package tim2

import "image/color"

type Picture struct {
	TotalSize      uint32    `json:"total_size"` // NOTE: total size is sum of clut size, image size, and picture header size
	ClutSize       uint32    `json:"clut_size"`
	ImageSize      uint32    `json:"image_size"`
	HeaderSize     uint16    `json:"header_size"`
	ClutColors     uint16    `json:"clut_colors"`
	PictureFormat  uint8     `json:"picture_format"`
	MipMapTextures uint8     `json:"mipmap_textures"`
	ClutType       ClutType  `json:"clut_type"`
	ImageType      ImageType `json:"image_type"`
	ImageWidth     uint16    `json:"image_width"`
	ImageHeight    uint16    `json:"image_height"`
	GsTex0         uint64    `json:"gs_tex0"`     // TODO: destruct bit
	GsTex1         uint64    `json:"gs_tex1"`     // TODO: destruct bit, NOTE: always 608
	GsRegs         uint32    `json:"gs_regs"`     // TODO: destruct bit, NOTE: always 0
	GsTexClut      uint32    `json:"ge_tex_clut"` // TODO: destruct bit, NOTE: always 0
	ImageData      []byte
	ClutData       []*color.RGBA
}
