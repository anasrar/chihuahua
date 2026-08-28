package mot

import "fmt"

const (
	ChannelTranslationX uint8 = 16
	ChannelTranslationY uint8 = 17
	ChannelTranslationZ uint8 = 18
	ChannelRotationX    uint8 = 19
	ChannelRotationY    uint8 = 20
	ChannelRotationZ    uint8 = 21
)

type Transform struct {
	Translation [3]float32 `json:"translation"`
	Rotation    [3]float32 `json:"rotation"`
}

func (self *Mot) QuantizeTransforms(boneTotal int) ([][]Transform, error) {
	if boneTotal < 0 {
		return nil, fmt.Errorf("bone total cannot be negative: %d", boneTotal)
	}

	frames := make([][]Transform, int(self.FrameTotal))
	for frame := range frames {
		frames[frame] = make([]Transform, boneTotal)
	}

	for _, record := range self.Records {
		if record.IsSentinel || record.Component() > 5 {
			continue
		}

		target := int(record.Target)
		if target >= boneTotal {
			return nil, fmt.Errorf("MOT target bone %d is outside bone total %d", target, boneTotal)
		}

		values := record.Quantize(self.FrameTotal)
		for frame, value := range values {
			switch record.Component() {
			case 0:
				frames[frame][target].Translation[0] = value
			case 1:
				frames[frame][target].Translation[1] = value
			case 2:
				frames[frame][target].Translation[2] = value
			case 3:
				frames[frame][target].Rotation[0] = value
			case 4:
				frames[frame][target].Rotation[1] = value
			case 5:
				frames[frame][target].Rotation[2] = value
			}
		}
	}

	return frames, nil
}
