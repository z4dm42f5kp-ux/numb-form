// Package match extracts legal forms from company names and decides
// whether a company already has a folder.
//
// In plain words: it reads the legal form (GmbH, S.A., Ltda) from a
// company name, and finds out whether a folder for that company already
// exists, even when the name is written slightly differently.
package match

import (
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// formDef is one legal form: its conventional spelling and the countries
// (ISO 3166-1 alpha-2) where it is used, most common first.
type formDef struct {
	canon     string
	countries []string
}

// Country groups used below.
var (
	dach     = []string{"DE", "AT", "CH"}
	latamSA  = []string{"ES", "MX", "AR", "CO", "PE", "CL", "EC", "VE", "UY", "PY", "BO", "GT", "CR", "PA", "DO", "HN", "SV", "NI", "BR", "PT", "CH", "FR", "BE", "LU", "MA", "SN", "CI", "CM"}
	hispanic = []string{"MX", "CO", "AR", "PE", "CL", "EC", "VE", "UY", "PY", "BO", "GT", "CR", "PA", "DO", "HN", "SV", "NI", "ES"}
	english  = []string{"GB", "IE", "NG", "KE", "GH", "ZA", "UG", "TZ", "ZM", "ZW", "MW", "BW", "NA", "MU", "IN", "SG", "HK", "MT", "CY", "US", "CA", "AU", "NZ"}
	us       = []string{"US"}
	fr       = []string{"FR", "BE", "LU", "CH", "MA", "TN", "DZ", "SN", "CI", "CM", "MG", "GA", "CG", "CD", "BF", "ML", "BJ", "TG", "NE"}
)

type compiled struct {
	def      formDef
	re       *regexp.Regexp // anchored at the end
	anywhere *regexp.Regexp // followed by more text
	key      string
}

var (
	formOnce sync.Once
	formREs  []compiled
)

func pattern(f string) string {
	var p strings.Builder
	for _, r := range f {
		switch {
		case r == '.':
			p.WriteString(`\.?\s*`)
		case r == ' ':
			p.WriteString(`\s*`)
		case r == ',':
			p.WriteString(`,?\s*`)
		case r == '(' || r == ')':
			p.WriteString(`[()]?\s*`)
		case r == '/':
			p.WriteString(`\s*/\s*`)
		case variants[fold(string(r))] != "":
			// "Anónima" also matches "Anonima" and "ANÓNIMA".
			p.WriteString("[" + variants[fold(string(r))] + "]")
		default:
			// After a letter followed by a dot, "\.?\s*" lets "S.A."
			// also match "S A", "S. A." and "SA".
			p.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	return p.String()
}

// variants lists, for each plain letter, the letter and its accented
// forms in both cases.
var variants = func() map[string]string {
	m := map[string]string{}
	for acc, plain := range accents {
		k := strings.ToLower(string(plain))
		if m[k] == "" {
			m[k] = k + strings.ToUpper(k)
		}
		m[k] += string(acc)
	}
	return m
}()

func compileForms() {
	defs := append([]formDef(nil), legalForms...)
	// Longest canonical spelling first, so "S.A.S." wins over "S.A.".
	sort.SliceStable(defs, func(i, j int) bool {
		return utf8.RuneCountInString(squash(defs[i].canon)) > utf8.RuneCountInString(squash(defs[j].canon))
	})
	byCanon := map[string]int{}
	for _, d := range defs {
		if i, ok := byCanon[d.canon]; ok {
			// Listed twice: merge the countries.
			for _, cc := range d.countries {
				if !contains(formREs[i].def.countries, cc) {
					formREs[i].def.countries = append(formREs[i].def.countries, cc)
				}
			}
			continue
		}
		byCanon[d.canon] = len(formREs)
		d.countries = append([]string(nil), d.countries...)
		p := pattern(d.canon)
		key := strings.ToUpper(squash(d.canon))
		flags := `(?i)`
		if ambiguous[key] {
			flags = ``
		}
		formREs = append(formREs, compiled{
			def:      d,
			key:      key,
			re:       regexp.MustCompile(flags + `(?:^|[\s,(])(` + p + `)[\s.,]*$`),
			anywhere: regexp.MustCompile(flags + `(?:^|[\s,(])(` + p + `)(?:[\s,()\-–/]|$)`),
		})
	}
}

// tail matches text after a legal form that does not belong to the
// company name: a branch, liquidation note, country or city in
// parentheses, or after a dash.
var tail = regexp.MustCompile(`(?i)^[\s,.]*(?:\(.*\)|[-–—/].*|(?:zweigniederlassung|niederlassung|filiale|sucursal|agencia|branch|succursale|in\s+liquidation|i\.\s*l\.|en\s+liquidaci[oó]n|en\s+reorganizaci[oó]n|em\s+liquida[cç][aã]o|in\s+liquidazione|zomac|me|epp)\b.*)?$`)

// Match is a legal form found in a name.
type Match struct {
	Form      string   // as written, with conventional capitals
	Base      string   // the name before the form
	Countries []string // where the form is used, most common first
}

// FindLegalForm finds the legal form of a company name: at the end of
// the name; followed only by a branch or location note such as
// "(Panamá)", "- Niederlassung Wien", ", Bogotá, Colombia" or " Paraguay";
// or written first ("Ets Foo", "Etablissements Foo").
func FindLegalForm(name string) (Match, bool) {
	s := strings.Trim(strings.Join(strings.Fields(name), " "), " ,;")
	if m, ok := findForm(s); ok {
		return m, true
	}
	rest := s
	for i := 0; i < 4; i++ {
		cut, ok := cutPlace(rest)
		if !ok {
			break
		}
		rest = cut
		if m, ok := findForm(rest); ok {
			return m, true
		}
	}
	return prefixForm(rest)
}

var rePrefix = regexp.MustCompile(`(?i)^(ets\.?|[ée]tablissements?)\s+(\S.*)$`)

// prefixForm finds a form written before the name, used in French-speaking
// Africa: "Ets Foo", "Etablissements Foo".
func prefixForm(s string) (Match, bool) {
	m := rePrefix.FindStringSubmatch(s)
	if m == nil {
		return Match{}, false
	}
	written := m[1]
	r := []rune(strings.ToLower(written)) // "ETS" gives "Ets"
	r[0] = unicode.ToUpper(r[0])
	written = string(r)
	return Match{Form: written, Base: strings.TrimSpace(m[2]), Countries: []string{"CM", "FR", "SN", "CI"}}, true
}

// findForm finds a form at the end of s or before a tail note.
func findForm(trimmed string) (Match, bool) {
	formOnce.Do(compileForms)
	// The longest form wins; among forms of that length, one spelled with
	// the same letters ("SARL" is S.A.R.L., not Sàrl).
	var first *Match
	firstLen := 0
	for _, c := range formREs {
		if first != nil && utf8.RuneCountInString(c.key) < firstLen {
			break
		}
		loc := c.re.FindStringSubmatchIndex(trimmed)
		if loc == nil {
			continue
		}
		m, ok := build(trimmed, loc[2], loc[3], c)
		if !ok {
			continue
		}
		if strings.EqualFold(squash(trimmed[loc[2]:loc[3]]), squash(c.def.canon)) {
			return m, true
		}
		if first == nil {
			first, firstLen = &m, utf8.RuneCountInString(c.key)
		}
	}
	if first != nil {
		return *first, true
	}
	// Not at the end: the last form followed only by a tail note.
	best, bestEnd := Match{}, -1
	for _, c := range formREs {
		for _, loc := range c.anywhere.FindAllStringSubmatchIndex(trimmed, -1) {
			if !tail.MatchString(trimmed[loc[3]:]) || loc[3] <= bestEnd {
				continue
			}
			if m, ok := build(trimmed, loc[2], loc[3], c); ok {
				best, bestEnd = m, loc[3]
			}
		}
	}
	return best, bestEnd >= 0
}

var reAndCo = regexp.MustCompile(`(?i)(?:&|\by|\be)$`)

func build(name string, start, end int, c compiled) (Match, bool) {
	written := strings.TrimRight(strings.TrimSpace(name[start:end]), " ,")
	base := strings.TrimRight(strings.TrimSpace(name[:start]), " ,(")
	// "Foo & Cía. Ltda.": "& Cía." belongs to the name, the form is "Ltda.".
	if reAndCo.MatchString(base) && strings.Contains(c.def.canon, " ") {
		if j := strings.LastIndexAny(written, " "); j > 0 && strings.HasPrefix(strings.ToLower(fold(c.def.canon)), "c") {
			for _, o := range formREs {
				if loc := o.re.FindStringSubmatchIndex(name[start+j:]); loc != nil {
					return build(name, start+j+loc[2], start+j+loc[3], o)
				}
			}
		}
	}
	if base == "" {
		// The whole name is a legal form, which is not a suffix.
		return Match{}, false
	}
	return Match{Form: recase(written, c.def.canon), Base: base, Countries: c.def.countries}, true
}

// LegalForm returns the legal form of name (for example "GmbH" or
// "S.A. de C.V.") and the name without it. It returns an empty form when
// no known form is present.
func LegalForm(name string) (form, base string) {
	if m, ok := FindLegalForm(name); ok {
		return m.Form, m.Base
	}
	return "", strings.TrimSpace(name)
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
