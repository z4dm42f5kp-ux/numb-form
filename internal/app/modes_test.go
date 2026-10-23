package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/odt"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

var fixedNow = func() time.Time { return time.Date(2026, 10, 16, 9, 30, 0, 0, time.Local) }

func direct(t *testing.T) *config.Config {
	c := config.Defaults()
	c.LogDir = filepath.Join(t.TempDir(), "logs")
	return &c
}

// Direct mode (the default) writes straight into the working folder,
// keeps the spreadsheet untouched and the log outside the folder.
func TestDirectMode(t *testing.T) {
	parent, root, _ := setup(t)
	sheet := filepath.Join(root, "3. Übersichten", "Kundenliste.xlsx")
	before, _ := os.ReadFile(sheet)
	var in []string
	in = append(in, "Max Muster", "")
	in = append(in, newCompany("Av. 1", "a")...)   // Foo C.A. de C.V.
	in = append(in, newCompany("Calle 2", "r")...) // Foo Co., Ltda rejected
	for i := 0; i < 3; i++ {                       // S.A.E, C.A., S.A.
		in = append(in, newCompany("Street", "a")...)
	}
	in = append(in, "y")                          // "Fooo Inc" is Foo Inc's folder
	in = append(in, newCompany("London", "a")...) // Foo Ltd
	cfg := direct(t)
	var out bytes.Buffer
	code := Run(Options{Parent: parent, AllowNonTTY: true, Config: cfg,
		In: strings.NewReader(strings.Join(in, "\n") + "\n"), Out: &out, Now: fixedNow})
	if code != ExitOK {
		t.Fatalf("exit %d\n%s", code, out.String())
	}
	if _, err := os.Stat(filepath.Join(root, "Foo C.A. de C.V", "Appendix A - Risk Assesment - Foo C.A. de C.V..docx")); err != nil {
		t.Fatalf("direct mode did not write into the working folder: %v\n%s", err, out.String())
	}
	if _, err := os.Stat(filepath.Join(root, "Foo Co., Ltda")); !os.IsNotExist(err) {
		t.Fatal("rejected folder left in the working folder")
	}
	if _, err := os.Stat(filepath.Join(root, "Foo Inc")); !os.IsNotExist(err) {
		t.Fatal("confirmed fuzzy match still created a folder")
	}
	after, _ := os.ReadFile(sheet)
	if !bytes.Equal(before, after) {
		t.Fatal("spreadsheet modified")
	}
	logs, _ := os.ReadDir(cfg.LogDir)
	if len(logs) != 1 {
		t.Fatalf("want one log folder, got %d", len(logs))
	}
	for _, e := range mustList(t, root) {
		if strings.HasSuffix(e, ".log") || strings.Contains(e, "numb-form-runs") || strings.Contains(e, ".numb-form-tmp") {
			t.Fatalf("run files left in the working folder: %s", e)
		}
	}
	// numb-commit has nothing to do for direct runs.
	out.Reset()
	if code := Commit(CommitOptions{RunDir: filepath.Join(cfg.LogDir, logs[0].Name()), In: strings.NewReader(""), Out: &out}); code != ExitOK {
		t.Fatalf("commit on a direct run: %d\n%s", code, out.String())
	}
}

func mustList(t *testing.T, root string) []string {
	var out []string
	_ = filepath.Walk(root, func(p string, _ os.FileInfo, err error) error {
		if err == nil {
			out = append(out, p)
		}
		return nil
	})
	return out
}

// The same flow works with an OpenDocument template and a CSV export.
func TestODTAndCSV(t *testing.T) {
	base := t.TempDir()
	parent := filepath.Join(base, "share")
	root := filepath.Join(parent, "Clients")
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(root, "Overview"), 0o750))
	must(os.WriteFile(filepath.Join(root, "Overview", "list.csv"), testfix.CSV([]string{"D1"}, []string{"Foo GmbH"}, ";", true), 0o600))
	must(os.WriteFile(filepath.Join(root, "Appendix A - Risk Assessment.odt"), testfix.ODT(t), 0o600))
	cfg := direct(t)
	cfg.OverviewDir = "overview"
	in := strings.Join(append([]string{"Max Muster", ""}, newCompany("Musterstr. 1", "a")...), "\n") + "\n"
	var out bytes.Buffer
	if code := Run(Options{Parent: parent, AllowNonTTY: true, Config: cfg, In: strings.NewReader(in), Out: &out, Now: fixedNow}); code != ExitOK {
		t.Fatalf("exit %d\n%s", code, out.String())
	}
	b, err := os.ReadFile(filepath.Join(root, "Foo GmbH", "Appendix A - Risk Assessment - Foo GmbH.odt"))
	if err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	got, err := (odt.Format{}).ReadBack(b)
	if err != nil || got[form.FieldAddress] != "Musterstr. 1" || got[form.FieldDate] != "16/10/2026" {
		t.Fatalf("got %v, %v", got, err)
	}
}

// Starting inside the working folder works as well as starting above it.
func TestStartInsideWorkingFolder(t *testing.T) {
	_, root, _ := setup(t)
	cfg := direct(t)
	var out bytes.Buffer
	// The input ends after the person's name; the run stops safely.
	if code := Run(Options{Parent: root, AllowNonTTY: true, Config: cfg, In: strings.NewReader("P\n"), Out: &out, Now: fixedNow}); code == ExitError {
		t.Fatalf("exit %d\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "Working folder: "+root) {
		t.Fatalf("wrong working folder:\n%s", out.String())
	}
}

// -y takes the similar folder "Fooo Inc" as Foo Inc's without asking.
func TestAssumeYes(t *testing.T) {
	parent, root, _ := setup(t)
	in := []string{"Max Muster", ""}
	in = append(in, newCompany("Av. 1", "a")...) // Foo C.A. de C.V.
	in = append(in, newCompany("Calle 2", "r")...)
	for i := 0; i < 3; i++ {
		in = append(in, newCompany("Street", "a")...)
	}
	in = append(in, newCompany("London", "a")...) // Foo Ltd; no question for Foo Inc
	var out bytes.Buffer
	code := Run(Options{Parent: parent, AllowNonTTY: true, AssumeYes: true, Config: direct(t),
		In: strings.NewReader(strings.Join(in, "\n") + "\n"), Out: &out, Now: fixedNow})
	if code != ExitOK || strings.Contains(out.String(), "Is the existing folder") {
		t.Fatalf("exit %d\n%s", code, out.String())
	}
	if _, err := os.Stat(filepath.Join(root, "Foo Inc")); !os.IsNotExist(err) {
		t.Fatal("-y still created a folder for Foo Inc")
	}
}

func TestWhere(t *testing.T) {
	if got := where("Customer ID", "D1001", 315); got != "Customer ID D1001, row 315" {
		t.Error(got)
	}
	if got := where("Customer ID", "", 7); got != "row 7" {
		t.Error(got)
	}
}

// Enter on a blank address, or on the legal form of a name without one,
// writes the standard text, as for the representatives.
func TestStandardTexts(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "Clients")
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(root, "Overview"), 0o750))
	must(os.WriteFile(filepath.Join(root, "Overview", "list.csv"), testfix.CSV([]string{"D1"}, []string{"Foo Trading"}, ";", false), 0o600))
	must(os.WriteFile(filepath.Join(root, "Appendix A - Risk Assessment.odt"), testfix.ODT(t), 0o600))
	cfg := direct(t)
	cfg.OverviewDir = "overview"
	var out bytes.Buffer
	in := "P\n\n\n\n\na\n" // person, use it, legal form, address, representatives, accept
	if code := Run(Options{Parent: root, AllowNonTTY: true, Config: cfg, In: strings.NewReader(in), Out: &out, Now: fixedNow}); code != ExitOK {
		t.Fatalf("exit %d\n%s", code, out.String())
	}
	b, err := os.ReadFile(filepath.Join(root, "Foo Trading", "Appendix A - Risk Assessment - Foo Trading.odt"))
	must(err)
	got, err := (odt.Format{}).ReadBack(b)
	must(err)
	if got[form.FieldLegalForm] != NoLegalFormText || got[form.FieldAddress] != NoAddressText || got[form.FieldReps] != NoRepsText {
		t.Fatalf("got %v", got)
	}
}
