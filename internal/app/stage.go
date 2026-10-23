package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
)

// stagePrefix starts the name of every staged copy, so that numb-commit
// only ever removes folders it made.
const stagePrefix = "numb-commit staged copy "

// stageDir picks the folder for the staged copy: in Downloads, never in
// the working copy or the real folder.
func stageDir(downloads, from, to string, now time.Time) (string, error) {
	if downloads == "" {
		var err error
		if downloads, err = fsutil.DownloadsDir(); err != nil {
			return "", err
		}
	}
	dir := filepath.Join(downloads, stagePrefix+now.Format("2006-01-02 15.04.05")+" "+fsutil.RandHex(2))
	if fsutil.IsUnder(dir, from) || fsutil.IsUnder(dir, to) {
		return "", fmt.Errorf("the staged copy would be inside %s or %s; set downloads to another folder", from, to)
	}
	return dir, nil
}

// stage copies exactly the files to add and to replace into dir, laid out
// as in the real folder, checks every SHA-256 and makes the copies
// read-only: the staged copy is what is reviewed and then copied, and
// edits belong in the working copy. Nothing touches the real folder.
func (q *request) stage(ctx context.Context, dir string) error {
	if err := os.MkdirAll(filepath.Dir(dir), 0o750); err != nil {
		return err
	}
	if err := os.Mkdir(dir, 0o750); err != nil {
		return err
	}
	q.Stage = dir
	for _, c := range q.Changes {
		if c.kind != kAdd && c.kind != kReplace {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		dst := filepath.Join(dir, filepath.FromSlash(c.rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
			return err
		}
		if err := copyChecked(filepath.Join(q.From, filepath.FromSlash(c.rel)), dst, c.sha); err != nil {
			return err
		}
		if err := os.Chmod(dst, 0o444); err != nil { // #nosec G302 -- read-only on purpose: the staged copy must not be edited.
			return err
		}
	}
	return nil
}

// unstage removes one file from the staged copy.
func (q *request) unstage(rel string) {
	if q.Stage == "" {
		return
	}
	p := filepath.Join(q.Stage, filepath.FromSlash(rel))
	_ = os.Chmod(p, 0o600)
	_ = os.Remove(p)
}

// removeStage deletes a staged copy, after checking that it is one.
func removeStage(dir string) error {
	if dir == "" {
		return nil
	}
	if !strings.HasPrefix(filepath.Base(dir), stagePrefix) {
		return errors.New("refusing to delete " + dir + ": not a staged copy")
	}
	// Read-only files cannot be deleted on Windows.
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.Type().IsRegular() {
			_ = os.Chmod(p, 0o600)
		}
		return nil
	})
	return os.RemoveAll(dir)
}
