package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/reportdate"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// formExts are the forms numb-date can read.
var formExts = map[string]bool{".docx": true, ".docm": true, ".odt": true, ".fodt": true, ".rtf": true}

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
	if err != nil || folder == "" {
		return planned{}, false, err
	}
	files := formFiles(folder, r.cfg.TemplateName)
	if len(files) == 0 {
		r.ui.Printf("  no form found in %s — skipped\n", ui.Safe(filepath.Base(folder)))
		r.logf("skip %q: no readable form in %s", row.Name, folder)
		return planned{}, false, nil
	}
	for _, f := range files {
		b, err := os.ReadFile(f.path) // #nosec G304 -- a form in the working folder, read only.
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
	r.ui.Printf("  no date found in %s\n", ui.Safe(files[0].rel))
	for {
		a, err := r.ui.Ask("  Date (" + reportdate.Layout + ", Enter = skip): ")
		if err != nil {
			return planned{}, false, err
		}
		if a = strings.TrimSpace(a); a == "" {
			r.logf("skip %q: no date in its form, none typed", row.Name)
			return planned{}, false, nil
		}
		if d, ok := reportdate.Parse(a); ok {
			r.logf("date %q: %s typed by the user", row.Name, d)
			return planned{row: row, date: d, from: "typed"}, true, nil
		}
		r.ui.Println("  Please type a date such as 16-10-2026.")
	}
}

// companyFolder returns the company's folder: an exact match, or a near
// match the user confirms. "" means none.
func (r *run) companyFolder(name string) (string, error) {
	exact, fuzzy := r.ix.Find(name)
	if len(exact) > 0 {
		return exact[0].Path, nil
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
	r.ui.Println("  no folder for this company — skipped")
	r.logf("skip %q: no folder", name)
	return "", nil
}
