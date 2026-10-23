package sheet

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Table is the company list of an .xlsx workbook with its report-date
// column, as numb-date needs it.
type Table struct {
	Sheet      string // sheet name
	Part       string // zip path of the sheet's XML
	HeaderRow  int
	NameCol    int
	DateCol    int
	NameHeader string // the header cells as written
	DateHeader string
	Rows       []TableRow
}

// TableRow is one company with its report-date cell.
type TableRow struct {
	Row     int
	Name    string
	NameRef string
	DateRef string
	Date    string // current text of the date cell, "" when empty
	ID      string // customer id from the idHeader column, "" when none
}

// NameHeaders are tried, in order, to find the company column after the
// configured header.
var NameHeaders = []string{"Customer Name", "Customer", "Company Name", "Company", "Client Name", "Client",
	"Kundenname", "Kunde", "Firmenname", "Firma", "Nombre del cliente", "Cliente", "Razón social", "Empresa"}

// DateHeaders are tried, in order, to find the report-date column.
var DateHeaders = []string{"Report created AB", "Report created at", "Report created on", "Report created",
	"Report creation date", "Date report created", "Report date", "Created", "Report erstellt am", "Report erstellt",
	"Bericht erstellt am", "Bericht erstellt", "Fecha del informe", "Fecha de creación del informe", "Informe creado"}

// looseDate accepts a header such as "Report-Created (AB)": it mentions a
// report and its creation or date.
func looseDate(h string) bool {
	h = normHeader(h)
	report := strings.Contains(h, "report") || strings.Contains(h, "bericht") || strings.Contains(h, "informe")
	when := strings.Contains(h, "creat") || strings.Contains(h, "date") || strings.Contains(h, "erstellt") ||
		strings.Contains(h, "fecha") || strings.Contains(h, "datum")
	return report && when
}

func looseName(h string) bool {
	h = normHeader(h)
	for _, w := range []string{"customer", "company", "client", "kunde", "firma", "cliente", "empresa"} {
		if strings.Contains(h, w) {
			return true
		}
	}
	return false
}

// findHeader returns the first cell matching one of the headers exactly,
// in the order of headers, then one matching loose.
func findHeader(cells []Cell, headers []string, loose func(string) bool) (Cell, bool) {
	for _, h := range headers {
		for _, c := range cells {
			if same(c.Val, h) {
				return c, true
			}
		}
	}
	for _, c := range cells {
		if loose(c.Val) {
			return c, true
		}
	}
	return Cell{}, false
}

func sortCells(cells []Cell) {
	sort.SliceStable(cells, func(i, j int) bool {
		if cells[i].Row != cells[j].Row {
			return cells[i].Row < cells[j].Row
		}
		return cells[i].Col < cells[j].Col
	})
}

type workbook struct {
	files  map[string]*zip.File
	sheets []sheetRef
	shared []string
}

func openWorkbook(b []byte) (*workbook, error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, fmt.Errorf("open workbook: %w", err)
	}
	wb := &workbook{files: map[string]*zip.File{}}
	for _, f := range zr.File {
		wb.files[f.Name] = f
	}
	if wb.shared, err = readShared(wb.files["xl/sharedStrings.xml"]); err != nil {
		return nil, err
	}
	if wb.sheets, err = readSheets(wb.files); err != nil {
		return nil, err
	}
	return wb, nil
}

// ReadTable finds, in an .xlsx or .xlsm workbook, the first sheet with a
// company-name header (header first, then NameHeaders) and a report-date
// header in the same row, and returns every company below it with its
// customer id from the idHeader column.
func ReadTable(b []byte, header, idHeader string) (*Table, error) {
	wb, err := openWorkbook(b)
	if err != nil {
		return nil, err
	}
	names := append([]string{header}, NameHeaders...)
	var found []string
	for _, sh := range wb.sheets {
		f := wb.files[sh.target]
		if f == nil {
			continue
		}
		cells, err := readCells(f, wb.shared)
		if err != nil {
			return nil, fmt.Errorf("sheet %q: %w", sh.name, err)
		}
		sortCells(cells)
		nh, ok := findHeader(cells, names, looseName)
		if !ok {
			continue
		}
		var hdr []Cell
		for _, c := range cells {
			if c.Row == nh.Row && c.Col != nh.Col {
				hdr = append(hdr, c)
			}
		}
		dh, ok := findHeader(hdr, DateHeaders, looseDate)
		if !ok {
			for _, c := range hdr {
				if v := strings.TrimSpace(c.Val); v != "" {
					found = append(found, fmt.Sprintf("%q", v))
				}
			}
			continue
		}
		t := &Table{Sheet: sh.name, Part: sh.target, HeaderRow: nh.Row, NameCol: nh.Col, DateCol: dh.Col,
			NameHeader: strings.TrimSpace(nh.Val), DateHeader: strings.TrimSpace(dh.Val)}
		idCol := -1
		for _, c := range hdr {
			if same(c.Val, idHeader) {
				idCol = c.Col
			}
		}
		dates, ids := map[int]string{}, map[int]string{}
		for _, c := range cells {
			switch {
			case c.Row <= nh.Row:
			case c.Col == dh.Col:
				dates[c.Row] = strings.TrimSpace(c.Val)
			case c.Col == idCol:
				ids[c.Row] = strings.TrimSpace(c.Val)
			}
		}
		for _, c := range cells {
			if c.Col != nh.Col || c.Row <= nh.Row || strings.TrimSpace(c.Val) == "" {
				continue
			}
			t.Rows = append(t.Rows, TableRow{Row: c.Row, Name: strings.TrimSpace(c.Val), NameRef: c.Ref,
				DateRef: Ref(dh.Col, c.Row), Date: dates[c.Row], ID: ids[c.Row]})
		}
		return t, nil
	}
	if len(found) > 0 {
		return nil, fmt.Errorf("no \"Report created\" column next to the company names; headings found: %s", strings.Join(found, ", "))
	}
	return nil, errors.New("no sheet has a company-name header (such as \"Customer Name\") with a \"Report created\" column")
}

// AllCells returns every cell of every sheet, keyed by sheet name, for
// comparing a workbook before and after an edit.
func AllCells(b []byte) (map[string][]Cell, error) {
	wb, err := openWorkbook(b)
	if err != nil {
		return nil, err
	}
	out := map[string][]Cell{}
	for _, sh := range wb.sheets {
		f := wb.files[sh.target]
		if f == nil {
			continue
		}
		cells, err := readCells(f, wb.shared)
		if err != nil {
			return nil, fmt.Errorf("sheet %q: %w", sh.name, err)
		}
		sortCells(cells)
		out[sh.name] = cells
	}
	return out, nil
}
