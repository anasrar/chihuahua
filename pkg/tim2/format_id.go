package tim2

type FormatId uint8

const (
	FormatId16Alignment  FormatId = 0x00
	FormatId128Alignment FormatId = 0x81
)

func (self FormatId) String() string {
	result := "Unknown"
	switch self {
	case 0x00:
		result = "16 Byte Alignment"
	case 0x01:
		result = "128 Byte Alignment"
	}

	return result
}
