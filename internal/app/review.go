package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
)

// copyToDownloads starts a review run: the whole working folder is copied
// to Downloads first and every change happens in the copy.
func (r *run) copyToDownloads(in *inputs) error {
	dl := r.cfg.Downloads
	var err error
	if dl == "" {
		if dl, err = fsutil.DownloadsDir(); err != nil {
			return err
		}
	}
	if st, err := os.Stat(dl); err != nil || !st.IsDir() {
		return fmt.Errorf("the Downloads folder %s was not found (set downloads in numb-form.conf, NUMBFORM_DOWNLOADS or -downloads)", dl)
	}
	if fsutil.IsUnder(dl, in.root) {
		return errors.New("the Downloads folder is inside the working folder")
	}
	runs := runlog.RunsDir(dl)
	r.warnPending(runs)
	st, err := fsutil.Measure(r.ctx, in.root)
	if err != nil {
		return err
	}
	r.ui.Printf("\nCopying %d files (%.1f MB) to Downloads before anything else.\n", st.Files, float64(st.Bytes)/1e6)
	if st.Bytes > BigCopyBytes {
		ok, err := r.ui.YesNo("That is a large copy. Continue?", false)
		if err != nil {
			return err
		}
		if !ok {
			return errAbort
		}
	}
	if err := r.openRun(runs, in); err != nil {
		return err
	}
	r.srcRoot = in.root
	r.workRoot = filepath.Join(r.runDir, filepath.Base(in.root))
	r.man.Mode, r.man.CopyRoot = "review", r.workRoot
	r.logf("copy   %s", r.workRoot)
	start := time.Now()
	cst, err := fsutil.CopyTree(r.ctx, in.root, r.workRoot, r.logf)
	if err != nil {
		r.logf("ERROR copy failed: %v; removing partial copy", err)
		_ = os.RemoveAll(r.workRoot)
		return fmt.Errorf("copy to Downloads failed: %w", err)
	}
	r.logf("copied %d files, %d folders, %d bytes in %s; skipped %d", cst.Files, cst.Dirs, cst.Bytes, time.Since(start).Round(time.Millisecond), len(cst.Skipped))
	r.ui.Printf("Copied to %s\n", r.workRoot)
	return r.man.Save(r.runDir)
}

func (r *run) warnPending(runs string) {
	ents, err := os.ReadDir(runs)
	if err != nil {
		return
	}
	n := 0
	for _, e := range ents {
		if m, err := runlog.Load(filepath.Join(runs, e.Name())); err == nil && !m.Committed {
			for _, c := range m.Companies {
				if c.Status == "accepted" {
					n++
					break
				}
			}
		}
	}
	if n > 0 {
		r.ui.Printf("note: %d earlier run(s) in %s were never committed. Their companies will be filled again here.\n", n, runs)
	}
}
