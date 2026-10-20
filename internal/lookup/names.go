package lookup

import (
	"strings"
	"unicode"

	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// normWords prepares text for "does this page name the company" checks:
// lower case, accents folded, dots dropped, other punctuation turned into
// spaces, and runs of single letters joined ("S. A. de C. V." becomes
// "sa de cv", "S.A." becomes "sa").
func normWords(s string) string {
	var b strings.Builder
	for _, r := range textutil.Fold(s) {
		switch {
		case r == '.' || r == '\'' || r == '’':
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		default:
			b.WriteByte(' ')
		}
	}
	var words []string
	for _, w := range strings.Fields(b.String()) {
		switch w {
		case "und", "and", "y", "e", "et", "u":
			continue // "&" is dropped as punctuation, so its words are too
		}
		words = append(words, w)
	}
	var out []string
	run := ""
	for _, w := range words {
		if len([]rune(w)) == 1 && unicode.IsLetter([]rune(w)[0]) {
			run += w
			continue
		}
		if run != "" {
			out = append(out, run)
			run = ""
		}
		out = append(out, w)
	}
	if run != "" {
		out = append(out, run)
	}
	return " " + strings.Join(out, " ") + " "
}

// longForms lists written-out legal forms that are equivalent to the
// short ones, in normWords form.
var longForms = map[string][]string{
	" gmbh ":     {" gesellschaft mit beschrankter haftung ", " gesellschaft mbh ", " ges mbh ", " gesmbh "},
	" ag ":       {" aktiengesellschaft "},
	" kg ":       {" kommanditgesellschaft "},
	" sa ":       {" sociedad anonima ", " sociedade anonima ", " societe anonyme ", " s a "},
	" sl ":       {" sociedad limitada ", " sociedad de responsabilidad limitada "},
	" ltda ":     {" limitada ", " ltd "},
	" lda ":      {" limitada "},
	" ltd ":      {" limited ", " ltda "},
	" inc ":      {" incorporated ", " corporation ", " corp "},
	" corp ":     {" corporation ", " incorporated ", " inc "},
	" llc ":      {" limited liability company ", " l l c "},
	" ca ":       {" compania anonima "},
	" sas ":      {" sociedad por acciones simplificada ", " societe par actions simplifiee "},
	" sa de cv ": {" sociedad anonima de capital variable ", " s a de c v "},
	" spa ":      {" sociedad por acciones ", " societa per azioni "},
	" srl ":      {" sociedad de responsabilidad limitada ", " societa a responsabilita limitata "},
	" pty ltd ":  {" proprietary limited ", " pty limited "},
}

// nameVariants returns the normalised ways a page may write the company.
func nameVariants(q Query) []string {
	full := normWords(q.Name)
	out := []string{full}
	if q.Form == "" {
		return out
	}
	base := normWords(q.Base)
	form := normWords(q.Form)
	out = append(out, base+strings.TrimLeft(form, " "))
	for _, l := range longForms[form] {
		out = append(out, base+strings.TrimLeft(l, " "))
	}
	return out
}

// Mentions reports whether normalised text names the company: the full
// name, or its base name followed by the same legal form written out.
// Names without a legal form need at least two words or eight letters,
// so that a common word is not taken for the company.
func Mentions(norm string, q Query) bool {
	if q.Form == "" {
		base := normWords(q.Base)
		if len(strings.Fields(base)) < 2 && len(strings.TrimSpace(base)) < 8 {
			return false
		}
	}
	for _, v := range nameVariants(q) {
		if strings.TrimSpace(v) != "" && strings.Contains(norm, v) {
			return true
		}
	}
	return false
}

// slugs returns plausible domain labels for a company name: "Foo Bar
// GmbH" gives "foobar" and "foo-bar", and "foo" for a long first word.
func slugs(q Query) []string {
	words := strings.Fields(normWords(q.Base))
	var keep []string
	for _, w := range words {
		switch w {
		case "the", "der", "die", "das", "el", "la", "los", "las", "le", "les", "o", "a", "de", "del", "do", "da", "dos", "und", "and", "y", "e", "et", "co", "cia":
			continue
		}
		if !isASCIIWord(w) {
			continue
		}
		keep = append(keep, w)
	}
	if len(keep) == 0 {
		return nil
	}
	var out []string
	add := func(s string) {
		if len(s) < 3 || len(s) > 40 {
			return
		}
		for _, x := range out {
			if x == s {
				return
			}
		}
		out = append(out, s)
	}
	add(strings.Join(keep, ""))
	if len(keep) > 1 {
		add(strings.Join(keep, "-"))
		if len(keep[0]) >= 5 {
			add(keep[0])
		}
	}
	return out
}

func isASCIIWord(w string) bool {
	for _, r := range w {
		if r > unicode.MaxASCII || !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

// countryTLDs maps a country to the domain endings its companies use.
var countryTLDs = map[string][]string{
	"DE": {".de"}, "AT": {".at"}, "CH": {".ch"}, "ES": {".es"}, "PT": {".pt"}, "BR": {".com.br"}, "MX": {".com.mx", ".mx"},
	"AR": {".com.ar"}, "CO": {".com.co", ".co"}, "CL": {".cl"}, "PE": {".com.pe", ".pe"}, "EC": {".com.ec", ".ec"},
	"VE": {".com.ve"}, "UY": {".com.uy"}, "PY": {".com.py"}, "BO": {".com.bo"}, "GT": {".com.gt"}, "CR": {".co.cr"},
	"PA": {".com.pa"}, "DO": {".com.do"}, "US": {".com"}, "GB": {".co.uk"}, "IE": {".ie"}, "ZA": {".co.za"},
	"NG": {".com.ng", ".ng"}, "KE": {".co.ke"}, "GH": {".com.gh"}, "EG": {".com.eg"}, "MA": {".ma"}, "FR": {".fr"},
	"BE": {".be"}, "NL": {".nl"}, "LU": {".lu"}, "IT": {".it"}, "PL": {".pl"}, "CZ": {".cz"}, "SK": {".sk"},
	"HU": {".hu"}, "SE": {".se"}, "DK": {".dk"}, "NO": {".no"}, "FI": {".fi"}, "IN": {".in", ".co.in"},
	"SG": {".com.sg"}, "AU": {".com.au"}, "CA": {".ca"}, "NZ": {".co.nz"},
}

// guessDomains builds candidate home pages from the name and the
// countries its legal form points to.
func guessDomains(q Query) []string {
	tlds := []string{}
	add := func(t string) {
		for _, x := range tlds {
			if x == t {
				return
			}
		}
		tlds = append(tlds, t)
	}
	for i, c := range q.Countries {
		if i >= 3 {
			break
		}
		for _, t := range countryTLDs[c] {
			add(t)
		}
	}
	add(".com")
	var out []string
	for _, s := range slugs(q) {
		for _, t := range tlds {
			out = append(out, "https://www."+s+t+"/")
		}
	}
	if len(out) > 8 {
		out = out[:8]
	}
	return out
}

// queryFor builds the lookup query from the literal spreadsheet name.
func queryFor(name string) Query {
	m, ok := match.FindLegalForm(name)
	if !ok {
		return Query{Name: name, Base: strings.TrimSpace(name)}
	}
	return Query{Name: name, Form: m.Form, Base: m.Base, Countries: m.Countries}
}
