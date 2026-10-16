// Command numb-commit copies the folders accepted in a numb-form run from
// the review copy in Downloads back into the real folder. It only adds
// new folders, never overwrites or deletes, and verifies every file with
// SHA-256.
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
	allowPipe := fs.Bool("allow-non-tty", false, "accept answers from a pipe (testing only)")
	version := fs.Bool("version", false, "print the version and exit")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "usage: numb-commit [-f conf] [-run dir] [-downloads dir]\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		return app.ExitUsage
	}
	if *version {
		fmt.Println("numb-commit", app.Version)
		return app.ExitOK
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return app.ExitUsage
	}
	if !*allowPipe && !ui.IsTerminal(os.Stdin) {
		fmt.Fprintln(os.Stderr, "numb-commit: the commit must be confirmed by a person; run it in an interactive terminal")
		return app.ExitUsage
	}
	cfg, _, err := config.Load(*conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "numb-commit:", err)
		return app.ExitUsage
	}
	if *downloads == "" {
		*downloads = cfg.Downloads
	}
	code := app.Commit(app.CommitOptions{
		Downloads: *downloads,
		RunDir:    *run,
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
