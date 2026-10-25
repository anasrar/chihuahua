package afs

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/anasrar/binarium"
	"github.com/anasrar/chihuahua/next/pkg/baseline"
)

func unmarshal(source string, stream io.ReadSeeker, offset int64, length int64) (*Container, error) {
	if _, err := stream.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	container := &Container{
		File: &baseline.File{
			Source: source,
			Offset: offset,
			Length: length,
		},
		Header:        &Header{},
		Entries:       []*Entry{},
		AttributeData: &AttributeData{},
	}
	if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, container); err != nil {
		return nil, err
	}

	if container.Header.Signature() != Signature {
		return nil, fmt.Errorf("AFS: signature 0x%X does not match", container.Header.Signature())
	}

	switch container.Header.Version() {
	case 0x00:
	case 0x20:
	default:
		return nil, fmt.Errorf("AFS: version 0x%X is not valid", uint8(container.Header.Version()))
	}

	for range container.EntryTotal {
		entry := Entry{}
		if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, &entry); err != nil {
			return nil, err
		}
		container.Entries = append(container.Entries, &entry)
	}

	entryBlockStartOffset := uint32(0)
	entryBlockEndOffset := uint32(0)
	for _, entry := range container.Entries {
		entryOffset := entry.Offset

		entry.Source = source
		entry.AbsoluteOffset = uint32(offset) + entryOffset
		isNull := entryOffset == 0
		entry.IsNull = isNull

		if isNull {
			continue
		}

		if entryBlockStartOffset == 0 {
			entryBlockStartOffset = entryOffset
		}
		entryBlockEndOffset = entryOffset + entry.Length
	}

	currentOffset, err := stream.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	alignment := MinEntryBlockAlignmentLength
	endInfoBlockOffset := uint32(currentOffset) + AttributeInfoLength

	for endInfoBlockOffset+alignment < entryBlockStartOffset {
		alignment <<= 1
	}

	if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, container.AttributeData); err != nil {
		return nil, err
	}

	isAttributeInfoValid := checkIsAttributeInfoValid(container.AttributeData.Offset, container.AttributeData.Length, uint32(length), container.EntryTotal, entryBlockEndOffset)

	attributesInfo := AttributesInfoInfoAtStart

	if isAttributeInfoValid {
		attributesInfo = AttributesInfoInfoAtStart
	} else {
		if _, err := stream.Seek(offset+int64(entryBlockStartOffset-AttributeInfoLength), io.SeekStart); err != nil {
			return nil, err
		}

		if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, &container.AttributeData); err != nil {
			return nil, err
		}

		isAttributeInfoValid = checkIsAttributeInfoValid(container.AttributeData.Offset, container.AttributeData.Length, uint32(length), container.EntryTotal, entryBlockEndOffset)

		if isAttributeInfoValid {
			attributesInfo = AttributesInfoInfoAtEnd
		}
	}

	if attributesInfo != AttributesInfoNoAttribute {
		if _, err := stream.Seek(offset+int64(container.AttributeData.Offset), io.SeekStart); err != nil {
			return nil, err
		}

		for _, entry := range container.Entries {
			if entry.IsNull {
				if _, err := stream.Seek(offset+int64(AttributeElementLength), io.SeekCurrent); err != nil {
					return nil, err
				}
				continue
			} else {
				description := EntryDescription{}
				if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, &description); err != nil {
					return nil, err
				}
				entry.Description = &description
			}
		}

	} else {
		for i, entry := range container.Entries {
			name := fmt.Sprintf("%0*d", 32, i)

			var buf [32]byte
			copy(buf[:], name)

			entry.Description.Name = buf
		}
	}

	return container, nil
}

func FromPath(source string, offset int64, size int64) (*Container, error) {
	file, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return unmarshal(source, file, offset, size)
}

func checkIsAttributeInfoValid(attributesOffset, attributesSize, afsFileSize, entryTotal, dataBlockEndOffset uint32) bool {
	if attributesOffset == 0 {
		return false
	}
	if attributesSize == 0 {
		return false
	}

	if attributesSize > afsFileSize-dataBlockEndOffset {
		return false
	}
	if attributesSize < entryTotal*AttributeElementLength {
		return false
	}
	if attributesOffset < dataBlockEndOffset {
		return false
	}
	if attributesOffset > afsFileSize-attributesSize {
		return false
	}

	return true
}
