package app

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/formats"
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

// isForm reports whether a file name is a risk assessment form: it starts
// with the template's name and has a template extension.
func isForm(name, templateBase string) bool {
	return templateBase != "" && strings.HasPrefix(strings.ToLower(name), strings.ToLower(templateBase)) && formats.IsTemplate(path.Ext(name))
}

// Origins of a form in the working copy.
const (
	byNumbForm = "numb-form" // exactly as numb-form filled it
	editedForm = "edited"    // section A readable, but not as numb-form fills it
	otherFile  = ""          // not a form, or not comparable
)

// originOf tells whether the file at p is a form numb-form filled from the
// template (of the working copy or of the real folder) and nobody changed
// since. numb-form marks nothing in its
// documents: the form is compared with the template filled with the
// values it holds.
func originOf(p, rel string, rc reviewConfig) string {
	if len(rc.tpls) == 0 || !isForm(path.Base(rel), rc.templateBase) {
		return otherFile
	}
	b, err := os.ReadFile(p) // #nosec G304 -- a form in the working copy.
	if err != nil || len(b) > 50<<20 {
		return otherFile
	}
	origin := otherFile
	for _, t := range rc.tpls {
		ok, err := t.AsFilled(b)
		switch {
		case err != nil:
		case ok:
			return byNumbForm
		default:
			origin = editedForm
		}
	}
	return origin
}

func originLabel(o string) string {
	switch o {
	case byNumbForm:
		return "made by numb-form, unchanged"
	case editedForm:
		return "a form edited after filling, or made by hand"
	}
	return ""
}

// reviewDups explains the new folders that are not copied because the
// real folder already has the company, and asks for a look when such a
// folder holds more than numb-form's form.
func reviewDups(v *reqReview, q *request) {
	type group struct {
		dup   string
		other []string
	}
	groups := map[string]*group{}
	var order []string
	for _, c := range q.of(kDup) {
		g := groups[c.top]
		if g == nil {
			g = &group{dup: c.dup}
			groups[c.top] = g
			order = append(order, c.top)
		}
		if c.origin != byNumbForm {
			g.other = append(g.other, path.Base(c.rel))
		}
	}
	for _, top := range order {
		g := groups[top]
		if len(g.other) == 0 {
			v.add("ok", "%q is not copied: the real folder already has %q for that company", top, g.dup)
			continue
		}
		v.add("warn", "%q is not copied because the real folder has %q, but it holds more than numb-form's form (%s): copy those files by hand if they belong there",
			top, g.dup, strings.Join(g.other, ", "))
	}
}
