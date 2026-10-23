package app

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// realFolder is a folder of the real working folder.
type realFolder struct{ rel, name string }

// listFolders returns the folders below root, ten levels deep at most,
// without the overview folder and hidden folders (as numb-form looks).
func listFolders(root, overview string) []realFolder {
	var out []realFolder
	rootDepth := strings.Count(filepath.Clean(root), string(filepath.Separator))
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() || p == root {
			return nil
		}
		if p == overview || strings.HasPrefix(d.Name(), ".") || strings.Count(p, string(filepath.Separator))-rootDepth > 10 {
			return filepath.SkipDir
		}
		r, _ := filepath.Rel(root, p)
		out = append(out, realFolder{rel: filepath.ToSlash(r), name: d.Name()})
		return nil
	})
	return out
}

// reIDPrefix splits a folder name such as "D10629 Foo S.A." or
// "D10629 - Foo S.A." into the customer id and the company name.
var reIDPrefix = regexp.MustCompile(`^\s*([A-Za-z]{0,3}[0-9]{3,}[A-Za-z]?)\s*[-_.]?\s+(\S.*)$`)

func keysMeet(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if x != "" && x == y {
				return true
			}
		}
	}
	return false
}

// companyFolder finds the folder that the real working folder already has
// for a company: one named after its customer id ("D10629 Foo S.A." for
// D10629), one whose name after a leading customer id is the company's,
// or one with the company's name in other spelling (case, accents,
// punctuation). It returns why it matched.
func companyFolder(folders []realFolder, company, id string) (realFolder, string, bool) {
	keys := textutil.MatchKeys(company)
	id = strings.TrimSpace(id)
	for _, f := range folders {
		if m := reIDPrefix.FindStringSubmatch(f.name); m != nil && id != "" && strings.EqualFold(m[1], id) {
			return f, "it starts with the customer id " + id, true
		}
	}
	for _, f := range folders {
		if m := reIDPrefix.FindStringSubmatch(f.name); m != nil && keysMeet(keys, textutil.MatchKeys(m[2])) {
			return f, "the same name after the customer id " + m[1], true
		}
	}
	for _, f := range folders {
		if keysMeet(keys, textutil.MatchKeys(f.name)) {
			return f, "the same name", true
		}
	}
	return realFolder{}, "", false
}
