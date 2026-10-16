package sheet

import (
	"errors"
	"strings"
	"unicode/utf16"
)

// BIFF8 record types used here ([MS-XLS] 2.3).
const (
	recFormula    = 0x0006
	recEOF        = 0x000A
	recFilePass   = 0x002F
	recContinue   = 0x003C
	recBoundSheet = 0x0085
	recRString    = 0x00D6
	recSST        = 0x00FC
	recLabelSST   = 0x00FD
	recLabel      = 0x0204
	recString     = 0x0207
	recBOF        = 0x0809
)

type record struct {
	typ  uint16
	data []byte
}

func records(b []byte) []record {
	var out []record
	for off := 0; off+4 <= len(b); {
		typ, n := u16(b, off), int(u16(b, off+2))
		off += 4
		if off+n > len(b) {
			break
		}
		out = append(out, record{typ: typ, data: b[off : off+n]})
		off += n
	}
	return out
}

// readXLS reads a legacy BIFF8 (Excel 97-2003) workbook.
func readXLS(b []byte, opt Options) ([]Row, []string, error) {
	c, err := openCFB(b)
	if err != nil {
		return nil, nil, err
	}
	wb, err := c.stream("Workbook")
	if err != nil {
		return nil, nil, errors.New("not an Excel 97-2003 workbook (Workbook stream missing; Excel 5 files are not supported)")
	}
	recs := records(wb)
	if len(recs) == 0 || recs[0].typ != recBOF || len(recs[0].data) < 2 || u16(recs[0].data, 0) != 0x0600 {
		return nil, nil, errors.New("only BIFF8 (Excel 97-2003) workbooks are supported")
	}
	type sheetInfo struct {
		name string
		pos  int
	}
	var sheets []sheetInfo
	var sst []string
	for i, r := range recs {
		switch r.typ {
		case recFilePass:
			return nil, nil, errors.New("the workbook is password protected")
		case recBoundSheet:
			if len(r.data) >= 8 && r.data[5] == 0 { // worksheet
				s, _ := shortString(r.data[6:])
				sheets = append(sheets, sheetInfo{name: s, pos: int(u32(r.data, 0))})
			}
		case recSST:
			parts := [][]byte{r.data}
			for j := i + 1; j < len(recs) && recs[j].typ == recContinue; j++ {
				parts = append(parts, recs[j].data)
			}
			sst = readSST(parts)
		}
	}
	var names []string
	var cells [][]Cell
	for _, sh := range sheets {
		if sh.pos < 0 || sh.pos+4 > len(wb) {
			continue
		}
		names = append(names, sh.name)
		cells = append(cells, sheetCells(records(wb[sh.pos:]), sst))
	}
	rows, err := fromSheets(names, cells, opt)
	return rows, nil, err
}

func sheetCells(recs []record, sst []string) []Cell {
	var out []Cell
	add := func(row, col int, v string) {
		if strings.TrimSpace(v) != "" {
			out = append(out, Cell{Col: col + 1, Row: row + 1, Ref: Ref(col+1, row+1), Val: v})
		}
	}
	pendRow, pendCol := -1, -1
	for i, r := range recs {
		if i > 0 && (r.typ == recEOF || r.typ == recBOF) {
			break
		}
		d := r.data
		switch r.typ {
		case recLabelSST:
			if len(d) >= 10 {
				if k := int(u32(d, 6)); k >= 0 && k < len(sst) {
					add(int(u16(d, 0)), int(u16(d, 2)), sst[k])
				}
			}
		case recLabel, recRString:
			if len(d) >= 9 {
				s, _ := longString(d[6:])
				add(int(u16(d, 0)), int(u16(d, 2)), s)
			}
		case recFormula:
			// A string result follows in a STRING record.
			if len(d) >= 14 && d[6] == 0 && d[12] == 0xFF && d[13] == 0xFF {
				pendRow, pendCol = int(u16(d, 0)), int(u16(d, 2))
			}
		case recString:
			if pendRow >= 0 {
				s, _ := longString(d)
				add(pendRow, pendCol, s)
				pendRow, pendCol = -1, -1
			}
		}
	}
	return out
}

func decodeChars(b []byte, n int, high bool) (string, int) {
	if high {
		n = min(n, len(b)/2)
		u := make([]uint16, n)
		for i := range u {
			u[i] = u16(b, 2*i)
		}
		return string(utf16.Decode(u)), 2 * n
	}
	n = min(n, len(b))
	r := make([]rune, n)
	for i := 0; i < n; i++ {
		r[i] = rune(b[i])
	}
	return string(r), n
}

// shortString decodes a ShortXLUnicodeString (1-byte length).
func shortString(b []byte) (string, int) {
	if len(b) < 2 {
		return "", len(b)
	}
	s, n := decodeChars(b[2:], int(b[0]), b[1]&1 != 0)
	return s, 2 + n
}

// longString decodes an XLUnicodeString (2-byte length) without rich text.
func longString(b []byte) (string, int) {
	if len(b) < 3 {
		return "", len(b)
	}
	cch, flags := int(u16(b, 0)), b[2]
	off := 3
	if flags&0x08 != 0 {
		off += 2
	}
	if flags&0x04 != 0 {
		off += 4
	}
	if off > len(b) {
		return "", len(b)
	}
	s, n := decodeChars(b[off:], cch, flags&1 != 0)
	return s, off + n
}

// sstReader walks the SST record and its CONTINUE records. A string's
// characters may be split across records; the continuation then starts
// with a new option byte.
type sstReader struct {
	parts [][]byte
	p, o  int
}

func (r *sstReader) left() int {
	if r.p >= len(r.parts) {
		return 0
	}
	return len(r.parts[r.p]) - r.o
}

func (r *sstReader) next() bool {
	for r.p < len(r.parts) && r.left() == 0 {
		r.p++
		r.o = 0
	}
	return r.p < len(r.parts)
}

func (r *sstReader) bytes(n int) ([]byte, error) {
	var out []byte
	for n > 0 {
		if !r.next() {
			return nil, errors.New("sst: truncated")
		}
		k := min(n, r.left())
		out = append(out, r.parts[r.p][r.o:r.o+k]...)
		r.o += k
		n -= k
	}
	return out, nil
}

func (r *sstReader) str() (string, error) {
	h, err := r.bytes(3)
	if err != nil {
		return "", err
	}
	cch, flags := int(u16(h, 0)), h[2]
	runs, ext := 0, 0
	if flags&0x08 != 0 {
		b, err := r.bytes(2)
		if err != nil {
			return "", err
		}
		runs = int(u16(b, 0))
	}
	if flags&0x04 != 0 {
		b, err := r.bytes(4)
		if err != nil {
			return "", err
		}
		ext = int(u32(b, 0))
	}
	high := flags&1 != 0
	var sb strings.Builder
	part := r.p
	for cch > 0 {
		if !r.next() {
			return "", errors.New("sst: truncated string")
		}
		if r.p != part {
			// The character array continues in a CONTINUE record, which
			// starts with a fresh option byte.
			part = r.p
			high = r.parts[r.p][0]&1 != 0
			r.o = 1
			if r.left() == 0 {
				continue
			}
		}
		width := 1
		if high {
			width = 2
		}
		k := min(cch, r.left()/width)
		if k == 0 {
			return "", errors.New("sst: bad continuation")
		}
		s, used := decodeChars(r.parts[r.p][r.o:], k, high)
		sb.WriteString(s)
		r.o += used
		cch -= k
	}
	if _, err := r.bytes(4*runs + ext); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func readSST(parts [][]byte) []string {
	if len(parts) == 0 || len(parts[0]) < 8 {
		return nil
	}
	unique := int(u32(parts[0], 4))
	r := &sstReader{parts: parts, o: 8}
	var out []string
	for i := 0; i < unique && i < 1<<20; i++ {
		s, err := r.str()
		if err != nil {
			break
		}
		out = append(out, s)
	}
	return out
}
