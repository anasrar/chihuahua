package baseline

type File struct {
	Source string `json:"source"`
	Offset int64  `json:"offset"`
	Length int64  `json:"length"`
}
