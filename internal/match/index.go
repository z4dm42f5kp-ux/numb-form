package match

import (
	"sort"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// Folder is a directory that may already belong to a company.
type Folder struct {
	Path string // absolute path
	Rel  string // path relative to the working root
	Name string // base name

	keys     []string
	baseKeys []string
	formKey  string
}

// Candidate is a folder that resembles, but does not exactly equal, a
// company name. The user must confirm it.
type Candidate struct {
	Folder Folder
	Score  float64
	Reason string
}

// Index holds every directory found below the working root.
type Index struct {
	folders []Folder
}

// Add registers a directory.
func (ix *Index) Add(path, rel, name string) {
	form, base := LegalForm(name)
	ix.folders = append(ix.folders, Folder{
		Path: path, Rel: rel, Name: name,
		keys:     textutil.MatchKeys(name),
		baseKeys: textutil.MatchKeys(base),
		formKey:  formKey(form),
	})
}

// Len returns the number of indexed folders.
func (ix *Index) Len() int { return len(ix.folders) }

func formKey(form string) string {
	if form == "" {
		return ""
	}
	return textutil.MatchKeys(form)[0]
}

func anyEqual(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if x != "" && x == y {
				return true
			}
		}
	}
	return false
}

// Find returns folders whose name equals the company name after the three
// variations (case, accents/umlauts, punctuation and spacing). When there
// is no exact match it returns fuzzy candidates for the user to confirm.
// Folders whose legal form differs from the company's legal form are never
// candidates: "Foo GmbH" and "Foo S.A." are different entities.
func (ix *Index) Find(company string) (exact []Folder, fuzzy []Candidate) {
	ckeys := textutil.MatchKeys(company)
	cform, cbase := LegalForm(company)
	cformKey := formKey(cform)
	cbaseKeys := textutil.MatchKeys(cbase)

	for _, f := range ix.folders {
		if anyEqual(ckeys, f.keys) {
			exact = append(exact, f)
		}
	}
	if len(exact) > 0 {
		return exact, nil
	}
	for _, f := range ix.folders {
		if cformKey != "" && f.formKey != "" && cformKey != f.formKey {
			continue
		}
		best, reason := 0.0, ""
		if anyEqual(cbaseKeys, f.baseKeys) {
			best, reason = 0.95, "same name, legal form missing on one side"
		}
		for _, a := range ckeys {
			for _, b := range f.keys {
				if len([]rune(a)) < 4 || len([]rune(b)) < 4 {
					continue
				}
				if s := textutil.Similarity(a, b); s >= 0.85 && s > best {
					best, reason = s, "spelling is very close"
				}
				if len(b) >= 5 && len(a) >= 5 && (strings.Contains(a, b) || strings.Contains(b, a)) && best < 0.9 {
					best, reason = 0.9, "one name contains the other"
				}
			}
		}
		if best > 0 {
			fuzzy = append(fuzzy, Candidate{Folder: f, Score: best, Reason: reason})
		}
	}
	sort.SliceStable(fuzzy, func(i, j int) bool { return fuzzy[i].Score > fuzzy[j].Score })
	if len(fuzzy) > 5 {
		fuzzy = fuzzy[:5]
	}
	return nil, fuzzy
}
