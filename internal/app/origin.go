package app

import (
	"path"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/formats"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
)

// maxForm bounds the forms numb-commit reads.
const maxForm = 50 << 20

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
	b, err := fsutil.ReadFileMax(p, maxForm)
	if err != nil {
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
