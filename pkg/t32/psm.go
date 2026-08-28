package t32

import "fmt"

type PixelStorageMode uint8

const (
	PSMCT32 PixelStorageMode = 0x00
	PSMT8   PixelStorageMode = 0x13
	PSMT4   PixelStorageMode = 0x14
)

func (p PixelStorageMode) String() string {
	switch p {
	case PSMCT32:
		return "PSMCT32"
	case PSMT8:
		return "PSMT8"
	case PSMT4:
		return "PSMT4"
	default:
		return fmt.Sprintf("PSM(%#x)", uint8(p))
	}
}
