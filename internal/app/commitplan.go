// Planning a commit: which folders numb-commit adds, and the question
// before it writes.

package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// planCommits lists the accepted folders that can be added to the real
// folder, with the files edited during review.
func planCommits(u *ui.UI, m *runlog.Manifest, lg *runlog.Logger) ([]plan, error) {
	existing, err := indexFolders(m.SourceRoot, filepath.Join(m.SourceRoot, filepath.Dir(m.Spreadsheet)))
	if err != nil {
		return nil, err
	}
	var plans []plan
	for i, c := range m.Companies {
		if c.Status != "accepted" || c.Committed {
			continue
		}
		p := plan{idx: i, src: filepath.Join(m.CopyRoot, c.Folder), dst: filepath.Join(m.SourceRoot, c.Folder), files: map[string]string{}}
		if strings.Contains(c.Folder, "..") || filepath.IsAbs(c.Folder) {
			return nil, fmt.Errorf("manifest entry %q has an unsafe folder name", c.Folder)
		}
		if _, err := os.Lstat(p.dst); err == nil {
			u.Printf("  skip %s: it already exists in the real folder\n", ui.Safe(c.Folder))
			lg.Printf("skip %s: exists", c.Folder)
			continue
		}
		if ex, _ := existing.Find(c.Name); len(ex) > 0 {
			u.Printf("  skip %s: the real folder now has %s\n", ui.Safe(c.Folder), ui.Safe(ex[0].Rel))
			lg.Printf("skip %s: matching folder %s appeared", c.Folder, ex[0].Rel)
			continue
		}
		if st, err := os.Stat(p.src); err != nil || !st.IsDir() {
			u.Printf("  skip %s: missing from the review copy\n", ui.Safe(c.Folder))
			lg.Printf("skip %s: missing in copy", c.Folder)
			continue
		}
		if err := hashPlan(&p, m.CopyRoot, c); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	if len(plans) == 0 {
		u.Println("Nothing left to commit in this run.")
	}
	return plans, nil
}

// hashPlan records the checksum of every file of a folder in the copy and
// notes the files that differ from what the run wrote.
func hashPlan(p *plan, copyRoot string, c runlog.Company) error {
	recorded := map[string]string{}
	for _, f := range c.Files {
		recorded[filepath.FromSlash(f.Rel)] = f.SHA256
	}
	return filepath.WalkDir(p.src, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.Type().IsRegular() {
			return err
		}
		h, _, err := fsutil.SHA256File(path)
		if err != nil {
			return err
		}
		r, _ := filepath.Rel(copyRoot, path)
		p.files[r] = h
		if want, ok := recorded[r]; ok && want != h {
			p.changed = append(p.changed, r)
		} else if !ok {
			p.changed = append(p.changed, r+" (added during review)")
		}
		return nil
	})
}

// confirmPlans shows what will be added and asks before anything is
// written.
func confirmPlans(u *ui.UI, m *runlog.Manifest, plans []plan, lg *runlog.Logger) error {
	u.Printf("\nThese new folders will be added to %s:\n", m.SourceRoot)
	for _, p := range plans {
		u.Printf("  + %s (%d file(s))\n", ui.Safe(m.Companies[p.idx].Folder), len(p.files))
		for _, c := range p.changed {
			u.Printf("      changed since the run: %s\n", ui.Safe(c))
		}
	}
	u.Println("Existing files and folders are never overwritten or deleted.")
	ok, err := u.YesNo("Commit now?", false)
	if err != nil {
		return err
	}
	if !ok {
		lg.Printf("user declined")
		return errAbort
	}
	return nil
}
