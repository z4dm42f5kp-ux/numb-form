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
	"strings"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/docx"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/lookup"
	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
	"github.com/z4dm42f5kp-ux/numb-form/internal/xlsx"
)

// Limits required by the project rules.
const (
	MaxRunTime     = 25 * time.Minute
	MaxScratchByte = 100 << 20 // bytes the run may write besides the copy
	BigCopyBytes   = 2 << 30
)

// Exact texts from the previous process; do not change them.
const (
	CautionText = "Caution: This field might differ from internal sources or be innacurate due to insufficient information online"
	NoRepsText  = "No representatives were found online, please consult other internal sources for more accurate information"
)

// Exit codes.
const (
	ExitOK          = 0
	ExitError       = 1
	ExitUsage       = 2
	ExitInterrupted = 3
)

// Options configures a run.
type Options struct {
	Parent      string // directory the program was started in
	Root        string // explicit working root; empty means auto-detect
	Downloads   string // override for the Downloads folder
	Offline     bool
	Timeout     time.Duration
	AllowNonTTY bool
	In          io.Reader
	Out         io.Writer
	Now         func() time.Time
	Lookup      lookup.Config
	Client      *lookup.Client
}

type run struct {
	opt      Options
	ctx      context.Context
	ui       *ui.UI
	log      *runlog.Logger
	man      *runlog.Manifest
	runDir   string
	copyRoot string
	template []byte
	tplName  string
	ix       *match.Index
	scratch  int64
	person   string
	date     string
}

var errAbort = errors.New("aborted by user")

// Run executes numb-form and returns the process exit code.
func Run(opt Options) int {
	if opt.Now == nil {
		opt.Now = time.Now
	}
	if opt.Timeout <= 0 || opt.Timeout > MaxRunTime {
		opt.Timeout = MaxRunTime
	}
	if opt.Client == nil {
		opt.Client = lookup.NewClient()
	}
	ctx, cancel := context.WithTimeout(context.Background(), opt.Timeout)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	r := &run{opt: opt, ctx: ctx, ui: ui.New(ctx, opt.In, opt.Out)}
	err := r.main()
	code := ExitOK
	switch {
	case err == nil:
	case errors.Is(err, context.DeadlineExceeded):
		r.ui.Printf("\nTime limit of %s reached. Stopped safely; accepted documents are kept.\n", opt.Timeout)
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
}

func (r *run) main() error {
	u := r.ui
	u.Printf("numb-form %s — risk assessment form filler\n\n", Version)

	root, err := r.pickRoot()
	if err != nil {
		return err
	}
	overview, err := findOverview(root)
	if err != nil {
		return err
	}
	sheet, others, err := findSpreadsheet(overview)
	if err != nil {
		return err
	}
	tplPath, err := findTemplate(root)
	if err != nil {
		return err
	}
	tpl, err := os.ReadFile(tplPath) // #nosec G304 -- user's own template.
	if err != nil {
		return err
	}
	if len(tpl) > 50<<20 {
		return errors.New("template is larger than 50 MB")
	}
	warns, err := docx.Check(tpl)
	if err != nil {
		return fmt.Errorf("template %s: %w", filepath.Base(tplPath), err)
	}
	r.template, r.tplName = tpl, filepath.Base(tplPath)
	sheetHash, _, err := fsutil.SHA256File(sheet)
	if err != nil {
		return err
	}
	u.Printf("Working folder: %s\nSpreadsheet:    %s (read only)\nTemplate:       %s\n", root, rel(root, sheet), r.tplName)
	for _, o := range others {
		u.Printf("note: %s also contains %q; it is ignored\n", filepath.Base(overview), o)
	}
	for _, w := range warns {
		u.Printf("note: template: %s\n", w)
	}

	r.person, err = r.askPerson()
	if err != nil {
		return err
	}
	r.date = r.opt.Now().Format("02/01/2006")

	if err := r.copyToDownloads(root); err != nil {
		return err
	}
	r.man.Spreadsheet = rel(root, sheet)
	r.man.SheetSHA256 = sheetHash
	r.man.Template = r.tplName
	r.man.Person = r.person
	copySheet := filepath.Join(r.copyRoot, rel(root, sheet))
	if h, _, err := fsutil.SHA256File(copySheet); err != nil || h != sheetHash {
		return errors.New("the copied spreadsheet differs from the original; stopping")
	}
	for _, w := range warns {
		r.logf("template note: %s", w)
	}

	rows, xw, err := xlsx.ReadCompanies(copySheet)
	if err != nil {
		return err
	}
	for _, w := range xw {
		r.logf("spreadsheet note: %s", w)
	}
	u.Printf("\n%d companies in %s (top to bottom).\n", len(rows), filepath.Base(sheet))
	r.logf("spreadsheet %s sha256=%s rows=%d", r.man.Spreadsheet, sheetHash, len(rows))

	r.ix, err = indexFolders(r.copyRoot, filepath.Join(r.copyRoot, rel(root, overview)))
	if err != nil {
		return err
	}
	r.logf("indexed %d existing folders", r.ix.Len())

	if !r.opt.Offline {
		r.ui.Printf("Online lookups: GLEIF, Wikidata%s, and the company's own legal notice. Only company names and websites are sent.\n",
			map[bool]string{true: ", Companies House", false: ""}[r.opt.Lookup.CompaniesHouseKey != ""])
	}

	var runErr error
	for i, row := range rows {
		if err := r.ctx.Err(); err != nil {
			runErr = err
			break
		}
		u.Printf("\n[%d/%d] %s  (%s!%s)\n", i+1, len(rows), ui.Safe(row.Name), row.Sheet, row.Cell)
		c, err := r.company(row)
		r.man.Companies = append(r.man.Companies, c)
		_ = r.man.Save(r.runDir)
		if err != nil {
			runErr = err
			break
		}
	}

	// The spreadsheet must be unchanged, in the original and in the copy.
	for _, p := range []string{sheet, copySheet} {
		if h, _, err := fsutil.SHA256File(p); err != nil || h != sheetHash {
			r.logf("ERROR spreadsheet %s changed during the run", p)
			return errors.New("the spreadsheet changed during the run; do not commit this run")
		}
	}
	r.logf("spreadsheet unchanged (sha256 verified)")
	if runErr != nil {
		r.summary()
		return runErr
	}
	r.man.Complete = true
	r.summary()
	return nil
}

func (r *run) summary() {
	counts := map[string]int{}
	for _, c := range r.man.Companies {
		counts[c.Status]++
	}
	r.ui.Printf("\nDone: %d accepted, %d already had a folder, %d rejected, %d skipped, %d errors.\n",
		counts["accepted"], counts["exists"], counts["rejected"], counts["skipped"], counts["error"])
	r.ui.Printf("Review copy: %s\nLog:         %s\n", r.copyRoot, filepath.Join(r.runDir, runlog.LogName))
	if counts["accepted"] > 0 {
		r.ui.Printf("\nNothing has been written to the real folder. After reviewing, run numb-commit to copy the accepted folders back.\n")
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

func (r *run) pickRoot() (string, error) {
	if r.opt.Root != "" {
		return filepath.Abs(r.opt.Root)
	}
	parent, err := filepath.Abs(r.opt.Parent)
	if err != nil {
		return "", err
	}
	dirs, err := subdirs(parent)
	if err != nil {
		return "", err
	}
	switch len(dirs) {
	case 0:
		return "", fmt.Errorf("no folder found inside %s; start numb-form in the folder above the working folder", parent)
	case 1:
		return dirs[0], nil
	}
	var withOverview []string
	for _, d := range dirs {
		if _, err := findOverview(d); err == nil {
			withOverview = append(withOverview, d)
		}
	}
	if len(withOverview) == 1 {
		return withOverview[0], nil
	}
	cands := withOverview
	if len(cands) == 0 {
		cands = dirs
	}
	r.ui.Println("Several folders could be the working folder:")
	for i, d := range cands {
		r.ui.Printf("  %d) %s\n", i+1, filepath.Base(d))
	}
	for {
		a, err := r.ui.Ask("Choose a number (or q to quit): ")
		if err != nil {
			return "", err
		}
		if a == "q" {
			return "", errAbort
		}
		var n int
		if _, err := fmt.Sscanf(a, "%d", &n); err == nil && n >= 1 && n <= len(cands) {
			return cands[n-1], nil
		}
	}
}

func (r *run) askPerson() (string, error) {
	for {
		a, err := r.ui.Ask("\nFull name of the person processing the form: ")
		if err != nil {
			return "", err
		}
		a = strings.Join(strings.Fields(a), " ")
		if a == "" {
			continue
		}
		ok, err := r.ui.YesNo(fmt.Sprintf("Use %q for every form in this run?", a), true)
		if err != nil {
			return "", err
		}
		if ok {
			return a, nil
		}
	}
}

func (r *run) copyToDownloads(root string) error {
	dl := r.opt.Downloads
	var err error
	if dl == "" {
		dl, err = fsutil.DownloadsDir()
		if err != nil {
			return err
		}
	}
	if st, err := os.Stat(dl); err != nil || !st.IsDir() {
		return fmt.Errorf("the Downloads folder %s was not found (set NUMBFORM_DOWNLOADS or use -downloads)", dl)
	}
	if rr, err := filepath.Rel(root, dl); err == nil && !strings.HasPrefix(rr, "..") {
		return errors.New("the Downloads folder is inside the working folder")
	}
	runs := runlog.RunsDir(dl)
	r.warnPending(runs)
	st, err := fsutil.Measure(r.ctx, root)
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
	stamp := r.opt.Now().Format("2006-01-02_150405")
	r.runDir = filepath.Join(runs, stamp+"_"+fsutil.RandHex(2))
	if err := os.MkdirAll(r.runDir, 0o750); err != nil {
		return err
	}
	r.log, err = runlog.Open(filepath.Join(r.runDir, runlog.LogName))
	if err != nil {
		return err
	}
	r.copyRoot = filepath.Join(r.runDir, filepath.Base(root))
	r.man = &runlog.Manifest{Version: 1, Tool: "numb-form " + Version, Started: r.opt.Now(), SourceRoot: root, CopyRoot: r.copyRoot}
	r.logf("source %s", root)
	r.logf("copy   %s", r.copyRoot)
	start := time.Now()
	cst, err := fsutil.CopyTree(r.ctx, root, r.copyRoot, r.logf)
	if err != nil {
		r.logf("ERROR copy failed: %v; removing partial copy", err)
		_ = os.RemoveAll(r.copyRoot)
		return fmt.Errorf("copy to Downloads failed: %w", err)
	}
	r.logf("copied %d files, %d folders, %d bytes in %s; skipped %d", cst.Files, cst.Dirs, cst.Bytes, time.Since(start).Round(time.Millisecond), len(cst.Skipped))
	r.ui.Printf("Copied to %s\n", r.copyRoot)
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
