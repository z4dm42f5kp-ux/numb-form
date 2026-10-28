package app

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/reportdate"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// formExts are the forms numb-date can read.
var formExts = map[string]bool{".docx": true, ".docm": true, ".odt": true, ".fodt": true, ".rtf": true}

// maxForm bounds the forms numb-form reads.
const maxForm = 50 << 20

type formFile struct {
	path, rel string
	named     bool // the name looks like the risk assessment form
	mod       int64
}

// formFiles lists the readable documents in a company folder and one
// level below it, the risk assessment form first, newest first.
func formFiles(folder, templateName string) []formFile {
	var out []formFile
	keys := textutil.MatchKeys(templateName)
	depth := strings.Count(filepath.Clean(folder), string(filepath.Separator))
	_ = filepath.WalkDir(folder, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if p != folder && (hidden(d.Name()) || strings.Count(p, string(filepath.Separator))-depth > 1) {
				return filepath.SkipDir
			}
			return nil
		}
		n := d.Name()
		if hidden(n) || !formExts[strings.ToLower(filepath.Ext(n))] || len(out) >= 20 {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > 50<<20 {
			return nil
		}
		f := formFile{path: p, rel: filepath.Join(filepath.Base(folder), mustRel(folder, p)), mod: info.ModTime().Unix()}
		low := strings.ToLower(n)
		f.named = strings.Contains(low, "risk") || strings.Contains(low, "appendix")
		for _, k := range textutil.MatchKeys(n) {
			for _, t := range keys {
				if t != "" && strings.Contains(k, t) {
					f.named = true
				}
			}
		}
		out = append(out, f)
		return nil
	})
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].named != out[j].named {
			return out[i].named
		}
		return out[i].mod > out[j].mod
	})
	return out
}

func mustRel(base, p string) string {
	r, err := filepath.Rel(base, p)
	if err != nil {
		return filepath.Base(p)
	}
	return r
}

// dateFor finds the company's folder and the date in its form. ok is
// false when the company is skipped.
func (r *run) dateFor(row sheet.TableRow) (planned, bool, error) {
	folder, err := r.companyFolder(row.Name)
	if err != nil {
		return planned{}, false, err
	}
	if folder == "" {
		return r.skipDate(row, "no folder for this company")
	}
	files := formFiles(folder, r.cfg.TemplateName)
	if len(files) == 0 {
		return r.skipDate(row, fmt.Sprintf("no form in the folder %s", filepath.Base(folder)))
	}
	for _, f := range files {
		b, err := fsutil.ReadFileMax(f.path, maxForm)
		if err != nil {
			r.logf("read %s: %v", f.path, err)
			continue
		}
		if d, how, ok := reportdate.FromForm(b, f.path); ok {
			r.ui.Printf("  date %s (%s in %s)\n", d, how, ui.Safe(f.rel))
			r.logf("date %q: %s from %s (%s)", row.Name, d, f.path, how)
			return planned{row: row, date: d, from: f.rel}, true, nil
		}
	}
	return r.skipDate(row, "no date in "+files[0].rel)
}

// skipDate leaves a company out, says why and remembers it for the list
// of skipped companies.
func (r *run) skipDate(row sheet.TableRow, why string) (planned, bool, error) {
	r.ui.Printf("  skipped: %s\n", ui.Safe(why))
	r.logf("skip %q: %s", row.Name, why)
	r.skipped = append(r.skipped, row.Name+" — "+why)
	return planned{}, false, nil
}

// companyFolder returns the company's folder: an exact match, or a near
// match the user confirms. "" means none.
func (r *run) companyFolder(name string) (string, error) {
	exact, fuzzy := r.ix.Find(name)
	if len(exact) > 0 {
		return exact[0].Path, nil
	}
	if len(fuzzy) > 0 && r.opt.AssumeYes {
		c := fuzzy[0]
		r.ui.Printf("  folder %s (similar name, -y)\n", ui.Safe(c.Folder.Rel))
		r.logf("fuzzy %q vs %s (%.2f): accepted with -y", name, c.Folder.Rel, c.Score)
		return c.Folder.Path, nil
	}
	for _, c := range fuzzy {
		ok, err := r.ui.YesNo(fmt.Sprintf("  Is the folder %q this company's folder? (%s)", ui.Safe(c.Folder.Rel), c.Reason), false)
		if err != nil {
			return "", err
		}
		r.logf("fuzzy %q vs %s (%.2f): user said %v", name, c.Folder.Rel, c.Score, ok)
		if ok {
			return c.Folder.Path, nil
		}
	}
	return "", nil
}
