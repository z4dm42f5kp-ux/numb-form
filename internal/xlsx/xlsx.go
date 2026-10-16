// Package xlsx reads company names from an .xlsx workbook. It never
// writes: the file is opened read-only and only parsed in memory.
package xlsx

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
)

// maxPart caps how many decompressed bytes are read from one workbook
// part, as a guard against zip bombs.
const maxPart = 256 << 20

// Row is one company taken from the workbook.
type Row struct {
	Sheet      string
	Cell       string // e.g. "C3"
	Name       string // literal cell text, trimmed of surrounding spaces
	CustomerID string // may be empty; never used for matching
}

// HeaderCompany is the header text that marks the company column.
const HeaderCompany = "company name"

// HeaderCustomer is the header text that marks the customer id column.
const HeaderCustomer = "customer id"

type cell struct {
	col, row int
	ref      string
	val      string
}

// ReadCompanies opens the workbook at p and returns the companies listed
// below the "Company Name" header, from top to bottom.
func ReadCompanies(p string) ([]Row, []string, error) {
	zr, err := zip.OpenReader(p)
	if err != nil {
		return nil, nil, fmt.Errorf("open workbook: %w", err)
	}
	defer zr.Close()
	files := map[string]*zip.File{}
	for _, f := range zr.File {
		files[f.Name] = f
	}
	var warnings []string

	shared, err := readShared(files["xl/sharedStrings.xml"])
	if err != nil {
		return nil, nil, err
	}
	sheets, err := readSheets(files)
	if err != nil {
		return nil, nil, err
	}
	for _, sh := range sheets {
		f := files[sh.target]
		if f == nil {
			warnings = append(warnings, fmt.Sprintf("sheet %q: part %s missing", sh.name, sh.target))
			continue
		}
		cells, err := readCells(f, shared)
		if err != nil {
			return nil, nil, fmt.Errorf("sheet %q: %w", sh.name, err)
		}
		rows, ok := extract(sh.name, cells)
		if !ok {
			continue
		}
		return rows, warnings, nil
	}
	return nil, warnings, errors.New(`no sheet has a "Company Name" header cell`)
}

func extract(sheet string, cells []cell) ([]Row, bool) {
	sort.Slice(cells, func(i, j int) bool {
		if cells[i].row != cells[j].row {
			return cells[i].row < cells[j].row
		}
		return cells[i].col < cells[j].col
	})
	hdrRow, nameCol, idCol := -1, -1, -1
	for _, c := range cells {
		if strings.EqualFold(strings.Join(strings.Fields(c.val), " "), HeaderCompany) {
			hdrRow, nameCol = c.row, c.col
			break
		}
	}
	if hdrRow < 0 {
		return nil, false
	}
	for _, c := range cells {
		if c.row == hdrRow && strings.EqualFold(strings.Join(strings.Fields(c.val), " "), HeaderCustomer) {
			idCol = c.col
		}
	}
	ids := map[int]string{}
	if idCol >= 0 {
		for _, c := range cells {
			if c.col == idCol && c.row > hdrRow {
				ids[c.row] = strings.TrimSpace(c.val)
			}
		}
	}
	var out []Row
	for _, c := range cells {
		if c.col != nameCol || c.row <= hdrRow {
			continue
		}
		name := strings.TrimSpace(c.val)
		if name == "" {
			continue
		}
		out = append(out, Row{Sheet: sheet, Cell: c.ref, Name: name, CustomerID: ids[c.row]})
	}
	return out, true
}

type sheetRef struct{ name, target string }

func readPart(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	b, err := io.ReadAll(io.LimitReader(rc, maxPart+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxPart {
		return nil, fmt.Errorf("%s is larger than %d bytes", f.Name, maxPart)
	}
	return b, nil
}

func readSheets(files map[string]*zip.File) ([]sheetRef, error) {
	wb := files["xl/workbook.xml"]
	rels := files["xl/_rels/workbook.xml.rels"]
	if wb == nil || rels == nil {
		return nil, errors.New("not an .xlsx workbook (xl/workbook.xml missing)")
	}
	rb, err := readPart(rels)
	if err != nil {
		return nil, err
	}
	var r struct {
		Rel []struct {
			ID     string `xml:"Id,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if err := xml.Unmarshal(rb, &r); err != nil {
		return nil, fmt.Errorf("workbook rels: %w", err)
	}
	targets := map[string]string{}
	for _, x := range r.Rel {
		t := x.Target
		if strings.HasPrefix(t, "/") {
			t = strings.TrimPrefix(t, "/")
		} else {
			t = path.Join("xl", t)
		}
		targets[x.ID] = t
	}
	wbb, err := readPart(wb)
	if err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(strings.NewReader(string(wbb)))
	var out []sheetRef
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("workbook: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "sheet" {
			continue
		}
		var name, id string
		for _, a := range se.Attr {
			switch {
			case a.Name.Local == "name":
				name = a.Value
			case a.Name.Local == "id" && strings.Contains(a.Name.Space, "relationships"):
				id = a.Value
			}
		}
		out = append(out, sheetRef{name: name, target: targets[id]})
	}
	return out, nil
}

func readShared(f *zip.File) ([]string, error) {
	if f == nil {
		return nil, nil
	}
	b, err := readPart(f)
	if err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(strings.NewReader(string(b)))
	var out []string
	var cur strings.Builder
	inSI, inT, inRPh := false, false, false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("shared strings: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "si":
				inSI = true
				cur.Reset()
			case "rPh":
				inRPh = true
			case "t":
				inT = inSI && !inRPh
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "si":
				out = append(out, cur.String())
				inSI = false
			case "rPh":
				inRPh = false
			case "t":
				inT = false
			}
		case xml.CharData:
			if inT {
				cur.Write(t)
			}
		}
	}
	return out, nil
}

func readCells(f *zip.File, shared []string) ([]cell, error) {
	b, err := readPart(f)
	if err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(strings.NewReader(string(b)))
	var cells []cell
	var cur *cell
	var typ string
	var buf strings.Builder
	inV, inIS, inT := false, false, false
	row, col := 0, 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "row":
				row++
				col = 0
				for _, a := range t.Attr {
					if a.Name.Local == "r" {
						if n, err := strconv.Atoi(a.Value); err == nil {
							row = n
						}
					}
				}
			case "c":
				col++
				typ = ""
				ref := ""
				for _, a := range t.Attr {
					switch a.Name.Local {
					case "r":
						ref = a.Value
					case "t":
						typ = a.Value
					}
				}
				c, r := col, row
				if ref != "" {
					if cc, rr, ok := parseRef(ref); ok {
						c, r = cc, rr
						col = cc
					}
				} else {
					ref = colName(c) + strconv.Itoa(r)
				}
				cur = &cell{col: c, row: r, ref: ref}
				buf.Reset()
			case "v":
				inV = cur != nil
			case "is":
				inIS = cur != nil
			case "t":
				inT = inIS
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "v":
				inV = false
			case "is":
				inIS = false
			case "t":
				inT = false
			case "c":
				if cur != nil {
					raw := buf.String()
					if typ == "s" {
						if i, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil && i >= 0 && i < len(shared) {
							raw = shared[i]
						} else {
							raw = ""
						}
					}
					cur.val = raw
					cells = append(cells, *cur)
				}
				cur = nil
			}
		case xml.CharData:
			if inV || inT {
				buf.Write(t)
			}
		}
	}
	return cells, nil
}

func parseRef(ref string) (col, row int, ok bool) {
	i := 0
	for i < len(ref) && ref[i] >= 'A' && ref[i] <= 'Z' {
		col = col*26 + int(ref[i]-'A'+1)
		i++
	}
	if i == 0 || i == len(ref) {
		return 0, 0, false
	}
	r, err := strconv.Atoi(ref[i:])
	if err != nil {
		return 0, 0, false
	}
	return col, r, true
}

func colName(c int) string {
	s := ""
	for c > 0 {
		c--
		s = string(rune('A'+c%26)) + s
		c /= 26
	}
	return s
}
