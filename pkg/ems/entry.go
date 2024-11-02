package ems

type Entry struct {
	Translation [3]float32 `json:"translation"`
}

func NewEntry(translation [3]float32) *Entry {
	return &Entry{
		Translation: translation,
	}
}
