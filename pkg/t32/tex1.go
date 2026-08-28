package t32

type GSTex1 struct {
	Raw  uint32 `json:"raw"`
	LCM  uint8  `json:"lod_calculation_method"`
	MXL  uint8  `json:"maximum_mipmap_level"`
	MMAG uint8  `json:"magnification_filter"`
	MMIN uint8  `json:"minification_filter"`
	MTBA uint8  `json:"mipmap_base_address_method"`
	L    uint8  `json:"lod_parameter"`
}

func DecodeGSTex1(raw uint32) GSTex1 {
	return GSTex1{
		Raw:  raw,
		LCM:  uint8(raw & 0x01),
		MXL:  uint8((raw >> 2) & 0x07),
		MMAG: uint8((raw >> 5) & 0x01),
		MMIN: uint8((raw >> 6) & 0x07),
		MTBA: uint8((raw >> 9) & 0x01),
		L:    uint8((raw >> 19) & 0x03),
	}
}
