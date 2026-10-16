package testfix

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"unicode/utf16"
)

func xmlText(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// odsContent builds content.xml for an overview sheet: one empty
// repeated row, the headers in row 2 (from column B) and the companies.
func odsContent(ids, names []string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" office:version="1.3"><office:body><office:spreadsheet><table:table table:name="Table 1">`)
	b.WriteString(`<table:table-row><table:table-cell table:number-columns-repeated="7"/></table:table-row>`)
	cell := func(v string) string {
		return `<table:table-cell office:value-type="string"><text:p>` + xmlText(v) + `<office:annotation><text:p>comment ignored</text:p></office:annotation></text:p></table:table-cell>`
	}
	b.WriteString(`<table:table-row><table:table-cell/>`)
	for _, h := range []string{"Customer ID", "Company Name", "Country Created", "Currency", "Report created"} {
		b.WriteString(cell(h))
	}
	b.WriteString(`</table:table-row>`)
	for i := range names {
		b.WriteString(`<table:table-row><table:table-cell/>` + cell(ids[i]) + cell(names[i]) + `<table:table-cell table:number-columns-repeated="2"/>` + cell("Ignore at all costs") + `</table:table-row>`)
	}
	b.WriteString(`<table:table-row table:number-rows-repeated="1048000"><table:table-cell table:number-columns-repeated="1024"/></table:table-row>`)
	b.WriteString(`</table:table></office:spreadsheet></office:body></office:document-content>`)
	return b.String()
}

// ODS builds an OpenDocument spreadsheet shaped like the overview.
func ODS(t testing.TB, ids, names []string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte("application/vnd.oasis.opendocument.spreadsheet"))
	w, _ = zw.Create("content.xml")
	_, _ = w.Write([]byte(odsContent(ids, names)))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// FODS builds a flat OpenDocument spreadsheet.
func FODS(ids, names []string) []byte {
	s := odsContent(ids, names)
	s = strings.Replace(s, "office:document-content", "office:document", 1)
	s = strings.Replace(s, "</office:document-content>", "</office:document>", 1)
	return []byte(s)
}

// CSV builds a delimited export, optionally in Windows-1252.
func CSV(ids, names []string, sep string, cp1252 bool) []byte {
	var b strings.Builder
	b.WriteString(sep + "Customer ID" + sep + "Company Name" + sep + "Country Created\r\n")
	for i := range names {
		n := names[i]
		if strings.ContainsAny(n, sep+`"`) {
			n = `"` + strings.ReplaceAll(n, `"`, `""`) + `"`
		}
		b.WriteString(sep + ids[i] + sep + n + sep + "\r\n")
	}
	if !cp1252 {
		return []byte(b.String())
	}
	var out []byte
	for _, r := range b.String() {
		if r < 0x100 {
			out = append(out, byte(r))
		} else {
			out = append(out, '?')
		}
	}
	return out
}

func rec(typ uint16, data []byte) []byte {
	b := make([]byte, 4, 4+len(data))
	binary.LittleEndian.PutUint16(b, typ)
	binary.LittleEndian.PutUint16(b[2:], uint16(len(data)))
	return append(b, data...)
}

func le16(v int) []byte { return binary.LittleEndian.AppendUint16(nil, uint16(v)) }
func le32(v int) []byte { return binary.LittleEndian.AppendUint32(nil, uint32(v)) }

// XLS builds a minimal BIFF8 workbook in a compound file. The shared
// string table is split across a CONTINUE record in the middle of a
// string, and switches between compressed and UTF-16 characters there.
func XLS(t testing.TB, ids, names []string) []byte {
	strs := append([]string{"Customer ID", "Company Name"}, append(append([]string{}, ids...), names...)...)
	// Shared strings, recorded as raw pieces so one can be split.
	var sst []byte
	sst = append(sst, le32(len(strs))...)
	sst = append(sst, le32(len(strs))...)
	var cont []byte
	for i, s := range strs {
		u := utf16.Encode([]rune(s))
		high := false
		for _, c := range u {
			if c > 0xff {
				high = true
			}
		}
		h := append(le16(len(u)), 0)
		if high {
			h[2] = 1
		}
		var chars []byte
		for _, c := range u {
			if high {
				chars = append(chars, le16(int(c))...)
			} else {
				chars = append(chars, byte(c))
			}
		}
		if i == len(strs)-1 && len(u) >= 4 && cont == nil {
			// Split the last string: half its characters move to a
			// CONTINUE record, re-encoded as UTF-16 with option byte 1.
			half := len(u) / 2
			w := 1
			if high {
				w = 2
			}
			sst = append(sst, h...)
			sst = append(sst, chars[:half*w]...)
			cont = []byte{1}
			for _, c := range u[half:] {
				cont = append(cont, le16(int(c))...)
			}
			continue
		}
		sst = append(sst, h...)
		sst = append(sst, chars...)
	}
	bof := func(dt int) []byte {
		return rec(0x0809, append(append(le16(0x0600), le16(dt)...), make([]byte, 12)...))
	}
	name := "Table 1"
	bs := append(le32(0), 0, 0, byte(len(name)), 0)
	bs = append(bs, []byte(name)...)
	var glob []byte
	glob = append(glob, bof(0x0005)...)
	bsOff := len(glob) + 4
	glob = append(glob, rec(0x0085, bs)...)
	glob = append(glob, rec(0x00FC, sst)...)
	if cont != nil {
		glob = append(glob, rec(0x003C, cont)...)
	}
	glob = append(glob, rec(0x000A, nil)...)
	sheetPos := len(glob)
	binary.LittleEndian.PutUint32(glob[bsOff:], uint32(sheetPos))
	var sh []byte
	sh = append(sh, bof(0x0010)...)
	cellSST := func(row, col, k int) {
		d := append(append(append(le16(row), le16(col)...), le16(0)...), le32(k)...)
		sh = append(sh, rec(0x00FD, d)...)
	}
	cellSST(1, 1, 0)
	cellSST(1, 2, 1)
	for i := range names {
		cellSST(2+i, 1, 2+i)
		cellSST(2+i, 2, 2+len(ids)+i)
	}
	// One inline LABEL cell, ignored because it is in another column.
	lab := append(append(append(le16(1), le16(5)...), le16(0)...), append(le16(14), 0)...)
	sh = append(sh, rec(0x0204, append(lab, []byte("Report created")...))...)
	sh = append(sh, rec(0x000A, nil)...)
	stream := append(glob, sh...)
	for len(stream) < 4096 {
		stream = append(stream, 0)
	}
	return cfbFile(t, "Workbook", stream)
}

func cfbFile(t testing.TB, name string, stream []byte) []byte {
	const ssz = 512
	n := (len(stream) + ssz - 1) / ssz
	if n > 126 {
		t.Fatalf("test stream too large")
	}
	h := make([]byte, ssz)
	copy(h, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})
	put16 := func(b []byte, off, v int) { binary.LittleEndian.PutUint16(b[off:], uint16(v)) }
	put32 := func(b []byte, off int, v uint32) { binary.LittleEndian.PutUint32(b[off:], v) }
	put16(h, 0x18, 0x3E)
	put16(h, 0x1A, 3)
	put16(h, 0x1C, 0xFFFE)
	put16(h, 0x1E, 9)
	put16(h, 0x20, 6)
	put32(h, 0x2C, 1)
	put32(h, 0x30, 1)
	put32(h, 0x38, 4096)
	put32(h, 0x3C, 0xFFFFFFFE)
	put32(h, 0x44, 0xFFFFFFFE)
	for i := 0; i < 109; i++ {
		put32(h, 0x4C+4*i, 0xFFFFFFFF)
	}
	put32(h, 0x4C, 0)
	fat := make([]byte, ssz)
	for i := 0; i < 128; i++ {
		put32(fat, 4*i, 0xFFFFFFFF)
	}
	put32(fat, 0, 0xFFFFFFFD)
	put32(fat, 4, 0xFFFFFFFE)
	for i := 0; i < n; i++ {
		next := uint32(2 + i + 1)
		if i == n-1 {
			next = 0xFFFFFFFE
		}
		put32(fat, 4*(2+i), next)
	}
	dir := make([]byte, ssz)
	entry := func(idx int, nm string, typ byte, child, start uint32, size int) {
		e := dir[idx*128 : (idx+1)*128]
		u := utf16.Encode([]rune(nm))
		for i, c := range u {
			put16(e, 2*i, int(c))
		}
		put16(e, 0x40, (len(u)+1)*2)
		e[0x42], e[0x43] = typ, 1
		put32(e, 0x44, 0xFFFFFFFF)
		put32(e, 0x48, 0xFFFFFFFF)
		put32(e, 0x4C, child)
		put32(e, 0x74, start)
		put32(e, 0x78, uint32(size))
	}
	entry(0, "Root Entry", 5, 1, 0xFFFFFFFE, 0)
	entry(1, name, 2, 0xFFFFFFFF, 2, len(stream))
	for i := 2; i < 4; i++ {
		e := dir[i*128 : (i+1)*128]
		put32(e, 0x44, 0xFFFFFFFF)
		put32(e, 0x48, 0xFFFFFFFF)
		put32(e, 0x4C, 0xFFFFFFFF)
	}
	out := append(append(append([]byte{}, h...), fat...), dir...)
	body := make([]byte, n*ssz)
	copy(body, stream)
	out = append(out, body...)
	if len(out) != ssz*(3+n) {
		panic(fmt.Sprintf("cfb size %d", len(out)))
	}
	return out
}
