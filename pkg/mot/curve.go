package mot

type Curve struct {
	FrameDelta      uint8 `json:"frame_delta"`
	ControlPoint    uint8 `json:"control_point"`
	ControlTangent0 uint8 `json:"control_tangent_0"`
	ControlTangent1 uint8 `json:"control_tangent_1"`
}

func NewCurve() *Curve {
	return &Curve{
		FrameDelta:      0,
		ControlPoint:    0,
		ControlTangent0: 0,
		ControlTangent1: 0,
	}
}
