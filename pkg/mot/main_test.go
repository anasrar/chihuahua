package mot

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

type testReadWriteSeeker struct {
	*bytes.Reader
}

func (self *testReadWriteSeeker) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func TestUnmarshalConstantHalfHermiteAndSentinel(t *testing.T) {
	data := make([]byte, 64)
	copy(data[0:4], "mtb3")
	binary.LittleEndian.PutUint16(data[4:6], 2)
	data[6] = 3
	data[7] = 1

	data[8] = 0xFF
	data[9] = 4
	binary.LittleEndian.PutUint16(data[10:12], 1)
	binary.LittleEndian.PutUint32(data[16:20], math.Float32bits(0.5))

	data[20] = 0
	data[21] = ChannelTranslationX
	binary.LittleEndian.PutUint16(data[22:24], 2)
	binary.LittleEndian.PutUint32(data[28:32], 44)
	binary.LittleEndian.PutUint16(data[46:48], 0x5E00) // position delta = 1
	copy(data[56:60], []byte{0, 0, 0, 0})
	copy(data[60:64], []byte{1, 2, 0, 0})

	data[32] = 0x7F
	data[33] = ChannelSentinel
	binary.LittleEndian.PutUint16(data[34:36], 0xFFFF)

	motion := New()
	motion.Size = uint32(len(data))
	require.NoError(t, motion.unmarshal(&testReadWriteSeeker{Reader: bytes.NewReader(data)}))

	require.Equal(t, uint8(1), motion.Loop)
	require.Len(t, motion.Records, 3)
	require.Equal(t, uint8(0), motion.Records[0].Target)
	require.Equal(t, float32(0.5), motion.Records[0].ConstantValue)
	require.Equal(t, uint32(44), motion.Records[1].DataOffset)
	require.Len(t, motion.Records[1].Curves, 2)
	require.Equal(t, []float32{0, 2}, motion.Records[1].Quantize(2))
	require.True(t, motion.Records[2].IsSentinel)
}
