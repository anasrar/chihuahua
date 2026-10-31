package tim2

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
		Picture: &Picture{},
	}
	if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, container.Header); err != nil {
		return nil, err
	}

	if container.Header.Signature != Signature {
		return nil, fmt.Errorf("TIM2: signature 0x%X does not match", container.Header.Signature)
	}

	if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, container.Picture); err != nil {
		return nil, err
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
