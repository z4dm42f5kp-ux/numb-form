// Command numb-form fills the risk assessment form for every company in
// the overview spreadsheet that does not have a folder yet.
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

const usage = `usage: numb-form [-f conf] [-offline] [-root dir] [-downloads dir] [-timeout d]
       numb-form [-f conf] check-template template [sample]
       numb-form [-f conf] config [init]
       numb-form keys

`

func realMain() int {
	fs := flag.NewFlagSet("numb-form", flag.ContinueOnError)
	conf := fs.String("f", "", "configuration file read after the system and per-user files")
	root := fs.String("root", "", "working folder (default: the single folder inside the current directory)")
	downloads := fs.String("downloads", "", "review mode: folder for the copy (default: the user's Downloads)")
	offline := fs.Bool("offline", false, "do not contact any online source")
	timeout := fs.Duration("timeout", 0, "stop safely after this long (at most 25m)")
	allowPipe := fs.Bool("allow-non-tty", false, "accept answers from a pipe (testing only)")
	version := fs.Bool("version", false, "print the version and exit")
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), usage)
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		return app.ExitUsage
	}
	if *version {
		fmt.Println("numb-form", app.Version)
		return app.ExitOK
	}
	args := fs.Args()
	if len(args) > 0 && args[0] == "keys" {
		return app.Keys(os.Stdout)
	}
	cfg, notes, err := config.Load(*conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "numb-form:", err)
		return app.ExitUsage
	}
	if len(args) > 0 {
		switch {
		case args[0] == "config" && len(args) == 1:
			return app.ShowConfig(os.Stdout, cfg, notes)
		case args[0] == "config" && len(args) == 2 && args[1] == "init":
			return app.InitConfig(os.Stdout)
		case args[0] == "check-template" && (len(args) == 2 || len(args) == 3):
			sample := ""
			if len(args) == 3 {
				sample = args[2]
			}
			return app.CheckTemplate(os.Stdout, cfg, args[1], sample)
		}
		fs.Usage()
		return app.ExitUsage
	}
	if *timeout > app.MaxRunTime {
		fmt.Fprintln(os.Stderr, "numb-form: -timeout may not exceed 25m")
		return app.ExitUsage
	}
	if cfg.Confirm == "each" && !*allowPipe && !ui.IsTerminal(os.Stdin) {
		fmt.Fprintln(os.Stderr, "numb-form: every document must be confirmed by a person; run it in an interactive terminal (docker/podman: -it, kubectl: -it), or set confirm = none")
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
		Config:      &cfg,
		Notes:       notes,
		Downloads:   *downloads,
		Offline:     *offline,
		Timeout:     *timeout,
		AllowNonTTY: *allowPipe,
		In:          os.Stdin,
		Out:         os.Stdout,
		Now:         time.Now,
		Lookup:      app.LookupConfig{CompaniesHouseKey: app.LookupKeyFromEnv()},
	})
	pause()
	return code
}

// pause keeps the console window open when the program was started by
// double-clicking it on Windows.
func pause() {
	if ui.OwnConsole() {
		fmt.Print("\nPress Enter to close.")
		_, _ = fmt.Scanln()
	}
}
