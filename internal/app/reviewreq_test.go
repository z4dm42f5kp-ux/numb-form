package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
)

// scanned makes a working copy (scanCopy) and its request.
func scanned(t *testing.T) (real, cp string, rc reviewConfig, q *request) {
	t.Helper()
	real, cp = scanCopy(t)
	ctx := context.Background()
	rc, err := reviewSetup(ctx, cp, real, testConfig(t, filepath.Dir(cp)))
	if err != nil {
		t.Fatal(err)
	}
	if q, err = scanRequest(ctx, cp, real, rc); err != nil {
		t.Fatal(err)
	}
	return real, cp, rc, q
}

// has reports whether the review has a finding of level that contains
// msg, and returns its fix.
func has(v *reqReview, level, msg string) (bool, string) {
	for _, f := range v.findings {
		if f.level == level && strings.Contains(f.msg, msg) {
			return true, f.fix
		}
	}
	return false, ""
}

func dump(v *reqReview) string {
	var b strings.Builder
	for _, f := range v.findings {
		b.WriteString(f.level + ": " + f.msg + " [" + f.fix + "]\n")
	}
	return b.String()
}

func TestReviewRequest(t *testing.T) {
	real, cp, rc, q := scanned(t)
	v := reviewRequest(q, rc)
	if v.count("fail") != 0 || !v.fixable() {
		t.Fatalf("a fresh request fails its review:\n%s", dump(v))
	}
	for _, want := range [][2]string{
		{"ok", "the spreadsheet only gains 1 report date(s)"},
		{"ok", "Foo Ltd: section A is filled for Foo Ltd"},
		{"warn", `new folder "Bar Ltd" matches no company in the Excel list`},
		{"ok", "2 file(s) to add and 1 to replace"},
	} {
		if ok, _ := has(v, want[0], want[1]); !ok {
			t.Errorf("no %s %q:\n%s", want[0], want[1], dump(v))
		}
	}

	// A form edited after the request is compared again.
	foo := "Foo Ltd/Appendix A - Risk Assesment - Foo Ltd.docx"
	p := filepath.Join(cp, filepath.FromSlash(foo))
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, append(b, 0), 0o600); err != nil {
		t.Fatal(err)
	}
	v = reviewRequest(q, rc)
	if ok, fix := has(v, "fail", foo+" changed after the request was made"); !ok || fix != fixRescan || !v.fixable() {
		t.Errorf("edited form:\n%s", dump(v))
	}
	if err := os.WriteFile(p, b, 0o600); err != nil {
		t.Fatal(err)
	}

	// A file that appeared in the real folder is never overwritten.
	if err := os.MkdirAll(filepath.Join(real, "Foo Ltd"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(real, filepath.FromSlash(foo)), []byte("someone else's"), 0o600); err != nil {
		t.Fatal(err)
	}
	v = reviewRequest(q, rc)
	if ok, fix := has(v, "fail", foo+" already exists in the real folder"); !ok || fix != fixDrop {
		t.Errorf("file in the real folder:\n%s", dump(v))
	}

	// A spreadsheet edited in the real folder needs a person.
	sp := filepath.Join(real, "3. Übersichten", "Kundenliste.xlsx")
	if err := os.WriteFile(sp, []byte("edited elsewhere"), 0o600); err != nil {
		t.Fatal(err)
	}
	v = reviewRequest(q, rc)
	if ok, _ := has(v, "fail", "changed after the request was made (someone else edited it?)"); !ok || v.fixable() {
		t.Errorf("spreadsheet edited in the real folder:\n%s", dump(v))
	}

	// After the commit, every file must be in the real folder.
	q.Status = "applied"
	v = reviewRequest(q, rc)
	if ok, _ := has(v, "fail", "Bar Ltd/Appendix A - Risk Assesment - Bar Ltd.docx is missing in the real folder"); !ok {
		t.Errorf("applied:\n%s", dump(v))
	}
	if ok, _ := has(v, "fail", "is not the committed version"); !ok {
		t.Errorf("applied:\n%s", dump(v))
	}
}

// Residue and the template are never copied, even when a request lists
// them.
func TestReviewRefusesResidue(t *testing.T) {
	_, _, rc, q := scanned(t)
	q.Changes = append(q.Changes, change{kind: kAdd, rel: "numb-form.log"}, change{kind: kAdd, rel: rc.template})
	v := reviewRequest(q, rc)
	for _, rel := range []string{"numb-form.log", rc.template} {
		if ok, fix := has(v, "fail", rel+" is listed for copying but must stay behind"); !ok || fix != fixDrop {
			t.Errorf("%s:\n%s", rel, dump(v))
		}
	}
}

// Only report dates may be written, and only into cells without one.
func TestReviewSheet(t *testing.T) {
	real, cp, rc, _ := scanned(t)
	rel := filepath.FromSlash(rc.sheetRel)
	old := filepath.Join(real, rel)
	b, err := os.ReadFile(old)
	if err != nil {
		t.Fatal(err)
	}
	tb, err := sheet.ReadTable(b, rc.header, rc.idHeader)
	if err != nil {
		t.Fatal(err)
	}
	check := func(base []byte, vals map[string]string, level, msg string) {
		t.Helper()
		before := filepath.Join(t.TempDir(), "before.xlsx")
		after := filepath.Join(t.TempDir(), "after.xlsx")
		nb, err := sheet.SetText(base, tb.Part, vals)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(before, base, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(after, nb, 0o600); err != nil {
			t.Fatal(err)
		}
		v := &reqReview{}
		reviewSheet(v, before, after, rc)
		if ok, _ := has(v, level, msg); !ok {
			t.Errorf("%v: no %s %q:\n%s", vals, level, msg, dump(v))
		}
	}
	r := tb.Rows[1]
	check(b, map[string]string{r.DateRef: "16-10-2026"}, "ok", "only gains 1 report date(s)")
	check(b, map[string]string{r.NameRef: "Renamed GmbH"}, "fail", "only report dates may change")
	check(b, map[string]string{r.DateRef: "soon"}, "warn", "not a DD-MM-YYYY date")
	dated, err := sheet.SetText(b, tb.Part, map[string]string{r.DateRef: "01-01-2026"})
	if err != nil {
		t.Fatal(err)
	}
	check(dated, map[string]string{r.DateRef: "16-10-2026"}, "fail", "already had a date")
	if got := (&request{From: cp}).staged("a/b"); got != filepath.Join(cp, "a", "b") {
		t.Errorf("staged without a stage: %s", got)
	}
	if got := (&request{From: cp, Stage: real}).staged("a/b"); got != filepath.Join(real, "a", "b") {
		t.Errorf("staged: %s", got)
	}
}
