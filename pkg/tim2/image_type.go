package tim2

type ImageType uint8

const (
	ImageType16BitColor  ImageType = 0x01
	ImageType24BitColor  ImageType = 0x02
	ImageType32BitColor  ImageType = 0x03
	ImageType4BitTexture ImageType = 0x04
	ImageType8BitTexture ImageType = 0x05
)

func (self ImageType) String() string {
	switch self {
	case ImageType16BitColor:
		return "16 Bit Color"
	case ImageType24BitColor:
		return "24 Bit Color"
	case ImageType32BitColor:
		return "32 Bit Color"
	case ImageType4BitTexture:
		return "4 Bit Texture"
	case ImageType8BitTexture:
		return "8 Bit Texture"
	default:
		return "Unknown"
	}
}
