package bone

type Bone struct {
	Name   string  `json:"name"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Z      float32 `json:"z"`
	Parent int16   `json:"parent"`
}

func New(name string, x, y, z float32, parent int16) *Bone {
	return &Bone{
		Name:   name,
		X:      x,
		Y:      y,
		Z:      z,
		Parent: parent,
	}
}
