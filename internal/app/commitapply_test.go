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
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// leftovers lists the temporary names numb-commit uses below root.
func leftovers(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && strings.HasPrefix(d.Name(), ".numb-") {
			out = append(out, p)
		}
		return nil
	})
	return out
}

// New folders appear whole and new files in existing folders are added,
// all from the staged copy and checked; nothing is overwritten.
func TestApplyAdds(t *testing.T) {
	ctx := context.Background()
	real, cp, rc, _ := scanned(t)
	if err := os.WriteFile(filepath.Join(cp, "Foo GmbH", "notes.txt"), []byte("new notes"), 0o600); err != nil {
		t.Fatal(err)
	}
	q, err := scanRequest(ctx, cp, real, rc)
	if err != nil {
		t.Fatal(err)
	}
	dir, err := stageDir(filepath.Dir(cp), cp, real, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := q.stage(ctx, dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = removeStage(dir) }()
	cfg := config.Defaults()
	cfg.LogDir = t.TempDir()
	logDir, lg, err := commitLog(cfg, CommitOptions{Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	u := ui.New(ctx, strings.NewReader(""), &out)
	sheetBefore := hashOf(filepath.Join(real, "3. Übersichten", "Kundenliste.xlsx"))
	if err := applyRequest(ctx, u, q, lg); err != nil {
		t.Fatal(err)
	}
	_ = lg.Close()
	for _, c := range q.of(kAdd) {
		if hashOf(filepath.Join(real, filepath.FromSlash(c.rel))) != c.sha {
			t.Errorf("%s not added", c.rel)
		}
	}
	if len(q.of(kAdd)) != 3 {
		t.Errorf("%d files to add, want 3", len(q.of(kAdd)))
	}
	if b, _ := os.ReadFile(filepath.Join(real, "Foo GmbH", "keep.txt")); string(b) != "existing" {
		t.Error("an existing file was overwritten")
	}
	if hashOf(filepath.Join(real, "3. Übersichten", "Kundenliste.xlsx")) != sheetBefore {
		t.Error("the spreadsheet was replaced")
	}
	if l := leftovers(t, real); len(l) > 0 {
		t.Errorf("temporary files left: %v", l)
	}
	log, _ := os.ReadFile(filepath.Join(logDir, runlog.CommitLogName))
	for _, want := range []string{"added folder Foo Ltd", "added folder Bar Ltd", "added Foo GmbH/notes.txt sha256="} {
		if !strings.Contains(string(log), want) {
			t.Errorf("log lacks %q:\n%s", want, log)
		}
	}
	if !strings.Contains(out.String(), "✓ Foo Ltd") {
		t.Errorf("output:\n%s", out.String())
	}

	// A second apply finds the folders there and copies nothing.
	lg, err = runlog.Open(filepath.Join(t.TempDir(), "again.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer lg.Close()
	if err := applyRequest(ctx, u, q, lg); err == nil || !strings.Contains(err.Error(), "appeared in the real folder") {
		t.Errorf("second apply: %v", err)
	}
	if l := leftovers(t, real); len(l) > 0 {
		t.Errorf("temporary files left: %v", l)
	}
}
