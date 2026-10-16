package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// CommitOptions configures numb-commit.
type CommitOptions struct {
	Downloads string
	RunDir    string
	In        io.Reader
	Out       io.Writer
	Now       func() time.Time
	Timeout   time.Duration
}

type plan struct {
	idx     int
	src     string
	dst     string
	files   map[string]string // rel path -> sha256 in the copy
	changed []string          // edited during review
}

// Commit copies the accepted folders of a reviewed run back into the real
// directory. It only adds new folders; it never overwrites or deletes.
func Commit(opt CommitOptions) int {
	if opt.Now == nil {
		opt.Now = time.Now
	}
	if opt.Timeout <= 0 || opt.Timeout > MaxRunTime {
		opt.Timeout = MaxRunTime
	}
	ctx, cancel := context.WithTimeout(context.Background(), opt.Timeout)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	u := ui.New(ctx, opt.In, opt.Out)
	err := commit(ctx, u, opt)
	switch {
	case err == nil:
		return ExitOK
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled), errors.Is(err, errAbort), errors.Is(err, ui.ErrClosed):
		u.Printf("\nStopped (%v). Folders already committed stay committed; nothing else changed.\n", err)
		return ExitInterrupted
	default:
		u.Printf("\nnumb-commit: %v\n", err)
		return ExitError
	}
}

func commit(ctx context.Context, u *ui.UI, opt CommitOptions) error {
	u.Printf("numb-commit %s — copy reviewed forms back into the real folder\n\n", Version)
	runDir := opt.RunDir
	if runDir == "" {
		dl := opt.Downloads
		var err error
		if dl == "" {
			if dl, err = fsutil.DownloadsDir(); err != nil {
				return err
			}
		}
		runs, err := pendingRuns(runlog.RunsDir(dl))
		if err != nil {
			return fmt.Errorf("no runs found in %s: %w", runlog.RunsDir(dl), err)
		}
		switch len(runs) {
		case 0:
			u.Println("There is nothing to commit.")
			return nil
		case 1:
			runDir = runs[0]
		default:
			u.Println("Runs waiting to be committed (newest first):")
			for i, r := range runs {
				u.Printf("  %d) %s\n", i+1, filepath.Base(r))
			}
			for runDir == "" {
				a, err := u.AskDefault("Choose a run", "1")
				if err != nil {
					return err
				}
				var n int
				if _, err := fmt.Sscanf(a, "%d", &n); err == nil && n >= 1 && n <= len(runs) {
					runDir = runs[n-1]
				}
			}
		}
	}
	m, err := runlog.Load(runDir)
	if err != nil {
		return err
	}
	if m.Mode == "direct" {
		u.Println("That run wrote directly into the working folder (mode = direct); there is nothing to commit.")
		return nil
	}
	lg, err := runlog.Open(filepath.Join(runDir, runlog.CommitLogName))
	if err != nil {
		return err
	}
	defer lg.Close()
	lg.Printf("commit started for run %s", runDir)

	if st, err := os.Stat(m.SourceRoot); err != nil || !st.IsDir() {
		return fmt.Errorf("real folder %s is not reachable", m.SourceRoot)
	}
	if !m.Complete {
		u.Printf("note: this run stopped early (%s). Only its accepted forms are committed.\n", m.Interrupted)
	}
	if h, _, err := fsutil.SHA256File(filepath.Join(m.SourceRoot, m.Spreadsheet)); err != nil || h != m.SheetSHA256 {
		u.Println("warning: the spreadsheet in the real folder has changed since the run.")
		ok, err := u.YesNo("Commit anyway?", false)
		if err != nil {
			return err
		}
		if !ok {
			return errAbort
		}
	}
	existing, err := indexFolders(m.SourceRoot, filepath.Join(m.SourceRoot, filepath.Dir(m.Spreadsheet)))
	if err != nil {
		return err
	}

	var plans []plan
	for i, c := range m.Companies {
		if c.Status != "accepted" || c.Committed {
			continue
		}
		p := plan{idx: i, src: filepath.Join(m.CopyRoot, c.Folder), dst: filepath.Join(m.SourceRoot, c.Folder), files: map[string]string{}}
		if strings.Contains(c.Folder, "..") || filepath.IsAbs(c.Folder) {
			return fmt.Errorf("manifest entry %q has an unsafe folder name", c.Folder)
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
		recorded := map[string]string{}
		for _, f := range c.Files {
			recorded[filepath.FromSlash(f.Rel)] = f.SHA256
		}
		err := filepath.WalkDir(p.src, func(path string, d fs.DirEntry, err error) error {
			if err != nil || !d.Type().IsRegular() {
				return err
			}
			h, _, err := fsutil.SHA256File(path)
			if err != nil {
				return err
			}
			r, _ := filepath.Rel(m.CopyRoot, path)
			p.files[r] = h
			if want, ok := recorded[r]; ok && want != h {
				p.changed = append(p.changed, r)
			} else if !ok {
				p.changed = append(p.changed, r+" (added during review)")
			}
			return nil
		})
		if err != nil {
			return err
		}
		plans = append(plans, p)
	}
	if len(plans) == 0 {
		u.Println("Nothing left to commit in this run.")
		return nil
	}
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

	done := 0
	for _, p := range plans {
		if err := ctx.Err(); err != nil {
			return err
		}
		name := m.Companies[p.idx].Folder
		if err := commitOne(ctx, p, lg); err != nil {
			u.Printf("  ! %s: %v\n", ui.Safe(name), err)
			lg.Printf("ERROR %s: %v", name, err)
			continue
		}
		m.Companies[p.idx].Committed = true
		done++
		_ = m.Save(runDir)
		u.Printf("  ✓ %s\n", ui.Safe(name))
		lg.Printf("committed %s (%d files, sha256 verified)", name, len(p.files))
	}
	all := true
	for _, c := range m.Companies {
		if c.Status == "accepted" && !c.Committed {
			all = false
		}
	}
	m.Committed = all
	m.CommittedAt = opt.Now()
	if err := m.Save(runDir); err != nil {
		return err
	}
	u.Printf("\nCommitted %d of %d folders. Log: %s\n", done, len(plans), filepath.Join(runDir, runlog.CommitLogName))
	if !all {
		return errors.New("some folders were not committed; see the log")
	}
	del, err := u.YesNo("Delete the review copy in Downloads now? (logs and manifest are kept)", true)
	if err != nil {
		return err
	}
	if del {
		if r, err := filepath.Rel(runDir, m.CopyRoot); err != nil || strings.HasPrefix(r, "..") || r == "." {
			return errors.New("refusing to delete: copy is not inside the run folder")
		}
		if err := os.RemoveAll(m.CopyRoot); err != nil {
			return err
		}
		m.CopyDeleted = true
		_ = m.Save(runDir)
		lg.Printf("deleted review copy %s", m.CopyRoot)
		u.Println("Review copy deleted.")
	}
	return nil
}
