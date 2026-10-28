// Command numb-commit copies the new forms from a working copy (or a
// review run) into the real working folder. It stages the files in
// Downloads, writes a request with the folder tree before and after, has
// it reviewed, and copies only after the person agrees. It only adds,
// never overwrites (except the overview spreadsheet) or deletes, and
// verifies every file with SHA-256.
//
// In plain words: it copies the new forms from your copy into the real
// working folder, but only after you have seen and accepted the list. It
// adds files and may write report dates into the overview spreadsheet;
// it never deletes anything.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/app"
	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	fs := flag.NewFlagSet("numb-commit", flag.ContinueOnError)
	conf := fs.String("f", "", "configuration file read after the system and per-user files")
	run := fs.String("run", "", "run folder to commit (default: the newest uncommitted run)")
	downloads := fs.String("downloads", "", "Downloads folder (default: the user's Downloads, or $NUMBFORM_DOWNLOADS)")
	from := fs.String("from", "", "working copy to commit (default: commit_from in the configuration, else asked)")
	to := fs.String("to", "", "real working folder to copy into (default: commit_to in the configuration, else asked)")
	allowPipe := fs.Bool("allow-non-tty", false, "accept answers from a pipe (testing only)")
	version := fs.Bool("version", false, "print the version and exit")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "usage: numb-commit [-f conf] [-from dir] [-to dir] [review]\n       numb-commit [-f conf] [-run dir] [-downloads dir]\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		return app.ExitUsage
	}
	if *version {
		fmt.Println("numb-commit", app.Version)
		return app.ExitOK
	}
	review := fs.NArg() == 1 && fs.Arg(0) == "review"
	if fs.NArg() != 0 && !review {
		fs.Usage()
		return app.ExitUsage
	}
	if !*allowPipe && !ui.IsTerminal(os.Stdin) {
		fmt.Fprintln(os.Stderr, "numb-commit: the commit must be confirmed by a person; run it in an interactive terminal")
		return app.ExitUsage
	}
	cfg, notes, err := config.Load(*conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "numb-commit:", err)
		return app.ExitUsage
	}
	if *downloads == "" {
		*downloads = cfg.Downloads
	}
	for _, n := range notes {
		fmt.Println("note:", n)
	}
	code := app.Commit(app.CommitOptions{
		Downloads: *downloads,
		RunDir:    *run,
		From:      *from,
		To:        *to,
		Review:    review,
		Config:    &cfg,
		In:        os.Stdin,
		Out:       os.Stdout,
		Now:       time.Now,
	})
	if ui.OwnConsole() {
		fmt.Print("\nPress Enter to close.")
		_, _ = fmt.Scanln()
	}
	return code
}
