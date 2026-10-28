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
	{"S.A.E.C.A.", []string{"PY"}}, {"S.A.I.C.A.", []string{"AR"}}, {"S.A.C.I.F.", []string{"AR"}}, {"S.A.I.C.", []string{"AR"}}, {"S.A.C.I.", []string{"AR", "CL", "PY"}},
	{"S.A.C.I.F.I.A.", []string{"AR"}}, {"S.A.C.I.F.I.", []string{"AR"}}, {"S.A.I.C.F.", []string{"AR"}}, {"S.A.C.I.A.", []string{"AR", "PY"}},
	{"S.A.I.C.F.I.A.", []string{"AR"}}, {"S.A.I.C.I.", []string{"AR"}}, {"S.A.C.", []string{"PE"}}, {"C.I.S.A.", []string{"PY"}}, {"S.A.C.I.A.G.", []string{"AR"}},
	{"S.R.L. de C.V.", []string{"MX"}}, {"S.A.S. BIC", []string{"CO"}}, {"S.A. BIC", []string{"CO"}}, {"S.A.S. E.S.P.", []string{"CO"}}, {"S.A. E.S.P.", []string{"CO"}},
	{"E.S.P.", []string{"CO"}}, {"C. Ltda.", []string{"EC"}}, {"Limitada", []string{"CO", "CL", "BR", "EC"}}, {"Compañía Limitada", []string{"EC", "CO"}},
	{"Sociedad por Acciones Simplificada", []string{"CO"}}, {"Sociedad Anónima Cerrada", []string{"PE"}}, {"Sociedad Comercial de Responsabilidad Limitada", []string{"PE"}},
	{"Empresa Individual de Responsabilidad Limitada", []string{"PE", "CL"}}, {"Sociedad en Comandita Simple", hispanic}, {"Sociedad Colectiva", hispanic},
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
	// French, Italian, Benelux, Nordic, Central and Eastern Europe.
	{"S.A.R.L.U.", fr}, {"SARLU", fr}, {"S.A.R.L.", fr}, {"G.I.E.", fr}, {"Société Anonyme", fr}, {"Société à Responsabilité Limitée", fr},
	{"Société par Actions Simplifiée", fr}, {"SUARL", []string{"SN", "CI", "CM"}}, {"S.à r.l.", []string{"LU"}}, {"EURL", []string{"FR"}}, {"SNC", fr},
	{"SA/NV", []string{"BE"}}, {"BV/SRL", []string{"BE"}}, {"B.V.", []string{"NL", "BE"}}, {"N.V.", []string{"NL", "BE", "CW"}}, {"BVBA", []string{"BE"}}, {"V.O.F.", []string{"NL"}},
	{"A/S", []string{"DK", "NO"}}, {"ApS", []string{"DK"}}, {"AB (publ)", []string{"SE"}}, {"AB", []string{"SE"}}, {"Oyj", []string{"FI"}}, {"Oy", []string{"FI"}}, {"ASA", []string{"NO"}}, {"AS", []string{"NO", "EE"}},
	{"Sp. z o.o.", []string{"PL"}}, {"S.A.S.", []string{"FR"}}, {"s.r.o.", []string{"CZ", "SK"}}, {"a.s.", []string{"CZ", "SK"}}, {"Kft.", []string{"HU"}}, {"Zrt.", []string{"HU"}}, {"Nyrt.", []string{"HU"}},
	{"d.o.o.", []string{"SI", "HR", "RS", "BA"}}, {"d.d.", []string{"SI", "HR"}}, {"OÜ", []string{"EE"}}, {"UAB", []string{"LT"}}, {"SIA", []string{"LV"}}, {"A.Ş.", []string{"TR"}},
	// English speaking, Africa and Asia.
	{"RF (Pty) Ltd", []string{"ZA"}}, {"(Pty) Ltd", []string{"ZA", "BW", "NA"}}, {"(Proprietary) Limited", []string{"ZA", "BW", "NA"}}, {"Proprietary Limited", []string{"ZA", "AU"}},
	{"Pty Ltd", []string{"AU", "ZA"}}, {"Pty. Ltd.", []string{"AU", "ZA"}}, {"Pvt. Ltd.", []string{"IN", "PK", "LK"}}, {"Private Limited", []string{"IN", "SG"}}, {"Pte. Ltd.", []string{"SG"}},
	{"Sdn. Bhd.", []string{"MY"}}, {"Bhd.", []string{"MY"}}, {"Co., Ltd.", []string{"CN", "JP", "KR", "TH", "VN", "TW"}}, {"K.K.", []string{"JP"}},
	{"Ltd/Gte", []string{"NG"}}, {"NPC", []string{"ZA"}}, {"Ltd. by Guarantee", []string{"NG", "GB"}},
	// Gulf and free zones.
	{"FZ-LLC", []string{"AE"}}, {"FZ LLC", []string{"AE"}}, {"LLC-FZ", []string{"AE"}}, {"L.L.C-FZ", []string{"AE"}}, {"FZCO", []string{"AE"}},
	{"FZE", []string{"AE"}}, {"FZC", []string{"AE"}}, {"DMCC", []string{"AE"}}, {"W.L.L.", []string{"BH", "QA", "KW"}}, {"P.J.S.C.", []string{"AE"}},
	{"P.S.C.", []string{"AE"}}, {"J.S.C.", []string{"AE", "SA", "RU"}}, {"S.A.O.G.", []string{"OM"}}, {"S.A.O.C.", []string{"OM"}}, {"S.P.C.", []string{"OM", "BH"}},
	{"K.S.C.C.", []string{"KW"}}, {"K.S.C.P.", []string{"KW"}}, {"B.S.C. (c)", []string{"BH"}}, {"Q.P.S.C.", []string{"QA"}}, {"Limited", english}, {"Ltd.", english}, {"PLC", english},
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
		switch c := letters[j]; {
		case strings.EqualFold(string(r), string(c)):
			out[i] = c
		case strings.EqualFold(fold(string(r)), fold(string(c))):
			// "CIA" for "Cía.": keep the letter as written, in the
			// conventional case.
			if unicode.IsUpper(c) {
				out[i] = unicode.ToUpper(r)
			} else {
				out[i] = unicode.ToLower(r)
			}
		default:
			return written
		}
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

// cutPlace removes a city or country written after the name: the last
// comma-separated part when it is a known place or looks like one (one
// to four capitalised words), or the last words when they are a known
// place. It reports false when nothing looks like a place.
func cutPlace(s string) (string, bool) {
	if i := strings.LastIndex(s, ","); i > 0 {
		seg := strings.Trim(s[i+1:], " .;")
		if seg == "" || knownPlace(seg) || placeLike(seg) {
			return strings.TrimRight(s[:i], " ,"), true
		}
		return "", false
	}
	w := strings.Fields(s)
	for k := min(4, len(w)-1); k >= 1; k-- {
		if knownPlace(strings.Join(w[len(w)-k:], " ")) {
			return strings.Join(w[:len(w)-k], " "), true
		}
	}
	return "", false
}

var particles = map[string]bool{"de": true, "del": true, "la": true, "las": true, "los": true, "el": true, "da": true, "do": true, "dos": true, "-": true, "–": true}

// placeLike accepts one to four capitalised words of letters, such as
// "Santo Domingo" or "Fernando De La Mora".
func placeLike(s string) bool {
	w := strings.Fields(s)
	if len(w) == 0 || len(w) > 5 || len(s) > 40 {
		return false
	}
	for _, x := range w {
		if particles[strings.ToLower(x)] {
			continue
		}
		r := []rune(x)
		if !unicode.IsUpper(r[0]) {
			return false
		}
		for _, c := range r {
			if !unicode.IsLetter(c) && c != '.' && c != '-' && c != '\'' {
				return false
			}
		}
	}
	return true
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
