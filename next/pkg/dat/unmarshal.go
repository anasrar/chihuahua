package dat

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
		Header:  &Header{},
		Entries: []*Entry{},
	}
	if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, container.Header); err != nil {
		return nil, err
	}

	for _, offset := range container.Header.Offsets {
		if int64(offset) > length {
			return nil, fmt.Errorf("DAT: offset exceeded size")
		}
	}

	for i := range container.Header.EntryTotal {
		entry := &Entry{}
		container.Entries = append(container.Entries, entry)

		entry.Source = source
		entryOffset := container.Header.Offsets[i]
		entry.IsNull = entryOffset == 0
		if entry.IsNull {
			continue
		}

		entry.Type = string(container.Header.Types[i][:])
		entry.Offset = entryOffset
		entry.AbsoluteOffset = uint32(offset) + entryOffset
	}

	for i, entry := range container.Entries {
		if entry.IsNull {
			continue
		}

		end := uint32(length)

		for j := i + 1; j < len(container.Entries); j++ {
			if container.Entries[j].IsNull {
				continue
			}

			end = container.Entries[j].Offset
			break
		}

		if end < container.Entries[i].Offset || end > uint32(length) {
			return nil, fmt.Errorf(
				"DAT: invalid offset for entry %d: offset %#x, end %#x",
				i, container.Entries[i].Offset, end,
			)
		}

		container.Entries[i].Length = end - container.Entries[i].Offset
	}

	return container, nil
}

func FromPath(source string, offset int64, length int64) (*Container, error) {
	file, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return unmarshal(source, file, offset, length)
}
