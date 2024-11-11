package mot

import (
	"fmt"

	"github.com/anasrar/chihuahua/pkg/utils"
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
	p0 *[2]float32,
	p1 *[2]float32,
) error {
	if index == int(self.CurveTotal-1) {
		return fmt.Errorf("Index is last curve")
	}

	position := utils.PgHalfFloat32FromUint16(self.Position)
	positionDelta := utils.PgHalfFloat32FromUint16(self.PositionDelta)

	currentCurve := self.Curves[index]
	nextCurve := self.Curves[index+1]

	currentPosition := (position + positionDelta*float32(currentCurve.ControlPoint))
	nextPosition := (position + positionDelta*float32(nextCurve.ControlPoint))

	*p0 = [2]float32{0, currentPosition}
	*p1 = [2]float32{float32(nextCurve.FrameDelta), nextPosition}

	return nil
}

func (self *Record) CurveToHermite(
	index int,
	p0 *[2]float32,
	m0 *float32,
	p1 *[2]float32,
	m1 *float32,
) error {
	if index == int(self.CurveTotal-1) {
		return fmt.Errorf("Index is last curve")
	}

	position := utils.PgHalfFloat32FromUint16(self.Position)
	positionDelta := utils.PgHalfFloat32FromUint16(self.PositionDelta)
	tangent1 := utils.PgHalfFloat32FromUint16(self.Tangent1)
	tangentDelta1 := utils.PgHalfFloat32FromUint16(self.TangentDelta0)
	tangent0 := utils.PgHalfFloat32FromUint16(self.Tangent1)
	tangentDelta0 := utils.PgHalfFloat32FromUint16(self.TangentDelta0)

	currentCurve := self.Curves[index]
	nextCurve := self.Curves[index+1]

	currentPosition := (position + positionDelta*float32(currentCurve.ControlPoint))
	*m0 = (tangent1 + tangentDelta1*float32(currentCurve.ControlTangent1))
	nextPosition := (position + positionDelta*float32(nextCurve.ControlPoint))
	*m1 = (tangent0 + tangentDelta0*float32(currentCurve.ControlTangent0))

	*p0 = [2]float32{0, currentPosition}
	*p1 = [2]float32{float32(nextCurve.FrameDelta), nextPosition}

	return nil
}

func (self *Record) CurveToBezier(
	index int,
	p0 *[2]float32,
	p1 *[2]float32,
	p2 *[2]float32,
	p3 *[2]float32,
	tangentScale float32,
) error {
	m0 := float32(0)
	m1 := float32(0)

	if err := self.CurveToHermite(index, p0, &m0, p3, &m1); err != nil {
		return err
	}

	m0t := m0 / 3
	m1t := m1 / 3

	*p1 = [2]float32{p0[0] + m0t, p0[1] + m0t}
	*p2 = [2]float32{p3[0] - m1t, p3[1] - m1t}

	return nil
}

func (self *Record) QuantizeLinear(frameTotal uint16) []float32 {
	result := []float32{}
	frame := uint16(0)

	for i := 0; i < int(self.CurveTotal-1); i++ {
		nextCurve := self.Curves[i+1]

		p0 := [2]float32{0, 0}
		p1 := [2]float32{0, 0}

		if err := self.CurveToLinear(i, &p0, &p1); err != nil {
			continue
		}

		total := float32(nextCurve.FrameDelta)

		for j := float32(0); j < total; j++ {
			result = append(
				result,
				utils.Lerp(p0[1], p1[1], j/total),
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

func (self *Record) QuantizeHermite(frameTotal uint16) []float32 {
	result := []float32{}
	frame := uint16(0)

	for i := 0; i < int(self.CurveTotal-1); i++ {
		nextCurve := self.Curves[i+1]

		p0 := [2]float32{0, 0}
		m0 := float32(0)
		p1 := [2]float32{0, 0}
		m1 := float32(0)

		if err := self.CurveToHermite(i, &p0, &m0, &p1, &m1); err != nil {
			continue
		}

		total := float32(nextCurve.FrameDelta)

		for j := float32(0); j < total; j++ {
			t := j / total
			t2 := t * t
			t3 := t2 * t

			result = append(
				result,
				(2*t3-3*t2+1)*p0[1]+(t3-2*t2+t)*m0+(-2*t3+3*t2)*p1[1]+(t3-t2)*m1,
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
