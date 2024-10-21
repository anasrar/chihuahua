package tm3

type Entry struct {
	Source string `json:"source"`
	Name   string `json:"name"`
	Size   uint32 `json:"size"`
	Offset uint32 `json:"offset"`
}
