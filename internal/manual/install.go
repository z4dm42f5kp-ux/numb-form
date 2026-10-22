package manual

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// manCmd lets "man numb-form" work in Command Prompt. PowerShell keeps
// "man" for its own help, so there the command is "numb-form man".
const manCmd = `@echo off
rem Installed by "numb-form man -install": "man numb-form" shows the numb-form manual.
if /i "%~1"=="numb-form" goto show
if /i "%~1"=="numb-commit" goto show
if /i "%~1"=="numb-date" goto show
echo man: no manual entry for %1. This man only knows numb-form: man numb-form
exit /b 1
:show
"%~dp0numb-form.exe" man %2 %3
`

// UnixDir is the per-user manual folder that man(1) searches for
// programs installed in ~/.local/bin.
func UnixDir() (string, error) {
	if d := os.Getenv("XDG_DATA_HOME"); d != "" && filepath.IsAbs(d) {
		return filepath.Join(d, "man", "man1"), nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".local", "share", "man", "man1"), nil
}

// Install makes "man numb-form" work: on Unix it writes numb-form.1 and
// numb-commit.1 to the per-user manual folder; on Windows it writes
// man.cmd next to numb-form.exe. It returns what was written.
func Install(roff string) ([]string, error) {
	if runtime.GOOS == "windows" {
		exe, err := os.Executable()
		if err != nil {
			return nil, err
		}
		p := filepath.Join(filepath.Dir(exe), "man.cmd")
		if err := os.WriteFile(p, []byte(manCmd), 0o644); err != nil { // #nosec G306 -- a launcher next to the program.
			return nil, err
		}
		return []string{p}, nil
	}
	dir, err := UnixDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil { // #nosec G301 -- manual pages are public.
		return nil, err
	}
	p1 := filepath.Join(dir, "numb-form.1")
	if err := os.WriteFile(p1, []byte(roff), 0o644); err != nil { // #nosec G306 -- manual pages are public.
		return nil, err
	}
	out := []string{p1}
	for _, n := range []string{"numb-commit.1", "numb-date.1"} {
		p := filepath.Join(dir, n)
		if err := os.WriteFile(p, []byte(".so man1/numb-form.1\n"), 0o644); err != nil { // #nosec G306
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// Installed reports whether the Unix manual page exists, so that an
// update can refresh it.
func Installed() bool {
	if runtime.GOOS == "windows" {
		return false
	}
	dir, err := UnixDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, "numb-form.1"))
	return !errors.Is(err, os.ErrNotExist) && err == nil
}

// Hint tells where the manual went and how to read it.
func Hint(paths []string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("Wrote %s.\nIn Command Prompt: man numb-form\nIn PowerShell:     numb-form man\n", paths[0])
	}
	return fmt.Sprintf("Wrote %s.\nRead it with: man numb-form\nIf man does not find it, add %s to MANPATH.\n", paths[0], filepath.Dir(filepath.Dir(paths[0])))
}
