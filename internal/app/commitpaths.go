package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/formats"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// Prompts for the two folders, unless -from/-to or commit_from/commit_to
// answer them.
const (
	askFrom = "Which directory should be committed (insert absolute path)?: "
	askTo   = "Which directory is the target?: "
)

// folderMode reports whether numb-commit copies a working copy into the
// real folder rather than committing a review run. It does unless -run
// is given, or nothing names a folder and either mode = review or review
// runs are waiting in Downloads.
func folderMode(opt CommitOptions) bool {
	if opt.RunDir != "" {
		return false
	}
	cfg := opt.Config
	if opt.Review || opt.From != "" || opt.To != "" || (cfg != nil && (cfg.CommitFrom != "" || cfg.CommitTo != "")) {
		return true
	}
	if cfg != nil && cfg.Mode == "review" {
		return false
	}
	dl := opt.Downloads
	if dl == "" {
		dl, _ = fsutil.DownloadsDir()
	}
	runs, _ := pendingRuns(runlog.RunsDir(dl))
	return len(runs) == 0
}

// cleanPath removes the quotes that "Copy as path" adds and surrounding
// spaces.
func cleanPath(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	return s
}

// pickFolder returns given after checking it, or asks until an answer
// passes the check. A given folder that fails is an error: it came from
// the command line or the configuration.
func pickFolder(u *ui.UI, given, prompt string, check func(string) error) (string, error) {
	if given != "" {
		p, err := filepath.Abs(cleanPath(given))
		if err != nil {
			return "", err
		}
		return p, check(p)
	}
	for {
		a, err := u.Ask(prompt)
		if err != nil {
			return "", err
		}
		p := cleanPath(a)
		switch {
		case p == "":
			continue
		case !filepath.IsAbs(p):
			u.Println(`  Give the whole path, for example C:\Users\you\Downloads\14. GWG or F:\14. GWG.`)
			continue
		}
		p = filepath.Clean(p)
		if err := check(p); err != nil {
			u.Printf("  %v\n", err)
			continue
		}
		return p, nil
	}
}

// folders resolves the working copy (from) and the real folder (to).
func (opt CommitOptions) folders(u *ui.UI) (from, to string, err error) {
	cfg := *opt.Config
	isWF := func(what string) func(string) error {
		return func(p string) error {
			if st, err := os.Stat(p); err != nil || !st.IsDir() {
				return fmt.Errorf("%s is not reachable (is the drive connected?)", p)
			}
			if !isWorkingFolder(p, cfg) {
				return fmt.Errorf("%s is not %s: it has no template and %q folder", p, what, cfg.OverviewDir)
			}
			return nil
		}
	}
	if from, err = pickFolder(u, firstOf(opt.From, cfg.CommitFrom), askFrom, isWF("a working folder")); err != nil {
		return "", "", err
	}
	checkTo := func(p string) error {
		switch {
		case strings.EqualFold(p, from):
			return errors.New("the target is the folder to commit")
		case fsutil.IsUnder(from, p) || fsutil.IsUnder(p, from):
			return errors.New("the two folders must not be inside each other")
		}
		return isWF("a working folder")(p)
	}
	if to, err = pickFolder(u, firstOf(opt.To, cfg.CommitTo), askTo, checkTo); err != nil {
		return "", "", err
	}
	return from, to, nil
}

func firstOf(a ...string) string {
	for _, s := range a {
		if s != "" {
			return s
		}
	}
	return ""
}

// reviewSetup reads what the scan and the review need from the working
// copy: the spreadsheet's companies and ids, and the templates.
func reviewSetup(ctx context.Context, from, to string, cfg config.Config) (rc reviewConfig, err error) {
	overview, err := findOverview(from, cfg.OverviewDir)
	if err != nil {
		return rc, err
	}
	sp, _, err := findSpreadsheet(overview)
	if err != nil {
		return rc, err
	}
	tpl, err := findTemplate(from, cfg.TemplateName)
	if err != nil {
		return rc, err
	}
	rc = reviewConfig{header: cfg.Header, idHeader: cfg.IDHeader, overviewDir: cfg.OverviewDir, sheetRel: rel(from, sp),
		template: filepath.ToSlash(rel(from, tpl)), templateBase: strings.TrimSuffix(filepath.Base(tpl), filepath.Ext(tpl)),
		names: map[string]string{}, ids: map[string]string{}}
	for _, p := range []string{tpl, filepath.Join(to, filepath.FromSlash(rc.template))} {
		if t, err := formats.LoadTemplate(ctx, p, cfg); err == nil {
			rc.tpls = append(rc.tpls, t)
		}
	}
	if b, err := os.ReadFile(sp); err == nil { // #nosec G304 -- the overview spreadsheet.
		if rows, _, err := sheet.Read(b, filepath.Ext(sp), sheet.Options{Header: cfg.Header, IDHeader: cfg.IDHeader, Legacy: cfg.LegacyFormats}); err == nil {
			for _, r := range rows {
				n, _ := fsutil.SafeName(r.Name, false)
				rc.names[strings.ToLower(n)] = r.Name
				rc.ids[strings.ToLower(n)] = r.CustomerID
			}
		}
	}
	return rc, nil
}
