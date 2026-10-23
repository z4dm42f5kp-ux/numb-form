package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

func commitLog(cfg config.Config, opt CommitOptions) (string, *runlog.Logger, error) {
	dir := cfg.LogDir
	if dir == "" {
		var err error
		if dir, err = config.StateDir(); err != nil {
			return "", nil, err
		}
	}
	dir = filepath.Join(dir, opt.Now().Format("2006-01-02_150405")+"_"+fsutil.RandHex(2)+"_commit")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", nil, err
	}
	lg, err := runlog.Open(filepath.Join(dir, runlog.CommitLogName))
	return dir, lg, err
}

// copyChecked copies one file to dst, which must not exist, through a
// temporary name, and checks its SHA-256.
func copyChecked(src, dst, sha string) error {
	tmp := filepath.Join(filepath.Dir(dst), ".numb-commit-"+fsutil.RandHex(4))
	if err := fsutil.CopyFile(src, tmp); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if h := hashOf(tmp); h != sha {
		_ = os.Remove(tmp)
		return fmt.Errorf("%s does not match the request (sha256 %s)", src, h)
	}
	if _, err := os.Lstat(dst); err == nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("%s appeared in the real folder", dst)
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// applyRequest copies new folders whole (built under a temporary name and
// renamed, so a folder appears complete or not at all), then new files.
func applyRequest(ctx context.Context, u *ui.UI, q *request, lg *runlog.Logger) error {
	groups, order := q.newFolders()
	for _, top := range order {
		if err := ctx.Err(); err != nil {
			return err
		}
		if top == "" {
			for _, c := range groups[top] {
				if err := copyChecked(q.staged(c.rel), filepath.Join(q.To, filepath.FromSlash(c.rel)), c.sha); err != nil {
					return err
				}
				lg.Printf("added %s sha256=%s", c.rel, c.sha)
				u.Printf("  ✓ %s\n", ui.Safe(filepath.FromSlash(c.rel)))
			}
			continue
		}
		final := filepath.Join(q.To, filepath.FromSlash(top))
		tmp := filepath.Join(filepath.Dir(final), ".numb-commit-"+fsutil.RandHex(4))
		if err := os.Mkdir(tmp, 0o755); err != nil { // #nosec G301 -- folders on a team share.
			return err
		}
		for _, c := range groups[top] {
			inside := strings.TrimPrefix(c.rel, top+"/")
			dst := filepath.Join(tmp, filepath.FromSlash(inside))
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { // #nosec G301
				_ = os.RemoveAll(tmp)
				return err
			}
			if err := copyChecked(q.staged(c.rel), dst, c.sha); err != nil {
				_ = os.RemoveAll(tmp)
				return err
			}
		}
		if _, err := os.Lstat(final); err == nil {
			_ = os.RemoveAll(tmp)
			return fmt.Errorf("%s appeared in the real folder", final)
		}
		if err := os.Rename(tmp, final); err != nil {
			_ = os.RemoveAll(tmp)
			return err
		}
		lg.Printf("added folder %s (%d files, sha256 verified)", top, len(groups[top]))
		u.Printf("  ✓ %s\n", ui.Safe(filepath.FromSlash(top)))
	}
	return nil
}
