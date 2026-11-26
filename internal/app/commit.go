package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// CommitOptions configures numb-commit.
type CommitOptions struct {
	Downloads string
	RunDir    string
	// Folder mode: copy a working copy (From, default commit_from, else
	// asked) into the real folder (To, default commit_to, else asked).
	From, To string
	Review   bool // only review the request
	Config   *config.Config
	In       io.Reader
	Out      io.Writer
	Now      func() time.Time
	Timeout  time.Duration
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
	ctx, cancel := runContext(opt.Timeout)
	defer cancel()
	// SIGTERM is also what Windows sends when the window is closed: stop
	// cleanly, so the staged copy is removed.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	u := ui.New(ctx, opt.In, opt.Out)
	if opt.Config == nil {
		c := config.Defaults()
		opt.Config = &c
	}
	var err error
	if folderMode(opt) {
		u.Printf("numb-commit %s: copy the new forms from a working copy into the real folder\n\n", Version)
		err = commitFolders(ctx, u, opt)
	} else {
		err = commit(ctx, u, opt)
	}
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
	runDir, err := chooseRun(u, opt)
	if err != nil || runDir == "" {
		return err
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
	plans, err := planCommits(u, m, lg)
	if err != nil || len(plans) == 0 {
		return err
	}
	if err := confirmPlans(u, m, plans, lg); err != nil {
		return err
	}
	return applyPlans(ctx, u, opt, runDir, m, plans, lg)
}

// chooseRun returns the run folder to commit: the one given, the only
// pending one, or the one the person picks. "" means nothing to commit.
func chooseRun(u *ui.UI, opt CommitOptions) (string, error) {
	if opt.RunDir != "" {
		return opt.RunDir, nil
	}
	dl := opt.Downloads
	var err error
	if dl == "" {
		if dl, err = fsutil.DownloadsDir(); err != nil {
			return "", err
		}
	}
	runs, err := pendingRuns(runlog.RunsDir(dl))
	if err != nil {
		return "", fmt.Errorf("no runs found in %s: %w", runlog.RunsDir(dl), err)
	}
	switch len(runs) {
	case 0:
		u.Println("There is nothing to commit.")
		return "", nil
	case 1:
		return runs[0], nil
	}
	u.Println("Runs waiting to be committed (newest first):")
	for i, r := range runs {
		u.Printf("  %d) %s\n", i+1, filepath.Base(r))
	}
	for {
		a, err := u.AskDefault("Choose a run", "1")
		if err != nil {
			return "", err
		}
		var n int
		if _, err := fmt.Sscanf(a, "%d", &n); err == nil && n >= 1 && n <= len(runs) {
			return runs[n-1], nil
		}
	}
}

// applyPlans copies the planned folders, records each one in the
// manifest, and offers to delete the review copy when all are in.
func applyPlans(ctx context.Context, u *ui.UI, opt CommitOptions, runDir string, m *runlog.Manifest, plans []plan, lg *runlog.Logger) error {
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
	return offerDeleteCopy(u, runDir, m, lg)
}

func offerDeleteCopy(u *ui.UI, runDir string, m *runlog.Manifest, lg *runlog.Logger) error {
	del, err := u.YesNo("Delete the review copy in Downloads now? (logs and manifest are kept)", true)
	if err != nil || !del {
		return err
	}
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
	return nil
}
