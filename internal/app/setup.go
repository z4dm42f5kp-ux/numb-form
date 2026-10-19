package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/formats"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
)

// inputs are the files a run works from.
type inputs struct {
	root, overview, sheet string
	sheetHash             string
	notes                 []string
}

func (r *run) inputs() (*inputs, error) {
	root, err := r.pickRoot()
	if err != nil {
		return nil, err
	}
	in := &inputs{root: root}
	if in.overview, err = findOverview(root, r.cfg.OverviewDir); err != nil {
		return nil, err
	}
	var others []string
	if in.sheet, others, err = findSpreadsheet(in.overview); err != nil {
		return nil, err
	}
	tplPath, err := findTemplate(root, r.cfg.TemplateName)
	if err != nil {
		return nil, err
	}
	if r.tpl, err = formats.LoadTemplate(r.ctx, tplPath, r.cfg); err != nil {
		return nil, fmt.Errorf("template %s: %w", filepath.Base(tplPath), err)
	}
	warns, err := r.tpl.Check()
	if err != nil {
		return nil, fmt.Errorf("template %s: %w", filepath.Base(tplPath), err)
	}
	if in.sheetHash, _, err = fsutil.SHA256File(in.sheet); err != nil {
		return nil, err
	}
	r.ui.Printf("Working folder: %s\nSpreadsheet:    %s (read only)\nTemplate:       %s\n", root, rel(root, in.sheet), r.tpl.Name)
	for _, o := range others {
		r.ui.Printf("note: %s also contains %q; it is ignored\n", filepath.Base(in.overview), o)
	}
	for _, n := range append(r.tpl.Notes, warns...) {
		r.ui.Printf("note: template: %s\n", n)
		in.notes = append(in.notes, "template: "+n)
	}
	return in, nil
}

func (r *run) pickRoot() (string, error) {
	if r.opt.Root != "" {
		return filepath.Abs(r.opt.Root)
	}
	parent, err := filepath.Abs(r.opt.Parent)
	if err != nil {
		return "", err
	}
	dirs, err := subdirs(parent)
	if err != nil {
		return "", err
	}
	switch len(dirs) {
	case 0:
		return "", fmt.Errorf("no folder found inside %s; start numb-form in the folder above the working folder", parent)
	case 1:
		return dirs[0], nil
	}
	var withOverview []string
	for _, d := range dirs {
		if _, err := findOverview(d, r.cfg.OverviewDir); err == nil {
			withOverview = append(withOverview, d)
		}
	}
	if len(withOverview) == 1 {
		return withOverview[0], nil
	}
	if r.cfg.Confirm == "none" {
		return "", errors.New("several folders could be the working folder; use -root")
	}
	cands := withOverview
	if len(cands) == 0 {
		cands = dirs
	}
	r.ui.Println("Several folders could be the working folder:")
	for i, d := range cands {
		r.ui.Printf("  %d) %s\n", i+1, filepath.Base(d))
	}
	for {
		a, err := r.ui.Ask("Choose a number (or q to quit): ")
		if err != nil {
			return "", err
		}
		if a == "q" {
			return "", errAbort
		}
		var n int
		if _, err := fmt.Sscanf(a, "%d", &n); err == nil && n >= 1 && n <= len(cands) {
			return cands[n-1], nil
		}
	}
}

func (r *run) askPerson() (string, error) {
	if p := strings.Join(strings.Fields(r.cfg.Person), " "); p != "" {
		r.ui.Printf("\nPerson processing the form (from the configuration): %s\n", p)
		return p, nil
	}
	for {
		a, err := r.ui.Ask("\nFull name of the person processing the form: ")
		if err != nil {
			return "", err
		}
		a = strings.Join(strings.Fields(a), " ")
		if a == "" {
			continue
		}
		ok, err := r.ui.YesNo(fmt.Sprintf("Use %q for every form in this run?", a), true)
		if err != nil {
			return "", err
		}
		if ok {
			return a, nil
		}
	}
}
