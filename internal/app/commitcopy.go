package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
)

func pendingRuns(runs string) ([]string, error) {
	ents, err := os.ReadDir(runs)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		m, err := runlog.Load(filepath.Join(runs, e.Name()))
		if err != nil || m.Committed || m.Mode == "direct" {
			continue
		}
		for _, c := range m.Companies {
			if c.Status == "accepted" && !c.Committed {
				out = append(out, filepath.Join(runs, e.Name()))
				break
			}
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(out)))
	return out, nil
}

func commitOne(ctx context.Context, p plan, lg *runlog.Logger) error {
	tmp := filepath.Join(filepath.Dir(p.dst), ".numb-commit-"+fsutil.RandHex(4))
	if _, err := fsutil.CopyTree(ctx, p.src, tmp, lg.Printf); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}
	verify := func(base string) error {
		for r, want := range p.files {
			relInFolder, _ := filepath.Rel(filepath.Base(p.src), r)
			h, _, err := fsutil.SHA256File(filepath.Join(base, relInFolder))
			if err != nil {
				return err
			}
			if h != want {
				return fmt.Errorf("checksum mismatch for %s", r)
			}
		}
		return nil
	}
	if err := verify(tmp); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}
	if _, err := os.Lstat(p.dst); err == nil {
		_ = os.RemoveAll(tmp)
		return errors.New("destination appeared while copying")
	}
	if err := os.Rename(tmp, p.dst); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}
	return verify(p.dst)
}
