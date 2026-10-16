package app

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

// isOverviewName matches "3. Übersichten", "Uebersicht", "3_Ubersichten"...
func isOverviewName(name string) bool {
	for _, k := range textutil.MatchKeys(name) {
		if strings.Contains(k, "bersicht") {
			return true
		}
	}
	return false
}

// findOverview returns the single overview directory below root (searched
// up to three levels deep).
func findOverview(root string) (string, error) {
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
		if isOverviewName(d.Name()) {
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
		return "", errors.New(`no folder like "3. Übersichten" found`)
	case 1:
		return found[0], nil
	}
	return "", fmt.Errorf("several overview folders found: %s", strings.Join(found, ", "))
}

// findSpreadsheet returns the single .xlsx in the overview directory and
// any other files that are present (reported, never touched).
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
		if strings.EqualFold(filepath.Ext(n), ".xlsx") {
			xl = append(xl, filepath.Join(dir, n))
		} else {
			other = append(other, n)
		}
	}
	switch len(xl) {
	case 0:
		return "", other, fmt.Errorf("no .xlsx file in %s", dir)
	case 1:
		return xl[0], other, nil
	}
	return "", other, fmt.Errorf("%d .xlsx files in %s; expected exactly one", len(xl), dir)
}

var templateKeys = map[string]bool{"appendixariskassesment": true, "appendixariskassessment": true}

// findTemplate returns "Appendix A - Risk Assesment.docx" at the top of root.
func findTemplate(root string) (string, error) {
	ents, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	var found []string
	for _, e := range ents {
		n := e.Name()
		if e.IsDir() || strings.HasPrefix(n, "~$") || !strings.EqualFold(filepath.Ext(n), ".docx") {
			continue
		}
		if templateKeys[textutil.MatchKeys(strings.TrimSuffix(n, filepath.Ext(n)))[0]] {
			found = append(found, filepath.Join(root, n))
		}
	}
	switch len(found) {
	case 0:
		return "", fmt.Errorf(`template "Appendix A - Risk Assesment.docx" not found in %s`, root)
	case 1:
		return found[0], nil
	}
	return "", fmt.Errorf("several templates found: %s", strings.Join(found, ", "))
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
