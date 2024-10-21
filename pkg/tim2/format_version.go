package tim2

type FormatVersion uint8

const (
	FormatVersionReserved            FormatVersion = 0x00
	FormatVersionPrivateIncompatible FormatVersion = 0x80
	FormatVersionPrivateCompatible   FormatVersion = 0xC0
)

func (self FormatVersion) String() string {
	result := "Unknown"
	if self < 0x80 {
		result = "Reserved"
	} else if self < 0xC0 {
		result = "Private Incompatible"
	} else if self <= 0xFF {
		result = "Private Compatible"
	}

	return result
}
