package mot

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCurveToHermiteUsesOutgoingAndNextIncomingTangents(t *testing.T) {
	record := &Record{
		CurveTotal:    2,
		PositionDelta: 0x5e00, // 1
		Tangent0:      0x5e00, // 1
		TangentDelta0: 0x6000, // 2
		Tangent1:      0x6200, // 4
		TangentDelta1: 0x6400, // 8
		Curves: []*Curve{
			{ControlPoint: 2, ControlTangent1: 3},
			{FrameDelta: 4, ControlPoint: 6, ControlTangent0: 5},
		},
	}

	var p0, p1 [2]float32
	var m0, m1 float32
	require.NoError(t, record.CurveToHermite(0, &p0, &m0, &p1, &m1))
	require.Equal(t, [2]float32{0, 2}, p0)
	require.Equal(t, float32(28), m0)
	require.Equal(t, [2]float32{4, 6}, p1)
	require.Equal(t, float32(11), m1)
}

func TestQuantizeHermiteUsesRelativeFramesAndHoldsEndpoints(t *testing.T) {
	record := &Record{
		PositionDelta: 0x5e00, // 1
		Curves: []*Curve{
			{FrameDelta: 2, ControlPoint: 0},
			{FrameDelta: 4, ControlPoint: 10},
		},
	}

	require.InDeltaSlice(t,
		[]float32{0, 0, 0, 1.5625, 5, 8.4375, 10, 10},
		record.QuantizeHermite(8),
		0.00001,
	)
}

func TestQuantizeHermiteOneKeyAndNoKeys(t *testing.T) {
	oneKey := &Record{
		PositionDelta: 0x5e00,
		Curves:        []*Curve{{FrameDelta: 2, ControlPoint: 7}},
	}
	require.Equal(t, []float32{7, 7, 7, 7}, oneKey.QuantizeHermite(4))
	require.Equal(t, []float32{0, 0, 0}, (&Record{}).QuantizeHermite(3))
}

func TestQuantizeConstant(t *testing.T) {
	record := &Record{Channel: 4, ConstantValue: 0.5}
	require.Equal(t, []float32{0.5, 0.5, 0.5}, record.Quantize(3))
}
