package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

// numb-typo corrects only the place parts of names without a folder, and
// only the name cells change.
func TestTypo(t *testing.T) {
	root := filepath.Join(t.TempDir(), "GWG")
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	names := []string{"Bar S.A. de C.V., Ciudad de Me", "Baz SAS, Bogta", "Qux Ltda, Medelin", "Quux GmbH"}
	must(os.MkdirAll(filepath.Join(root, "3. Übersichten"), 0o750))
	must(os.MkdirAll(filepath.Join(root, "Qux Ltda, Medelin"), 0o750)) // has a folder: left alone
	sp := filepath.Join(root, "3. Übersichten", "Kundenliste.xlsx")
	must(os.WriteFile(sp, testfix.XLSX(t, []string{"D1", "D2", "D3", "D4"}, names), 0o600))
	must(os.WriteFile(filepath.Join(root, "Appendix A - Risk Assesment.docx"), testfix.Template(t), 0o600))
	before, _ := os.ReadFile(sp)
	var out bytes.Buffer
	code := Typo(Options{Parent: root, AllowNonTTY: true, Config: direct(t), In: strings.NewReader("a\na\n"), Out: &out, Now: fixedNow})
	if code != ExitOK {
		t.Fatalf("exit %d\n%s", code, out.String())
	}
	after, _ := os.ReadFile(sp)
	tb, err := sheet.ReadNames(after, "Customer Name", "Customer ID")
	must(err)
	want := []string{"Bar S.A. de C.V., Ciudad de México", "Baz SAS, Bogotá", "Qux Ltda, Medelin", "Quux GmbH"}
	vals := map[string]string{}
	for i, r := range tb.Rows {
		if r.Name != want[i] {
			t.Errorf("row %d = %q, want %q\n%s", i, r.Name, want[i], out.String())
		}
		if r.Name != names[i] {
			vals[r.NameRef] = r.Name
		}
	}
	if err := sheet.CheckEdit(before, after, tb.Sheet, vals); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "1 company name(s) with a possible typo already have a folder") {
		t.Errorf("folder note missing:\n%s", out.String())
	}
}
