// Package sheet reads company names from the overview spreadsheet. It
// never writes: files are opened read-only and parsed in memory.
//
// Supported: .xlsx and .xlsm (macros are never run), .ods and .fods,
// .csv and .tsv, and, only when legacy formats are enabled, .xls.
// Apple .numbers files are converted by the caller first.
package sheet

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// maxPart caps how many bytes are read from a file or from one part of
// an archive, as a guard against zip bombs and runaway inputs.
const maxPart = 256 << 20

// Row is one company taken from the spreadsheet.
type Row struct {
	Sheet      string
	Cell       string // e.g. "C3"
	Name       string // literal cell text, trimmed of surrounding spaces
	CustomerID string // may be empty; never used for matching
}

// Cell is one non-empty cell. Col and Row start at 1.
type Cell struct {
	Col, Row int
	Ref      string
	Val      string
}

// Options configures reading.
type Options struct {
	Header   string // column header of the company names (default "Customer Name")
	IDHeader string // column header of the customer ids (default "Customer ID")
	Legacy   bool   // allow .xls
}

func (o Options) header() string {
	if o.Header == "" {
		return "Customer Name"
	}
	return o.Header
}

func (o Options) idHeader() string {
	if o.IDHeader == "" {
		return "Customer ID"
	}
	return o.IDHeader
}

// Extensions lists every spreadsheet extension Read understands.
var Extensions = []string{".xlsx", ".xlsm", ".ods", ".fods", ".csv", ".tsv", ".xls"}

// Supported reports whether ext (with dot, any case) can be read.
func Supported(ext string) bool {
	ext = strings.ToLower(ext)
	for _, e := range Extensions {
		if e == ext {
			return true
		}
	}
	return false
}

// IsLegacy reports whether ext is a legacy binary format.
func IsLegacy(ext string) bool { return strings.EqualFold(ext, ".xls") }

// ReadFile reads the companies from the file at p.
func ReadFile(p string, opt Options) ([]Row, []string, error) {
	f, err := os.Open(p) // #nosec G304 -- the user's own spreadsheet, read only.
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxPart+1))
	if err != nil {
		return nil, nil, err
	}
	if len(b) > maxPart {
		return nil, nil, fmt.Errorf("%s is larger than %d MB", filepath.Base(p), maxPart>>20)
	}
	return Read(b, filepath.Ext(p), opt)
}

// Read parses spreadsheet bytes of the given extension.
func Read(b []byte, ext string, opt Options) ([]Row, []string, error) {
	switch strings.ToLower(ext) {
	case ".xlsx", ".xlsm":
		return readXLSX(b, opt)
	case ".ods":
		return readODS(b, false, opt)
	case ".fods":
		return readODS(b, true, opt)
	case ".csv":
		return readDelimited(b, 0, opt)
	case ".tsv":
		return readDelimited(b, '\t', opt)
	case ".xls":
		if !opt.Legacy {
			return nil, nil, errors.New(".xls is a legacy format; an administrator must set legacy_formats = yes in the system configuration")
		}
		return readXLS(b, opt)
	}
	return nil, nil, fmt.Errorf("unsupported spreadsheet format %q", ext)
}

// normHeader folds case and spacing and drops a trailing colon or
// asterisk, so "Customer Name", "customer  name" and "Customer Name:" match.
func normHeader(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	s = strings.TrimRight(s, ":* ")
	return strings.ToLower(s)
}

func same(a, b string) bool { return normHeader(a) != "" && normHeader(a) == normHeader(b) }

// Extract finds the company-name header among cells and returns the
// non-empty cells below it, from top to bottom.
func Extract(sheet string, cells []Cell, opt Options) ([]Row, bool) {
	sort.SliceStable(cells, func(i, j int) bool {
		if cells[i].Row != cells[j].Row {
			return cells[i].Row < cells[j].Row
		}
		return cells[i].Col < cells[j].Col
	})
	hdrRow, nameCol, idCol := -1, -1, -1
	for _, c := range cells {
		if same(c.Val, opt.header()) {
			hdrRow, nameCol = c.Row, c.Col
			break
		}
	}
	if hdrRow < 0 {
		return nil, false
	}
	ids := map[int]string{}
	for _, c := range cells {
		if c.Row == hdrRow && same(c.Val, opt.idHeader()) {
			idCol = c.Col
		}
	}
	var out []Row
	for _, c := range cells {
		if c.Row <= hdrRow {
			continue
		}
		if c.Col == idCol {
			ids[c.Row] = strings.TrimSpace(c.Val)
		}
	}
	for _, c := range cells {
		if c.Col != nameCol || c.Row <= hdrRow {
			continue
		}
		name := strings.TrimSpace(c.Val)
		if name == "" {
			continue
		}
		out = append(out, Row{Sheet: sheet, Cell: c.Ref, Name: name, CustomerID: ids[c.Row]})
	}
	return out, true
}

// fromSheets returns the rows of the first sheet that has the header.
func fromSheets(names []string, cells [][]Cell, opt Options) ([]Row, error) {
	for i := range names {
		if rows, ok := Extract(names[i], cells[i], opt); ok {
			return rows, nil
		}
	}
	return nil, headerError(opt.header(), cells)
}

// headerError says which header was missing and lists the text cells of
// the first rows, which are usually the column headings, so the user can
// see what to set as header.
func headerError(want string, sheets [][]Cell) error {
	const maxRows, maxHints, maxLen = 3, 12, 40
	seen := map[string]bool{}
	var hints []string
	for _, cells := range sheets {
		sorted := append([]Cell(nil), cells...)
		sort.SliceStable(sorted, func(i, j int) bool {
			if sorted[i].Row != sorted[j].Row {
				return sorted[i].Row < sorted[j].Row
			}
			return sorted[i].Col < sorted[j].Col
		})
		rows, last := 0, -1
		for _, c := range sorted {
			if c.Row != last {
				if rows == maxRows {
					break
				}
				rows, last = rows+1, c.Row
			}
			v := strings.Join(strings.Fields(c.Val), " ")
			if v == "" || seen[v] || len(hints) == maxHints {
				continue
			}
			if _, err := strconv.ParseFloat(v, 64); err == nil {
				continue
			}
			if r := []rune(v); len(r) > maxLen {
				v = string(r[:maxLen]) + "..."
			}
			seen[v] = true
			hints = append(hints, strconv.Quote(v))
		}
		if len(hints) > 0 {
			break
		}
	}
	msg := fmt.Sprintf("no sheet has a %q header cell (the heading above the company names)", want)
	if len(hints) == 0 {
		return errors.New(msg)
	}
	return fmt.Errorf("%s.\nHeadings found at the top of the sheet: %s\nSave the right one once, for example:  numb-form config set header %s",
		msg, strings.Join(hints, ", "), hints[len(hints)-1])
}

// Ref builds an A1-style reference from 1-based column and row.
func Ref(col, row int) string { return colName(col) + strconv.Itoa(row) }
