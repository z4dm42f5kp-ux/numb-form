package match

import (
	"regexp"
	"strings"
	"unicode"
)

// Name is a company name from the spreadsheet split into its parts:
// "Acme SAS, Bogota, Colombia" has the core "Acme", the legal form
// "SAS" and the places "Bogota" and "Colombia".
type Name struct {
	Raw       string   // the spreadsheet cell
	Core      string   // without legal form and places
	Form      string   // legal form as written, conventional capitals
	Prefix    bool     // the form comes first ("Ets Zorbu")
	Places    []string // cities and countries written after the name
	Countries []string // ISO codes: from the places first, then the form
	// Confident is set when the country comes from a place or a country
	// word in the name, or from a legal form used in at most two
	// countries; "S.A." alone says little.
	Confident bool
}

// prefixForms are legal forms written before the name.
var rePrefixForm = regexp.MustCompile(`(?i)^(ets\.?|[ée]tablissements?|ste\.?|sté)\s+`)

// Clean returns the name without places: "Acme SAS".
func (n Name) Clean() string {
	switch {
	case n.Form == "":
		return n.Core
	case n.Prefix:
		return n.Form + " " + n.Core
	}
	return n.Core + " " + n.Form
}

// ParseName splits a spreadsheet company name.
func ParseName(raw string) Name {
	n := Name{Raw: strings.TrimSpace(raw)}
	s := strings.Join(strings.Fields(n.Raw), " ")
	s = strings.Trim(s, " ,;")
	if m := rePrefixForm.FindStringSubmatch(s); m != nil {
		n.Form, n.Prefix = "Ets", true
		s = strings.TrimSpace(s[len(m[0]):])
	}
	// Places: trailing comma-separated parts that are a known city or
	// country, or short capitalised words after a legal form.
	parts := strings.Split(s, ",")
	for len(parts) > 1 {
		last := strings.Trim(strings.TrimSpace(parts[len(parts)-1]), ". ")
		rest := strings.Join(parts[:len(parts)-1], ",")
		cc, known := Place(last)
		_, formBefore := FindLegalForm(strings.Trim(rest, " ,"))
		if last == "" {
			parts = parts[:len(parts)-1]
			continue
		}
		if !known && !(formBefore && placeLike(last) && !isForm(last)) {
			break
		}
		n.Places = append([]string{last}, n.Places...)
		if cc != "" {
			n.Countries = appendUnique(n.Countries, cc)
			n.Confident = true
		}
		parts = parts[:len(parts)-1]
	}
	s = strings.Trim(strings.Join(parts, ","), " ,")
	if !n.Prefix {
		if m, ok := FindLegalForm(s); ok {
			n.Form, n.Core = m.Form, m.Base
			for _, c := range m.Countries {
				n.Countries = appendUnique(n.Countries, c)
			}
			if len(m.Countries) <= 2 {
				n.Confident = true
			}
		} else {
			n.Core = s
		}
	} else {
		n.Core = s
		n.Countries = appendUnique(n.Countries, "CM")
		n.Countries = appendUnique(n.Countries, "FR")
	}
	n.Core = strings.Trim(n.Core, " ,;-")
	// A country written inside the name ("Zorbu Chile") is a hint too.
	for _, w := range strings.Fields(n.Core) {
		if cc, ok := countryWord(w); ok {
			n.Countries = append([]string{cc}, removeItem(n.Countries, cc)...)
			n.Confident = true
		}
	}
	return n
}

func isForm(s string) bool {
	m, ok := FindLegalForm("X " + s)
	return ok && m.Base == "X"
}

// placeLike accepts one to four capitalised words of letters.
func placeLike(s string) bool {
	w := strings.Fields(s)
	if len(w) == 0 || len(w) > 4 {
		return false
	}
	for _, x := range w {
		r := []rune(x)
		if !unicode.IsUpper(r[0]) && !particle[strings.ToLower(x)] {
			return false
		}
		for _, c := range r {
			if !unicode.IsLetter(c) && c != '-' && c != '.' {
				return false
			}
		}
	}
	return true
}

var particle = map[string]bool{"de": true, "del": true, "la": true, "las": true, "los": true, "el": true, "da": true, "do": true, "dos": true}

func appendUnique(list []string, s string) []string {
	for _, x := range list {
		if x == s {
			return list
		}
	}
	return append(list, s)
}

func removeItem(list []string, s string) []string {
	var out []string
	for _, x := range list {
		if x != s {
			out = append(out, x)
		}
	}
	return out
}

// Queries returns web searches for the company, in the order a person
// would try them: the literal cell, the clean name in quotes, the clean
// name with its places, the core name with its country, and spellings
// without accents or with joined initials ("A B C" becomes "ABC").
func (n Name) Queries() []string {
	var out []string
	seen := map[string]bool{}
	add := func(q string) {
		q = strings.Join(strings.Fields(q), " ")
		k := strings.ToLower(q)
		if q == "" || seen[k] {
			return
		}
		seen[k] = true
		out = append(out, q)
	}
	clean := n.Clean()
	add(strings.Trim(n.Raw, " ,;"))
	add(`"` + clean + `"`)
	if len(n.Places) > 0 {
		add(clean + " " + strings.Join(n.Places, " "))
	}
	if n.Confident && len(n.Countries) > 0 && n.Core != "" {
		add(`"` + n.Core + `" ` + CountryName(n.Countries[0]))
	}
	if a := ASCII(clean); a != clean {
		add(`"` + a + `"`)
	}
	if j := JoinInitials(clean); j != clean {
		add(`"` + j + `"`)
	}
	return out
}

var reInitials = regexp.MustCompile(`\b(?:\p{Lu}\.?\s){2,}\p{Lu}\.?(?:\s|$)`)

// JoinInitials writes spaced initials together: "A B C Trading" gives
// "ABC Trading".
func JoinInitials(s string) string {
	return strings.TrimSpace(reInitials.ReplaceAllStringFunc(s, func(m string) string {
		return strings.NewReplacer(" ", "", ".", "").Replace(m) + " "
	}))
}

// ASCII removes accents: "Químicas" gives "Quimicas".
func ASCII(s string) string {
	var b strings.Builder
	for _, r := range s {
		if v, ok := accents[r]; ok {
			b.WriteRune(v)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

var accents = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e', 'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
	'ó': 'o', 'ò': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u', 'ñ': 'n', 'ç': 'c',
	'Á': 'A', 'À': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'É': 'E', 'È': 'E', 'Ê': 'E', 'Í': 'I', 'Ó': 'O', 'Ô': 'O', 'Õ': 'O',
	'Ö': 'O', 'Ú': 'U', 'Ü': 'U', 'Ñ': 'N', 'Ç': 'C',
}
