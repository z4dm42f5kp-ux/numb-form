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

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// DateLogName is the log of a numb-date run.
const DateLogName = "numb-date.log"

// Date runs numb-date: for every company whose "Report created" cell has
// no digits, it reads the date from the company's form and, after the
// user agrees, writes it into that cell as DD-MM-YYYY. It is the only
// program that writes to the spreadsheet, and only to those cells.
func Date(opt Options) int {
	cfg := opt.defaults()
	ctx, cancel := runContext(cfg.Timeout)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	r := &run{opt: opt, cfg: cfg, ctx: ctx, ui: ui.New(ctx, opt.In, opt.Out)}
	err := r.dateMain()
	code := ExitOK
	switch {
	case err == nil:
	case errors.Is(err, context.DeadlineExceeded):
		r.ui.Printf("\nTime limit of %s reached. Stopped safely.\n", cfg.Timeout)
		code = ExitInterrupted
	case errors.Is(err, context.Canceled), errors.Is(err, errAbort), errors.Is(err, ui.ErrClosed):
		r.ui.Printf("\nStopped (%v).\n", err)
		code = ExitInterrupted
	default:
		r.ui.Printf("\nnumb-date: %v\n", err)
		code = ExitError
	}
	if r.log != nil {
		r.logf("run finished: %v", err)
		_ = r.log.Close()
	}
	return code
}

type planned struct {
	row  sheet.TableRow
	date string
	from string
}

func hasDigit(s string) bool { return strings.ContainsAny(s, "0123456789") }

func (r *run) dateMain() error {
	r.ui.Printf("numb-date %s — fills the report dates in the overview spreadsheet\n\n", Version)
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
		return fmt.Errorf("numb-date writes only .xlsx and .xlsm spreadsheets; %s is %s", filepath.Base(path), ext)
	}
	if err := notOpen(path); err != nil {
		return err
	}
	b, err := readSheetFile(path)
	if err != nil {
		return err
	}
	have := fsutil.SHA256Bytes(b)
	t, err := sheet.ReadTable(b, r.cfg.Header, r.cfg.IDHeader)
	if err != nil {
		return err
	}
	if err := r.openDateLog(root); err != nil {
		return err
	}
	r.ui.Printf("Working folder: %s\nSpreadsheet:    %s\nSheet %q: companies in column %q, dates in column %q\n",
		root, rel(root, path), t.Sheet, t.NameHeader, t.DateHeader)
	r.logf("spreadsheet %s sha256=%s sheet=%q name=%q date=%q rows=%d", path, have, t.Sheet, t.NameHeader, t.DateHeader, len(t.Rows))
	r.ix, err = indexFolders(root, overview)
	if err != nil {
		return err
	}
	r.ix.Threshold = r.cfg.Fuzzy

	var plan []planned
	counts := map[string]int{}
	for i, row := range t.Rows {
		if err := r.ctx.Err(); err != nil {
			return err
		}
		r.ui.Printf("\n[%d/%d] %s  (%s)\n", i+1, len(t.Rows), ui.Safe(row.Name), ui.Safe(where(r.cfg.IDHeader, row.ID, row.Row)))
		if hasDigit(row.Date) {
			r.ui.Printf("  already has a date: %s\n", ui.Safe(row.Date))
			counts["dated"]++
			continue
		}
		p, ok, err := r.dateFor(row)
		if err != nil {
			return err
		}
		if ok {
			plan = append(plan, p)
		}
	}
	counts["skipped"] = len(r.skipped)
	if len(r.skipped) > 0 {
		r.ui.Printf("\nSkipped (%d), fill these by hand if needed:\n", len(r.skipped))
		for _, s := range r.skipped {
			r.ui.Printf("  %s\n", ui.Safe(s))
		}
	}
	if len(plan) == 0 {
		r.ui.Printf("\nNothing to write: %d already had a date, %d skipped.\n", counts["dated"], counts["skipped"])
		return nil
	}
	return r.writeDates(path, b, have, t, plan, counts)
}

// readSheetFile reads the spreadsheet, at most 256 MB.
func readSheetFile(path string) ([]byte, error) {
	f, err := os.Open(path) // #nosec G304 -- the overview spreadsheet.
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, 256<<20+1))
	if err != nil {
		return nil, err
	}
	if len(b) > 256<<20 {
		return nil, errors.New("the spreadsheet is larger than 256 MB")
	}
	return b, nil
}

// notOpen refuses when Excel or LibreOffice has the file open: both leave
// a lock file next to it.
func notOpen(path string) error {
	dir, base := filepath.Dir(path), filepath.Base(path)
	for _, lock := range []string{"~$" + base, ".~lock." + base + "#"} {
		if _, err := os.Lstat(filepath.Join(dir, lock)); err == nil {
			return fmt.Errorf("%s is open in Excel or LibreOffice; close it and start numb-date again", base)
		}
	}
	return nil
}

func (r *run) openDateLog(root string) error {
	dir := r.cfg.LogDir
	if dir == "" {
		var err error
		if dir, err = config.StateDir(); err != nil {
			return err
		}
	}
	r.runDir = filepath.Join(dir, r.opt.Now().Format("2006-01-02_150405")+"_"+fsutil.RandHex(2)+"_date")
	if fsutil.IsUnder(r.runDir, root) {
		return fmt.Errorf("the log folder %s is inside the working folder; set log_dir elsewhere", dir)
	}
	if err := os.MkdirAll(r.runDir, 0o750); err != nil {
		return err
	}
	var err error
	r.log, err = runlog.Open(filepath.Join(r.runDir, DateLogName))
	return err
}

// writeDates shows the plan, writes it after the user accepts it (or an
// edited one), reads the spreadsheet back and lets the user accept or
// undo.
func (r *run) writeDates(path string, before []byte, have string, t *sheet.Table, plan []planned, counts map[string]int) error {
	plan, ok, err := r.reviewDates(filepath.Base(path), t.DateHeader, plan)
	if err != nil {
		return err
	}
	if !ok {
		r.ui.Println("Nothing was written.")
		r.logf("user declined writing %d dates", len(plan))
		return nil
	}
	vals := map[string]string{}
	for _, p := range plan {
		vals[p.row.DateRef] = p.date
	}
	if err := notOpen(path); err != nil {
		return err
	}
	after, err := sheet.SetText(before, t.Part, vals)
	if err != nil {
		return err
	}
	if err := sheet.CheckEdit(before, after, t.Sheet, vals); err != nil {
		return fmt.Errorf("self-check failed, nothing was written: %w", err)
	}
	backup := filepath.Join(r.runDir, "before-"+filepath.Base(path))
	if err := fsutil.WriteNew(backup, before); err != nil {
		return fmt.Errorf("could not keep a copy of the original: %w", err)
	}
	if err := fsutil.ReplaceVerified(path, have, after); err != nil {
		_ = os.Remove(backup)
		return err
	}
	wrote := fsutil.SHA256Bytes(after)
	r.logf("wrote %d dates: sha256 %s -> %s; original kept at %s", len(plan), have, wrote, backup)
	for _, p := range plan {
		r.logf("  %s %q = %s (%s)", p.row.DateRef, p.row.Name, p.date, p.from)
	}

	// Read the file back from disk and show what the cells now hold.
	got, err := readSheetFile(path)
	if err != nil {
		return err
	}
	nt, err := sheet.ReadTable(got, r.cfg.Header, r.cfg.IDHeader)
	if err != nil {
		return err
	}
	now := map[string]string{}
	for _, row := range nt.Rows {
		now[row.DateRef] = row.Date
	}
	r.ui.Printf("\nWritten. %s now reads:\n", filepath.Base(path))
	for _, p := range plan {
		r.ui.Printf("  %-6s %s  %s\n", p.row.DateRef, ui.Safe(now[p.row.DateRef]), ui.Safe(p.row.Name))
	}
	ch, err := r.ui.Choose("Is this correct? [a]ccept, [u]ndo: ", "au", 0)
	if err != nil {
		r.ui.Printf("\nThe dates were written; the original spreadsheet is kept at %s\n", backup)
		return err
	}
	if ch == 'u' {
		if err := fsutil.ReplaceVerified(path, wrote, before); err != nil {
			r.ui.Printf("Could not undo: %v\nThe original spreadsheet is kept at %s\n", err, backup)
			return err
		}
		_ = os.Remove(backup)
		r.logf("undone: sha256 back to %s", have)
		r.ui.Println("Undone: the spreadsheet is exactly as before (SHA-256 checked).")
		return nil
	}
	_ = os.Remove(backup)
	r.logf("accepted")
	r.ui.Printf("\nDone: %d date(s) written, %d already had a date, %d skipped.\nLog: %s\n",
		len(plan), counts["dated"], len(r.skipped), filepath.Join(r.runDir, DateLogName))
	return nil
}
