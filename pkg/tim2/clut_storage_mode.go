package tim2

type ClutStorageMode uint8

const (
	ClutStorageMode1 ClutStorageMode = 0x00
	ClutStorageMode2 ClutStorageMode = 0x01
)

func (self ClutStorageMode) String() string {
	switch self {
	case ClutStorageMode1:
		return "CSM1"
	case ClutStorageMode2:
		return "CSM2"
	default:
		return "CSM1"
	}
}
