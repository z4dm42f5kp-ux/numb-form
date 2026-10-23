package app

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/docx"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

// scanCopy copies the real folder to Downloads and changes the copy as a
// run and a person would: a form filled by numb-form for Foo Ltd, a form
// edited by hand for Bar Ltd, a report date in the spreadsheet, a changed
// existing file, a changed template and residue.
func scanCopy(t *testing.T) (real, cp string) {
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
	fill := func(name string) []byte {
		doc, _, err := docx.Format{}.Fill(testfix.Template(t), form.Values{form.FieldName: name, form.FieldLegalForm: "Ltd",
			form.FieldAddress: "London", form.FieldRegistry: RegistryText, form.FieldReps: NoRepsText, form.FieldDate: "16/10/2026", form.FieldPerson: "P"})
		must(err)
		return doc
	}
	edited := fill("Bar Ltd")
	zr, err := zip.NewReader(bytes.NewReader(edited), int64(len(edited)))
	must(err)
	x, err := form.ZipEntry(zr, "word/document.xml")
	must(err)
	edited, err = form.ReplaceZipEntry(edited, "word/document.xml", append(x, '\n'))
	must(err)
	for name, doc := range map[string][]byte{"Foo Ltd": fill("Foo Ltd"), "Bar Ltd": edited} {
		must(os.Mkdir(filepath.Join(cp, name), 0o750))
		must(os.WriteFile(filepath.Join(cp, name, "Appendix A - Risk Assesment - "+name+".docx"), doc, 0o600))
	}
	must(os.WriteFile(filepath.Join(cp, "Foo Ltd", "~$pendix A.docx"), []byte("lock"), 0o600))
	must(os.WriteFile(filepath.Join(cp, "numb-form.log"), []byte("log"), 0o600))
	must(os.WriteFile(filepath.Join(cp, "Appendix A - Risk Assesment.docx"), []byte("changed template"), 0o600))
	must(os.WriteFile(filepath.Join(cp, "Foo GmbH", "keep.txt"), []byte("changed in the copy"), 0o600))
	sp := filepath.Join(cp, "3. Übersichten", "Kundenliste.xlsx")
	b, err := os.ReadFile(sp)
	must(err)
	tb, err := sheet.ReadTable(b, "Customer Name", "Customer ID")
	must(err)
	nb, err := sheet.SetText(b, tb.Part, map[string]string{tb.Rows[0].DateRef: "16-10-2026"})
	must(err)
	must(os.WriteFile(sp, nb, 0o600))
	return real, cp
}

// The scan adds new files, replaces only the spreadsheet, never copies
// residue or the template and never overwrites a changed file. It tells
// numb-form's unedited forms from edited ones.
func TestScanRequest(t *testing.T) {
	real, cp := scanCopy(t)
	ctx := context.Background()
	rc, err := reviewSetup(ctx, cp, real, config.Defaults())
	if err != nil {
		t.Fatal(err)
	}
	q, err := scanRequest(ctx, cp, real, rc)
	if err != nil {
		t.Fatal(err)
	}
	type want struct{ kind, origin, top string }
	wants := map[string]want{
		"Foo Ltd/Appendix A - Risk Assesment - Foo Ltd.docx": {kAdd, byNumbForm, "Foo Ltd"},
		"Bar Ltd/Appendix A - Risk Assesment - Bar Ltd.docx": {kAdd, editedForm, "Bar Ltd"},
		"3. Übersichten/Kundenliste.xlsx":                    {kReplace, "", ""},
		"Foo GmbH/keep.txt":                                  {kDiffers, otherFile, ""},
		"Appendix A - Risk Assesment.docx":                   {kKeep, "", ""},
		"numb-form.log":                                      {kKeep, "", ""},
		"Foo Ltd/~$pendix A.docx":                            {kKeep, "", ""},
	}
	for _, c := range q.Changes {
		w, ok := wants[c.rel]
		if !ok {
			t.Errorf("unexpected %s %s", c.kind, c.rel)
			continue
		}
		delete(wants, c.rel)
		if c.kind != w.kind || c.origin != w.origin || c.top != w.top {
			t.Errorf("%s: got %s/%q/%q, want %s/%q/%q", c.rel, c.kind, c.origin, c.top, w.kind, w.origin, w.top)
		}
		if c.kind == kDup {
			t.Errorf("%s: the real folder has no folder for this company", c.rel)
		}
	}
	for rel := range wants {
		t.Errorf("%s missing from the request", rel)
	}
	if q.Sheet != "3. Übersichten/Kundenliste.xlsx" {
		t.Errorf("sheet %q", q.Sheet)
	}
	// FOO S.A.S, Archive, Foo Ltda, Fooo Inc and the template stay as they are.
	if q.Others != 5 {
		t.Errorf("others at the top: %d, want 5", q.Others)
	}
	groups, order := q.newFolders()
	if len(order) != 2 || order[0] != "Bar Ltd" || order[1] != "Foo Ltd" || len(groups["Foo Ltd"]) != 1 {
		t.Errorf("new folders %v %v", order, groups)
	}
	if originLabel(byNumbForm) != "made by numb-form, unchanged" || originLabel(otherFile) != "" {
		t.Error("origin labels")
	}
	if got := firstMissing(real, "Foo Ltd/x/y.docx"); got != "Foo Ltd" {
		t.Errorf("firstMissing: %q", got)
	}
	if got := firstMissing(real, "Foo GmbH/y.docx"); got != "" {
		t.Errorf("firstMissing of an existing folder: %q", got)
	}
	// Scanning changes nothing in either folder.
	if _, err := os.Stat(filepath.Join(real, "Foo Ltd")); !os.IsNotExist(err) {
		t.Error("the scan wrote into the real folder")
	}
}

func TestResidue(t *testing.T) {
	for rel, want := range map[string]bool{
		RequestName: true, "a/~$x.docx": true, "a/.~lock.x#": true, ".hidden": true, "a/run.log": true,
		"Thumbs.db": true, "a/Appendix A - Risk Assesment - Foo.docx": false, "a/notes.txt": false,
	} {
		if got := residue(rel, false) != ""; got != want {
			t.Errorf("residue(%q) = %v", rel, got)
		}
	}
	if residue("numb-form-runs", true) == "" || residue("Foo GmbH", true) != "" {
		t.Error("residue of folders")
	}
}
