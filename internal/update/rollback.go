package update

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Rollback switches back to the version the last update replaced, which
// that update kept as name.previous. The version rolled back from is
// kept the same way, so a second rollback returns to it. It returns the
// version now installed.
func Rollback(ctx context.Context, o Options) (string, error) {
	if o.Probe == nil {
		o.Probe = probe
	}
	ext := ""
	if o.GOOS == "windows" {
		ext = ".exe"
	}
	prev := PreviousPath(o.Dir, Programs[0], ext)
	if _, err := os.Stat(prev); err != nil {
		return "", errors.New("no previous version is kept here (it is kept by numb-form update); install a specific release instead, for example: numb-form update v0.1.1")
	}
	// Ask the kept program for its version before swapping: it must start.
	// On Windows a program only starts with an .exe name, so it is copied
	// to a temporary name first.
	check := prev
	if ext != "" {
		check = filepath.Join(o.Dir, Programs[0]+".check"+ext)
		b, err := os.ReadFile(prev) // #nosec G304 -- our own kept program.
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(check, b, 0o755); err != nil { // #nosec G306 -- programs must be executable.
			return "", err
		}
		defer os.Remove(check)
	}
	v, err := o.Probe(ctx, check)
	if err != nil {
		return "", fmt.Errorf("the previous numb-form does not start (%w); nothing was changed", err)
	}
	for _, name := range Programs {
		cur := filepath.Join(o.Dir, name+ext)
		old := PreviousPath(o.Dir, name, ext)
		if _, err := os.Stat(old); err != nil {
			continue // nothing kept for this program
		}
		tmp := filepath.Join(o.Dir, name+".swap"+ext)
		_ = os.Remove(tmp)
		if err := os.Rename(cur, tmp); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("roll back %s: %w (close every numb-form window and try again)", filepath.Base(cur), err)
		}
		if err := os.Rename(old, cur); err != nil {
			_ = os.Rename(tmp, cur)
			return "", fmt.Errorf("roll back %s: %w", filepath.Base(cur), err)
		}
		if err := os.Rename(tmp, old); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("keep %s for a later rollback: %w", filepath.Base(cur), err)
		}
	}
	return v, nil
}
