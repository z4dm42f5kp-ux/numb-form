package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The staged copy holds exactly the files to add and to replace, checked
// and read-only, in Downloads; the review then reads it.
func TestStage(t *testing.T) {
	real, cp, rc, q := scanned(t)
	dl := filepath.Dir(cp)
	if _, err := stageDir(dl, dl, real, time.Now()); err == nil {
		t.Error("a staged copy inside the working copy was allowed")
	}
	dir, err := stageDir(dl, cp, real, time.Date(2026, 10, 23, 14, 5, 12, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(filepath.Base(dir), stagePrefix+"2026-10-23 14.05.12 ") || filepath.Dir(dir) != dl {
		t.Fatalf("staged copy at %s", dir)
	}
	if err := q.stage(context.Background(), dir); err != nil {
		t.Fatal(err)
	}
	if q.Stage != dir {
		t.Errorf("stage %q", q.Stage)
	}
	staged := 0
	for _, c := range q.Changes {
		p := filepath.Join(dir, filepath.FromSlash(c.rel))
		st, err := os.Stat(p)
		if c.kind != kAdd && c.kind != kReplace {
			if err == nil {
				t.Errorf("%s (%s) was staged", c.rel, c.kind)
			}
			continue
		}
		staged++
		if err != nil || hashOf(p) != c.sha {
			t.Errorf("%s not staged correctly: %v", c.rel, err)
			continue
		}
		if st.Mode().Perm()&0o200 != 0 {
			t.Errorf("%s is writable in the staged copy", c.rel)
		}
	}
	if staged != 3 {
		t.Errorf("%d files staged, want 3", staged)
	}
	if v := reviewRequest(q, rc); v.count("fail") != 0 {
		t.Errorf("review of the staged copy:\n%s", dump(v))
	}

	// An edit in the working copy after staging is found.
	foo := "Foo Ltd/Appendix A - Risk Assesment - Foo Ltd.docx"
	if err := os.WriteFile(filepath.Join(cp, filepath.FromSlash(foo)), []byte("edited"), 0o600); err != nil {
		t.Fatal(err)
	}
	if ok, fix := has(reviewRequest(q, rc), "fail", foo+" changed in the working copy after it was staged"); !ok || fix != fixRescan {
		t.Error("an edit after staging was not found")
	}

	q.unstage(foo)
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(foo))); !os.IsNotExist(err) {
		t.Error("unstage left the file")
	}
	if err := removeStage(cp); err == nil {
		t.Error("removeStage deleted a folder that is not a staged copy")
	}
	if _, err := os.Stat(cp); err != nil {
		t.Fatal(err)
	}
	if err := removeStage(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("the staged copy is still there")
	}
}

// copyChecked never leaves a temporary file and never overwrites.
func TestCopyChecked(t *testing.T) {
	d := t.TempDir()
	src := filepath.Join(d, "src.txt")
	if err := os.WriteFile(src, []byte("form"), 0o600); err != nil {
		t.Fatal(err)
	}
	sha := hashOf(src)
	dst := filepath.Join(d, "dst.txt")
	if err := copyChecked(src, dst, "0000"); err == nil {
		t.Error("a wrong checksum was accepted")
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Error("copied despite a wrong checksum")
	}
	if err := copyChecked(src, dst, sha); err != nil || hashOf(dst) != sha {
		t.Fatalf("copy: %v", err)
	}
	if err := copyChecked(src, dst, sha); err == nil {
		t.Error("an existing file was overwritten")
	}
	es, _ := os.ReadDir(d)
	if len(es) != 2 {
		t.Errorf("left behind: %v", es)
	}
}

// Only staged copies older than a day are removed; other folders and
// recent staged copies stay.
func TestRemoveOldStages(t *testing.T) {
	dl := t.TempDir()
	now := time.Now()
	old := now.Add(-48 * time.Hour)
	mk := func(name string, at time.Time) string {
		t.Helper()
		p := filepath.Join(dl, name)
		if err := os.MkdirAll(filepath.Join(p, "Foo"), 0o750); err != nil {
			t.Fatal(err)
		}
		f := filepath.Join(p, "Foo", "form.docx")
		if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, at, at); err != nil {
			t.Fatal(err)
		}
		return p
	}
	gone := mk(stagePrefix+"2026-10-21 10.00.00 abcd", old)
	recent := mk(stagePrefix+"2026-10-23 10.00.00 ef01", now)
	other := mk("14. GWG", old)
	got := removeOldStages(dl, now)
	if len(got) != 1 || got[0] != filepath.Base(gone) {
		t.Errorf("removed %v", got)
	}
	if _, err := os.Stat(gone); !os.IsNotExist(err) {
		t.Error("the old staged copy is still there")
	}
	for _, p := range []string{recent, other} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s was removed", p)
		}
	}
}
