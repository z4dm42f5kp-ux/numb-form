package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/formats"
	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// ignoredDirs are system folders that are never the working root.
var ignoredDirs = map[string]bool{
	"$recycle.bin": true, "system volume information": true, "numb-form-runs": true,
	"lost+found": true, ".git": true, "node_modules": true,
}

func hidden(name string) bool {
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "~$") || ignoredDirs[strings.ToLower(name)]
}

// subdirs lists the visible directories directly inside dir.
func subdirs(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() && !hidden(e.Name()) {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(out)
	return out, nil
}

// isOverviewName reports whether a folder name contains the configured
// pattern, ignoring case, accents and punctuation, with both umlaut
// spellings: "Übersicht" matches "3. Übersichten", "Uebersicht" and
// "3_Ubersichten".
func isOverviewName(name, pattern string) bool {
	if pattern == "" {
		pattern = "Übersicht"
	}
	for _, k := range textutil.MatchKeys(name) {
		for _, p := range textutil.MatchKeys(pattern) {
			if p != "" && strings.Contains(k, p) {
				return true
			}
		}
	}
	return false
}

// findOverview returns the single overview directory below root (searched
// up to three levels deep).
func findOverview(root, pattern string) (string, error) {
	var found []string
	rootDepth := strings.Count(filepath.Clean(root), string(filepath.Separator))
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable folders are skipped
		}
		if !d.IsDir() || p == root {
			return nil
		}
		if hidden(d.Name()) {
			return filepath.SkipDir
		}
		if strings.Count(p, string(filepath.Separator))-rootDepth > 3 {
			return filepath.SkipDir
		}
		if isOverviewName(d.Name(), pattern) {
			found = append(found, p)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	switch len(found) {
	case 0:
		return "", fmt.Errorf("no folder whose name contains %q found (overview_dir in numb-form.conf)", pattern)
	case 1:
		return found[0], nil
	}
	return "", fmt.Errorf("several overview folders found: %s", strings.Join(found, ", "))
}

// findSpreadsheet returns the single spreadsheet in the overview directory
// and any other files that are present (reported, never touched).
func findSpreadsheet(dir string) (string, []string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return "", nil, err
	}
	var xl, other []string
	for _, e := range ents {
		n := e.Name()
		if e.IsDir() || strings.HasPrefix(n, "~$") || strings.HasPrefix(n, ".") {
			continue
		}
		if formats.IsSheet(filepath.Ext(n)) {
			xl = append(xl, filepath.Join(dir, n))
		} else {
			other = append(other, n)
		}
	}
	switch len(xl) {
	case 0:
		return "", other, fmt.Errorf("no spreadsheet (%s) in %s", strings.Join(formats.SheetExts, " "), dir)
	case 1:
		return xl[0], other, nil
	}
	return "", other, fmt.Errorf("%d spreadsheets in %s; expected exactly one", len(xl), dir)
}

// findTemplate returns the template at the top of root: a file named
// like the configured template name (case, punctuation and the
// Assesment/Assessment spelling ignored) with a supported extension. When
// several formats exist, the most direct one wins (.docx first).
func findTemplate(root, name string) (string, error) {
	want := map[string]bool{}
	for _, k := range textutil.MatchKeys(name) {
		want[k] = true
		want[strings.ReplaceAll(k, "assesment", "assessment")] = true
		want[strings.ReplaceAll(k, "assessment", "assesment")] = true
	}
	ents, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	var found []string
	for _, e := range ents {
		n := e.Name()
		ext := filepath.Ext(n)
		if e.IsDir() || strings.HasPrefix(n, "~$") || !formats.IsTemplate(ext) {
			continue
		}
		if want[textutil.MatchKeys(strings.TrimSuffix(n, ext))[0]] {
			found = append(found, filepath.Join(root, n))
		}
	}
	if len(found) == 0 {
		return "", fmt.Errorf("template %q (%s) not found in %s", name, strings.Join(formats.TemplateExts, " "), root)
	}
	best := found[0]
	for _, ext := range formats.TemplateExts {
		var same []string
		for _, f := range found {
			if strings.EqualFold(filepath.Ext(f), ext) {
				same = append(same, f)
			}
		}
		if len(same) > 1 {
			return "", fmt.Errorf("several templates found: %s", strings.Join(same, ", "))
		}
		if len(same) == 1 {
			best = same[0]
			break
		}
	}
	return best, nil
}

// indexFolders registers every directory below root except the overview
// directory, without following symbolic links.
func indexFolders(root, overview string) (*match.Index, error) {
	ix := &match.Index{}
	rootDepth := strings.Count(filepath.Clean(root), string(filepath.Separator))
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() || p == root {
			return nil
		}
		if p == overview || strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		if strings.Count(p, string(filepath.Separator))-rootDepth > 10 {
			return filepath.SkipDir
		}
		rel, _ := filepath.Rel(root, p)
		ix.Add(p, rel, d.Name())
		return nil
	})
	return ix, err
}
