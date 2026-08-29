package t32

import "fmt"

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

func CompressPS2Alpha(alpha uint8) uint8 {
	if alpha == 0xff {
		return 0x80
	}
	return uint8((uint16(alpha) + 1) / 2)
}

func ClutPermutation(index int) int {
	return (index & 0xe7) | ((index & 0x08) << 1) | ((index & 0x10) >> 1)
}

func EncodeCSM1RGBA32(palette []RGBA) ([]byte, error) {
	if len(palette) != 256 {
		return nil, fmt.Errorf("CSM1 palette has %d entries, expected 256", len(palette))
	}
	payload := make([]byte, 256*4)
	for logicalIndex, entry := range palette {
		storedIndex := ClutPermutation(logicalIndex)
		position := storedIndex * 4
		payload[position+0] = entry.R
		payload[position+1] = entry.G
		payload[position+2] = entry.B
		payload[position+3] = CompressPS2Alpha(entry.A)
	}
	return payload, nil
}
