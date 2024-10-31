package mot

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/x448/float16"
)

const (
	TangentScale = float32(0.01)
)

type Record struct {
	IsNull             bool   `json:"is_null"`
	Target             uint8  `json:"target"`
	Channel            uint8  `json:"channel"`
	CurveTotal         uint16 `json:"curve_total"`
	UseGlobalTransform uint32 `json:"use_global_transform"`

	Position      uint16   `json:"position"`
	PositionDelta uint16   `json:"position_delta"`
	Tangent0      uint16   `json:"tangent_0"`
	TangentDelta0 uint16   `json:"tangent_delta_0"`
	Tangent1      uint16   `json:"tangent_1"`
	TangentDelta1 uint16   `json:"tangent_delta_1"`
	Curves        []*Curve `json:"curves"`
}

func (self *Record) CurveToLinear(
	index int,
	p0 *rl.Vector2,
	p1 *rl.Vector2,
) error {
	if index == int(self.CurveTotal-1) {
		return fmt.Errorf("Index is last curve")
	}

	position := float16.Frombits(self.Position).Float32()
	positionDelta := float16.Frombits(self.PositionDelta).Float32()

	currentCurve := self.Curves[index]
	nextCurve := self.Curves[index+1]

	currentPosition := (position + positionDelta*float32(currentCurve.ControlPoint))
	nextPosition := (position + positionDelta*float32(nextCurve.ControlPoint))

	*p0 = rl.NewVector2(0, currentPosition)
	*p1 = rl.NewVector2(float32(nextCurve.FrameDelta), nextPosition)

	return nil
}

func (self *Record) CurveToHermite(
	index int,
	p0 *rl.Vector2,
	m0 *float32,
	p1 *rl.Vector2,
	m1 *float32,
) error {
	if index == int(self.CurveTotal-1) {
		return fmt.Errorf("Index is last curve")
	}

	position := float16.Frombits(self.Position).Float32()
	positionDelta := float16.Frombits(self.PositionDelta).Float32()
	tangent1 := float16.Frombits(self.Tangent1).Float32()
	tangentDelta1 := float16.Frombits(self.TangentDelta0).Float32()
	tangent0 := float16.Frombits(self.Tangent1).Float32()
	tangentDelta0 := float16.Frombits(self.TangentDelta0).Float32()

	currentCurve := self.Curves[index]
	nextCurve := self.Curves[index+1]

	currentPosition := (position + positionDelta*float32(currentCurve.ControlPoint))
	*m0 = (tangent1 + tangentDelta1*float32(currentCurve.ControlTangent1)) * TangentScale
	nextPosition := (position + positionDelta*float32(nextCurve.ControlPoint))
	*m1 = (tangent0 + tangentDelta0*float32(currentCurve.ControlTangent0)) * TangentScale

	*p0 = rl.NewVector2(0, currentPosition)
	*p1 = rl.NewVector2(float32(nextCurve.FrameDelta), nextPosition)

	return nil
}

func (self *Record) CurveToBezier(
	index int,
	p0 *rl.Vector2,
	p1 *rl.Vector2,
	p2 *rl.Vector2,
	p3 *rl.Vector2,
) error {
	m0 := float32(0)
	m1 := float32(0)

	if err := self.CurveToHermite(index, p0, &m0, p3, &m1); err != nil {
		return err
	}

	*p1 = rl.Vector2AddValue(rl.NewVector2(p1.X, p1.Y), m0/3)
	*p2 = rl.Vector2SubtractValue(rl.NewVector2(p3.X, p3.Y), m1/3)

	return nil
}

func (self *Record) QuantizeLinear(frameTotal uint16) []float32 {
	result := []float32{}
	frame := uint16(0)

	for i := 0; i < int(self.CurveTotal-1); i++ {
		nextCurve := self.Curves[i+1]

		p0 := rl.Vector2Zero()
		p1 := rl.Vector2Zero()

		if err := self.CurveToLinear(i, &p0, &p1); err != nil {
			continue
		}

		total := float32(nextCurve.FrameDelta)

		for j := float32(0); j < total; j++ {
			result = append(
				result,
				rl.Lerp(p0.Y, p1.Y, j/total),
			)
			frame++
		}

	}

	d := frameTotal - frame

	if d > 0 {
		lastFrame := result[len(result)-1]

		for i := uint16(0); i < d; i++ {
			result = append(result, lastFrame)
		}
	}

	return result
}

func NewRecord() *Record {
	return &Record{
		Target:             0,
		Channel:            0,
		CurveTotal:         0,
		UseGlobalTransform: 0,
		Position:           0,
		PositionDelta:      0,
		Tangent0:           0,
		TangentDelta0:      0,
		Tangent1:           0,
		TangentDelta1:      0,
		Curves:             []*Curve{},
	}
}
