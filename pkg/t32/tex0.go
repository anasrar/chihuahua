package t32

type GSTex0 struct {
	Raw  uint64           `json:"raw"`
	TBP0 uint16           `json:"texture_base_pointer"`
	TBW  uint8            `json:"texture_buffer_width"`
	PSM  PixelStorageMode `json:"pixel_storage_mode"`
	TW   uint8            `json:"texture_width_exponent"`
	TH   uint8            `json:"texture_height_exponent"`
	TCC  uint8            `json:"texture_color_component"`
	TFX  uint8            `json:"texture_function"`
	CBP  uint16           `json:"clut_base_pointer"`
	CPSM PixelStorageMode `json:"clut_pixel_storage_mode"`
	CSM  uint8            `json:"clut_storage_mode"`
	CSA  uint8            `json:"clut_entry_offset"`
	CLD  uint8            `json:"clut_load_control"`
}

func (t GSTex0) Width() int {
	return 1 << t.TW
}

func (t GSTex0) Height() int {
	return 1 << t.TH
}

func DecodeGSTex0(raw uint64) GSTex0 {
	return GSTex0{
		Raw:  raw,
		TBP0: uint16(raw & 0x3fff),
		TBW:  uint8((raw >> 14) & 0x3f),
		PSM:  PixelStorageMode((raw >> 20) & 0x3f),
		TW:   uint8((raw >> 26) & 0x0f),
		TH:   uint8((raw >> 30) & 0x0f),
		TCC:  uint8((raw >> 34) & 0x01),
		TFX:  uint8((raw >> 35) & 0x03),
		CBP:  uint16((raw >> 37) & 0x3fff),
		CPSM: PixelStorageMode((raw >> 51) & 0x0f),
		CSM:  uint8((raw >> 55) & 0x01),
		CSA:  uint8((raw >> 56) & 0x1f),
		CLD:  uint8((raw >> 61) & 0x07),
	}
}

func EncodeGSTex0(t GSTex0) uint64 {
	return uint64(t.TBP0&0x3fff) |
		uint64(t.TBW&0x3f)<<14 |
		uint64(t.PSM&0x3f)<<20 |
		uint64(t.TW&0x0f)<<26 |
		uint64(t.TH&0x0f)<<30 |
		uint64(t.TCC&0x01)<<34 |
		uint64(t.TFX&0x03)<<35 |
		uint64(t.CBP&0x3fff)<<37 |
		uint64(t.CPSM&0x0f)<<51 |
		uint64(t.CSM&0x01)<<55 |
		uint64(t.CSA&0x1f)<<56 |
		uint64(t.CLD&0x07)<<61
}
