package xlsx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

func TestReadCompanies(t *testing.T) {
	names := []string{"Foo C.A. de C.V.", "Foo Co., Ltda", "Bär & Söhne GmbH", "  Spaced Inc  "}
	ids := []string{"D1", "D2", "D3", "D4"}
	p := filepath.Join(t.TempDir(), "book.xlsx")
	if err := os.WriteFile(p, testfix.XLSX(t, ids, names), 0o600); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(p)
	rows, _, err := ReadCompanies(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != len(names) {
		t.Fatalf("got %d rows: %+v", len(rows), rows)
	}
	want := []string{"Foo C.A. de C.V.", "Foo Co., Ltda", "Bär & Söhne GmbH", "Spaced Inc"}
	for i, r := range rows {
		if r.Name != want[i] || r.CustomerID != ids[i] || r.Cell != "C"+string(rune('3'+i)) {
			t.Errorf("row %d = %+v", i, r)
		}
	}
	after, _ := os.ReadFile(p)
	if string(before) != string(after) {
		t.Fatal("workbook modified")
	}
}

func TestParseRef(t *testing.T) {
	for ref, want := range map[string][2]int{"A1": {1, 1}, "Z9": {26, 9}, "AA10": {27, 10}, "XFD1048576": {16384, 1048576}} {
		c, r, ok := parseRef(ref)
		if !ok || c != want[0] || r != want[1] {
			t.Errorf("parseRef(%q) = %d, %d, %v", ref, c, r, ok)
		}
		if colName(c) != ref[:len(ref)-len(itoa(r))] {
			t.Errorf("colName(%d) = %q", c, colName(c))
		}
	}
}

func itoa(n int) string {
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func FuzzReadCompanies(f *testing.F) {
	f.Add(testfix.XLSX(f, []string{"D1"}, []string{"Foo"}))
	f.Add([]byte("PK\x03\x04garbage"))
	f.Fuzz(func(t *testing.T, b []byte) {
		p := filepath.Join(t.TempDir(), "f.xlsx")
		if err := os.WriteFile(p, b, 0o600); err != nil {
			t.Fatal(err)
		}
		_, _, _ = ReadCompanies(p) // must not panic
	})
}
