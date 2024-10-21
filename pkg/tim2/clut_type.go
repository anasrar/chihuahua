package tim2

import "fmt"

type ClutType uint8

func (self ClutType) String() string {
	format := "Unknown"
	switch self & 0x1F {
	case 0:
		format = "None"
	case 1:
		format = ImageType16BitColor.String()
	case 2:
		format = ImageType24BitColor.String()
	case 3:
		format = ImageType32BitColor.String()
	}

	storage := "Unknown"
	switch self >> 6 & 0x1 {
	case 0:
		storage = "CSM1"
	case 1:
		storage = "CSM2"
	}

	return fmt.Sprintf("Format: %s, Storage: %s", format, storage)
}
