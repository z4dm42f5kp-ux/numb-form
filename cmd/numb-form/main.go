// Command numb-form fills the risk assessment form for every company in
// the overview spreadsheet that does not have a folder yet. It works on
// a copy of the tree in the Downloads folder; numb-commit copies the
// reviewed result back.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/app"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "keys":
			return app.Keys(os.Stdout)
		case "check-template":
			if len(os.Args) < 3 || len(os.Args) > 4 {
				fmt.Fprintln(os.Stderr, "usage: numb-form check-template template.docx [sample-output.docx]")
				return app.ExitUsage
			}
			sample := ""
			if len(os.Args) == 4 {
				sample = os.Args[3]
			}
			return app.CheckTemplate(os.Stdout, os.Args[2], sample)
		}
	}
	fs := flag.NewFlagSet("numb-form", flag.ContinueOnError)
	root := fs.String("root", "", "working folder (default: the single folder inside the current directory)")
	downloads := fs.String("downloads", "", "Downloads folder for the review copy (default: the user's Downloads, or $NUMBFORM_DOWNLOADS)")
	offline := fs.Bool("offline", false, "do not contact any online source; type every value yourself")
	timeout := fs.Duration("timeout", app.MaxRunTime, "stop safely after this long (at most 25m)")
	allowPipe := fs.Bool("allow-non-tty", false, "accept answers from a pipe (testing only)")
	version := fs.Bool("version", false, "print the version and exit")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "usage: numb-form [-offline] [-root dir] [-downloads dir] [-timeout d]\n       numb-form check-template template.docx [sample.docx]\n       numb-form keys\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		return app.ExitUsage
	}
	if *version {
		fmt.Println("numb-form", app.Version)
		return app.ExitOK
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return app.ExitUsage
	}
	if *timeout > app.MaxRunTime {
		fmt.Fprintln(os.Stderr, "numb-form: -timeout may not exceed 25m")
		return app.ExitUsage
	}
	if !*allowPipe && !ui.IsTerminal(os.Stdin) {
		fmt.Fprintln(os.Stderr, "numb-form: every document must be confirmed by a person; run it in an interactive terminal (docker/podman: -it, kubectl: -it)")
		return app.ExitUsage
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "numb-form:", err)
		return app.ExitError
	}
	code := app.Run(app.Options{
		Parent:      cwd,
		Root:        *root,
		Downloads:   *downloads,
		Offline:     *offline,
		Timeout:     *timeout,
		AllowNonTTY: *allowPipe,
		In:          os.Stdin,
		Out:         os.Stdout,
		Now:         time.Now,
		Lookup:      lookupConfig(),
	})
	pause()
	return code
}

func lookupConfig() (c app.LookupConfig) {
	c.CompaniesHouseKey = app.LookupKeyFromEnv()
	return c
}

// pause keeps the console window open when the program was started by
// double-clicking it in Explorer.
func pause() {
	if ui.OwnConsole() {
		fmt.Print("\nPress Enter to close.")
		_, _ = fmt.Scanln()
	}
}
