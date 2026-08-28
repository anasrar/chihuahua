package t32

type GSUpload struct {
	DestinationBase        uint16            `json:"destination_base"`
	DestinationBufferWidth uint8             `json:"destination_buffer_width"`
	DestinationPSM         PixelStorageMode  `json:"destination_pixel_storage_mode"`
	DestinationX           uint16            `json:"destination_x"`
	DestinationY           uint16            `json:"destination_y"`
	Width                  uint16            `json:"width"`
	Height                 uint16            `json:"height"`
	Direction              TransferDirection `json:"direction"`
	DataOffset             uint64            `json:"data_offset"`
	DataSize               uint64            `json:"data_size"`
}

func DecodeGSUpload(packet GIFUploadPacket, dataOffset, dataSize uint64) GSUpload {
	bitbltbuf := packet.BITBLTBUF.Value
	trxpos := packet.TRXPOS.Value
	trxreg := packet.TRXREG.Value
	trxdir := packet.TRXDIR.Value
	return GSUpload{
		DestinationBase:        uint16((bitbltbuf >> 32) & 0x3fff),
		DestinationBufferWidth: uint8((bitbltbuf >> 48) & 0x3f),
		DestinationPSM:         PixelStorageMode((bitbltbuf >> 56) & 0x3f),
		DestinationX:           uint16((trxpos >> 32) & 0x7ff),
		DestinationY:           uint16((trxpos >> 48) & 0x7ff),
		Width:                  uint16(trxreg & 0xfff),
		Height:                 uint16((trxreg >> 32) & 0xfff),
		Direction:              TransferDirection(trxdir & 0x03),
		DataOffset:             dataOffset,
		DataSize:               dataSize,
	}
}
