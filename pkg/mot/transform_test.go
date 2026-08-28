package mot

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuantizeTransformsMapsChannelsToBones(t *testing.T) {
	motion := &Mot{
		FrameTotal: 2,
		Records: []*Record{
			{
				Target:        0,
				Channel:       4,
				ConstantValue: 0.5,
			},
			{
				Target:        1,
				Channel:       ChannelTranslationX,
				PositionDelta: 0x5e00,
				Curves:        []*Curve{{ControlPoint: 3}},
			},
			{
				Target:        1,
				Channel:       ChannelRotationZ,
				PositionDelta: 0x5e00,
				Curves:        []*Curve{{ControlPoint: 2}},
			},
		},
	}

	frames, err := motion.QuantizeTransforms(2)
	require.NoError(t, err)
	require.Len(t, frames, 2)
	for _, frame := range frames {
		require.Equal(t, [3]float32{0, 0.5, 0}, frame[0].Rotation)
		require.Equal(t, [3]float32{3, 0, 0}, frame[1].Translation)
		require.Equal(t, [3]float32{0, 0, 2}, frame[1].Rotation)
	}
}
