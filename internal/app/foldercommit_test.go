package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/docx"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
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
	// The record of the numb-form run that wrote the form, committed
	// already, so it does not wait for numb-commit.
	runDir := filepath.Join(runlog.RunsDir(dl), "2026-10-16_120000_ab")
	must(os.MkdirAll(runDir, 0o750))
	man := &runlog.Manifest{Version: 1, Tool: "numb-form", Complete: true, Committed: true, Companies: []runlog.Company{
		{Name: "Foo Ltd", Status: "accepted", Committed: true, Files: []runlog.File{{Rel: "Foo Ltd/Appendix A - Risk Assesment - Foo Ltd.docx",
			SHA256: fsutil.SHA256Bytes(doc), Size: int64(len(doc))}}},
	}}
	must(man.Save(runDir))
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
	code := Commit(CommitOptions{From: cp, To: real, Downloads: filepath.Dir(cp), Review: review, Config: &cfg, In: strings.NewReader(answers), Out: &out})
	return code, out.String()
}

// noStage fails when a staged copy was left in Downloads.
func noStage(t *testing.T, dl string) {
	t.Helper()
	es, _ := os.ReadDir(dl)
	for _, e := range es {
		if strings.HasPrefix(e.Name(), stagePrefix) {
			t.Errorf("staged copy %s left behind", e.Name())
		}
	}
}

func TestFolderCommit(t *testing.T) {
	real, cp := workingCopy(t, dateOf("Foo Ltd"))
	// A staged copy left by a numb-commit killed two days ago goes too.
	old := filepath.Join(filepath.Dir(cp), stagePrefix+"2026-10-21 10.00.00 abcd")
	if err := os.MkdirAll(filepath.Join(old, "Foo"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(old, time.Now().Add(-48*time.Hour), time.Now().Add(-48*time.Hour)); err != nil {
		t.Fatal(err)
	}
	tplBefore, _, _ := fsutil.SHA256File(filepath.Join(real, "Appendix A - Risk Assesment.docx"))
	code, out := runCommit(t, cp, real, "y\n", false)
	if code != ExitOK || !strings.Contains(out, "Verdict: OK to commit") || !strings.Contains(out, "staged in") {
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
	if st, err := os.Stat(filepath.Join(real, "Foo Ltd", "Appendix A - Risk Assesment - Foo Ltd.docx")); err != nil || st.Mode().Perm()&0o200 == 0 {
		t.Errorf("committed form is not writable: %v", err)
	}
	noStage(t, filepath.Dir(cp))
	req, _ := os.ReadFile(filepath.Join(cp, RequestName))
	if !strings.HasPrefix(string(req), "\ufeff") {
		t.Error("the request has no byte order mark")
	}
	for _, want := range []string{"Status: applied", "New folder  Foo Ltd", "made by numb-form, unchanged", "After the commit:",
		"Foo Ltd   ", "+ new folder", "~ replaced"} {
		if !strings.Contains(string(req), want) {
			t.Errorf("request lacks %q:\n%s", want, req)
		}
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
	if code != ExitError || !strings.Contains(out, "only report dates may change") || !strings.Contains(out, "Verdict: do not commit.") {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if !reflect.DeepEqual(treeHashes(t, real), before) {
		t.Fatal("the real folder changed")
	}
	noStage(t, filepath.Dir(cp))
}

// A file that differs from the real folder is never overwritten.
func TestFolderCommitKeepsDifferingFile(t *testing.T) {
	real, cp := workingCopy(t, dateOf("Foo Ltd"))
	if err := os.WriteFile(filepath.Join(cp, "Foo GmbH", "keep.txt"), []byte("edited in the copy"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, out := runCommit(t, cp, real, "y\n", false)
	if code != ExitOK || !strings.Contains(out, "≠ unchanged") {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if b, _ := os.ReadFile(filepath.Join(real, "Foo GmbH", "keep.txt")); string(b) != "existing" {
		t.Error("an existing file was overwritten")
	}
	if _, err := os.Stat(filepath.Join(real, "Foo Ltd")); err != nil {
		t.Error("the new folder was not copied")
	}
}

// Without -from, -to or the configuration, numb-commit asks for both
// folders, insists on absolute paths and accepts quoted ones.
func TestFolderCommitAsksForFolders(t *testing.T) {
	real, cp := workingCopy(t, dateOf("Foo Ltd"))
	cfg := config.Defaults()
	cfg.LogDir = filepath.Join(t.TempDir(), "logs")
	opt := CommitOptions{Downloads: filepath.Dir(cp), Config: &cfg}
	if !folderMode(opt) {
		t.Fatal("no review runs waiting, but not in folder mode")
	}
	var out bytes.Buffer
	opt.In, opt.Out = strings.NewReader("14. GWG\n\""+cp+"\"\n"+cp+"\n"+real+"\nn\n"), &out
	if code := Commit(opt); code != ExitOK {
		t.Fatalf("exit %d\n%s", code, out.String())
	}
	for _, want := range []string{askFrom, "Give the whole path", askTo, "the target is the folder to commit", "Nothing was copied"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
	if _, err := os.Stat(filepath.Join(real, "Foo Ltd")); !os.IsNotExist(err) {
		t.Error("copied although the answer was no")
	}
	noStage(t, filepath.Dir(cp))
}

// An unchanged working copy has nothing to commit, and a review without
// a request says so.
func TestFolderCommitNothingToDo(t *testing.T) {
	_, real, dl := setup(t)
	cp := filepath.Join(dl, filepath.Base(real))
	if _, err := fsutil.CopyTree(context.Background(), real, cp, func(string, ...any) {}); err != nil {
		t.Fatal(err)
	}
	if code, out := runCommit(t, cp, real, "", true); code != ExitError || !strings.Contains(out, "no request to review") {
		t.Fatalf("review without a request: exit %d\n%s", code, out)
	}
	if code, out := runCommit(t, cp, real, "", false); code != ExitOK || !strings.Contains(out, "Nothing to commit") {
		t.Fatalf("exit %d\n%s", code, out)
	}
}

func TestFolderMode(t *testing.T) {
	dl := t.TempDir()
	rv, to := review(), config.Defaults()
	to.CommitTo = "F:/14. GWG"
	for _, c := range []struct {
		opt  CommitOptions
		want bool
	}{
		{CommitOptions{Downloads: dl, Config: &to, RunDir: "run"}, false},
		{CommitOptions{Downloads: dl, Config: rv}, false},
		{CommitOptions{Downloads: dl, Config: rv, From: "a"}, true},
		{CommitOptions{Downloads: dl, Config: rv, Review: true}, true},
		{CommitOptions{Downloads: dl, Config: &to}, true},
		{CommitOptions{Downloads: dl}, true},
	} {
		if got := folderMode(c.opt); got != c.want {
			t.Errorf("folderMode(%+v) = %v", c.opt, got)
		}
	}
}
