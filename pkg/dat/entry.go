package dat

type Entry struct {
	Source string `json:"source"`
	Type   string `json:"type"`
	Size   uint32 `json:"size"`
	Offset uint32 `json:"offset"`
	IsNull bool   `json:"is_null"`
}
