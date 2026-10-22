package sheet

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

func TestReadTable(t *testing.T) {
	b := testfix.XLSX(t, fxIDs, fxNames)
	tb, err := ReadTable(b, "Customer Name")
	if err != nil {
		t.Fatal(err)
	}
	if tb.Sheet != "Table 1" || tb.NameHeader != "Customer Name" || tb.DateHeader != "Report created" || len(tb.Rows) != len(fxNames) {
		t.Fatalf("table %+v", tb)
	}
	if r := tb.Rows[0]; r.Name != fxNames[0] || r.NameRef != "C3" || r.DateRef != "F3" || r.Date != "Ignore at all costs" {
		t.Fatalf("row %+v", r)
	}
	// Another spelling of both headers.
	b2 := rezip(t, b, "Report created", "REPORT CREATED AB:")
	if tb, err := ReadTable(b2, "Kunde"); err != nil || tb.DateHeader != "REPORT CREATED AB:" {
		t.Fatalf("%v %+v", err, tb)
	}
	b3 := rezip(t, b, "Report created", "Currency 2")
	if _, err := ReadTable(b3, "Customer Name"); err == nil || !strings.Contains(err.Error(), "Report created") {
		t.Fatalf("missing date column: %v", err)
	}
}

// rezip replaces text in every entry of a zip archive.
func rezip(t *testing.T, b []byte, old, new string) []byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		d, _ := io.ReadAll(rc)
		rc.Close()
		w, _ := zw.Create(f.Name)
		_, _ = w.Write([]byte(strings.ReplaceAll(string(d), old, new)))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestSetCell(t *testing.T) {
	x := `<sheetData><row r="2" spans="2:3"><c r="B2" t="s"><v>0</v></c><c r="D2"><v>7</v></c></row>` +
		`<row r="3"><c r="B3" s="4"/><c r="C3" s="5" t="s"><v>1</v></c></row><row r="4"/>` +
		`<row r="5"><c r="C5"><f>A1</f><v>1</v></c></row></sheetData>`
	cases := []struct{ ref, want string }{
		{"C2", `<row r="2" spans="2:3"><c r="B2" t="s"><v>0</v></c><c r="C2" t="inlineStr"><is><t>16-10-2026</t></is></c><c r="D2">`},
		{"E2", `<row r="2" spans="2:5">`},
		{"B3", `<c r="B3" s="4" t="inlineStr"><is><t>16-10-2026</t></is></c><c r="C3" s="5" t="s">`},
		{"C3", `<c r="C3" s="5" t="inlineStr"><is><t>16-10-2026</t></is></c></row>`},
		{"A4", `<row r="4"><c r="A4" t="inlineStr"><is><t>16-10-2026</t></is></c></row>`},
	}
	for _, c := range cases {
		got, err := setCell(x, c.ref, "16-10-2026")
		if err != nil || !strings.Contains(got, c.want) {
			t.Errorf("%s: %v\n%s", c.ref, err, got)
		}
	}
	if _, err := setCell(x, "C5", "x"); err == nil {
		t.Error("formula cell overwritten")
	}
	if _, err := setCell(x, "A9", "x"); err == nil {
		t.Error("missing row accepted")
	}
	p := `<x:sheetData><x:row r="2"><x:c r="B2"/></x:row></x:sheetData>`
	if got, err := setCell(p, "B2", "a<b"); err != nil || !strings.Contains(got, `<x:c r="B2" t="inlineStr"><x:is><x:t>a&lt;b</x:t></x:is></x:c>`) {
		t.Errorf("prefixed: %v %s", err, got)
	}
}

// Only the target cells change; every other zip entry keeps its bytes.
func TestSetTextChangesOnlyTheCells(t *testing.T) {
	b := testfix.XLSX(t, fxIDs, fxNames)
	tb, err := ReadTable(b, "Customer Name")
	if err != nil {
		t.Fatal(err)
	}
	vals := map[string]string{"F3": "16-10-2026", "F5": "01-02-2026"}
	out, err := SetText(b, tb.Part, vals)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckEdit(b, out, tb.Sheet, vals); err != nil {
		t.Fatal(err)
	}
	if err := CheckEdit(b, out, tb.Sheet, map[string]string{"F3": "16-10-2026"}); err == nil {
		t.Error("an unexpected change was not noticed")
	}
	nt, err := ReadTable(out, "Customer Name")
	if err != nil || nt.Rows[0].Date != "16-10-2026" || nt.Rows[1].Date != "Ignore at all costs" || nt.Rows[2].Date != "01-02-2026" {
		t.Fatalf("%v %+v", err, nt)
	}
	entries := func(b []byte) map[string]string {
		zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{}
		for _, f := range zr.File {
			rc, _ := f.Open()
			d, _ := io.ReadAll(rc)
			rc.Close()
			m[f.Name] = string(d)
		}
		return m
	}
	be, ae := entries(b), entries(out)
	if len(be) != len(ae) {
		t.Fatalf("entries %d -> %d", len(be), len(ae))
	}
	for k, v := range be {
		if k != tb.Part && ae[k] != v {
			t.Errorf("%s changed", k)
		}
	}
}
