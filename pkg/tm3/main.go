package tm3

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/anasrar/chihuahua/pkg/buffer"
	"github.com/anasrar/chihuahua/pkg/utils"
)

const (
	Signature       uint32 = 0x00334D54
	EntryNameLength uint64 = 0x8
)

type Tm3 struct {
	Offset     uint32   `json:"offset"`
	Size       uint32   `json:"size"`
	EntryTotal uint32   `json:"entry_total"`
	Entries    []*Entry `json:"entries"`
}

func (self *Tm3) unmarshal(source string, stream io.ReadWriteSeeker) error {
	if self.Size == 0 {
		size, _ := buffer.Seek(stream, 0, buffer.SeekEnd)
		self.Size = uint32(size) - self.Offset
	}

	if _, err := buffer.Seek(stream, int64(self.Offset), buffer.SeekStart); err != nil {
		return err
	}

	signature := uint32(0)
	if _, err := buffer.ReadUint32LE(stream, &signature); err != nil {
		return err
	}

	if signature != Signature {
		return fmt.Errorf("TM3 signature not match")
	}

	if _, err := buffer.ReadUint32LE(stream, &self.EntryTotal); err != nil {
		return err
	}

	if _, err := buffer.Seek(stream, 8, buffer.SeekCurrent); err != nil {
		return err
	}

	for range self.EntryTotal {
		entry := Entry{
			Source: source,
			Name:   "\x00\x00\x00\x00\x00\x00\x00\x00",
			Size:   0,
			Offset: 0,
		}

		if _, err := buffer.ReadUint32LE(stream, &entry.Offset); err != nil {
			return err
		}

		entry.Offset += self.Offset

		self.Entries = append(self.Entries, &entry)
	}

	if self.EntryTotal&0x1 == 1 {
		if _, err := buffer.Seek(stream, 4, buffer.SeekCurrent); err != nil {
			return err
		}
	}

	for _, entry := range self.Entries {
		if _, err := buffer.ReadString(stream, &entry.Name, EntryNameLength); err != nil {
			return err
		}
	}

	for i, entry := range self.Entries {

		if i == int(self.EntryTotal-1) {
			entry.Size = (self.Offset + self.Size) - entry.Offset
		} else {
			entry.Size = self.Entries[i+1].Offset - entry.Offset
		}
	}

	return nil
}

func pad(entryTotal uint32) uint32 {
	// NOTE: entry total should even
	if entryTotal&0x1 == 1 {
		entryTotal += 1
	}

	result := uint32(16)     // NOTE: signature(4) + entry total (4) + unknown (8)
	result += entryTotal * 4 // NOTE: offset relative to signature (uint32)
	result += entryTotal * 8 // NOTE: name (char[8])

	if result < 128 {
		return 128
	}

	result = uint32(math.Ceil(float64(result)/64) * 64)

	return result
}

func (self *Tm3) Pack(
	ctx context.Context,
	output string,
	onStart,
	onDone func(total uint32, current uint32, name string),
) error {
	p := pad(self.EntryTotal)

	packFile, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer packFile.Close()

	if _, err := buffer.WriteBytes(packFile, make([]byte, p)); err != nil {
		return err
	}

	for i, entry := range self.Entries {
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

	if _, err := buffer.WriteUint32LE(packFile, Signature); err != nil {
		return err
	}

	if _, err := buffer.WriteUint32LE(packFile, self.EntryTotal); err != nil {
		return err
	}

	// NOTE: unknown padding
	{
		if _, err := buffer.WriteUint32LE(packFile, 4); err != nil {
			return err
		}
		if _, err := buffer.WriteUint32LE(packFile, 0); err != nil {
			return err
		}
	}

	for _, entry := range self.Entries {
		if _, err := buffer.WriteUint32LE(packFile, entry.Offset); err != nil {
			return err
		}
	}

	if self.EntryTotal&0x1 == 1 {
		if _, err := buffer.Seek(packFile, 4, buffer.SeekCurrent); err != nil {
			return err
		}
	}

	for _, entry := range self.Entries {
		if _, err := buffer.WriteString(packFile, entry.Name); err != nil {
			return err
		}
	}

	return nil
}

func (self *Tm3) Unpack(
	ctx context.Context,
	dir string,
	onStart,
	onDone func(total uint32, current uint32, name string),
) error {
	total := len(self.Entries)
	for i, entry := range self.Entries {
		normalizeName := utils.FilterUnprintableString(entry.Name)
		filename := fmt.Sprintf("%s_%03d.tm3", normalizeName, i)

		onStart(uint32(total), uint32(i+1), filename)

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

		unpackFile, err := os.OpenFile(fmt.Sprintf("%s/%s", dir, filename), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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

func (self *Tm3) AddEntryFromPathWithName(
	source string,
	name string,
) error {
	if len(name) > 8 {
		name = name[:8]
	} else if len(name) != 8 {
		d := 8 - len(name)
		for range d {
			name += "\x00"
		}
	}

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
			Name:   name,
			Size:   uint32(size),
			Offset: 0,
		},
	)

	self.EntryTotal += 1

	return nil
}

func New() *Tm3 {
	return &Tm3{
		Offset:     0,
		Size:       0,
		EntryTotal: 0,
		Entries:    []*Entry{},
	}
}

func FromPathWithOffsetSize(dat *Tm3, filePath string, offset uint32, size uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	dat.Offset = offset
	dat.Size = size
	return dat.unmarshal(filePath, file)
}

func FromPath(dat *Tm3, filePath string) error {
	return FromPathWithOffsetSize(dat, filePath, 0, 0)
}
