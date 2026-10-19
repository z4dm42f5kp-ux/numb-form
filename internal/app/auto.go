package app

import (
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/lookup"
	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
)

// cautionFor decides on the caution text without asking.
func (r *run) cautionFor(unverified bool) bool {
	switch r.cfg.Caution {
	case "always":
		return true
	case "never":
		return false
	}
	return unverified
}

func addCaution(v string, yes bool) string {
	if yes && v != "" && v != NoRepsText {
		return v + " " + CautionText
	}
	return v
}

// autoValues chooses every value without asking (confirm = none), using
// only the sources and trust level allowed by the configuration. Every
// choice is written to the log.
func (r *run) autoValues(name string, g gathered) form.Values {
	lf, _ := match.LegalForm(name)
	vals := form.Values{
		form.FieldName:      name,
		form.FieldLegalForm: lf,
		form.FieldDate:      r.date,
		form.FieldPerson:    r.person,
	}
	if as := g.addresses(); len(as) > 0 {
		a := as[0]
		vals[form.FieldAddress] = addCaution(a.Text, r.cautionFor(a.Source.Kind > lookup.KindRegulatory))
		r.logf("auto address for %q from %s", name, a.Source)
	} else {
		r.logf("auto address for %q: none found, left blank", name)
	}
	people := g.people()
	switch {
	case len(people) > 0:
		unverified := false
		for _, p := range people {
			if !p.Verified() {
				unverified = true
			}
		}
		vals[form.FieldReps] = addCaution(FormatPeople(people), r.cautionFor(unverified))
		r.logf("auto representatives for %q: %d person(s), unverified=%v", name, len(people), unverified)
	case r.cfg.RepsFallback == "no-reps-text":
		vals[form.FieldReps] = NoRepsText
		r.logf("auto representatives for %q: none found", name)
	}
	r.ui.Printf("  Filled automatically: %s\n", strings.Join(nonEmpty(vals), ", "))
	return vals
}

func nonEmpty(v form.Values) []string {
	var out []string
	for _, f := range form.AllFields {
		if v[f] != "" {
			out = append(out, f.Title())
		}
	}
	return out
}
