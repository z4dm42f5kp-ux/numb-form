package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

func TestCompanyFolder(t *testing.T) {
	folders := []realFolder{
		{rel: "D10629 Foo S.A.", name: "D10629 Foo S.A."},
		{rel: "Archive/X20001 - Bar GmbH", name: "X20001 - Bar GmbH"},
		{rel: "Baz Ltda.", name: "Baz Ltda."},
	}
	for _, c := range []struct{ company, id, want string }{
		{"Foo S.A.", "D10629", "D10629 Foo S.A."},
		{"Foo S.A.", "", "D10629 Foo S.A."},
		{"Foo Sociedad", "d10629", "D10629 Foo S.A."}, // the id decides
		{"Bar GmbH", "X99999", "Archive/X20001 - Bar GmbH"},
		{"BAZ LTDA", "", "Baz Ltda."},
		{"Foo GmbH", "", ""},
		{"Qux S.A.", "D1062", ""},
	} {
		f, why, ok := companyFolder(folders, c.company, c.id)
		if f.rel != c.want || ok != (c.want != "") {
			t.Errorf("%s (%s): got %q (%s), want %q", c.company, c.id, f.rel, why, c.want)
		}
	}
}

func TestTree(t *testing.T) {
	q := &request{To: "F", Others: 3, Changes: []change{
		{kind: kReplace, rel: "3. Übersichten/Liste.xlsx"},
		{kind: kAdd, rel: "Bar Ltda/Appendix A - Bar Ltda.docx", top: "Bar Ltda", origin: byNumbForm},
		{kind: kDup, rel: "Foo S.A/Appendix A - Foo S.A.docx", top: "Foo S.A", dup: "D10629 Foo S.A."},
		{kind: kKeep, rel: "numb-form.log"},
	}}
	got := q.tree()
	before := "Before: F\n├── 3. Übersichten\n│   └── Liste.xlsx\n├── D10629 Foo S.A.\n└── … 3 more, unchanged\n"
	if !strings.HasPrefix(got, before) {
		t.Fatalf("before:\n%s", got)
	}
	after := strings.SplitN(got, "After the commit:\n", 2)[1]
	lines := strings.Split(strings.TrimSpace(after), "\n")
	want := []string{"├── 3. Übersichten", "│   └── Liste.xlsx   ", "├── Bar Ltda   ", "│   └── Appendix A - Bar Ltda.docx   + new, made by numb-form, unchanged",
		"├── D10629 Foo S.A.   ", "└── … 3 more, unchanged"}
	if len(lines) != len(want) {
		t.Fatalf("after:\n%s", after)
	}
	for i, w := range want {
		if !strings.HasPrefix(lines[i], w) {
			t.Errorf("line %d = %q, want it to start with %q", i, lines[i], w)
		}
	}
	if strings.Contains(got, "numb-form.log") {
		t.Error("residue in the tree")
	}
}

// The automatic review corrects what it can: a file edited in the working
// copy after staging is staged again, and a file that appeared in the
// real folder is taken out of the request instead of being overwritten.
func TestAutoReviewCorrects(t *testing.T) {
	ctx := context.Background()
	real, cp := workingCopy(t, dateOf("Foo Ltd"))
	cfg := config.Defaults()
	rc, err := reviewSetup(ctx, cp, real, cfg)
	if err != nil {
		t.Fatal(err)
	}
	q, err := scanRequest(ctx, cp, real, rc)
	if err != nil {
		t.Fatal(err)
	}
	dir, err := stageDir(t.TempDir(), cp, real, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := q.stage(ctx, dir); err != nil {
		t.Fatal(err)
	}
	if v := reviewRequest(q, rc); v.count("fail") != 0 {
		t.Fatalf("fresh request fails: %+v", v.findings)
	}
	form := filepath.Join("Foo Ltd", "Appendix A - Risk Assesment - Foo Ltd.docx")
	b, _ := os.ReadFile(filepath.Join(cp, form))
	if err := os.WriteFile(filepath.Join(cp, form), append(b, 0), 0o600); err != nil {
		t.Fatal(err)
	}
	v := reviewRequest(q, rc)
	if v.count("fail") != 1 || !v.fixable() || v.findings[0].fix != fixRescan {
		t.Fatalf("edit after staging: %+v", v.findings)
	}
	if q, err = correct(ctx, q, v, rc); err != nil {
		t.Fatal(err)
	}
	if v := reviewRequest(q, rc); v.count("fail") != 0 {
		t.Fatalf("after staging again: %+v", v.findings)
	}
	if hashOf(q.staged(filepath.ToSlash(form))) != hashOf(filepath.Join(cp, form)) {
		t.Fatal("the staged copy has the old version")
	}

	if err := os.WriteFile(filepath.Join(real, "Foo GmbH", "new.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cp, "Foo GmbH", "new.txt"), []byte("y"), 0o600); err != nil {
		t.Fatal(err)
	}
	q.Changes = append(q.Changes, change{kind: kAdd, rel: "Foo GmbH/new.txt", sha: hashOf(filepath.Join(cp, "Foo GmbH", "new.txt"))})
	if err := os.MkdirAll(filepath.Dir(q.staged("Foo GmbH/new.txt")), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(q.staged("Foo GmbH/new.txt"), []byte("y"), 0o600); err != nil {
		t.Fatal(err)
	}
	v = reviewRequest(q, rc)
	if v.count("fail") != 1 || v.findings[0].fix != fixDrop {
		t.Fatalf("file in the real folder: %+v", v.findings)
	}
	if q, err = correct(ctx, q, v, rc); err != nil {
		t.Fatal(err)
	}
	if v := reviewRequest(q, rc); v.count("fail") != 0 {
		t.Fatalf("after dropping: %+v", v.findings)
	}
	if _, err := os.Stat(q.staged("Foo GmbH/new.txt")); !os.IsNotExist(err) {
		t.Error("the dropped file is still staged")
	}
	if err := removeStage(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("the staged copy was not removed")
	}
	if removeStage(cp) == nil {
		t.Fatal("removed a folder that is not a staged copy")
	}
}

// numb-typo's place corrections pass the spreadsheet review; any other
// change to a name fails.
func TestReviewSheetPlaceFixes(t *testing.T) {
	dir := t.TempDir()
	old := testfix.XLSX(t, []string{"D00000", "D00000"}, []string{"Foo SAS, Bogta", "Foo GmbH"})
	tb, err := sheet.ReadTable(old, "Customer Name", "Customer ID")
	if err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(dir, "old.xlsx")
	if err := os.WriteFile(oldPath, old, 0o600); err != nil {
		t.Fatal(err)
	}
	rc := reviewConfig{header: "Customer Name", idHeader: "Customer ID"}
	for to, fails := range map[string]int{"Foo SAS, Bogotá": 0, "Bar SAS, Bogotá": 1} {
		nb, err := sheet.SetText(old, tb.Part, map[string]string{tb.Rows[0].NameRef: to})
		if err != nil {
			t.Fatal(err)
		}
		newPath := filepath.Join(dir, "new.xlsx")
		if err := os.WriteFile(newPath, nb, 0o600); err != nil {
			t.Fatal(err)
		}
		var v reqReview
		reviewSheet(&v, oldPath, newPath, rc)
		if v.count("fail") != fails {
			t.Errorf("%q: %+v", to, v.findings)
		}
	}
}
