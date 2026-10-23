package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

var errReview = errors.New("the review found problems; nothing was copied")

// folderMode reports whether numb-commit copies a working copy into the
// real folder (commit_to or -to) rather than committing a review run.
func folderMode(opt CommitOptions) bool {
	return opt.Review || opt.To != "" || (opt.RunDir == "" && opt.Config != nil && opt.Config.CommitTo != "")
}

// folders resolves the working copy (from) and the real folder (to).
func (opt CommitOptions) folders() (from, to string, err error) {
	cfg := *opt.Config
	from = opt.From
	if from == "" {
		from = opt.Parent
	}
	if from, err = filepath.Abs(from); err != nil {
		return "", "", err
	}
	if !isWorkingFolder(from, cfg) {
		return "", "", fmt.Errorf("%s is not a working folder (template and %q folder); start numb-commit inside the copy, or use -from", from, cfg.OverviewDir)
	}
	to = opt.To
	if to == "" {
		to = cfg.CommitTo
	}
	if to == "" {
		return "", "", errors.New(`say once where the real folder is, for example:  numb-form config set commit_to "F:\14. GWG"`)
	}
	if to, err = filepath.Abs(to); err != nil {
		return "", "", err
	}
	switch {
	case strings.EqualFold(from, to):
		return "", "", errors.New("the working copy and the real folder are the same folder")
	case fsutil.IsUnder(from, to) || fsutil.IsUnder(to, from):
		return "", "", errors.New("the working copy and the real folder must not be inside each other")
	}
	if st, err := os.Stat(to); err != nil || !st.IsDir() {
		return "", "", fmt.Errorf("the real folder %s is not reachable (is the network drive connected?)", to)
	}
	if !isWorkingFolder(to, cfg) {
		return "", "", fmt.Errorf("%s is not a working folder (no template and %q folder)", to, cfg.OverviewDir)
	}
	return from, to, nil
}

// reviewSetup reads what a review needs from the working copy.
func reviewSetup(from string, cfg config.Config) (sheetRel string, rc reviewConfig, err error) {
	overview, err := findOverview(from, cfg.OverviewDir)
	if err != nil {
		return "", rc, err
	}
	sp, _, err := findSpreadsheet(overview)
	if err != nil {
		return "", rc, err
	}
	tpl, err := findTemplate(from, cfg.TemplateName)
	if err != nil {
		return "", rc, err
	}
	rc = reviewConfig{header: cfg.Header, idHeader: cfg.IDHeader, overviewDir: cfg.OverviewDir, template: filepath.ToSlash(rel(from, tpl)),
		templateBase: strings.TrimSuffix(filepath.Base(tpl), filepath.Ext(tpl)), names: map[string]string{}}
	if b, err := os.ReadFile(sp); err == nil { // #nosec G304 -- the overview spreadsheet.
		if rows, _, err := sheet.Read(b, filepath.Ext(sp), sheet.Options{Header: cfg.Header, IDHeader: cfg.IDHeader, Legacy: cfg.LegacyFormats}); err == nil {
			for _, r := range rows {
				n, _ := fsutil.SafeName(r.Name, false)
				rc.names[strings.ToLower(n)] = r.Name
			}
		}
	}
	return rel(from, sp), rc, nil
}

func printReview(u *ui.UI, v *reqReview) {
	u.Println("\nReview:")
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

// commitFolders runs numb-commit for a working copy: make the request,
// review it, and after the user agrees, apply it.
func commitFolders(ctx context.Context, u *ui.UI, opt CommitOptions) error {
	cfg := *opt.Config
	from, to, err := opt.folders()
	if err != nil {
		return err
	}
	sheetRel, rc, err := reviewSetup(from, cfg)
	if err != nil {
		return err
	}
	reqPath := filepath.Join(from, RequestName)
	u.Printf("Working copy: %s\nReal folder:  %s\n", from, to)
	if opt.Review {
		q, err := readRequest(reqPath)
		if err != nil {
			return fmt.Errorf("no request to review (%w); run numb-commit first", err)
		}
		if !strings.EqualFold(q.From, from) || !strings.EqualFold(q.To, to) {
			return fmt.Errorf("the request is for %s -> %s", q.From, q.To)
		}
		u.Printf("Request:      %s (%s, made %s)\n", reqPath, q.Status, q.Made.Format("02/01/2006 15:04"))
		v := reviewRequest(q, rc)
		printReview(u, v)
		if v.count("fail") > 0 {
			return errReview
		}
		return nil
	}

	u.Println("\nComparing the two folders…")
	q, err := scanRequest(ctx, from, to, sheetRel, rc.template)
	if err != nil {
		return err
	}
	q.Made = opt.Now()
	if err := q.write(reqPath); err != nil {
		return err
	}
	text, _ := os.ReadFile(reqPath) // #nosec G304 -- written just now.
	u.Printf("\n%s", ui.Safe(strings.SplitN(string(text), requestData, 2)[0]))
	u.Printf("This list is saved as %s\n", reqPath)
	if len(q.of("add")) == 0 && len(q.of("replace")) == 0 {
		u.Println("\nNothing to commit: the real folder already has everything.")
		return nil
	}
	v := reviewRequest(q, rc)
	printReview(u, v)
	if v.count("fail") > 0 {
		return errReview
	}
	ok, err := u.YesNo(fmt.Sprintf("\nCopy this into %s now?", to), false)
	if err != nil {
		return err
	}
	if !ok {
		u.Println("Nothing was copied. The request stays for later: run numb-commit again.")
		return nil
	}
	logDir, lg, err := commitLog(cfg, opt)
	if err != nil {
		return err
	}
	defer lg.Close()
	lg.Printf("commit %s -> %s, request made %s", from, to, q.Made.Format("2006-01-02 15:04:05"))
	if err := applyRequest(ctx, u, q, logDir, lg); err != nil {
		lg.Printf("ERROR %v", err)
		return fmt.Errorf("%w (log: %s)", err, filepath.Join(logDir, runlog.CommitLogName))
	}
	q.Status = "applied"
	if err := q.write(reqPath); err != nil {
		return err
	}
	_ = fsutil.CopyFile(reqPath, filepath.Join(logDir, RequestName))
	after := reviewRequest(q, rc)
	printReview(u, after)
	u.Printf("\nLog: %s\n", filepath.Join(logDir, runlog.CommitLogName))
	if after.count("fail") > 0 {
		return errors.New("the real folder does not match the request; see the log")
	}
	return nil
}

func commitLog(cfg config.Config, opt CommitOptions) (string, *runlog.Logger, error) {
	dir := cfg.LogDir
	if dir == "" {
		var err error
		if dir, err = config.StateDir(); err != nil {
			return "", nil, err
		}
	}
	dir = filepath.Join(dir, opt.Now().Format("2006-01-02_150405")+"_"+fsutil.RandHex(2)+"_commit")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", nil, err
	}
	lg, err := runlog.Open(filepath.Join(dir, runlog.CommitLogName))
	return dir, lg, err
}

// copyChecked copies one file to dst, which must not exist, through a
// temporary name, and checks its SHA-256.
func copyChecked(src, dst, sha string) error {
	tmp := filepath.Join(filepath.Dir(dst), ".numb-commit-"+fsutil.RandHex(4))
	if err := fsutil.CopyFile(src, tmp); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if h := hashOf(tmp); h != sha {
		_ = os.Remove(tmp)
		return fmt.Errorf("%s does not match the request (sha256 %s)", src, h)
	}
	if _, err := os.Lstat(dst); err == nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("%s appeared in the real folder", dst)
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// applyRequest copies new folders whole (built under a temporary name and
// renamed, so a folder appears complete or not at all), then new files,
// then replaces the spreadsheet after keeping its old version.
func applyRequest(ctx context.Context, u *ui.UI, q *request, logDir string, lg *runlog.Logger) error {
	groups, order := q.newFolders()
	for _, top := range order {
		if err := ctx.Err(); err != nil {
			return err
		}
		if top == "" {
			for _, c := range groups[top] {
				if err := copyChecked(filepath.Join(q.From, filepath.FromSlash(c.rel)), filepath.Join(q.To, filepath.FromSlash(c.rel)), c.sha); err != nil {
					return err
				}
				lg.Printf("added %s sha256=%s", c.rel, c.sha)
				u.Printf("  ✓ %s\n", ui.Safe(filepath.FromSlash(c.rel)))
			}
			continue
		}
		final := filepath.Join(q.To, filepath.FromSlash(top))
		tmp := filepath.Join(filepath.Dir(final), ".numb-commit-"+fsutil.RandHex(4))
		if err := os.Mkdir(tmp, 0o755); err != nil { // #nosec G301 -- folders on a team share.
			return err
		}
		for _, c := range groups[top] {
			inside := strings.TrimPrefix(c.rel, top+"/")
			dst := filepath.Join(tmp, filepath.FromSlash(inside))
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { // #nosec G301
				_ = os.RemoveAll(tmp)
				return err
			}
			if err := copyChecked(filepath.Join(q.From, filepath.FromSlash(c.rel)), dst, c.sha); err != nil {
				_ = os.RemoveAll(tmp)
				return err
			}
		}
		if _, err := os.Lstat(final); err == nil {
			_ = os.RemoveAll(tmp)
			return fmt.Errorf("%s appeared in the real folder", final)
		}
		if err := os.Rename(tmp, final); err != nil {
			_ = os.RemoveAll(tmp)
			return err
		}
		lg.Printf("added folder %s (%d files, sha256 verified)", top, len(groups[top]))
		u.Printf("  ✓ %s\n", ui.Safe(filepath.FromSlash(top)))
	}
	for _, c := range q.of("replace") {
		dst := filepath.Join(q.To, filepath.FromSlash(c.rel))
		data, err := os.ReadFile(filepath.Join(q.From, filepath.FromSlash(c.rel))) // #nosec G304 -- listed in the request.
		if err != nil {
			return err
		}
		if fsutil.SHA256Bytes(data) != c.sha {
			return fmt.Errorf("%s changed after the request was made", c.rel)
		}
		old, err := os.ReadFile(dst) // #nosec G304 -- listed in the request.
		if err != nil {
			return err
		}
		if fsutil.SHA256Bytes(old) != c.old {
			return fmt.Errorf("%s in the real folder changed after the request was made", c.rel)
		}
		keep := filepath.Join(logDir, "replaced", filepath.FromSlash(c.rel))
		if err := os.MkdirAll(filepath.Dir(keep), 0o750); err != nil {
			return err
		}
		if err := fsutil.WriteNew(keep, old); err != nil {
			return err
		}
		if err := fsutil.ReplaceVerified(dst, c.old, data); err != nil {
			return err
		}
		lg.Printf("replaced %s sha256 %s -> %s; old version kept at %s", c.rel, c.old, c.sha, keep)
		u.Printf("  ✓ %s (old version kept in the log folder)\n", ui.Safe(filepath.FromSlash(c.rel)))
	}
	return nil
}
