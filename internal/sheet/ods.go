package sheet

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	nsTable  = "urn:oasis:names:tc:opendocument:xmlns:table:1.0"
	nsText   = "urn:oasis:names:tc:opendocument:xmlns:text:1.0"
	nsOffice = "urn:oasis:names:tc:opendocument:xmlns:office:1.0"
)

// maxRepeat bounds row and column repetition counts. Empty filler rows
// are often repeated a million times; they are counted, never expanded.
const maxRepeat = 1 << 20

// readODS reads OpenDocument spreadsheets: content.xml inside an .ods
// archive, or the whole file for flat .fods.
func readODS(b []byte, flat bool, opt Options) ([]Row, []string, error) {
	content := b
	if !flat {
		zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
		if err != nil {
			return nil, nil, fmt.Errorf("open spreadsheet: %w", err)
		}
		var f *zip.File
		for _, e := range zr.File {
			if e.Name == "content.xml" {
				f = e
			}
		}
		if f == nil {
			return nil, nil, errors.New("not an OpenDocument spreadsheet (content.xml missing)")
		}
		if content, err = readPart(f); err != nil {
			return nil, nil, err
		}
	}
	names, cells, err := odsCells(content)
	if err != nil {
		return nil, nil, err
	}
	rows, err := fromSheets(names, cells, opt)
	return rows, nil, err
}

func repeat(attrs []xml.Attr, name string) int {
	for _, a := range attrs {
		if a.Name.Space == nsTable && a.Name.Local == name {
			if n, err := strconv.Atoi(a.Value); err == nil && n > 0 {
				return min(n, maxRepeat)
			}
		}
	}
	return 1
}

func odsCells(content []byte) ([]string, [][]Cell, error) {
	dec := xml.NewDecoder(bytes.NewReader(content))
	var (
		names   []string
		sheets  [][]Cell
		row     int
		col     int
		inCell  bool
		cellRep int
		paras   int
		buf     strings.Builder
		pending int // extra copies of the current row
		skip    int // depth inside cell comments (office:annotation)
	)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("spreadsheet xml: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Space == nsOffice && t.Name.Local == "annotation" || skip > 0 {
				skip++
				continue
			}
			switch {
			case t.Name.Space == nsTable && t.Name.Local == "table":
				name := ""
				for _, a := range t.Attr {
					if a.Name.Space == nsTable && a.Name.Local == "name" {
						name = a.Value
					}
				}
				names = append(names, name)
				sheets = append(sheets, nil)
				row = 0
			case t.Name.Space == nsTable && t.Name.Local == "table-row":
				row++
				col = 0
				// The cells are recorded for the first copy of a repeated
				// row; the counter then skips the copies.
				pending = repeat(t.Attr, "number-rows-repeated") - 1
			case t.Name.Space == nsTable && (t.Name.Local == "table-cell" || t.Name.Local == "covered-table-cell"):
				inCell, cellRep, paras = true, repeat(t.Attr, "number-columns-repeated"), 0
				buf.Reset()
			case inCell && t.Name.Space == nsText && t.Name.Local == "p":
				if paras > 0 {
					buf.WriteByte('\n')
				}
				paras++
			case inCell && t.Name.Space == nsText && t.Name.Local == "s":
				n := 1
				for _, a := range t.Attr {
					if a.Name.Space == nsText && a.Name.Local == "c" {
						if v, err := strconv.Atoi(a.Value); err == nil && v > 0 && v < 1000 {
							n = v
						}
					}
				}
				buf.WriteString(strings.Repeat(" ", n))
			case inCell && t.Name.Space == nsText && (t.Name.Local == "tab" || t.Name.Local == "line-break"):
				buf.WriteByte(' ')
			}
		case xml.EndElement:
			if skip > 0 {
				skip--
				continue
			}
			if t.Name.Space == nsTable && t.Name.Local == "table-row" {
				row += pending
				pending = 0
			}
			if t.Name.Space == nsTable && (t.Name.Local == "table-cell" || t.Name.Local == "covered-table-cell") {
				v := buf.String()
				if strings.TrimSpace(v) != "" && len(sheets) > 0 {
					reps := min(cellRep, 64) // a name repeated across columns is still one name per column
					for k := 0; k < reps; k++ {
						c := col + 1 + k
						sheets[len(sheets)-1] = append(sheets[len(sheets)-1], Cell{Col: c, Row: row, Ref: Ref(c, row), Val: v})
					}
				}
				col += cellRep
				inCell = false
			}
		case xml.CharData:
			if inCell && paras > 0 && skip == 0 {
				buf.Write(t)
			}
		}
	}
	return names, sheets, nil
}
