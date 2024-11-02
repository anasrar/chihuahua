package oms

type Entry struct {
	Name        string     `json:"name"`
	Translation [3]float32 `json:"translation"`
}

func NewEntry(name string, translation [3]float32) *Entry {
	return &Entry{
		Name:        name,
		Translation: translation,
	}
}
