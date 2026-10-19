package app

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/docx"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

var names = []string{
	"Foo C.A. de C.V.", "Foo Co., Ltda", "Foo GmbH", "Foo S.A.S.", "Foo Corp.", "Foo S.A.E",
	"Foo C.A.", "Foo S.A.", "Foo Inc", "Foo Ltd", "Foo Ltda",
}

func treeHashes(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		r, _ := filepath.Rel(root, p)
		if d.IsDir() {
			out[r+"/"] = "dir"
			return nil
		}
		h, _, err := fsutil.SHA256File(p)
		out[r] = h
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func setup(t *testing.T) (parent, root, dl string) {
	t.Helper()
	base := t.TempDir()
	parent = filepath.Join(base, "share")
	root = filepath.Join(parent, "Kunden 2026")
	dl = filepath.Join(base, "Downloads")
	ids := make([]string, len(names))
	for i := range ids {
		ids[i] = "D00000"
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range []string{"3. Übersichten", "Foo GmbH", "FOO S.A.S", "Archive/foo-corp", "Foo Ltda", "Fooo Inc"} {
		must(os.MkdirAll(filepath.Join(root, d), 0o750))
	}
	must(os.MkdirAll(dl, 0o750))
	must(os.WriteFile(filepath.Join(root, "3. Übersichten", "Kundenliste.xlsx"), testfix.XLSX(t, ids, names), 0o600))
	must(os.WriteFile(filepath.Join(root, "Appendix A - Risk Assesment.docx"), testfix.Template(t), 0o600))
	must(os.WriteFile(filepath.Join(root, "Foo GmbH", "keep.txt"), []byte("existing"), 0o600))
	return parent, root, dl
}

func review() *config.Config {
	c := config.Defaults()
	c.Mode = "review"
	return &c
}

// newCompany answers the offline dialogue for one company whose name has
// a legal form (so it is not asked): address "m", the address, no
// caution, no representatives found.
func newCompany(addr string, final ...string) []string {
	a := []string{"", addr, "", ""}
	return append(a, final...)
}

func TestRunAndCommit(t *testing.T) {
	parent, root, dl := setup(t)
	before := treeHashes(t, root)

	var in []string
	in = append(in, "Max Muster", "")                      // person, confirm
	in = append(in, newCompany("Av. 1, Caracas", "a")...)  // Foo C.A. de C.V.
	in = append(in, newCompany("Calle 2, Bogotá", "a")...) // Foo Co., Ltda
	// Foo GmbH, Foo S.A.S., Foo Corp. exist.
	in = append(in, newCompany("Cairo 3", "a")...)                                // Foo S.A.E
	in = append(in, newCompany("Caracas 4", "e", "3", "Edited Street 9", "r")...) // Foo C.A. rejected
	in = append(in, newCompany("Asunción 5", "e", "4", "Registro 123", "a")...)   // Foo S.A. (registry number typed)
	in = append(in, "n")                                                          // "Fooo Inc" is not Foo Inc
	in = append(in, newCompany("NY 6", "a")...)                                   // Foo Inc
	in = append(in, newCompany("London 7", "a")...)                               // Foo Ltd
	// Foo Ltda exists.

	var out bytes.Buffer
	now := func() time.Time { return time.Date(2026, 10, 16, 9, 30, 0, 0, time.Local) }
	code := Run(Options{
		Parent: parent, Downloads: dl, Offline: true, AllowNonTTY: true, Config: review(),
		In: strings.NewReader(strings.Join(in, "\n") + "\n"), Out: &out, Now: now,
	})
	if code != ExitOK {
		t.Fatalf("exit %d\n%s", code, out.String())
	}
	if after := treeHashes(t, root); len(after) != len(before) {
		t.Fatalf("real tree changed during the run:\n%v\n%v", before, after)
	} else {
		for k, v := range before {
			if after[k] != v {
				t.Fatalf("real tree changed: %s", k)
			}
		}
	}

	runs, _ := os.ReadDir(runlog.RunsDir(dl))
	if len(runs) != 1 {
		t.Fatalf("want one run, got %d", len(runs))
	}
	runDir := filepath.Join(runlog.RunsDir(dl), runs[0].Name())
	m, err := runlog.Load(runDir)
	if err != nil {
		t.Fatal(err)
	}
	status := map[string]string{}
	for _, c := range m.Companies {
		status[c.Name] = c.Status
	}
	want := map[string]string{
		"Foo C.A. de C.V.": "accepted", "Foo Co., Ltda": "accepted", "Foo GmbH": "exists", "Foo S.A.S.": "exists",
		"Foo Corp.": "exists", "Foo S.A.E": "accepted", "Foo C.A.": "rejected", "Foo S.A.": "accepted",
		"Foo Inc": "accepted", "Foo Ltd": "accepted", "Foo Ltda": "exists",
	}
	for k, v := range want {
		if status[k] != v {
			t.Errorf("%s: status %q, want %q\n%s", k, status[k], v, out.String())
		}
	}
	if !m.Complete {
		t.Fatal("manifest not complete")
	}
	if _, err := os.Stat(filepath.Join(m.CopyRoot, "Foo C.A")); !os.IsNotExist(err) {
		t.Fatal("rejected folder was not removed")
	}

	// Inspect one filled document in the copy.
	doc := filepath.Join(m.CopyRoot, "Foo S.A", "Appendix A - Risk Assesment - Foo S.A..docx")
	b, err := os.ReadFile(doc)
	if err != nil {
		t.Fatal(err)
	}
	got, err := (docx.Format{}).ReadBack(b)
	if err != nil {
		t.Fatal(err)
	}
	check := form.Values{
		form.FieldName: "Foo S.A.", form.FieldLegalForm: "S.A.", form.FieldAddress: "Asunción 5",
		form.FieldRegistry: "Registro 123", form.FieldReps: NoRepsText, form.FieldDate: "16/10/2026", form.FieldPerson: "Max Muster",
	}
	for f, v := range check {
		if got[f] != v {
			t.Errorf("%s = %q, want %q", f.Title(), got[f], v)
		}
	}
	if lf, _ := (docx.Format{}).ReadBack(mustRead(t, filepath.Join(m.CopyRoot, "Foo C.A. de C.V", "Appendix A - Risk Assesment - Foo C.A. de C.V..docx"))); lf[form.FieldLegalForm] != "C.A. de C.V." {
		t.Errorf("legal form = %q", lf[form.FieldLegalForm])
	}

	// Commit: confirm, then delete the review copy.
	out.Reset()
	code = Commit(CommitOptions{Downloads: dl, In: strings.NewReader("y\ny\n"), Out: &out, Now: now})
	if code != ExitOK {
		t.Fatalf("commit exit %d\n%s", code, out.String())
	}
	for _, f := range []string{"Foo C.A. de C.V", "Foo Co., Ltda", "Foo S.A.E", "Foo S.A", "Foo Inc", "Foo Ltd"} {
		if _, err := os.Stat(filepath.Join(root, f)); err != nil {
			t.Errorf("not committed: %s", f)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "Foo C.A")); !os.IsNotExist(err) {
		t.Error("rejected company was committed")
	}
	if b, _ := os.ReadFile(filepath.Join(root, "Foo GmbH", "keep.txt")); string(b) != "existing" {
		t.Error("existing folder modified")
	}
	sheetAfter, _, _ := fsutil.SHA256File(filepath.Join(root, "3. Übersichten", "Kundenliste.xlsx"))
	if sheetAfter != before[filepath.Join("3. Übersichten", "Kundenliste.xlsx")] {
		t.Error("spreadsheet modified")
	}
	if _, err := os.Stat(m.CopyRoot); !os.IsNotExist(err) {
		t.Error("review copy not deleted")
	}
	m2, _ := runlog.Load(runDir)
	if !m2.Committed {
		t.Error("manifest not marked committed")
	}
	// A second commit has nothing to do; a second run finds every folder.
	out.Reset()
	if code := Commit(CommitOptions{Downloads: dl, In: strings.NewReader(""), Out: &out, Now: now}); code != ExitOK {
		t.Fatalf("second commit exit %d\n%s", code, out.String())
	}
	out.Reset()
	in2 := "Max Muster\n\n" + newCompanyLine("Caracas 4") // only the rejected Foo C.A. is new again
	code = Run(Options{Parent: parent, Downloads: dl, Offline: true, AllowNonTTY: true, Config: review(), In: strings.NewReader(in2), Out: &out, Now: now})
	if code != ExitOK {
		t.Fatalf("second run exit %d\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "1 accepted, 10 already had a folder") {
		t.Fatalf("unexpected second run summary:\n%s", out.String())
	}
}

func newCompanyLine(addr string) string { return strings.Join(newCompany(addr, "a"), "\n") + "\n" }

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestTimeoutStopsSafely(t *testing.T) {
	parent, _, dl := setup(t)
	pr, pw, err := os.Pipe() // never written: the dialogue blocks until the deadline
	if err != nil {
		t.Fatal(err)
	}
	defer pr.Close()
	defer pw.Close()
	var out bytes.Buffer
	code := Run(Options{Parent: parent, Downloads: dl, Offline: true, AllowNonTTY: true, Config: review(), In: pr, Out: &out, Timeout: 300 * time.Millisecond})
	if code != ExitInterrupted {
		t.Fatalf("exit %d\n%s", code, out.String())
	}
}
