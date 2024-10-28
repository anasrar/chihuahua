package mdb

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"

	"github.com/anasrar/chihuahua/pkg/bone"
	"github.com/anasrar/chihuahua/pkg/buffer"
)

const (
	Signature uint32  = 0x0062646D
	RoomScale float32 = 0.01
)

type Mdb struct {
	Offset            uint32          `json:"offset"`
	BoneTotal         uint16          `json:"bone_total"`
	Bones             []*bone.Bone    `json:"bones"`
	VertexBufferTotal uint16          `json:"vertex_buffer_total"`
	VertexBuffers     []*VertexBuffer `json:"vertex_buffers"`
}

func (self *Mdb) unmarshal(stream io.ReadWriteSeeker) error {
	if _, err := buffer.Seek(stream, int64(self.Offset), buffer.SeekStart); err != nil {
		return err
	}

	signature := uint32(0)
	if _, err := buffer.ReadUint32LE(stream, &signature); err != nil {
		return err
	}

	if signature != Signature {
		return fmt.Errorf("MDB signature not match")
	}

	boneOffset := uint32(0)
	if _, err := buffer.ReadUint32LE(stream, &boneOffset); err != nil {
		return err
	}
	boneOffset += self.Offset

	if _, err := buffer.ReadUint16LE(stream, &self.BoneTotal); err != nil {
		return err
	}

	if _, err := buffer.ReadUint16LE(stream, &self.VertexBufferTotal); err != nil {
		return err
	}

	// TODO: research this padding
	if _, err := buffer.Seek(stream, 18, buffer.SeekCurrent); err != nil {
		return err
	}

	flag := uint16(0)
	if _, err := buffer.ReadUint16LE(stream, &flag); err != nil {
		return err
	}
	isRoom := flag == 0

	vertexBufferOffsets := []uint64{}
	vertexBufferOffset := uint32(0)
	for range self.VertexBufferTotal {
		if _, err := buffer.ReadUint32LE(stream, &vertexBufferOffset); err != nil {
			return err
		}
		vertexBufferOffsets = append(vertexBufferOffsets, uint64(self.Offset+vertexBufferOffset))
	}

	if _, err := buffer.Seek(stream, int64(boneOffset), buffer.SeekStart); err != nil {
		return err
	}
	for i := range self.BoneTotal {
		x := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &x); err != nil {
			return err
		}
		y := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &y); err != nil {
			return err
		}
		z := float32(0)
		if _, err := buffer.ReadFloat32LE(stream, &z); err != nil {
			return err
		}
		// TODO: research this padding
		if _, err := buffer.Seek(stream, 2, buffer.SeekCurrent); err != nil {
			return err
		}
		parent := int16(0)
		if _, err := buffer.ReadInt16LE(stream, &parent); err != nil {
			return err
		}
		parent += 1

		self.Bones = append(
			self.Bones,
			bone.New(
				strconv.Itoa(int(i)),
				x,
				y,
				z,
				parent,
			),
		)
	}

	for _, vertexBufferOffset := range vertexBufferOffsets {
		if _, err := buffer.Seek(stream, int64(vertexBufferOffset), buffer.SeekStart); err != nil {
			return err
		}

		positionsOffset := uint32(0)
		if _, err := buffer.ReadUint32LE(stream, &positionsOffset); err != nil {
			return err
		}

		normalsOffset := uint32(0)
		if _, err := buffer.ReadUint32LE(stream, &normalsOffset); err != nil {
			return err
		}

		uvsOffset := uint32(0)
		if _, err := buffer.ReadUint32LE(stream, &uvsOffset); err != nil {
			return err
		}

		colorsOffset := uint32(0)
		if _, err := buffer.ReadUint32LE(stream, &colorsOffset); err != nil {
			return err
		}

		weightsOffset := uint32(0)
		if _, err := buffer.ReadUint32LE(stream, &weightsOffset); err != nil {
			return err
		}

		verticesTotal := uint16(0)
		if _, err := buffer.ReadUint16LE(stream, &verticesTotal); err != nil {
			return err
		}

		material := uint16(0)
		if _, err := buffer.ReadUint16LE(stream, &material); err != nil {
			return err
		}

		vb := VertexBuffer{
			Vertices: [][3]float32{},
			Indices:  [][3]int16{},
			Normals:  [][3]float32{},
			Uvs:      [][2]float32{},
			Weights:  [][4]float32{},
			Joints:   [][4]uint8{},
			Material: material,
		}

		if _, err := buffer.Seek(stream, int64(vertexBufferOffset)+int64(positionsOffset), buffer.SeekStart); err != nil {
			return err
		}

		for k := range verticesTotal {
			kk := int16(k)

			if isRoom {
				x := int16(0)
				if _, err := buffer.ReadInt16LE(stream, &x); err != nil {
					return err
				}
				y := int16(0)
				if _, err := buffer.ReadInt16LE(stream, &y); err != nil {
					return err
				}
				z := int16(0)
				if _, err := buffer.ReadInt16LE(stream, &z); err != nil {
					return err
				}

				vb.Vertices = append(
					vb.Vertices,
					[3]float32{float32(x) * RoomScale, float32(y) * RoomScale, float32(z) * RoomScale},
				)

				flag := uint16(0)
				if _, err := buffer.ReadUint16LE(stream, &flag); err != nil {
					return err
				}
				if flag == 32768 {
					continue
				} else if flag == 0 {
					vb.Indices = append(
						vb.Indices,
						[3]int16{kk - 2, kk - 1, kk},
					)
				} else if flag == 1 {
					vb.Indices = append(
						vb.Indices,
						[3]int16{kk - 1, kk - 2, kk},
					)
				}

			} else {
				x := float32(0)
				if _, err := buffer.ReadFloat32LE(stream, &x); err != nil {
					return err
				}
				y := float32(0)
				if _, err := buffer.ReadFloat32LE(stream, &y); err != nil {
					return err
				}
				z := float32(0)
				if _, err := buffer.ReadFloat32LE(stream, &z); err != nil {
					return err
				}

				vb.Vertices = append(vb.Vertices, [3]float32{x, y, z})

				flag := int32(0)
				if _, err := buffer.ReadInt32LE(stream, &flag); err != nil {
					return err
				}
				if flag == 32768 {
					continue
				} else if flag == 0 {
					vb.Indices = append(
						vb.Indices,
						[3]int16{kk - 2, kk - 1, kk},
					)
				} else if flag == 1 {
					vb.Indices = append(
						vb.Indices,
						[3]int16{kk - 1, kk - 2, kk},
					)
				}

			}

		}

		if normalsOffset != 0 {
			if _, err := buffer.Seek(stream, int64(vertexBufferOffset)+int64(normalsOffset), buffer.SeekStart); err != nil {
				return err
			}

			xyzw := make([]byte, 4)
			for range verticesTotal {
				if _, err := buffer.ReadBytes(stream, xyzw); err != nil {
					return err
				}
				x := ((float64(xyzw[0]) / 127.5) - 1) * -1
				y := ((float64(xyzw[1]) / 127.5) - 1) * -1
				z := ((float64(xyzw[2]) / 127.5) - 1) * -1
				l := math.Sqrt(x*x + y*y + z*z)
				x /= l
				y /= l
				z /= l
				vb.Normals = append(vb.Normals, [3]float32{
					float32(x),
					float32(y),
					float32(z),
				})
			}
		}

		if _, err := buffer.Seek(stream, int64(vertexBufferOffset)+int64(uvsOffset), buffer.SeekStart); err != nil {
			return err
		}

		for range verticesTotal {
			_u := int16(0)
			if _, err := buffer.ReadInt16LE(stream, &_u); err != nil {
				return err
			}
			_v := int16(0)
			if _, err := buffer.ReadInt16LE(stream, &_v); err != nil {
				return err
			}

			u := (float32(_u) / 4096)
			v := (float32(_v) / 4096) + 1
			vb.Uvs = append(vb.Uvs, [2]float32{u, v})
		}

		if colorsOffset != 0 {
			if _, err := buffer.Seek(stream, int64(vertexBufferOffset)+int64(colorsOffset), buffer.SeekStart); err != nil {
				return err
			}
			// TODO: vertex colors
		}

		if weightsOffset != 0 {
			if _, err := buffer.Seek(stream, int64(vertexBufferOffset)+int64(weightsOffset), buffer.SeekStart); err != nil {
				return err
			}

			xyzw := make([]byte, 4)
			for range verticesTotal {
				if _, err := buffer.ReadBytes(stream, xyzw); err != nil {
					return err
				}
				vb.Joints = append(vb.Joints, [4]byte{
					((xyzw[1] + 1) / 4) + 1,
					((xyzw[2] + 1) / 4) + 1,
					((xyzw[3] + 1) / 4) + 1,
					0,
				})

				if _, err := buffer.ReadBytes(stream, xyzw); err != nil {
					return err
				}
				vb.Weights = append(vb.Weights, [4]float32{
					float32(xyzw[0]) / 100,
					float32(xyzw[1]) / 100,
					float32(xyzw[2]) / 100,
					0,
				})
			}
		}

		self.VertexBuffers = append(self.VertexBuffers, &vb)
	}

	return nil
}

func New() *Mdb {
	return &Mdb{
		Offset:        0,
		Bones:         []*bone.Bone{},
		VertexBuffers: []*VertexBuffer{},
	}
}

func FromStreamWithOffset(mdb *Mdb, stream io.ReadWriteSeeker, offset uint32) error {
	mdb.Offset = offset
	return mdb.unmarshal(stream)
}

func FromStream(mdb *Mdb, stream io.ReadWriteSeeker) error {
	return FromStreamWithOffset(mdb, stream, 0)
}

func FromPathWithOffset(mdb *Mdb, filePath string, offset uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	mdb.Offset = offset
	return mdb.unmarshal(file)
}

func FromPath(mdb *Mdb, filePath string) error {
	return FromPathWithOffset(mdb, filePath, 0)
}
