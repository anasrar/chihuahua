package tim2

import "fmt"

type ClutType uint8

func (self ClutType) Format() ImageType {
	switch self & 0x1F {
	case 0:
		return ImageTypeNone
	case 1:
		return ImageType16BitColor
	case 2:
		return ImageType24BitColor
	case 3:
		return ImageType32BitColor
	default:
		return ImageTypeNone
	}
}

func (self ClutType) CompoundMode() bool {
	return ((self >> 7) & 0x1) == 1
}

func (self ClutType) StorageMode() ClutStorageMode {
	switch self >> 6 & 0x1 {
	case 0:
		return ClutStorageMode1
	case 1:
		return ClutStorageMode2
	default:
		return ClutStorageMode1
	}
}

func (self ClutType) String() string {
	format := self.Format()

	storage := self.StorageMode()

	return fmt.Sprintf("Format: %s, Compound Mode: %v, Storage: %s", format, self.CompoundMode(), storage)
}
