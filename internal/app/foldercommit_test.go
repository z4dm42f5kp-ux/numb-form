package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/docx"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

// workingCopy copies the real folder to Downloads, adds a filled form for
// Foo Ltd, writes its report date into the spreadsheet (as numb-date
// does), changes the template and leaves residue that must stay behind.
func workingCopy(t *testing.T, edit func(sheetBytes []byte, tb *sheet.Table) map[string]string) (real, cp string) {
	t.Helper()
	_, real, dl := setup(t)
	cp = filepath.Join(dl, filepath.Base(real))
	if _, err := fsutil.CopyTree(context.Background(), real, cp, func(string, ...any) {}); err != nil {
		t.Fatal(err)
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	doc, _, err := docx.Format{}.Fill(testfix.Template(t), form.Values{form.FieldName: "Foo Ltd", form.FieldLegalForm: "Ltd",
		form.FieldAddress: "London", form.FieldRegistry: RegistryText, form.FieldReps: NoRepsText, form.FieldDate: "16/10/2026", form.FieldPerson: "P"})
	must(err)
	must(os.Mkdir(filepath.Join(cp, "Foo Ltd"), 0o750))
	must(os.WriteFile(filepath.Join(cp, "Foo Ltd", "Appendix A - Risk Assesment - Foo Ltd.docx"), doc, 0o600))
	must(os.WriteFile(filepath.Join(cp, "Foo Ltd", "~$pendix A.docx"), []byte("lock"), 0o600))
	must(os.WriteFile(filepath.Join(cp, "numb-form.log"), []byte("log"), 0o600))
	must(os.WriteFile(filepath.Join(cp, "Appendix A - Risk Assesment.docx"), []byte("changed template"), 0o600))
	sp := filepath.Join(cp, "3. Übersichten", "Kundenliste.xlsx")
	b, err := os.ReadFile(sp)
	must(err)
	tb, err := sheet.ReadTable(b, "Customer Name", "Customer ID")
	must(err)
	nb, err := sheet.SetText(b, tb.Part, edit(b, tb))
	must(err)
	must(os.WriteFile(sp, nb, 0o600))
	return real, cp
}

func dateOf(name string) func([]byte, *sheet.Table) map[string]string {
	return func(_ []byte, tb *sheet.Table) map[string]string {
		for _, r := range tb.Rows {
			if r.Name == name {
				return map[string]string{r.DateRef: "16-10-2026"}
			}
		}
		return nil
	}
}

func runCommit(t *testing.T, cp, real, answers string, review bool) (int, string) {
	t.Helper()
	cfg := config.Defaults()
	cfg.LogDir = filepath.Join(t.TempDir(), "logs")
	var out bytes.Buffer
	code := Commit(CommitOptions{Parent: cp, To: real, Review: review, Config: &cfg, In: strings.NewReader(answers), Out: &out})
	return code, out.String()
}

func TestFolderCommit(t *testing.T) {
	real, cp := workingCopy(t, dateOf("Foo Ltd"))
	tplBefore, _, _ := fsutil.SHA256File(filepath.Join(real, "Appendix A - Risk Assesment.docx"))
	code, out := runCommit(t, cp, real, "y\n", false)
	if code != ExitOK || !strings.Contains(out, "Verdict: OK to commit") {
		t.Fatalf("exit %d\n%s", code, out)
	}
	for _, f := range []string{"Foo Ltd/Appendix A - Risk Assesment - Foo Ltd.docx", "3. Übersichten/Kundenliste.xlsx"} {
		a, _, _ := fsutil.SHA256File(filepath.Join(cp, filepath.FromSlash(f)))
		b, _, _ := fsutil.SHA256File(filepath.Join(real, filepath.FromSlash(f)))
		if a != b {
			t.Errorf("%s not committed", f)
		}
	}
	if h, _, _ := fsutil.SHA256File(filepath.Join(real, "Appendix A - Risk Assesment.docx")); h != tplBefore {
		t.Error("the template was committed")
	}
	for _, f := range []string{"numb-form.log", "Foo Ltd/~$pendix A.docx", RequestName} {
		if _, err := os.Stat(filepath.Join(real, filepath.FromSlash(f))); !os.IsNotExist(err) {
			t.Errorf("%s was committed", f)
		}
	}
	req, _ := os.ReadFile(filepath.Join(cp, RequestName))
	if !strings.Contains(string(req), "Status: applied") || !strings.Contains(string(req), "New folder  Foo Ltd") {
		t.Errorf("request:\n%s", req)
	}
	// Reviewing afterwards checks the result.
	if code, out := runCommit(t, cp, real, "", true); code != ExitOK || !strings.Contains(out, "every listed file is in the real folder") {
		t.Fatalf("review after commit: %d\n%s", code, out)
	}
}

// A spreadsheet change outside the report dates fails the review, and
// nothing is copied.
func TestFolderCommitRefusesOtherSheetChanges(t *testing.T) {
	real, cp := workingCopy(t, func(_ []byte, tb *sheet.Table) map[string]string {
		return map[string]string{tb.Rows[0].NameRef: "Renamed GmbH"}
	})
	before := treeHashes(t, real)
	code, out := runCommit(t, cp, real, "y\n", false)
	if code != ExitError || !strings.Contains(out, "only report dates may change") {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if after := treeHashes(t, real); len(after) != len(before) {
		t.Fatal("the real folder changed")
	}
}
