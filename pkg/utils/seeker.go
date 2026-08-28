package utils

import (
	"fmt"
	"io"
)

func SeekerSize(stream io.ReadSeeker) (uint64, error) {
	position, err := stream.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, fmt.Errorf("seek stream end: %w", err)
	}
	if position < 0 {
		return 0, fmt.Errorf("stream returned negative size %d", position)
	}
	return uint64(position), nil
}

func SeekAbsolute(stream io.ReadSeeker, offset uint64) error {
	const maxInt64 = uint64(1<<63 - 1)
	if offset > maxInt64 {
		return fmt.Errorf("stream offset %#x exceeds int64", offset)
	}
	if _, err := stream.Seek(int64(offset), io.SeekStart); err != nil {
		return fmt.Errorf("seek stream offset %#x: %w", offset, err)
	}
	return nil
}
