package mdb

type VertexBuffer struct {
	Vertices [][3]float32 `json:"vertices"`
	Indices  [][3]int16   `json:"indices"`
	Normals  [][3]float32 `json:"normals"`
	Uvs      [][2]float32 `json:"uvs"`
	Joints   [][4]uint8   `json:"joints"`
	Weights  [][4]float32 `json:"weights"`
	Material uint16       `json:"material"`
	// TODO: vertex color
}
