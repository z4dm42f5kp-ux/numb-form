package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// TypoLogName is the log of a numb-typo run.
const TypoLogName = "numb-typo.log"

// Typo runs numb-typo: it corrects misspelt or cut-off cities and
// countries written after a comma in the company names of the overview
// spreadsheet, after the user agrees to each change. Only those cells of
// the company-name column change.
func Typo(opt Options) int {
	cfg := opt.defaults()
	ctx, cancel := runContext(cfg.Timeout)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	r := &run{opt: opt, cfg: cfg, ctx: ctx, ui: ui.New(ctx, opt.In, opt.Out)}
	err := r.typoMain()
	code := ExitOK
	switch {
	case err == nil:
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled), errors.Is(err, errAbort), errors.Is(err, ui.ErrClosed):
		r.ui.Printf("\nStopped (%v).\n", err)
		code = ExitInterrupted
	default:
		r.ui.Printf("\nnumb-typo: %v\n", err)
		code = ExitError
	}
	if r.log != nil {
		r.logf("run finished: %v", err)
		_ = r.log.Close()
	}
	return code
}

type rename struct {
	row   sheet.TableRow
	to    string
	fixes []match.PlaceFix
}

func describe(fixes []match.PlaceFix) string {
	var parts []string
	for _, f := range fixes {
		if f.To != "" {
			parts = append(parts, fmt.Sprintf("%s → %s (%s)", f.From, f.To, f.How))
		}
	}
	return strings.Join(parts, "; ")
}

func (r *run) typoMain() error {
	r.ui.Printf("numb-typo %s — corrects misspelt or cut-off places in the company names\n\n", Version)
	for _, n := range r.opt.Notes {
		r.ui.Printf("note: %s\n", n)
	}
	root, err := r.pickRoot()
	if err != nil {
		return err
	}
	overview, err := findOverview(root, r.cfg.OverviewDir)
	if err != nil {
		return err
	}
	path, _, err := findSpreadsheet(overview)
	if err != nil {
		return err
	}
	if ext := strings.ToLower(filepath.Ext(path)); ext != ".xlsx" && ext != ".xlsm" {
		return fmt.Errorf("numb-typo writes only .xlsx and .xlsm spreadsheets; %s is %s", filepath.Base(path), ext)
	}
	if err := notOpen(path); err != nil {
		return err
	}
	b, err := readSheetFile(path)
	if err != nil {
		return err
	}
	have := fsutil.SHA256Bytes(b)
	t, err := sheet.ReadNames(b, r.cfg.Header, r.cfg.IDHeader)
	if err != nil {
		return err
	}
	if err := r.openLog(root, "_typo", TypoLogName); err != nil {
		return err
	}
	r.ui.Printf("Working folder: %s\nSpreadsheet:    %s\nSheet %q: companies in column %q\n", root, rel(root, path), t.Sheet, t.NameHeader)
	r.logf("spreadsheet %s sha256=%s sheet=%q name=%q rows=%d", path, have, t.Sheet, t.NameHeader, len(t.Rows))
	r.ix, err = indexFolders(root, overview)
	if err != nil {
		return err
	}
	r.ix.Threshold = r.cfg.Fuzzy

	var plan []rename
	withFolder := 0
	for _, row := range t.Rows {
		if err := r.ctx.Err(); err != nil {
			return err
		}
		fixes := match.FixPlaces(row.Raw)
		if len(fixes) == 0 {
			continue
		}
		exact, fuzzy := r.ix.Find(row.Name)
		if len(exact) > 0 || len(fuzzy) > 0 {
			// Renaming it would part it from its folder and form.
			withFolder++
			r.logf("keep %s %q: it has a folder; fixes would be %s", row.NameRef, row.Name, describe(fixes))
			continue
		}
		r.ui.Printf("\n%s  (%s)\n", ui.Safe(row.Name), ui.Safe(where(r.cfg.IDHeader, row.ID, row.Row)))
		for i, f := range fixes {
			if f.To != "" {
				continue
			}
			if fixes[i].To, err = r.pickPlace(f); err != nil {
				return err
			}
		}
		to := match.Apply(row.Raw, fixes)
		if to == row.Raw {
			r.ui.Println("  left as it is")
			continue
		}
		r.ui.Printf("  → %s\n", ui.Safe(strings.TrimSpace(to)))
		plan = append(plan, rename{row: row, to: to, fixes: fixes})
	}
	if withFolder > 0 {
		r.ui.Printf("\n%d company name(s) with a possible typo already have a folder and are left as they are (see the log).\n", withFolder)
	}
	if len(plan) == 0 {
		r.ui.Println("\nNothing to correct.")
		return nil
	}
	return r.writeRenames(path, b, have, t, plan)
}

// pickPlace asks which place a part could be.
func (r *run) pickPlace(f match.PlaceFix) (string, error) {
	r.ui.Printf("  %q (%s) could be:\n", f.From, f.How)
	for i, c := range f.Choices {
		r.ui.Printf("    %d) %s\n", i+1, ui.Safe(c))
	}
	for {
		a, err := r.ui.Ask("  Number, or Enter to leave it: ")
		if err != nil || a == "" {
			return "", err
		}
		if n, err := strconv.Atoi(a); err == nil && n >= 1 && n <= len(f.Choices) {
			return f.Choices[n-1], nil
		}
	}
}

// writeRenames lists the corrections, lets the user take all, choose one
// by one or none, and writes them.
func (r *run) writeRenames(path string, before []byte, have string, t *sheet.Table, plan []rename) error {
	r.ui.Printf("\nCorrections for %s (column %q):\n", filepath.Base(path), t.NameHeader)
	for _, p := range plan {
		r.ui.Printf("  %-6s %s\n         %s\n", p.row.NameRef, ui.Safe(p.row.Name), ui.Safe(strings.TrimSpace(p.to)))
	}
	ch, err := r.ui.Choose(fmt.Sprintf("\nWrite these %d correction(s)? [a]ll, [o]ne by one, [n]one: ", len(plan)), "aon", 'n')
	if err != nil {
		return err
	}
	var edits []cellEdit
	for _, p := range plan {
		take := ch == 'a'
		if ch == 'o' {
			if take, err = r.ui.YesNo(fmt.Sprintf("  %s → %s?", ui.Safe(p.row.Name), ui.Safe(strings.TrimSpace(p.to))), false); err != nil {
				return err
			}
		}
		if take {
			edits = append(edits, cellEdit{ref: p.row.NameRef, value: p.to, label: describe(p.fixes)})
		}
	}
	if len(edits) == 0 {
		r.ui.Println("Nothing was written.")
		r.logf("user wrote no corrections")
		return nil
	}
	accepted, err := r.writeCells(path, before, have, t.Sheet, t.Part, edits)
	if err != nil || !accepted {
		return err
	}
	r.ui.Printf("\nDone: %d name(s) corrected.\nLog: %s\n", len(edits), filepath.Join(r.runDir, TypoLogName))
	return nil
}
