package iso

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/anasrar/chihuahua/next/pkg/baseline"
)

type source struct {
	stream io.ReadSeeker
	offset int64
	length int64
}

func newSource(stream io.ReadSeeker, offset, length int64) (*source, error) {
	end, err := stream.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if offset < 0 || offset > end {
		return nil, fmt.Errorf("ISO: offset %d is outside the file (size %d)", offset, end)
	}
	size := end - offset
	if length > 0 && length < size {
		size = length
	}
	return &source{stream: stream, offset: offset, length: size}, nil
}

func (stream *source) readAt(p []byte, offset int64) error {
	if offset < 0 || offset+int64(len(p)) > stream.length {
		return io.ErrUnexpectedEOF
	}
	if _, err := stream.stream.Seek(stream.offset+offset, io.SeekStart); err != nil {
		return err
	}
	_, err := io.ReadFull(stream.stream, p)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return err
}

type sectionReader struct {
	source   *source
	off, n   int64
	consumed int64
}

func (reader *sectionReader) Read(p []byte) (int, error) {
	if reader.consumed >= reader.n {
		return 0, io.EOF
	}
	pos := reader.off + reader.consumed
	if pos >= reader.source.length {
		return 0, io.ErrUnexpectedEOF
	}
	if rem := reader.n - reader.consumed; int64(len(p)) > rem {
		p = p[:rem]
	}
	if rem := reader.source.length - pos; int64(len(p)) > rem {
		p = p[:rem]
	}
	if _, err := reader.source.stream.Seek(reader.source.offset+pos, io.SeekStart); err != nil {
		return 0, err
	}
	n, err := reader.source.stream.Read(p)
	reader.consumed += int64(n)
	if err == io.EOF {
		if n > 0 {
			return n, nil
		}
		return 0, io.ErrUnexpectedEOF
	}
	return n, err
}

func readSector(s *source, lba int64, n int) ([]byte, error) {
	buf := make([]byte, n*SECTOR_SIZE)
	if err := s.readAt(buf, lba*SECTOR_SIZE); err != nil {
		return nil, fmt.Errorf("ISO: read sector %d %w", lba, err)
	}
	return buf, nil
}

func findISOPVD(s *source) []byte {
	for lba := int64(16); lba < 32; lba++ {
		b, err := readSector(s, lba, 1)
		if err != nil || string(b[1:6]) != "CD001" {
			return nil
		}
		if b[0] == 1 {
			return b
		}
		if b[0] == 255 {
			return nil
		}
	}
	return nil
}

type longAD struct {
	length  uint32
	block   uint32
	partRef uint16
}

func parseLongAD(p []byte) longAD {
	return longAD{le.Uint32(p[0:]) & 0x3FFFFFFF, le.Uint32(p[4:]), le.Uint16(p[8:])}
}

type udfReader struct {
	s         *source
	total     int64
	partStart map[uint16]int64
	partMap   []uint16
	root      longAD
	haveLVD   bool
	vol       Volume
}

func tagID(b []byte) (uint16, bool) {
	if len(b) < 16 {
		return 0, false
	}
	var sum byte
	for i := range 16 {
		if i != 4 {
			sum += b[i]
		}
	}
	id := le.Uint16(b)
	return id, sum == b[4] && id != 0
}

func (u *udfReader) resolve(partRef uint16, block uint32) (int64, error) {
	if int(partRef) >= len(u.partMap) {
		return 0, fmt.Errorf("ISO: invalid partition reference %d", partRef)
	}
	start, ok := u.partStart[u.partMap[partRef]]
	if !ok {
		return 0, fmt.Errorf("ISO: partition %d not found", u.partMap[partRef])
	}
	return start + int64(block), nil
}

func readUDF(s *source) (*Image, error) {
	u := &udfReader{s: s, total: s.length / SECTOR_SIZE, partStart: map[uint16]int64{}}
	var avdp []byte
	for _, lba := range []int64{256, u.total - 1, u.total - 257} {
		if lba <= 0 {
			continue
		}
		b, err := readSector(s, lba, 1)
		if err != nil {
			continue
		}
		if id, ok := tagID(b); ok && id == 2 {
			avdp = b
			break
		}
	}
	if avdp == nil {
		return nil, fmt.Errorf("ISO: UDF anchor (AVDP) not found")
	}
	if err := u.parseVDS(int64(le.Uint32(avdp[20:])), le.Uint32(avdp[16:]), 0); err != nil {
		return nil, err
	}
	if !u.haveLVD || len(u.partStart) == 0 {
		if err := u.parseVDS(int64(le.Uint32(avdp[28:])), le.Uint32(avdp[24:]), 0); err != nil {
			return nil, err
		}
	}
	if !u.haveLVD || len(u.partStart) == 0 {
		return nil, fmt.Errorf("ISO: logical volume / partition descriptor missing")
	}
	sec, err := u.resolve(u.root.partRef, u.root.block)
	if err != nil {
		return nil, err
	}
	b, err := readSector(s, sec, 1)
	if err != nil {
		return nil, err
	}
	if id, ok := tagID(b); !ok || id != 256 {
		return nil, fmt.Errorf("ISO: file set descriptor not found")
	}
	u.vol.FileSetID = decodeDString(b[304:336])
	for _, c := range []string{u.vol.LogicalVolumeID, u.vol.VolumeID, u.vol.FileSetID} {
		if c != "" {
			u.vol.Name = c
			break
		}
	}
	if pvd := findISOPVD(s); pvd != nil {
		u.vol.SystemID = strings.TrimRight(string(pvd[8:40]), " \x00")
		if u.vol.Name == "" {
			u.vol.Name = strings.TrimRight(string(pvd[40:72]), " \x00")
		}
	}
	rootRef := parseLongAD(b[400:])
	root := u.buildNode(rootRef, "/", "/", map[int64]bool{})
	return &Image{Volume: u.vol, Sectors: u.total, Root: root}, nil
}

func (u *udfReader) parseVDS(lba int64, length uint32, depth int) error {
	n := int(length / SECTOR_SIZE)
	if n == 0 || n > 256 {
		n = 16
	}
	for i := 0; i < n; i++ {
		b, err := readSector(u.s, lba+int64(i), 1)
		if err != nil {
			return err
		}
		id, ok := tagID(b)
		if !ok {
			return nil
		}
		switch id {
		case 1:
			u.vol.VolumeID = decodeDString(b[24:56])
		case 3:
			if depth < 8 {
				return u.parseVDS(int64(le.Uint32(b[24:])), le.Uint32(b[20:]), depth+1)
			}
			return nil
		case 5:
			u.partStart[le.Uint16(b[22:])] = int64(le.Uint32(b[188:]))
		case 6:
			if bs := le.Uint32(b[212:]); bs != SECTOR_SIZE {
				return fmt.Errorf("ISO: unsupported logical block size %d", bs)
			}
			u.vol.LogicalVolumeID = decodeDString(b[84:212])
			u.root = parseLongAD(b[248:])
			nmaps := int(le.Uint32(b[268:]))
			off := 440
			u.partMap = u.partMap[:0]
			for j := 0; j < nmaps && off+2 <= len(b); j++ {
				typ, l := b[off], int(b[off+1])
				if l == 0 || off+l > len(b) {
					break
				}
				switch typ {
				case 1:
					u.partMap = append(u.partMap, le.Uint16(b[off+4:]))
				case 2:
					u.partMap = append(u.partMap, le.Uint16(b[off+38:]))
				}
				off += l
			}
			u.haveLVD = true
		case 8:
			return nil
		}
	}
	return nil
}

type udfEntry struct {
	icbSector   int64
	fileType    uint8
	size        int64
	mtime       time.Time
	extents     []Extent
	embedded    []byte
	hasEmbedded bool
	dataOffset  int64
}

func (u *udfReader) loadEntry(ref longAD) (*udfEntry, error) {
	for range 8 {
		sec, err := u.resolve(ref.partRef, ref.block)
		if err != nil {
			return nil, err
		}
		b, err := readSector(u.s, sec, 1)
		if err != nil {
			return nil, err
		}
		id, ok := tagID(b)
		if !ok {
			return nil, fmt.Errorf("ISO: invalid ICB tag at sector %d", sec)
		}
		switch id {
		case 259:
			ref = parseLongAD(b[36:])
		case 261, 266:
			return u.parseEntry(b, sec, id, ref.partRef)
		default:
			return nil, fmt.Errorf("ISO: unexpected tag %d at sector %d", id, sec)
		}
	}
	return nil, fmt.Errorf("ISO: ICB indirection too deep")
}

func (u *udfReader) parseEntry(b []byte, sec int64, id uint16, partRef uint16) (*udfEntry, error) {
	e := &udfEntry{icbSector: sec, fileType: b[27]}
	adType := le.Uint16(b[34:]) & 7
	e.size = int64(le.Uint64(b[56:]))
	var lea, lad, base int
	if id == 261 {
		e.mtime = parseUDFTime(b[84:96])
		lea, lad, base = int(le.Uint32(b[168:])), int(le.Uint32(b[172:])), 176
	} else {
		e.mtime = parseUDFTime(b[92:104])
		lea, lad, base = int(le.Uint32(b[208:])), int(le.Uint32(b[212:])), 216
	}
	start := base + lea
	if start > len(b) {
		return nil, fmt.Errorf("ISO: corrupt file entry")
	}
	if start+lad > len(b) {
		lad = len(b) - start
	}
	ads := b[start : start+lad]
	switch adType {
	case 3:
		e.hasEmbedded = true
		e.embedded = append([]byte(nil), ads...)
		if int64(len(e.embedded)) > e.size {
			e.embedded = e.embedded[:e.size]
		}
		e.dataOffset = sec*SECTOR_SIZE + int64(start)
	case 0, 1:
		ex, err := u.collectExtents(ads, adType, partRef)
		if err != nil {
			return nil, err
		}
		e.extents = ex
	default:
		return nil, fmt.Errorf("ISO: unsupported allocation descriptor type %d", adType)
	}
	return e, nil
}

func (u *udfReader) collectExtents(ads []byte, adType uint16, partRef uint16) ([]Extent, error) {
	var out []Extent
	step := 8
	if adType == 1 {
		step = 16
	}
	for range 256 {
		var next *longAD
		for off := 0; off+step <= len(ads); off += step {
			raw := le.Uint32(ads[off:])
			length, typ := raw&0x3FFFFFFF, raw>>30
			if length == 0 {
				return out, nil
			}
			block := le.Uint32(ads[off+4:])
			pr := partRef
			if adType == 1 {
				pr = le.Uint16(ads[off+8:])
			}
			if typ == 3 {
				next = &longAD{length, block, pr}
				break
			}
			sec, err := u.resolve(pr, block)
			if err != nil {
				return nil, err
			}
			out = append(out, Extent{LBA: sec, Offset: sec * SECTOR_SIZE, Length: int64(length), Sparse: typ != 0})
		}
		if next == nil {
			return out, nil
		}
		sec, err := u.resolve(next.partRef, next.block)
		if err != nil {
			return nil, err
		}
		b, err := readSector(u.s, sec, 1)
		if err != nil {
			return nil, err
		}
		if id, ok := tagID(b); !ok || id != 258 {
			return nil, fmt.Errorf("ISO: bad allocation extent descriptor")
		}
		l := int(le.Uint32(b[20:]))
		if 24+l > len(b) {
			l = len(b) - 24
		}
		ads = b[24 : 24+l]
	}
	return out, fmt.Errorf("ISO: too many allocation extents")
}

func (u *udfReader) readData(e *udfEntry) ([]byte, error) {
	if e.hasEmbedded {
		return e.embedded, nil
	}
	buf := make([]byte, 0, e.size)
	remaining := e.size
	for _, x := range e.extents {
		if remaining <= 0 {
			break
		}
		n := min(x.Length, remaining)
		chunk := make([]byte, n)
		if !x.Sparse {
			if err := u.s.readAt(chunk, x.Offset); err != nil {
				return nil, err
			}
		}
		buf = append(buf, chunk...)
		remaining -= n
	}
	return buf, nil
}

func (u *udfReader) buildNode(ref longAD, name, path string, seen map[int64]bool) *Node {
	n := &Node{Name: name, Path: path}
	e, err := u.loadEntry(ref)
	if err != nil {
		n.Error = err.Error()
		return n
	}
	n.Length, n.Date, n.Extents = e.size, e.mtime, e.extents
	switch {
	case e.hasEmbedded:
		n.Offset = e.dataOffset
	case len(e.extents) > 0:
		n.Offset = e.extents[0].Offset
	default:
		n.Offset = e.icbSector * SECTOR_SIZE
	}
	n.LBA = n.Offset / SECTOR_SIZE
	n.IsDirectory = e.fileType == 4
	n.IsFile = !n.IsDirectory
	if !n.IsDirectory || seen[e.icbSector] {
		return n
	}
	seen[e.icbSector] = true
	data, err := u.readData(e)
	if err != nil {
		n.Error = err.Error()
		return n
	}
	for pos := 0; pos+38 <= len(data); {
		if id, ok := tagID(data[pos:]); !ok || id != 257 {
			break
		}
		chars := data[pos+18]
		lfi := int(data[pos+19])
		liu := int(le.Uint16(data[pos+36:]))
		total := (38 + liu + lfi + 3) &^ 3
		if pos+38+liu+lfi > len(data) {
			break
		}
		if chars&0x0C == 0 && lfi > 0 {
			cn := decodeName(data[pos+38+liu : pos+38+liu+lfi])
			n.Children = append(n.Children, u.buildNode(parseLongAD(data[pos+20:]), cn, strings.TrimSuffix(path, "/")+"/"+cn, seen))
		}
		pos += total
	}
	return n
}

func decodeDString(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	n := int(b[len(b)-1])
	if n == 0 || n > len(b)-1 {
		return ""
	}
	return strings.TrimRight(decodeName(b[:n]), "\x00 ")
}

func decodeName(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	comp, s := b[0], b[1:]
	if comp == 16 {
		u := make([]uint16, len(s)/2)
		for i := range u {
			u[i] = be.Uint16(s[i*2:])
		}
		return string(utf16.Decode(u))
	}
	r := make([]rune, len(s))
	for i, c := range s {
		r[i] = rune(c)
	}
	return string(r)
}

func parseUDFTime(b []byte) time.Time {
	timezone := le.Uint16(b[0:])
	typ, off := timezone>>12, int(timezone&0x0FFF)
	if off&0x800 != 0 {
		off -= 0x1000
	}
	loc := time.UTC
	if typ == 1 && off != -2047 {
		loc = time.FixedZone("", off*60)
	}
	y, mo, d := int(le.Uint16(b[2:])), int(b[4]), int(b[5])
	h, mi, s := int(b[6]), int(b[7]), int(b[8])
	if y == 0 || mo < 1 || mo > 12 || d < 1 || d > 31 || h > 23 || mi > 59 || s > 59 {
		return time.Time{}
	}
	ns := int(b[9])*10_000_000 + int(b[10])*100_000 + int(b[11])*1_000
	return time.Date(y, time.Month(mo), d, h, mi, s, ns, loc)
}

func unmarshal(source string, stream io.ReadSeeker, offset int64, length int64) (*Container, error) {
	container := &Container{
		File: &baseline.File{
			Source: source,
			Offset: offset,
			Length: length,
		},
	}

	src, err := newSource(stream, offset, length)
	if err != nil {
		return nil, err
	}
	img, err := readUDF(src)
	if err != nil {
		return nil, err
	}
	container.Image = img

	return container, nil
}

func FromPath(source string, offset int64, length int64) (*Container, error) {
	file, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return unmarshal(source, file, offset, length)
}
