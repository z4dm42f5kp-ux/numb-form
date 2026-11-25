package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

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

// The automatic review corrects and reviews again until it passes, and
// stops at once on what needs a person.
func TestAutoReviewRounds(t *testing.T) {
	ctx := context.Background()
	real, cp := workingCopy(t, dateOf("Foo Ltd"))
	rc, err := reviewSetup(ctx, cp, real, testConfig(t, filepath.Dir(cp)))
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
	defer func() { _ = removeStage(dir) }()
	if err := q.stage(ctx, dir); err != nil {
		t.Fatal(err)
	}
	reqPath := filepath.Join(cp, RequestName)
	form := filepath.Join(cp, "Foo Ltd", "Appendix A - Risk Assesment - Foo Ltd.docx")
	b, _ := os.ReadFile(form)
	if err := os.WriteFile(form, append(b, 0), 0o600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	u := ui.New(ctx, strings.NewReader(""), &out)
	if q, err = autoReview(ctx, u, q, rc, reqPath); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	for _, want := range []string{"Review, round 1", "corrected 1 point(s)", "Review, round 2"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
	if last := out.String()[strings.Index(out.String(), "Review, round 2"):]; strings.Contains(last, "do not commit") {
		t.Errorf("round 2 still fails:\n%s", last)
	}
	if req, err := readRequest(reqPath); err != nil || req.Stage != q.Stage {
		t.Errorf("request not updated: %v", err)
	}

	// Someone else changed the real spreadsheet: a person must look.
	if err := os.WriteFile(filepath.Join(real, "3. Übersichten", "Kundenliste.xlsx"), []byte("edited elsewhere"), 0o600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if _, err := autoReview(ctx, u, q, rc, reqPath); err != errReview || !strings.Contains(out.String(), "These need a person") || strings.Contains(out.String(), "round 2") {
		t.Errorf("%v\n%s", err, out.String())
	}
}
