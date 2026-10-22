package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/docx"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

// dateSetup adds a filled form to the existing "Foo Ltda" folder.
func dateSetup(t *testing.T) (parent, sheetPath string) {
	t.Helper()
	parent, root, _ := setup(t)
	b, _, err := docx.Format{}.Fill(testfix.Template(t), form.Values{form.FieldName: "Foo Ltda", form.FieldDate: "16/10/2026"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Foo Ltda", "Appendix A - Risk Assesment - Foo Ltda.docx"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	return parent, filepath.Join(root, "3. Übersichten", "Kundenliste.xlsx")
}

func runDate(t *testing.T, parent string, answers ...string) string {
	t.Helper()
	var out bytes.Buffer
	code := Date(Options{Parent: parent, AllowNonTTY: true, Config: direct(t),
		In: strings.NewReader(strings.Join(answers, "\n") + "\n"), Out: &out, Now: fixedNow})
	if code != ExitOK {
		t.Fatalf("exit %d\n%s", code, out.String())
	}
	return out.String()
}

// Only the empty date cell of the company with a dated form changes.
func TestDateFillsOnlyTheDateCell(t *testing.T) {
	parent, path := dateSetup(t)
	before, _ := os.ReadFile(path)
	// "Fooo Inc" is not Foo Inc's folder; write; accept.
	out := runDate(t, parent, "n", "y", "a")
	after, _ := os.ReadFile(path)
	tb, err := sheet.ReadTable(after, "Customer Name")
	if err != nil {
		t.Fatal(err)
	}
	ref := ""
	for _, r := range tb.Rows {
		if r.Name == "Foo Ltda" {
			ref = r.DateRef
			if r.Date != "16-10-2026" {
				t.Errorf("Foo Ltda date %q\n%s", r.Date, out)
			}
		}
	}
	if err := sheet.CheckEdit(before, after, tb.Sheet, map[string]string{ref: "16-10-2026"}); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	ents, _ := os.ReadDir(filepath.Dir(path))
	if len(ents) != 1 {
		t.Errorf("files left next to the spreadsheet: %d", len(ents))
	}
	// A second run finds the date and writes nothing.
	out = runDate(t, parent, "n")
	if !strings.Contains(out, "Nothing to write") {
		t.Fatalf("second run:\n%s", out)
	}
	if h, _, _ := fsutil.SHA256File(path); h != fsutil.SHA256Bytes(after) {
		t.Error("second run changed the spreadsheet")
	}
}

func TestDateUndoAndDecline(t *testing.T) {
	parent, path := dateSetup(t)
	h0, _, _ := fsutil.SHA256File(path)
	runDate(t, parent, "n", "y", "u") // undo after writing
	if h, _, _ := fsutil.SHA256File(path); h != h0 {
		t.Fatal("undo did not restore the original")
	}
	runDate(t, parent, "n", "n") // decline before writing
	if h, _, _ := fsutil.SHA256File(path); h != h0 {
		t.Fatal("declined run changed the spreadsheet")
	}
}

func TestDateRefusesOpenSpreadsheet(t *testing.T) {
	parent, path := dateSetup(t)
	lock := filepath.Join(filepath.Dir(path), "~$"+filepath.Base(path))
	if err := os.WriteFile(lock, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if code := Date(Options{Parent: parent, AllowNonTTY: true, Config: direct(t), In: strings.NewReader(""), Out: &out, Now: fixedNow}); code != ExitError || !strings.Contains(out.String(), "open in Excel") {
		t.Fatalf("exit %d\n%s", code, out.String())
	}
}
