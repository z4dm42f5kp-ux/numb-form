package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

func (r *run) write(target string, vals form.Values) error {
	out, warns, err := r.tpl.Fill(r.ctx, vals)
	if err != nil {
		return err
	}
	for _, w := range warns {
		r.logf("fill note: %s", w)
	}
	r.written += int64(len(out))
	if limit := int64(r.cfg.MaxWriteMB) << 20; r.written > limit {
		return fmt.Errorf("this run would write more than %d MB of documents; stopping", r.cfg.MaxWriteMB)
	}
	return fsutil.WriteNew(target, out)
}

// show re-reads the written file from disk and prints what is in it.
func (r *run) show(target string) (form.Values, error) {
	b, err := os.ReadFile(target) // #nosec G304 -- file written by this run.
	if err != nil {
		return nil, err
	}
	got, err := r.tpl.ReadBack(r.ctx, b)
	if err != nil {
		return nil, err
	}
	r.ui.Printf("\n  Written to %s\n  Section A now reads:\n", filepath.Base(target))
	for i, f := range form.AllFields {
		v := got[f]
		if v == "" {
			v = "(blank)"
		}
		r.ui.Printf("    %d. %s: %s\n", i+1, f.Title(), ui.Safe(v))
	}
	r.ui.Println("  Section B and everything else is unchanged.")
	return got, nil
}

func (r *run) editField(vals form.Values) error {
	for {
		a, err := r.ui.Ask("  Field number to change (1-7, Enter to go back): ")
		if err != nil {
			return err
		}
		if a == "" {
			return nil
		}
		n, err := strconv.Atoi(a)
		if err != nil || n < 1 || n > len(form.AllFields) {
			continue
		}
		f := form.AllFields[n-1]
		v, err := r.ui.Ask(fmt.Sprintf("  New value for %s (empty = blank): ", f.Title()))
		if err != nil {
			return err
		}
		vals[f] = textutil.OneLine(v)
		r.logf("user edited %s", f.Title())
		return nil
	}
}
