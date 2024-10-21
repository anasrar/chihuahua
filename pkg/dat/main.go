package dat

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/anasrar/chihuahua/pkg/buffer"
	"github.com/anasrar/chihuahua/pkg/utils"
)

const (
	EntryTypeLength uint64 = 4
)

type Dat struct {
	Offset     uint32   `json:"offset"`
	Size       uint32   `json:"size"`
	EntryTotal uint32   `json:"entry_total"`
	Entries    []*Entry `json:"entries"`
}

func (self *Dat) unmarshal(source string, stream io.ReadWriteSeeker) error {
	if self.Size == 0 {
		size, _ := buffer.Seek(stream, 0, buffer.SeekEnd)
		self.Size = uint32(size) - self.Offset
	}

	if _, err := buffer.Seek(stream, int64(self.Offset), buffer.SeekStart); err != nil {
		return err
	}

	if _, err := buffer.ReadUint32LE(stream, &self.EntryTotal); err != nil {
		return err
	}

	for range self.EntryTotal {
		entry := Entry{
			Source: source,
			Type:   "\x00\x00\x00\x00",
			Size:   0,
			Offset: 0,
			IsNull: false,
		}

		if _, err := buffer.ReadUint32LE(stream, &entry.Offset); err != nil {
			return err
		}

		entry.IsNull = entry.Offset == 0

		if !entry.IsNull {
			entry.Offset += self.Offset
		}

		self.Entries = append(self.Entries, &entry)
	}

	for _, entry := range self.Entries {
		if _, err := buffer.ReadString(stream, &entry.Type, EntryTypeLength); err != nil {
			return err
		}
	}

	for i, entry := range self.Entries {
		if entry.IsNull {
			continue
		}

		if i == int(self.EntryTotal-1) {
			entry.Size = (self.Offset + self.Size) - entry.Offset
		} else {
			index := i + 1
			offset := uint32(0)
			for {
				nextEntry := self.Entries[index]
				if !nextEntry.IsNull {
					offset = nextEntry.Offset
					break
				}

				if index == int(self.EntryTotal-1) {
					offset = self.Size
					break
				}

				index += 1
			}

			entry.Size = offset - entry.Offset
		}
	}

	return nil
}

func (self *Dat) Pack(
	ctx context.Context,
	output string,
	onStart,
	onDone func(total uint32, current uint32, name string),
) error {
	pad := uint(math.Ceil(float64(self.EntryTotal*2+1)/8)*8) * 4

	packFile, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer packFile.Close()

	if _, err := buffer.WriteBytes(packFile, make([]byte, pad)); err != nil {
		return err
	}

	for i, entry := range self.Entries {
		if entry.IsNull {
			continue
		}

		onStart(self.EntryTotal, uint32(i+1), utils.Basename(entry.Source))
		position := uint64(0)
		if _, err := buffer.Position(packFile, &position); err != nil {
			return err
		}
		entry.Offset = uint32(position)

		entryFile, err := os.Open(entry.Source)
		if err != nil {
			return err
		}
		defer entryFile.Close()

		buf := make([]byte, entry.Size)

		if _, err := entryFile.Read(buf); err != nil {
			return err
		}

		if _, err := packFile.Write(buf); err != nil {
			return err
		}

		onDone(self.EntryTotal, uint32(i+1), utils.Basename(entry.Source))

		select {
		case <-ctx.Done():
			return fmt.Errorf("Canceled")
		default:
		}

	}

	if _, err := buffer.Seek(packFile, 0, buffer.SeekStart); err != nil {
		return err
	}

	if _, err := buffer.WriteUint32LE(packFile, self.EntryTotal); err != nil {
		return err
	}

	for _, entry := range self.Entries {
		if _, err := buffer.WriteUint32LE(packFile, entry.Offset); err != nil {
			return err
		}
	}

	for _, entry := range self.Entries {
		if _, err := buffer.WriteString(packFile, entry.Type); err != nil {
			return err
		}
	}

	return nil
}

func (self *Dat) Unpack(
	ctx context.Context,
	dir string,
	onStart,
	onDone func(total uint32, current uint32, name string),
) error {
	total := len(self.Entries)
	for i, entry := range self.Entries {
		if entry.IsNull {
			continue
		}

		normalizeType := utils.FilterUnprintableString(entry.Type)
		filename := fmt.Sprintf("%s_%03d.%s", normalizeType, i, strings.ToLower(normalizeType))

		onStart(uint32(total), uint32(i+1), filename)

		target := fmt.Sprintf("%s/%s", dir, normalizeType)
		if err := os.MkdirAll(target, os.ModePerm); err != nil {
			return err
		}

		sourceFile, err := os.Open(entry.Source)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		if _, err := sourceFile.Seek(int64(entry.Offset), io.SeekStart); err != nil {
			return err
		}

		b := make([]byte, entry.Size)

		if _, err := sourceFile.Read(b); err != nil {
			return err
		}

		unpackFile, err := os.OpenFile(fmt.Sprintf("%s/%s", target, filename), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer unpackFile.Close()

		if _, err := unpackFile.Write(b); err != nil {
			return err
		}

		onDone(uint32(total), uint32(i+1), filename)

		select {
		case <-ctx.Done():
			return fmt.Errorf("Canceled")
		default:
		}
	}

	return nil
}

func (self *Dat) AddNullEntry() {
	self.Entries = append(
		self.Entries,
		&Entry{
			Source: "",
			Type:   "\x00\x00\x00\x00",
			Size:   0,
			Offset: 0,
			IsNull: true,
		},
	)

	self.EntryTotal += 1
}

func (self *Dat) AddEntryFromPathWithType(
	source string,
	t string,
) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	self.Entries = append(
		self.Entries,
		&Entry{
			Source: source,
			Type:   t,
			Size:   uint32(size),
			Offset: 0,
			IsNull: size == 0,
		},
	)

	self.EntryTotal += 1

	return nil
}

func New() *Dat {
	return &Dat{
		Offset:     0,
		Size:       0,
		EntryTotal: 0,
		Entries:    []*Entry{},
	}
}

func FromPathWithOffsetSize(dat *Dat, filePath string, offset uint32, size uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	dat.Offset = offset
	dat.Size = size
	return dat.unmarshal(filePath, file)
}

func FromPath(dat *Dat, filePath string) error {
	return FromPathWithOffsetSize(dat, filePath, 0, 0)
}
