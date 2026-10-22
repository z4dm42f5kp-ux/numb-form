// Package app implements the numb-form and numb-commit programs.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/formats"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/lookup"
	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// MaxRunTime is the hard limit for one run.
const MaxRunTime = config.MaxTimeout

// BigCopyBytes is the review-copy size above which the user is asked.
const BigCopyBytes = 2 << 30

// Exact texts from the previous process; do not change them.
const (
	CautionText = "Caution: This field might differ from internal sources or be innacurate due to insufficient information online"
	NoRepsText  = "No representatives were found online, please consult other internal sources for more accurate information"
	// RegistryText always fills the Commercial Registry Number field; the
	// user may replace it with the number when editing the form.
	// TraceName is the debug file written next to the log of every run.
	TraceName    = "output.txt"
	RegistryText = "No commercial registry number was found online consult internal sources for further information"
)

// Exit codes.
const (
	ExitOK          = 0
	ExitError       = 1
	ExitUsage       = 2
	ExitInterrupted = 3
)

// Options configures a run. Fields left empty fall back to Config.
type Options struct {
	Parent      string // directory the program was started in
	Root        string // explicit working root; empty means auto-detect
	Config      *config.Config
	Notes       []string // notes from loading the configuration
	Downloads   string   // overrides config downloads
	Offline     bool     // overrides config online
	Timeout     time.Duration
	AllowNonTTY bool
	In          io.Reader
	Out         io.Writer
	Now         func() time.Time
	Lookup      lookup.Config
	Client      *lookup.Client
}

type run struct {
	trace        *lookup.Trace // output.txt: every lookup step, for debugging
	traceFile    *os.File
	browserNoted bool // the browser note was shown
	opt          Options
	cfg          config.Config
	ctx          context.Context
	ui           *ui.UI
	log          *runlog.Logger
	man          *runlog.Manifest
	runDir       string
	srcRoot      string // the real working folder
	workRoot     string // where folders are created: the copy (review) or srcRoot (direct)
	tpl          *formats.Template
	ix           *match.Index
	written      int64
	person       string
	date         string
	online       bool
}

var errAbort = errors.New("aborted by user")

func (o *Options) defaults() config.Config {
	c := config.Defaults()
	if o.Config != nil {
		c = *o.Config
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Downloads != "" {
		c.Downloads = o.Downloads
	}
	if o.Offline {
		c.Online = false
	}
	if o.Timeout > 0 && o.Timeout < c.Timeout {
		c.Timeout = o.Timeout
	}
	if c.Timeout <= 0 || c.Timeout > MaxRunTime {
		c.Timeout = MaxRunTime
	}
	if o.Client == nil {
		o.Client = lookup.NewClient()
	}
	return c
}

// Run executes numb-form and returns the process exit code.
func Run(opt Options) int {
	cfg := opt.defaults()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	r := &run{opt: opt, cfg: cfg, ctx: ctx, ui: ui.New(ctx, opt.In, opt.Out), online: cfg.Online}
	err := r.main()
	code := ExitOK
	switch {
	case err == nil:
	case errors.Is(err, context.DeadlineExceeded):
		r.ui.Printf("\nTime limit of %s reached. Stopped safely; accepted documents are kept.\n", cfg.Timeout)
		code = ExitInterrupted
	case errors.Is(err, context.Canceled), errors.Is(err, errAbort), errors.Is(err, ui.ErrClosed):
		r.ui.Printf("\nStopped (%v). Accepted documents are kept.\n", err)
		code = ExitInterrupted
	default:
		r.ui.Printf("\nnumb-form: %v\n", err)
		code = ExitError
	}
	if err != nil && r.man != nil {
		r.man.Interrupted = err.Error()
	}
	r.finish()
	return code
}

func (r *run) logf(format string, a ...any) { r.log.Printf(format, a...) }

func (r *run) finish() {
	if r.man != nil && r.runDir != "" {
		r.man.Finished = r.opt.Now()
		if err := r.man.Save(r.runDir); err != nil {
			r.ui.Printf("warning: could not save manifest: %v\n", err)
		}
		r.logf("run finished")
	}
	if r.log != nil {
		_ = r.log.Close()
	}
	if r.traceFile != nil {
		r.trace.Ev("end")
		_ = r.traceFile.Close()
	}
}

func (r *run) main() error {
	r.ui.Printf("numb-form %s — risk assessment form filler (mode: %s, confirm: %s)\n\n", Version, r.cfg.Mode, r.cfg.Confirm)
	for _, n := range r.opt.Notes {
		r.ui.Printf("note: %s\n", n)
	}
	in, err := r.inputs()
	if err != nil {
		return err
	}
	if r.person, err = r.askPerson(); err != nil {
		return err
	}
	layout, _ := config.GoDateLayout(r.cfg.DateFormat)
	r.date = r.opt.Now().Format(layout)

	if r.cfg.Mode == "review" {
		err = r.copyToDownloads(in)
	} else {
		err = r.startDirect(in)
	}
	if err != nil {
		return err
	}
	workSheet := filepath.Join(r.workRoot, rel(r.srcRoot, in.sheet))
	if h, _, err := fsutil.SHA256File(workSheet); err != nil || h != in.sheetHash {
		return errors.New("the spreadsheet changed before it could be read; stopping")
	}
	for _, n := range in.notes {
		r.logf("note: %s", n)
	}
	rows, warns, legacy, err := formats.LoadSheet(r.ctx, workSheet, r.cfg)
	if err != nil {
		return err
	}
	for _, w := range warns {
		r.logf("spreadsheet note: %s", w)
	}
	if (legacy || r.tpl.Legacy) && r.online {
		r.online = false
		r.ui.Println("note: a legacy format is in use, so this run is fully offline.")
		r.logf("legacy format: online lookups disabled")
	}
	r.ui.Printf("\n%d companies in %s (top to bottom).\n", len(rows), filepath.Base(in.sheet))
	r.logf("spreadsheet %s sha256=%s rows=%d", r.man.Spreadsheet, in.sheetHash, len(rows))

	r.ix, err = indexFolders(r.workRoot, filepath.Join(r.workRoot, rel(r.srcRoot, in.overview)))
	if err != nil {
		return err
	}
	r.ix.Threshold = r.cfg.Fuzzy
	r.logf("indexed %d existing folders", r.ix.Len())
	if r.online {
		r.ui.Println("Online lookups send only company names and websites.")
	}

	var runErr error
	for i, row := range rows {
		if err := r.ctx.Err(); err != nil {
			runErr = err
			break
		}
		r.ui.Printf("\n[%d/%d] %s  (%s!%s)\n", i+1, len(rows), ui.Safe(row.Name), row.Sheet, row.Cell)
		c, err := r.company(row)
		r.man.Companies = append(r.man.Companies, c)
		_ = r.man.Save(r.runDir)
		if err != nil {
			runErr = err
			break
		}
	}
	for _, p := range []string{in.sheet, workSheet} {
		if h, _, err := fsutil.SHA256File(p); err != nil || h != in.sheetHash {
			r.logf("ERROR spreadsheet %s changed during the run", p)
			return errors.New("the spreadsheet changed during the run; check the log")
		}
	}
	r.logf("spreadsheet unchanged (sha256 verified)")
	r.summary()
	if runErr != nil {
		return runErr
	}
	r.man.Complete = true
	return nil
}

func (r *run) summary() {
	counts := map[string]int{}
	for _, c := range r.man.Companies {
		counts[c.Status]++
	}
	r.ui.Printf("\nDone: %d accepted, %d already had a folder, %d rejected, %d need review, %d skipped, %d errors.\n",
		counts["accepted"], counts["exists"], counts["rejected"], counts["needs-review"], counts["skipped"], counts["error"])
	if r.cfg.Mode == "review" {
		r.ui.Printf("Review copy: %s\n", r.workRoot)
	}
	r.ui.Printf("Log:         %s\n", filepath.Join(r.runDir, runlog.LogName))
	if r.trace != nil {
		r.ui.Printf("Debug file:  %s  (send this file when something did not work)\n", filepath.Join(r.runDir, TraceName))
	}
	if r.cfg.Mode == "review" && counts["accepted"] > 0 {
		r.ui.Println("\nNothing has been written to the real folder. After reviewing, run numb-commit to copy the accepted folders back.")
	}
	r.logf("summary %v", counts)
}

func rel(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return r
}

// startDirect prepares a direct run: logs go to the per-user state
// folder and folders are created in the real working folder.
func (r *run) startDirect(in *inputs) error {
	dir := r.cfg.LogDir
	if dir == "" {
		var err error
		if dir, err = config.StateDir(); err != nil {
			return err
		}
	}
	r.srcRoot, r.workRoot = in.root, in.root
	if err := r.openRun(dir, in); err != nil {
		return err
	}
	r.man.Mode = "direct"
	r.logf("direct mode: folders are created in %s", in.root)
	return r.man.Save(r.runDir)
}

func (r *run) openRun(dir string, in *inputs) error {
	stamp := r.opt.Now().Format("2006-01-02_150405")
	r.runDir = filepath.Join(dir, stamp+"_"+fsutil.RandHex(2))
	if fsutil.IsUnder(r.runDir, in.root) {
		return fmt.Errorf("the log folder %s is inside the working folder; set log_dir elsewhere", dir)
	}
	if err := os.MkdirAll(r.runDir, 0o750); err != nil {
		return err
	}
	var err error
	if r.log, err = runlog.Open(filepath.Join(r.runDir, runlog.LogName)); err != nil {
		return err
	}
	// output.txt: a compact record of every lookup step, made for
	// debugging rather than for reading.
	if f, err := os.OpenFile(filepath.Join(r.runDir, TraceName), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600); err == nil {
		r.traceFile, r.trace = f, lookup.NewTrace(f)
		r.trace.Ev("run", "v", Version, "os", runtime.GOOS, "mode", r.cfg.Mode, "confirm", r.cfg.Confirm, "search", r.cfg.Search, "browser", r.cfg.Browser, "online", r.online)
	}
	r.man = &runlog.Manifest{
		Version: 1, Tool: "numb-form " + Version, Started: r.opt.Now(), SourceRoot: in.root,
		Spreadsheet: rel(in.root, in.sheet), SheetSHA256: in.sheetHash, Template: r.tpl.Name, Person: r.person,
	}
	r.logf("source %s", in.root)
	if r.cfg.Confirm == "none" {
		r.logf("confirm = none: documents are accepted automatically after the read-back check")
	}
	return nil
}
