package t32

type RGBA struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

func ExpandPS2Alpha(alpha uint8) uint8 {
	if alpha >= 0x80 {
		return 0xff
	}
	return alpha * 2
}

func ClutPermutation(index int) int {
	return (index & 0xe7) | ((index & 0x08) << 1) | ((index & 0x10) >> 1)
}
