package t32

import (
	"errors"
	"sync"
)

const (
	HeaderSize       = 0x20
	DescriptorSize   = 0x10
	TextureLinkSize  = 0x08
	ImageDMAOffset   = HeaderSize + DescriptorSize
	DMATagSize       = 0x10
	UploadPacketSize = 0x60

	GIFRegisterAD        = 0x0e
	GIFRegisterBITBLTBUF = 0x50
	GIFRegisterTRXPOS    = 0x51
	GIFRegisterTRXREG    = 0x52
	GIFRegisterTRXDIR    = 0x53
)

var (
	ErrNotIndexed     = errors.New("texture is not PSMT8 or PSMT4 indexed data")
	ErrUnsupportedT32 = errors.New("unsupported T32 variant")
	ErrContainerIsNil = errors.New("Container is nil")
	gsMutex           sync.Mutex
)
