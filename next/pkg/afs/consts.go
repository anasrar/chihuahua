package afs

const (
	Signature                    uint32 = 0x534641
	MinEntryBlockAlignmentLength uint32 = 0x800
	AttributeInfoLength          uint32 = 0x8
	AttributeElementLength       uint32 = 0x30
	DateLayoutFormat             string = "2006-01-02 15:04:05"
	HeaderLength                 uint32 = 0x8
	EntryInfoElementLength       uint32 = 0x8
	AlignmentLength              uint32 = 0x800
)

type Version uint8

const (
	Version00 Version = 0x00
	Version20 Version = 0x20
)

type AttributesInfo int

const (
	AttributesInfoNoAttribute AttributesInfo = iota
	AttributesInfoInfoAtStart
	AttributesInfoInfoAtEnd
)
