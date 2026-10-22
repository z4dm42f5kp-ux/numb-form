package sheet

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// SetText returns a copy of an .xlsx workbook in which the cells vals
// (A1 reference -> text) of the sheet part hold that text. Only those
// cells change: each is spliced into the sheet's XML, keeps its own style
// index and gets no new formatting; every other part of the archive is
// copied byte for byte. Cells holding a formula are refused.
func SetText(b []byte, part string, vals map[string]string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, fmt.Errorf("open workbook: %w", err)
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	done := false
	for _, f := range zr.File {
		if f.Name != part {
			if err := w.Copy(f); err != nil {
				return nil, err
			}
			continue
		}
		src, err := readPart(f)
		if err != nil {
			return nil, err
		}
		x := string(src)
		for ref, v := range vals {
			if x, err = setCell(x, ref, v); err != nil {
				return nil, fmt.Errorf("cell %s: %w", ref, err)
			}
		}
		hdr := f.FileHeader
		hdr.Method = zip.Deflate
		hdr.CompressedSize64, hdr.UncompressedSize64, hdr.CRC32 = 0, 0, 0
		hdr.Flags &^= 0x8 // the writer sets the data-descriptor flag itself
		hdr.Extra = nil   // old size fields would be stale
		dst, err := w.CreateHeader(&hdr)
		if err != nil {
			return nil, err
		}
		if _, err := io.WriteString(dst, x); err != nil {
			return nil, err
		}
		done = true
	}
	if !done {
		return nil, fmt.Errorf("the workbook has no part %s", part)
	}
	if err := w.SetComment(zr.Comment); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

var (
	reRowTag  = regexp.MustCompile(`<(\w+:)?row\b[^>]*>`)
	reCellTag = regexp.MustCompile(`<(\w+:)?c\b[^>]*?(/?)>`)
	reAttr    = regexp.MustCompile(`\s(\w+)="([^"]*)"`)
)

func attrs(tag string) map[string]string {
	m := map[string]string{}
	for _, a := range reAttr.FindAllStringSubmatch(tag, -1) {
		m[a[1]] = a[2]
	}
	return m
}

func xmlText(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			if r >= 0x20 || r == '\t' {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// setCell puts text into cell ref of a sheet's XML.
func setCell(x, ref, text string) (string, error) {
	col, row, ok := parseRef(ref)
	if !ok {
		return "", errors.New("not a cell reference")
	}
	rs := strconv.Itoa(row)
	// The row element.
	start, open := -1, ""
	for _, m := range reRowTag.FindAllStringIndex(x, -1) {
		if attrs(x[m[0]:m[1]])["r"] == rs {
			start, open = m[0], x[m[0]:m[1]]
			break
		}
	}
	if start < 0 {
		return "", errors.New("its row is not in the sheet")
	}
	prefix := reRowTag.FindStringSubmatch(open)[1]
	newCell := func(style string) string {
		s := ""
		if style != "" {
			s = ` s="` + style + `"`
		}
		return "<" + prefix + `c r="` + ref + `"` + s + ` t="inlineStr"><` + prefix + "is><" + prefix + "t>" +
			xmlText(text) + "</" + prefix + "t></" + prefix + "is></" + prefix + "c>"
	}
	bodyStart := start + len(open)
	if strings.HasSuffix(open, "/>") { // an empty row
		return x[:start] + strings.TrimSuffix(open, "/>") + ">" + newCell("") + "</" + prefix + "row>" + x[bodyStart:], nil
	}
	end := strings.Index(x[bodyStart:], "</"+prefix+"row>")
	if end < 0 {
		return "", errors.New("its row is not closed")
	}
	end += bodyStart
	body := x[bodyStart:end]
	for _, m := range reCellTag.FindAllStringSubmatchIndex(body, -1) {
		tag := body[m[0]:m[1]]
		a := attrs(tag)
		c, _, ok := parseRef(a["r"])
		if !ok {
			return "", errors.New("the sheet has cells without references; save it once in Excel")
		}
		cellEnd := m[1]
		if m[4] == m[5] { // not self-closing: find </c>
			i := strings.Index(body[m[1]:], "</"+prefix+"c>")
			if i < 0 {
				return "", errors.New("a cell in its row is not closed")
			}
			cellEnd = m[1] + i + len("</"+prefix+"c>")
		}
		switch {
		case c == col:
			if strings.Contains(body[m[0]:cellEnd], "<"+prefix+"f") {
				return "", errors.New("it holds a formula; it is left unchanged")
			}
			body = body[:m[0]] + newCell(a["s"]) + body[cellEnd:]
			return x[:bodyStart] + body + x[end:], nil
		case c > col:
			body = body[:m[0]] + newCell("") + body[m[0]:]
			return x[:start] + widenSpans(open, col) + body + x[end:], nil
		}
	}
	return x[:start] + widenSpans(open, col) + body + newCell("") + x[end:], nil
}

// widenSpans keeps a row's optional spans="first:last" hint covering col.
func widenSpans(open string, col int) string {
	sp, ok := attrs(open)["spans"]
	if !ok {
		return open
	}
	a, b, ok := strings.Cut(sp, ":")
	lo, err1 := strconv.Atoi(a)
	hi, err2 := strconv.Atoi(b)
	if !ok || err1 != nil || err2 != nil || (col >= lo && col <= hi) {
		return open
	}
	lo, hi = min(lo, col), max(hi, col)
	return strings.Replace(open, `spans="`+sp+`"`, `spans="`+strconv.Itoa(lo)+":"+strconv.Itoa(hi)+`"`, 1)
}

// CheckEdit compares two versions of a workbook: the cells in vals must
// read as given in after, and every other cell of every sheet must be
// unchanged.
func CheckEdit(before, after []byte, sheet string, vals map[string]string) error {
	bc, err := AllCells(before)
	if err != nil {
		return err
	}
	ac, err := AllCells(after)
	if err != nil {
		return fmt.Errorf("the edited workbook cannot be read: %w", err)
	}
	if len(bc) != len(ac) {
		return errors.New("the number of sheets changed")
	}
	for name, cells := range bc {
		got := map[string]string{}
		for _, c := range ac[name] {
			got[c.Ref] = c.Val
		}
		want := map[string]string{}
		for _, c := range cells {
			want[c.Ref] = c.Val
		}
		if name == sheet {
			for ref, v := range vals {
				want[ref] = v
			}
		}
		for ref, v := range want {
			if got[ref] != v {
				return fmt.Errorf("sheet %q cell %s reads %q, want %q", name, ref, got[ref], v)
			}
		}
		for ref, v := range got {
			if _, ok := want[ref]; !ok && v != "" {
				return fmt.Errorf("sheet %q cell %s appeared", name, ref)
			}
		}
	}
	return nil
}
