package buffer

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/x448/float16"
)

type SeekMode int

const (
	SeekStart   SeekMode = io.SeekStart
	SeekEnd     SeekMode = io.SeekEnd
	SeekCurrent SeekMode = io.SeekCurrent
)

func check(stream io.ReadWriteSeeker) error {
	if stream == nil {
		return fmt.Errorf("Stream is nil")
	}

	return nil
}

func Seek(stream io.ReadWriteSeeker, offset int64, whence SeekMode) (uint64, error) {
	if err := check(stream); err != nil {
		return 0, err
	}

	position, err := stream.Seek(offset, int(whence))
	if err != nil {
		return 0, err
	}
	return uint64(position), nil
}

func Position(stream io.ReadWriteSeeker, position *uint64) (uint64, error) {
	if err := check(stream); err != nil {
		return 0, err
	}

	pos, err := stream.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	*position = uint64(pos)

	return uint64(pos), nil
}

func ReadBytes(stream io.ReadWriteSeeker, b []byte) (uint64, error) {
	if err := check(stream); err != nil {
		return 0, err
	}

	position, err := stream.Read(b)
	if err != nil {
		return 0, err
	}
	return uint64(position), nil
}

func ReadString(stream io.ReadWriteSeeker, str *string, size uint64) (uint64, error) {
	bytes := make([]byte, size)
	position, err := ReadBytes(stream, bytes)
	if err != nil {
		return 0, err
	}

	*str = string(bytes)

	return position, nil
}

func ReadNumberFactory(stream io.ReadWriteSeeker, n any, endian binary.ByteOrder) (uint64, error) {
	if err := check(stream); err != nil {
		return 0, err
	}

	err := binary.Read(stream, endian, n)
	if err != nil {
		return 0, err
	}

	position := uint64(0)
	if _, err := Position(stream, &position); err != nil {
		return 0, err
	}

	return position, nil
}

func ReadUint8(stream io.ReadWriteSeeker, n *uint8) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.LittleEndian)
}

func ReadInt8(stream io.ReadWriteSeeker, n *int8) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.LittleEndian)
}

func ReadUint16LE(stream io.ReadWriteSeeker, n *uint16) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.LittleEndian)
}

func ReadUint16BE(stream io.ReadWriteSeeker, n *uint16) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.BigEndian)
}

func ReadInt16LE(stream io.ReadWriteSeeker, n *int16) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.LittleEndian)
}

func ReadInt16BE(stream io.ReadWriteSeeker, n *int16) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.BigEndian)
}

func ReadUint32LE(stream io.ReadWriteSeeker, n *uint32) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.LittleEndian)
}

func ReadUint32BE(stream io.ReadWriteSeeker, n *uint32) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.BigEndian)
}

func ReadInt32LE(stream io.ReadWriteSeeker, n *int32) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.LittleEndian)
}

func ReadInt32BE(stream io.ReadWriteSeeker, n *int32) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.BigEndian)
}

func ReadUint64LE(stream io.ReadWriteSeeker, n *uint64) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.LittleEndian)
}

func ReadUint64BE(stream io.ReadWriteSeeker, n *uint64) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.BigEndian)
}

func ReadInt64LE(stream io.ReadWriteSeeker, n *int64) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.LittleEndian)
}

func ReadInt64BE(stream io.ReadWriteSeeker, n *int64) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.BigEndian)
}

func ReadFloat16LE(stream io.ReadWriteSeeker, n *float16.Float16) (uint64, error) {
	var short uint16
	position, err := ReadUint16LE(stream, &short)
	if err != nil {
		return 0, err
	}
	*n = float16.Frombits(short)
	return position, nil
}

func ReadFloat16BE(stream io.ReadWriteSeeker, n *float16.Float16) (uint64, error) {
	var short uint16
	position, err := ReadUint16BE(stream, &short)
	if err != nil {
		return 0, err
	}
	*n = float16.Frombits(short)
	return position, nil
}

func ReadFloat32LE(stream io.ReadWriteSeeker, n *float32) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.LittleEndian)
}

func ReadFloat32BE(stream io.ReadWriteSeeker, n *float32) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.BigEndian)
}

func ReadFloat64LE(stream io.ReadWriteSeeker, n *float64) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.LittleEndian)
}

func ReadFloat64BE(stream io.ReadWriteSeeker, n *float64) (uint64, error) {
	return ReadNumberFactory(stream, n, binary.BigEndian)
}

func WriteBytes(stream io.ReadWriteSeeker, b []byte) (uint64, error) {
	if err := check(stream); err != nil {
		return 0, err
	}

	if _, err := stream.Write(b); err != nil {
		return 0, err
	}

	position := uint64(0)
	if _, err := Position(stream, &position); err != nil {
		return 0, err
	}

	return position, nil
}

func WriteString(stream io.ReadWriteSeeker, str string) (uint64, error) {
	return WriteBytes(stream, []byte(str))
}

func WriteNumberFactory(stream io.ReadWriteSeeker, n any, endian binary.ByteOrder) (uint64, error) {
	if err := check(stream); err != nil {
		return 0, err
	}

	err := binary.Write(stream, endian, n)
	if err != nil {
		return 0, err
	}

	position := uint64(0)
	if _, err := Position(stream, &position); err != nil {
		return 0, err
	}

	return position, nil
}

func WriteUint8(stream io.ReadWriteSeeker, n uint8) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.LittleEndian)
}

func WriteInt8(stream io.ReadWriteSeeker, n int8) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.LittleEndian)
}

func WriteUint16LE(stream io.ReadWriteSeeker, n uint16) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.LittleEndian)
}

func WriteUint16BE(stream io.ReadWriteSeeker, n uint16) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.BigEndian)
}

func WriteInt16LE(stream io.ReadWriteSeeker, n int16) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.LittleEndian)
}

func WriteInt16BE(stream io.ReadWriteSeeker, n int16) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.BigEndian)
}

func WriteUint32LE(stream io.ReadWriteSeeker, n uint32) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.LittleEndian)
}

func WriteUint32BE(stream io.ReadWriteSeeker, n uint32) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.BigEndian)
}

func WriteInt32LE(stream io.ReadWriteSeeker, n int32) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.LittleEndian)
}

func WriteInt32BE(stream io.ReadWriteSeeker, n int32) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.BigEndian)
}

func WriteUint64LE(stream io.ReadWriteSeeker, n uint64) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.LittleEndian)
}

func WriteUint64BE(stream io.ReadWriteSeeker, n uint64) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.BigEndian)
}

func WriteInt64LE(stream io.ReadWriteSeeker, n int64) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.LittleEndian)
}

func WriteInt64BE(stream io.ReadWriteSeeker, n int64) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.BigEndian)
}

func Writeloat16LE(stream io.ReadWriteSeeker, n float16.Float16) (uint64, error) {
	return WriteUint16LE(stream, n.Bits())
}

func WriteFloat16BE(stream io.ReadWriteSeeker, n float16.Float16) (uint64, error) {
	return WriteUint16BE(stream, n.Bits())
}

func WriteFloat32LE(stream io.ReadWriteSeeker, n float32) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.LittleEndian)
}

func WriteFloat32BE(stream io.ReadWriteSeeker, n float32) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.BigEndian)
}

func WriteFloat64LE(stream io.ReadWriteSeeker, n float64) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.LittleEndian)
}

func WriteFloat64BE(stream io.ReadWriteSeeker, n float64) (uint64, error) {
	return WriteNumberFactory(stream, n, binary.BigEndian)
}
