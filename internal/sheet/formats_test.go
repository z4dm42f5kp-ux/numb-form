package sheet

import (
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

var (
	fxNames = []string{"Foo C.A. de C.V.", "Foo Co., Ltda", "Bär & Söhne GmbH", "Łódź Sp. z o.o. – Zażółć"}
	fxIDs   = []string{"D1", "D2", "D3", "D4"}
)

func check(t *testing.T, label string, rows []Row, err error, want []string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", label, err)
	}
	if len(rows) != len(want) {
		t.Fatalf("%s: got %d rows %+v", label, len(rows), rows)
	}
	for i, r := range rows {
		if r.Name != want[i] {
			t.Errorf("%s row %d: %q, want %q", label, i, r.Name, want[i])
		}
		if r.CustomerID != fxIDs[i] {
			t.Errorf("%s row %d: id %q", label, i, r.CustomerID)
		}
	}
}

func TestODSAndFODS(t *testing.T) {
	rows, _, err := Read(testfix.ODS(t, fxIDs, fxNames), ".ods", Options{})
	check(t, "ods", rows, err, fxNames)
	if rows[0].Cell != "C3" {
		t.Errorf("ods cell %q, want C3", rows[0].Cell)
	}
	rows, _, err = Read(testfix.FODS(fxIDs, fxNames), ".fods", Options{})
	check(t, "fods", rows, err, fxNames)
}

func TestDelimited(t *testing.T) {
	rows, _, err := Read(testfix.CSV(fxIDs, fxNames, ";", false), ".csv", Options{})
	check(t, "csv utf-8 ;", rows, err, fxNames)
	rows, _, err = Read(testfix.CSV(fxIDs, fxNames, "\t", false), ".tsv", Options{})
	check(t, "tsv", rows, err, fxNames)
	latin := []string{"Foo C.A. de C.V.", "Foo Co., Ltda", "Bär & Söhne GmbH", "?ód? Sp. z o.o. ? Za?ó??"}
	rows, warn, err := Read(testfix.CSV(fxIDs, fxNames, ",", true), ".csv", Options{})
	check(t, "csv cp1252 ,", rows, err, latin)
	if len(warn) == 0 {
		t.Error("expected an encoding warning")
	}
}

func TestXLS(t *testing.T) {
	b := testfix.XLS(t, fxIDs, fxNames)
	if _, _, err := Read(b, ".xls", Options{}); err == nil {
		t.Fatal(".xls read without legacy_formats")
	}
	rows, _, err := Read(b, ".xls", Options{Legacy: true})
	check(t, "xls", rows, err, fxNames)
}

func TestCustomHeader(t *testing.T) {
	csv := []byte("Kunde;Firmenname\nD1;Foo GmbH\n")
	rows, _, err := Read(csv, ".csv", Options{Header: "Firmenname", IDHeader: "Kunde"})
	if err != nil || len(rows) != 1 || rows[0].Name != "Foo GmbH" || rows[0].CustomerID != "D1" {
		t.Fatalf("rows %+v err %v", rows, err)
	}
}

func FuzzRead(f *testing.F) {
	f.Add(testfix.XLS(f, fxIDs, fxNames), ".xls")
	f.Add(testfix.ODS(f, fxIDs, fxNames), ".ods")
	f.Add(testfix.FODS(fxIDs, fxNames), ".fods")
	f.Add(testfix.CSV(fxIDs, fxNames, ";", true), ".csv")
	f.Add(testfix.XLSX(f, fxIDs, fxNames), ".xlsx")
	f.Fuzz(func(t *testing.T, b []byte, ext string) {
		_, _, _ = Read(b, ext, Options{Legacy: true}) // must not panic or hang
	})
}
