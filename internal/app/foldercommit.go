package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// commitFolders runs numb-commit for a working copy: compare it with the
// real folder, write the request and review it. This version copies
// nothing: the request shows what a commit would change.
func commitFolders(ctx context.Context, u *ui.UI, opt CommitOptions) error {
	from, to, err := opt.folders(u)
	if err != nil {
		return err
	}
	cfg := *opt.Config
	if opt.Downloads != "" {
		cfg.Downloads = opt.Downloads // the run records of -downloads
	}
	rc, err := reviewSetup(ctx, from, to, cfg)
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
	if err := q.write(reqPath); err != nil {
		return err
	}
	u.Printf("\n%s\nSaved as %s\n", ui.Safe(q.text()), reqPath)
	if len(q.of(kAdd)) == 0 && len(q.of(kReplace)) == 0 {
		u.Println("\nNothing to commit: the real folder already has everything.")
		return nil
	}
	v := reviewRequest(q, rc)
	printReview(u, "numb-commit's own checks", v)
	u.Printf("\nThis version of numb-commit only shows the request: nothing was copied into %s.\n", to)
	if v.count("fail") > 0 {
		return errReview
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
