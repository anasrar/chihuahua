package t32

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/anasrar/binarium"
	"github.com/anasrar/chihuahua/pkg/utils"
)

type DMAID uint8

const (
	DMAIDReference DMAID = 3
	DMAIDEnd       DMAID = 7
)

func (d DMAID) String() string {
	switch d {
	case DMAIDReference:
		return "REF"
	case DMAIDEnd:
		return "END"
	default:
		return fmt.Sprintf("DMA_ID(%d)", uint8(d))
	}
}

type DMATag struct {
	Word0 uint64 `json:"word_0"`
	Word1 uint64 `json:"word_1"`
}

func (d DMATag) QuadwordCount() uint16 {
	return uint16(d.Word0 & 0xffff)
}

func (d DMATag) ID() DMAID {
	return DMAID((d.Word0 >> 28) & 0x07)
}

func (d DMATag) InterruptRequest() bool {
	return ((d.Word0 >> 31) & 1) != 0
}

func (d DMATag) Address() uint32 {
	return uint32((d.Word0 >> 32) & 0x7fffffff)
}

type DMAChain struct {
	Offset    uint32        `json:"offset"`
	Transfers []DMATransfer `json:"transfers"`
	EndTag    DMATag        `json:"end_tag"`
}

type DMATransfer struct {
	Tag          DMATag          `json:"tag"`
	PacketOffset uint64          `json:"packet_offset"`
	Packet       GIFUploadPacket `json:"packet"`
	Upload       GSUpload        `json:"upload"`
	Data         []byte          `json:"data,omitempty"`
}

func ReadDMATransfer(
	stream io.ReadSeeker,
	streamSize uint64,
	baseOffset uint64,
	packetOffset uint64,
	tag DMATag,
) (DMATransfer, error) {
	transfer := DMATransfer{Tag: tag, PacketOffset: packetOffset}
	absolutePacketOffset := baseOffset + packetOffset
	if err := ValidateRegion(streamSize, absolutePacketOffset, UploadPacketSize, "GIF upload packet"); err != nil {
		return transfer, err
	}
	if err := unmarshalAt(stream, absolutePacketOffset, &transfer.Packet); err != nil {
		return transfer, fmt.Errorf("GIF upload packet: %w", err)
	}
	if err := ValidateGIFUploadPacket(transfer.Packet); err != nil {
		return transfer, err
	}

	payloadQWC := uint64(transfer.Packet.ImageTag.NLoop())
	payloadOffset := packetOffset + UploadPacketSize
	payloadSize := payloadQWC * 16
	if uint64(tag.QuadwordCount()) != 6+payloadQWC {
		return transfer, fmt.Errorf(
			"DMA QWC is %#x, expected GIF header plus payload %#x",
			tag.QuadwordCount(),
			6+payloadQWC,
		)
	}
	if err := ValidateRegion(
		streamSize,
		baseOffset+payloadOffset,
		payloadSize,
		"GIF image payload",
	); err != nil {
		return transfer, err
	}

	transfer.Upload = DecodeGSUpload(transfer.Packet, payloadOffset, payloadSize)
	transfer.Data = make([]byte, payloadSize)
	if err := utils.SeekAbsolute(stream, baseOffset+payloadOffset); err != nil {
		return transfer, err
	}
	if _, err := io.ReadFull(stream, transfer.Data); err != nil {
		return transfer, fmt.Errorf("GIF image payload: %w", err)
	}
	return transfer, nil
}

func ReadDMAChain(
	stream io.ReadSeeker,
	streamSize uint64,
	baseOffset uint64,
	chainOffset uint32,
	transferTotal uint16,
	name string,
) (DMAChain, error) {
	chain := DMAChain{
		Offset:    chainOffset,
		Transfers: make([]DMATransfer, 0, transferTotal),
	}

	chainSize := (uint64(transferTotal) + 1) * DMATagSize
	if err := ValidateRegion(
		streamSize,
		baseOffset+uint64(chainOffset),
		chainSize,
		name+" DMA chain",
	); err != nil {
		return chain, err
	}

	for i := 0; i < int(transferTotal); i++ {
		tagOffset := baseOffset + uint64(chainOffset) + uint64(i*DMATagSize)
		var tag DMATag
		if err := unmarshalAt(stream, tagOffset, &tag); err != nil {
			return chain, fmt.Errorf("%s DMA tag %d: %w", name, i, err)
		}
		if tag.ID() != DMAIDReference {
			return chain, fmt.Errorf("%s DMA tag %d has ID %d, expected REF", name, i, tag.ID())
		}

		packetOffset := uint64(chainOffset) + uint64(tag.Address())
		transfer, err := ReadDMATransfer(stream, streamSize, baseOffset, packetOffset, tag)
		if err != nil {
			return chain, fmt.Errorf("%s DMA transfer %d: %w", name, i, err)
		}
		chain.Transfers = append(chain.Transfers, transfer)
	}

	endOffset := baseOffset + uint64(chainOffset) + uint64(transferTotal)*DMATagSize
	if err := unmarshalAt(stream, endOffset, &chain.EndTag); err != nil {
		return chain, fmt.Errorf("%s DMA terminator: %w", name, err)
	}
	if chain.EndTag.ID() != DMAIDEnd {
		return chain, fmt.Errorf("%s DMA terminator has ID %d, expected END", name, chain.EndTag.ID())
	}
	return chain, nil
}

func ValidateRegion(streamSize, offset, size uint64, name string) error {
	if offset > streamSize || size > streamSize-offset {
		return fmt.Errorf(
			"%s [%#x,%#x) is outside stream size %#x",
			name,
			offset,
			offset+size,
			streamSize,
		)
	}
	return nil
}

func unmarshalAt(stream io.ReadSeeker, offset uint64, target any) error {
	if err := utils.SeekAbsolute(stream, offset); err != nil {
		return err
	}
	return binarium.UnmarshalWithReader(stream, binary.LittleEndian, target)
}
