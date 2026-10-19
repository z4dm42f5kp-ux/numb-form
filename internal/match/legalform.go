// Package match extracts legal forms from company names and decides
// whether a company already has a folder.
package match

import (
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"
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

// legalForms lists conventional spellings. Matching tolerates missing or
// extra dots and spaces and any letter case; the text returned keeps the
// punctuation written in the name but uses conventional capitals.
var legalForms = []formDef{
	// German speaking.
	{"GmbH & Co. KGaA", []string{"DE"}}, {"GmbH & Co. KG", dach}, {"GmbH & Co. OHG", []string{"DE"}},
	{"SE & Co. KGaA", []string{"DE"}}, {"SE & Co. KG", []string{"DE"}}, {"AG & Co. KGaA", []string{"DE"}}, {"AG & Co. KG", []string{"DE"}},
	{"UG (haftungsbeschränkt) & Co. KG", []string{"DE"}}, {"UG (haftungsbeschränkt)", []string{"DE"}},
	{"Gesellschaft mit beschränkter Haftung", dach}, {"Aktiengesellschaft", dach}, {"Kommanditgesellschaft", dach},
	{"gGmbH", []string{"DE"}}, {"GmbH", dach}, {"GesmbH", []string{"AT"}}, {"Ges.m.b.H.", []string{"AT"}}, {"mbH", []string{"DE"}},
	{"KGaA", []string{"DE"}}, {"AG", dach}, {"KG", dach}, {"OHG", []string{"DE"}}, {"OG", []string{"AT"}}, {"GbR", []string{"DE"}},
	{"e.K.", []string{"DE"}}, {"e.U.", []string{"AT"}}, {"e.V.", []string{"DE"}}, {"eG", []string{"DE"}}, {"SE", []string{"DE", "AT", "FR", "NL", "LU"}},
	{"Sàrl", []string{"CH", "LU"}}, {"Genossenschaft", []string{"CH", "DE", "AT"}},
	// Spanish and Portuguese speaking.
	{"S.A.P.I. de C.V.", []string{"MX"}}, {"S.A.P.I.B. de C.V.", []string{"MX"}}, {"S.A.B. de C.V.", []string{"MX"}},
	{"S. de R.L. de C.V.", []string{"MX"}}, {"S. de R.L. de C.V.", []string{"MX"}}, {"S. en C. de C.V.", []string{"MX"}},
	{"S.C. de R.L. de C.V.", []string{"MX"}}, {"S.A. de C.V.", []string{"MX", "SV", "HN"}}, {"C.A. de C.V.", []string{"MX", "VE"}},
	{"S. de R.L.", []string{"MX", "HN", "PA", "GT"}}, {"S. en C.S.", []string{"CO"}}, {"S. en C.", []string{"CO", "MX"}},
	{"S.A.E.C.A.", []string{"PY"}}, {"S.A.I.C.A.", []string{"AR"}}, {"S.A.C.I.F.", []string{"AR"}}, {"S.A.I.C.", []string{"AR"}}, {"S.A.C.I.", []string{"AR", "CL"}},
	{"S.A.P.I.", []string{"MX"}}, {"S.A.S.U.", []string{"FR"}}, {"S.A.S.", []string{"CO", "FR", "AR"}}, {"S.A.C.", []string{"PE"}}, {"S.A.A.", []string{"PE"}},
	{"S.A.E.", []string{"EG", "ES"}}, {"S.A.U.", []string{"ES", "AR"}}, {"S.L.U.", []string{"ES"}}, {"S.L.N.E.", []string{"ES"}}, {"S.L.L.", []string{"ES"}}, {"S.L.P.", []string{"ES"}},
	{"S.Coop.", []string{"ES"}}, {"S.C.A.", []string{"ES", "AR", "BE", "LU"}}, {"S.C.S.", []string{"AR", "BE", "LU"}}, {"S.C.R.L.", []string{"PE", "BE"}},
	{"E.I.R.L.", []string{"PE", "CL"}}, {"S.R.L.", []string{"AR", "PE", "BO", "UY", "PY", "IT", "RO", "DO", "CR", "BE"}}, {"SpA", []string{"CL", "IT"}}, {"S.p.A.", []string{"IT", "CL"}},
	{"S.r.l.", []string{"IT"}}, {"S.L.", []string{"ES"}}, {"S.C.", []string{"MX", "PE"}}, {"A.C.", []string{"MX"}}, {"S/A", []string{"BR"}},
	{"Sociedad Anónima de Capital Variable", []string{"MX", "SV"}}, {"Sociedad Anónima", hispanic}, {"Sociedad Limitada", []string{"ES"}},
	{"Sociedad de Responsabilidad Limitada", []string{"MX", "AR", "PE"}}, {"Sociedade Anónima", []string{"PT"}}, {"Sociedade Anônima", []string{"BR"}},
	{"Sociedade Limitada", []string{"BR"}}, {"Compañía Anónima", []string{"VE", "EC"}}, {"Companhia", []string{"BR"}},
	{"S.A.", latamSA}, {"C.A.", []string{"VE", "EC"}}, {"Cía. Ltda.", []string{"EC", "CO"}}, {"Cia. Ltda.", []string{"BR", "EC"}}, {"Ltda.", []string{"BR", "CO", "CL", "EC"}},
	{"Unipessoal Lda.", []string{"PT"}}, {"Lda.", []string{"PT"}}, {"EIRELI", []string{"BR"}}, {"S.G.P.S., S.A.", []string{"PT"}}, {"SGPS", []string{"PT"}},
	{"ME", []string{"BR"}}, {"EPP", []string{"BR"}},
	// French, Italian, Benelux, Nordic, Central and Eastern Europe.
	{"S.A.R.L.U.", fr}, {"SARLU", fr}, {"S.A.R.L.", fr}, {"SUARL", []string{"SN", "CI", "CM"}}, {"S.à r.l.", []string{"LU"}}, {"EURL", []string{"FR"}}, {"SNC", fr},
	{"SA/NV", []string{"BE"}}, {"BV/SRL", []string{"BE"}}, {"B.V.", []string{"NL", "BE"}}, {"N.V.", []string{"NL", "BE", "CW"}}, {"BVBA", []string{"BE"}}, {"V.O.F.", []string{"NL"}},
	{"A/S", []string{"DK", "NO"}}, {"ApS", []string{"DK"}}, {"AB (publ)", []string{"SE"}}, {"AB", []string{"SE"}}, {"Oyj", []string{"FI"}}, {"Oy", []string{"FI"}}, {"ASA", []string{"NO"}}, {"AS", []string{"NO", "EE"}},
	{"Sp. z o.o.", []string{"PL"}}, {"S.A.S.", []string{"FR"}}, {"s.r.o.", []string{"CZ", "SK"}}, {"a.s.", []string{"CZ", "SK"}}, {"Kft.", []string{"HU"}}, {"Zrt.", []string{"HU"}}, {"Nyrt.", []string{"HU"}},
	{"d.o.o.", []string{"SI", "HR", "RS", "BA"}}, {"d.d.", []string{"SI", "HR"}}, {"OÜ", []string{"EE"}}, {"UAB", []string{"LT"}}, {"SIA", []string{"LV"}}, {"A.Ş.", []string{"TR"}},
	// English speaking, Africa and Asia.
	{"RF (Pty) Ltd", []string{"ZA"}}, {"(Pty) Ltd", []string{"ZA", "BW", "NA"}}, {"(Proprietary) Limited", []string{"ZA", "BW", "NA"}}, {"Proprietary Limited", []string{"ZA", "AU"}},
	{"Pty Ltd", []string{"AU", "ZA"}}, {"Pty. Ltd.", []string{"AU", "ZA"}}, {"Pvt. Ltd.", []string{"IN", "PK", "LK"}}, {"Private Limited", []string{"IN", "SG"}}, {"Pte. Ltd.", []string{"SG"}},
	{"Sdn. Bhd.", []string{"MY"}}, {"Bhd.", []string{"MY"}}, {"Co., Ltd.", []string{"CN", "JP", "KR", "TH", "VN", "TW"}}, {"K.K.", []string{"JP"}},
	{"Ltd/Gte", []string{"NG"}}, {"NPC", []string{"ZA"}}, {"Limited", english}, {"Ltd.", english}, {"PLC", english},
	{"PLLC", us}, {"L.L.C.", us}, {"LLC", us}, {"LLLP", us}, {"LLP", []string{"US", "GB"}}, {"L.P.", us}, {"LP", us}, {"P.C.", us}, {"N.A.", us},
	{"Incorporated", us}, {"Inc.", us}, {"Corporation", us}, {"Corp.", us}, {"Co.", us},
}

// ambiguous forms are short words that also occur in ordinary names; they
// are recognised only when written in capitals.
var ambiguous = map[string]bool{"ME": true, "EPP": true, "AS": true, "AB": true, "SE": true, "LP": true, "OG": true, "KG": true, "AG": true, "SA": true, "CA": true, "SC": true, "AC": true, "NA": true, "PC": true, "CO": true, "OY": true, "EG": true, "DD": true}

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
		default:
			// After a letter followed by a dot, "\.?\s*" lets "S.A."
			// also match "S A", "S. A." and "SA".
			p.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	return p.String()
}

func compileForms() {
	defs := append([]formDef(nil), legalForms...)
	// Longest canonical spelling first, so "S.A.S." wins over "S.A.".
	sort.SliceStable(defs, func(i, j int) bool { return len(squash(defs[i].canon)) > len(squash(defs[j].canon)) })
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

func squash(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '.' || r == ' ' || r == ',' || r == '(' || r == ')' || r == '/' {
			return -1
		}
		return r
	}, s)
}

// recase gives written the capitals of canon, keeping written's
// punctuation: "GMBH" becomes "GmbH", "s.a" becomes "S.A".
func recase(written, canon string) string {
	var letters []rune
	for _, r := range canon {
		if unicode.IsLetter(r) {
			letters = append(letters, r)
		}
	}
	out := []rune(written)
	j := 0
	for i, r := range out {
		if !unicode.IsLetter(r) {
			continue
		}
		if j >= len(letters) {
			return written
		}
		if !strings.EqualFold(string(r), string(letters[j])) {
			return written
		}
		out[i] = letters[j]
		j++
	}
	if j != len(letters) {
		return written
	}
	return string(out)
}

// tail matches text after a legal form that does not belong to the
// company name: a branch, liquidation note, country or city in
// parentheses, or after a dash.
var tail = regexp.MustCompile(`(?i)^[\s,.]*(?:\(.*\)|[-–—/].*|(?:zweigniederlassung|niederlassung|filiale|sucursal|branch|succursale|in\s+liquidation|i\.\s*l\.|en\s+liquidaci[oó]n|em\s+liquida[cç][aã]o|in\s+liquidazione)\b.*)?$`)

// Match is a legal form found in a name.
type Match struct {
	Form      string   // as written, with conventional capitals
	Base      string   // the name before the form
	Countries []string // where the form is used, most common first
}

// FindLegalForm finds the legal form of a company name: at the end of
// the name, or followed only by a branch or location note such as
// "(Panamá)" or "- Niederlassung Wien".
func FindLegalForm(name string) (Match, bool) {
	formOnce.Do(compileForms)
	trimmed := strings.TrimSpace(name)
	for _, c := range formREs {
		loc := c.re.FindStringSubmatchIndex(trimmed)
		if loc == nil {
			continue
		}
		if m, ok := build(trimmed, loc[2], loc[3], c); ok {
			return m, true
		}
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

func build(name string, start, end int, c compiled) (Match, bool) {
	written := strings.TrimRight(strings.TrimSpace(name[start:end]), " ,")
	base := strings.TrimRight(strings.TrimSpace(name[:start]), " ,(")
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
