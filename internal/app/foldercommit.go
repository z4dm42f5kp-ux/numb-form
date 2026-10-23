package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

var errReview = errors.New("the review found problems; nothing was copied")

func printReview(u *ui.UI, title string, v *reqReview) {
	u.Printf("\n%s:\n", title)
	marks := map[string]string{"ok": "  ok    ", "warn": "  check ", "fail": "  FAIL  "}
	for _, f := range v.findings {
		u.Printf("%s%s\n", marks[f.level], ui.Safe(f.msg))
	}
	switch {
	case v.count("fail") > 0:
		u.Println("Verdict: do not commit.")
	case v.count("warn") > 0:
		u.Println("Verdict: commit after checking the lines marked check.")
	default:
		u.Println("Verdict: OK to commit.")
	}
}

// commitFolders runs numb-commit for a working copy: compare, stage the
// files to copy in Downloads, write the request, review it, and after the
// person agrees, copy the staged files into the real folder.
func commitFolders(ctx context.Context, u *ui.UI, opt CommitOptions) error {
	from, to, err := opt.folders(u)
	if err != nil {
		return err
	}
	rc, err := reviewSetup(ctx, from, to, *opt.Config)
	if err != nil {
		return err
	}
	reqPath := filepath.Join(from, RequestName)
	u.Printf("\nWorking copy: %s\nReal folder:  %s\n", from, to)
	if opt.Review {
		return reviewOnly(u, reqPath, from, to, rc)
	}

	u.Println("\nComparing the two folders…")
	q, err := scanRequest(ctx, from, to, rc)
	if err != nil {
		return err
	}
	q.Made = opt.Now()
	if len(q.of(kAdd)) == 0 && len(q.of(kReplace)) == 0 {
		if err := q.write(reqPath); err != nil {
			return err
		}
		u.Printf("\n%s\nNothing to commit: the real folder already has everything.\n", ui.Safe(q.text()))
		return nil
	}
	dir, err := stageDir(opt.Downloads, from, to, opt.Now())
	if err != nil {
		return err
	}
	for _, n := range removeOldStages(filepath.Dir(dir), opt.Now()) {
		u.Printf("Removed %q, left in Downloads by an earlier numb-commit.\n", n)
	}
	defer func() {
		if _, err := os.Stat(dir); err == nil && removeStage(dir) == nil {
			u.Println("The staged copy in Downloads was removed.")
		}
	}()
	if err := q.stage(ctx, dir); err != nil {
		return err
	}
	if err := q.write(reqPath); err != nil {
		return err
	}
	u.Printf("\n%s\n", ui.Safe(q.text()))
	u.Printf("Saved as %s\nThe files to copy are staged in %s; nothing is in %s yet.\n", reqPath, dir, to)

	v := reviewRequest(q, rc)
	printReview(u, "numb-commit's own checks", v)
	if v.count("fail") > 0 {
		return errReview
	}
	u.Println("Look at the staged copy and the request now. To change something, answer no, change the working copy and run numb-commit again.")
	ok, err := u.YesNo(fmt.Sprintf("\nCopy this into %s now?", to), false)
	if err != nil {
		return err
	}
	if !ok {
		q.Stage = ""
		_ = q.write(reqPath)
		u.Println("Nothing was copied. The request stays in the working copy for reference.")
		return nil
	}
	return applyAndCheck(ctx, u, opt, q, rc, reqPath)
}

// applyAndCheck copies the staged files and reviews the real folder.
func applyAndCheck(ctx context.Context, u *ui.UI, opt CommitOptions, q *request, rc reviewConfig, reqPath string) error {
	logDir, lg, err := commitLog(*opt.Config, opt)
	if err != nil {
		return err
	}
	defer lg.Close()
	lg.Printf("commit %s -> %s through %s, request made %s", q.From, q.To, q.Stage, q.Made.Format("2006-01-02 15:04:05"))
	if err := applyRequest(ctx, u, q, logDir, lg); err != nil {
		lg.Printf("ERROR %v", err)
		return fmt.Errorf("%w (log: %s)", err, filepath.Join(logDir, runlog.CommitLogName))
	}
	q.Status, q.Stage = "applied", ""
	if err := q.write(reqPath); err != nil {
		return err
	}
	_ = fsutil.CopyFile(reqPath, filepath.Join(logDir, RequestName))
	after := reviewRequest(q, rc)
	printReview(u, "Review of the real folder", after)
	u.Printf("\nLog: %s\n", filepath.Join(logDir, runlog.CommitLogName))
	if after.count("fail") > 0 {
		return errors.New("the real folder does not match the request; see the log")
	}
	return nil
}

// reviewOnly reviews the request in the working copy without copying.
func reviewOnly(u *ui.UI, reqPath, from, to string, rc reviewConfig) error {
	q, err := readRequest(reqPath)
	if err != nil {
		return fmt.Errorf("no request to review (%w); run numb-commit first", err)
	}
	if !strings.EqualFold(q.From, from) || !strings.EqualFold(q.To, to) {
		return fmt.Errorf("the request is for %s -> %s", q.From, q.To)
	}
	if q.Stage != "" {
		if _, err := os.Stat(q.Stage); err != nil {
			q.Stage = "" // removed: review the working copy instead
		}
	}
	u.Printf("Request:      %s (%s, made %s)\n", reqPath, q.Status, q.Made.Format("02/01/2006 15:04"))
	v := reviewRequest(q, rc)
	printReview(u, "Review", v)
	if v.count("fail") > 0 {
		return errReview
	}
	return nil
}
