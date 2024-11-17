package graphicsynthesizer

var GsMem = make([]byte, 1024*1024*4)

var block32 = [32]int{0, 1, 4, 5, 16, 17, 20, 21, 2, 3, 6,
	7, 18, 19, 22, 23, 8, 9, 12, 13, 24, 25,
	28, 29, 10, 11, 14, 15, 26, 27, 30, 31}

var columnWord32 = [16]int{0, 1, 4, 5, 8, 9, 12, 13, 2, 3, 6, 7, 10, 11, 14, 15}

func WriteTexPSMCT32(dbp int, dbw int, dsax int, dsay int, rrw int, rrh int, data []byte) {
	src := 0
	startBlockPos := dbp * 64

	for y := dsay; y < dsay+rrh; y++ {
		for x := dsax; x < dsax+rrw; x++ {
			pageX := x / 64
			pageY := y / 32
			page := pageX + pageY*dbw

			px := x - (pageX * 64)
			py := y - (pageY * 32)

			blockX := px / 8
			blockY := py / 8
			block := block32[blockX+blockY*8]

			bx := px - blockX*8
			by := py - blockY*8

			column := by / 2

			cx := bx
			cy := by - column*2
			cw := columnWord32[cx+cy*8]

			pos := (startBlockPos + page*2048 + block*64 + column*16 + cw) * 4
			GsMem[pos] = data[src]
			GsMem[pos+1] = data[src+1]
			GsMem[pos+2] = data[src+2]
			GsMem[pos+3] = data[src+3]
			src += 4
		}
	}
}

var block8 = [32]int{0, 1, 4, 5, 16, 17, 20, 21, 2, 3, 6, 7, 18, 19, 22, 23,
	8, 9, 12, 13, 24, 25, 28, 29, 10, 11, 14, 15, 26, 27, 30, 31}

var columnWord8 = [2][64]int{
	{0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
		2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,

		8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
		10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7},
	{8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
		10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,

		0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
		2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15}}

var columnByte8 = [64]int{0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2,
	0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2,

	1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3, 3, 3, 3, 3,
	1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3, 3, 3, 3, 3}

func ReadTexPSMT8(dbp int, dbw int, dsax int, dsay int, rrw int, rrh int, data []byte) {
	dbw >>= 1
	src := 0
	startBlockPos := dbp * 64

	for y := dsay; y < dsay+rrh; y++ {
		for x := dsax; x < dsax+rrw; x++ {
			pageX := x / 128
			pageY := y / 64
			page := pageX + pageY*dbw

			px := x - (pageX * 128)
			py := y - (pageY * 64)

			blockX := px / 16
			blockY := py / 16
			block := block8[blockX+blockY*8]

			bx := px - blockX*16
			by := py - blockY*16

			column := by / 4

			cx := bx
			cy := by - column*4
			cw := columnWord8[column&1][cx+cy*16]
			cb := columnByte8[cx+cy*16]

			pos := startBlockPos + page*2048 + block*64 + column*16 + cw
			data[src] = GsMem[pos*4+cb]

			src++
		}
	}
}

var Block4 = [32]int{0, 2, 8, 10, 1, 3, 9, 11, 4, 6, 12,
	14, 5, 7, 13, 15, 16, 18, 24, 26, 17, 19,
	25, 27, 20, 22, 28, 30, 21, 23, 29, 31}

var ColumnWord4 = [2][128]int{
	{0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
		0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
		2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
		2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,

		8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
		8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
		10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
		10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7},
	{8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
		8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
		10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
		10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,

		0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
		0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
		2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
		2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15}}

var ColumnByte4 = [128]int{
	0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2, 4, 4, 4, 4, 4, 4,
	4, 4, 6, 6, 6, 6, 6, 6, 6, 6, 0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2,
	2, 2, 2, 2, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 6, 6, 6, 6, 6,

	1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3, 3, 3, 3, 3, 5, 5, 5, 5, 5, 5,
	5, 5, 7, 7, 7, 7, 7, 7, 7, 7, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3,
	3, 3, 3, 3, 5, 5, 5, 5, 5, 5, 5, 5, 7, 7, 7, 7, 7, 7, 7, 7}

func ReadTexPSMT4(dbp int, dbw int, dsax int, dsay int, rrw int, rrh int, data []byte) {
	dbw >>= 1
	src := 0
	startBlockPos := dbp * 64

	odd := false

	for y := dsay; y < dsay+rrh; y++ {
		for x := dsax; x < dsax+rrw; x++ {
			pageX := x / 128
			pageY := y / 128
			page := pageX + pageY*dbw

			px := x - (pageX * 128)
			py := y - (pageY * 128)

			blockX := px / 32
			blockY := py / 16
			block := Block4[blockX+blockY*4]

			bx := px - blockX*32
			by := py - blockY*16

			column := by / 4

			cx := bx
			cy := by - column*4
			cw := ColumnWord4[column&1][cx+cy*32]
			cb := ColumnByte4[cx+cy*32]
			pos := startBlockPos + page*2048 + block*64 + column*16 + cw
			pix := GsMem[pos*4+(cb>>1)]

			if (cb & 1) != 0 {
				if odd {
					data[src] = ((data[src]) & 0x0f) | (pix & 0xf0)
				} else {
					data[src] = ((data[src]) & 0xf0) | ((pix >> 4) & 0x0f)
				}
			} else {
				if odd {
					data[src] = (data[src] & 0x0f) | (pix<<4)&0xf0
				} else {
					data[src] = (data[src] & 0xf0) | (pix & 0x0f)
				}
			}

			if odd {
				src++
			}

			odd = !odd
		}
	}
}

func Unswizzle4(data []byte, width int, height int) []byte {
	result := make([]byte, len(data))
	rrw := width / 2
	rrh := height / 4
	WriteTexPSMCT32(0, rrw/64, 0, 0, rrw, rrh, data)
	ReadTexPSMT4(0, width/64, 0, 0, width, height, result)
	return result
}

func Unswizzle8(data []byte, width int, height int) []byte {
	result := make([]byte, len(data))
	rrw := width / 2
	rrh := height / 2
	WriteTexPSMCT32(0, rrw/64, 0, 0, rrw, rrh, data)
	ReadTexPSMT8(0, width/64, 0, 0, width, height, result)
	return result
}

func Swizzle8(data []byte, width, height int) []byte {
	result := make([]byte, len(data))
	for y := range height {
		for x := range width {
			block := (y & ^0xF)*width + (x & ^0xF)*2
			selector := (((y + 2) >> 2) & 0x1) * 4
			pos := (((y & ^0x3) >> 1) + (y & 1)) & 0x7
			location := pos*width*2 + ((x+selector)&0x7)*4
			num := ((y >> 1) & 1) + ((x >> 2) & 2)
			swizzle := block + location + num
			result[swizzle] = data[y*width+x]
		}
	}
	return result
}
