package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/update"
)

// goarm is the 32-bit ARM level of release builds (see build-release.sh).
const goarm = "7"

// Update replaces the installed programs with the newest release of the
// channel ("" means the channel this build came from).
func Update(out io.Writer, cfg config.Config, offline bool, channel string) int {
	if offline || !cfg.Online {
		fmt.Fprintln(out, "update: online access is turned off (-offline or online = no in the configuration)")
		return ExitUsage
	}
	if channel == "" {
		channel = update.DefaultChannel(Version)
	}
	exe, err := os.Executable()
	if err == nil {
		exe, err = filepath.EvalSymlinks(exe)
	}
	if err != nil {
		fmt.Fprintln(out, "update:", err)
		return ExitError
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	fmt.Fprintf(out, "Installed: numb-form %s in %s\n", Version, filepath.Dir(exe))
	v, err := update.Run(ctx, update.Options{
		Channel: channel, Dir: filepath.Dir(exe), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GOARM: goarm,
		Current: Version, Out: out,
	})
	if err != nil {
		fmt.Fprintln(out, "update:", err)
		return ExitError
	}
	if v == Version {
		fmt.Fprintf(out, "Already up to date: numb-form %s (%s).\n", v, channel)
		return ExitOK
	}
	fmt.Fprintf(out, "Updated to numb-form %s (%s). Your settings are kept.\n", v, channel)
	return ExitOK
}

// CleanupOldPrograms removes what an earlier update left on Windows.
func CleanupOldPrograms() {
	if runtime.GOOS != "windows" {
		return
	}
	if exe, err := os.Executable(); err == nil {
		update.Cleanup(filepath.Dir(exe))
	}
}
