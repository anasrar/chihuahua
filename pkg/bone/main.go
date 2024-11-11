package bone

type Bone struct {
	Name        string     `json:"name"`
	Translation [3]float32 `json:"translation"`
	Rotation    [3]float32 `json:"rotation"`
	Parent      int16      `json:"parent"`
}

func New(
	name string,
	translationX,
	translationY,
	translationZ float32,
	rotationX,
	rotationY,
	rotationZ float32,
	parent int16,
) *Bone {
	return &Bone{
		Name:        name,
		Translation: [3]float32{translationX, translationY, translationZ},
		Rotation:    [3]float32{rotationX, rotationY, rotationZ},
		Parent:      parent,
	}
}
