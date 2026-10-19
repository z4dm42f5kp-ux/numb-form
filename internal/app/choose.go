package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/lookup"
	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// askValues lets the user decide on every value (confirm = each).
func (r *run) askValues(name string, g gathered) (form.Values, error) {
	vals := form.Values{form.FieldName: name, form.FieldDate: r.date, form.FieldPerson: r.person}
	lf, _ := match.LegalForm(name)
	r.ui.Println()
	if len(g.res.FormHints) > 0 {
		r.ui.Printf("  Register lists the legal form as: %s\n", ui.Safe(strings.Join(g.res.FormHints, "; ")))
	}
	a, err := r.ui.AskDefault("  Legal form (Enter = keep, '-' = blank)", lf)
	if err != nil {
		return nil, err
	}
	if a == "-" {
		a = ""
	}
	vals[form.FieldLegalForm] = textutil.OneLine(a)
	if vals[form.FieldAddress], err = r.chooseAddress(g.addresses()); err != nil {
		return nil, err
	}
	if vals[form.FieldReps], err = r.chooseReps(g.people()); err != nil {
		return nil, err
	}
	return vals, nil
}

func (r *run) chooseAddress(cands []lookup.Address) (string, error) {
	r.ui.Println("\n  Headquarters / Address candidates (most authoritative first):")
	if len(cands) > 9 {
		cands = cands[:9]
	}
	for i, a := range cands {
		r.ui.Printf("    %d) %s\n       source: %s\n", i+1, ui.Safe(a.Text), ui.Safe(a.Source.String()))
	}
	def := "m"
	if len(cands) == 0 {
		r.ui.Println("    (none found)")
		if r.cfg.AddressFallback == "blank" {
			def = "b"
		}
	} else {
		def = "1"
	}
	for {
		a, err := r.ui.AskDefault("  Choose a number, 'm' to type it, 'b' to leave blank", def)
		if err != nil {
			return "", err
		}
		switch a {
		case "b":
			return "", nil
		case "m":
			v, err := r.ui.Ask("  Address on one line: ")
			if err != nil {
				return "", err
			}
			if v = textutil.OneLine(v); v == "" {
				continue
			}
			return r.withCaution(v, false, "address")
		}
		n, err := strconv.Atoi(a)
		if err != nil || n < 1 || n > len(cands) {
			continue
		}
		c := cands[n-1]
		v, err := r.ui.AskDefault("  Address (Enter = keep, or correct it)", c.Text)
		if err != nil {
			return "", err
		}
		return r.withCaution(textutil.OneLine(v), c.Source.Kind > lookup.KindRegulatory, "address")
	}
}

// withCaution appends the caution text as the configuration says; with
// caution = auto the user decides, def being the suggested answer.
func (r *run) withCaution(v string, def bool, what string) (string, error) {
	switch r.cfg.Caution {
	case "always":
		return addCaution(v, true), nil
	case "never":
		return v, nil
	}
	ok, err := r.ui.YesNo(fmt.Sprintf("  Add the caution text to the %s (not fully verified)?", what), def)
	if err != nil {
		return "", err
	}
	return addCaution(v, ok), nil
}

// FormatPeople renders "Jane Doe (General Manager, Legal Representative), John Roe (Director)".
func FormatPeople(ps []lookup.Person) string {
	var parts []string
	for _, p := range ps {
		roles := append(append([]string(nil), p.Roles...), p.Untranslated...)
		if len(roles) == 0 {
			parts = append(parts, p.Name)
			continue
		}
		parts = append(parts, p.Name+" ("+strings.Join(roles, ", ")+")")
	}
	return strings.Join(parts, ", ")
}

func (r *run) listPeople(people []lookup.Person) []string {
	var verified []string
	for i, p := range people {
		state := "unverified"
		if p.Verified() {
			state = "verified"
			verified = append(verified, strconv.Itoa(i+1))
		}
		var src []string
		for _, s := range p.Sources {
			src = append(src, s.String())
		}
		r.ui.Printf("    %d) %s  [%s]\n       source: %s\n", i+1, ui.Safe(FormatPeople([]lookup.Person{p})), state, ui.Safe(strings.Join(src, "; ")))
		if len(p.Untranslated) > 0 {
			r.ui.Printf("       note: role not translated automatically: %s\n", ui.Safe(strings.Join(p.Untranslated, ", ")))
		}
	}
	if len(people) == 0 {
		r.ui.Println("    (none found)")
	}
	return verified
}

func pick(a string, people []lookup.Person) []lookup.Person {
	if a == "a" {
		return people
	}
	var sel []lookup.Person
	for _, s := range strings.Split(a, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || n < 1 || n > len(people) {
			return nil
		}
		sel = append(sel, people[n-1])
	}
	return sel
}

func (r *run) chooseReps(people []lookup.Person) (string, error) {
	r.ui.Println("\n  Authorized representative candidates (most authoritative first):")
	if len(people) > 20 {
		people = people[:20]
	}
	verified := r.listPeople(people)
	def := "n"
	switch {
	case len(verified) > 0:
		def = strings.Join(verified, ",")
	case len(people) > 0:
		def = "m"
	}
	for {
		a, err := r.ui.AskDefault("  Numbers to include (e.g. 1,3), 'a' all, 'm' type them, 'n' none found", def)
		if err != nil {
			return "", err
		}
		switch a {
		case "n":
			if len(people) > 0 {
				ok, err := r.ui.YesNo("  Candidates were found. Use the 'No representatives were found online' text anyway?", false)
				if err != nil {
					return "", err
				}
				if !ok {
					continue
				}
			}
			if r.cfg.RepsFallback == "blank" {
				return "", nil
			}
			return NoRepsText, nil
		case "m":
			r.ui.Println("  Format: Full Name (Role 1, Role 2), Full Name (Role) — roles in English.")
			v, err := r.ui.Ask("  Representatives: ")
			if err != nil {
				return "", err
			}
			if v = textutil.OneLine(v); v == "" {
				continue
			}
			return r.withCaution(v, true, "representatives")
		}
		sel := pick(a, people)
		if len(sel) == 0 {
			continue
		}
		unverified := false
		for _, p := range sel {
			if !p.Verified() {
				unverified = true
			}
		}
		v, err := r.ui.AskDefault("  Line to write (Enter = keep, or correct it)", FormatPeople(sel))
		if err != nil {
			return "", err
		}
		return r.withCaution(textutil.OneLine(v), unverified, "representatives")
	}
}
