// Command numb-typo corrects misspelt or cut-off cities and countries
// written after a comma in the company names of the overview spreadsheet
// ("Foo S.A. de C.V., Ciudad de Me" becomes "..., Ciudad de México"),
// after the user agrees. Only those cells change.
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
	fs := flag.NewFlagSet("numb-typo", flag.ContinueOnError)
	conf := fs.String("f", "", "configuration file read after the system and per-user files")
	root := fs.String("root", "", "working folder (default: the current folder if it is one, else the single folder inside it)")
	timeout := fs.Duration("timeout", 0, "stop safely after this long, such as 30m (default: timeout in the configuration, none: no limit)")
	allowPipe := fs.Bool("allow-non-tty", false, "accept answers from a pipe (testing only)")
	version := fs.Bool("version", false, "print the version and exit")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "usage: numb-typo [-f conf] [-root dir] [-timeout d]\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		return app.ExitUsage
	}
	if *version {
		fmt.Println("numb-typo", app.Version)
		return app.ExitOK
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return app.ExitUsage
	}
	if !*allowPipe && !ui.IsTerminal(os.Stdin) {
		fmt.Fprintln(os.Stderr, "numb-typo: every correction must be confirmed by a person; run it in an interactive terminal")
		return app.ExitUsage
	}
	cfg, notes, err := config.Load(*conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "numb-typo:", err)
		return app.ExitUsage
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "numb-typo:", err)
		return app.ExitError
	}
	code := app.Typo(app.Options{
		Parent: cwd, Root: *root, Config: &cfg, Notes: notes, Timeout: *timeout,
		AllowNonTTY: *allowPipe, In: os.Stdin, Out: os.Stdout, Now: time.Now,
	})
	if ui.OwnConsole() {
		fmt.Print("\nPress Enter to close.")
		_, _ = fmt.Scanln()
	}
	return code
}
