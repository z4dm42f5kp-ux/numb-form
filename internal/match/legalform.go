// Package match extracts legal forms from company names and decides
// whether a company already has a folder.
package match

import (
	"regexp"
	"sort"
	"strings"
	"sync"
)

// legalForms lists conventional spellings. Matching is tolerant of
// missing dots and spaces; the text returned is the spelling found in
// the company name itself, so conventional casing is preserved.
var legalForms = []string{
	// German speaking.
	"GmbH & Co. KGaA", "GmbH & Co. KG", "UG (haftungsbeschränkt)", "gGmbH", "GmbH",
	"KGaA", "AG", "KG", "OHG", "GbR", "e.K.", "e.V.", "SE",
	// Spanish and Portuguese speaking.
	"S.A.P.I. de C.V.", "S.A.B. de C.V.", "S. de R.L. de C.V.", "S.A. de C.V.", "C.A. de C.V.",
	"S. en C.", "S.A.E.C.A.", "S.A.S.", "S.A.C.", "S.A.A.", "S.A.E.", "S.A.U.", "S.L.U.", "S.L.N.E.",
	"S.R.L.", "S.L.", "S.C.", "S.A.", "C.A.", "Ltda.", "Ltda", "Cía. Ltda.", "E.I.R.L.", "EIRELI", "ME", "EPP",
	// French, Italian, Benelux, Nordic, Central Europe.
	"S.à r.l.", "SARL", "EURL", "SNC", "S.p.A.", "S.a.s.", "B.V.", "N.V.", "BVBA", "A/S", "ApS", "AB", "Oy", "Oyj", "AS", "ASA",
	"Sp. z o.o.", "s.r.o.", "a.s.", "Kft.", "Zrt.", "d.o.o.",
	// English speaking.
	"Pty Ltd", "Pty. Ltd.", "Pvt. Ltd.", "Limited", "Ltd.", "Ltd", "PLC", "LLC", "L.L.C.", "LLP", "LP",
	"Inc.", "Inc", "Corporation", "Corp.", "Corp", "Co.",
}

var (
	formOnce sync.Once
	formREs  []*regexp.Regexp
)

func compileForms() {
	forms := append([]string(nil), legalForms...)
	// Longest canonical spelling first, so "S.A.S." wins over "S.A.".
	sort.SliceStable(forms, func(i, j int) bool { return len(squash(forms[i])) > len(squash(forms[j])) })
	for _, f := range forms {
		var p strings.Builder
		for _, r := range f {
			switch r {
			case '.':
				p.WriteString(`\.?`)
			case ' ':
				p.WriteString(`\s*`)
			case ',':
				p.WriteString(`,?`)
			default:
				p.WriteString(regexp.QuoteMeta(string(r)))
			}
		}
		flags := `(?i)`
		if isShortUpper(f) {
			// "AG", "SE", "AS", "ME"... must be written in capitals,
			// otherwise ordinary words such as "Me" would match.
			flags = ``
		}
		formREs = append(formREs, regexp.MustCompile(flags+`(?:^|[\s,])(`+p.String()+`)\s*$`))
	}
}

func isShortUpper(f string) bool {
	return len(f) <= 3 && strings.ToUpper(f) == f && !strings.ContainsAny(f, ". ")
}

func squash(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '.' || r == ' ' || r == ',' {
			return -1
		}
		return r
	}, s)
}

// LegalForm returns the legal-form suffix of name exactly as written in
// name (for example "GmbH" or "S.A. de C.V."), and the name without it.
// It returns an empty form when no known suffix is present.
func LegalForm(name string) (form, base string) {
	formOnce.Do(compileForms)
	trimmed := strings.TrimSpace(name)
	for _, re := range formREs {
		loc := re.FindStringSubmatchIndex(trimmed)
		if loc == nil {
			continue
		}
		form = strings.TrimSpace(trimmed[loc[2]:loc[3]])
		base = strings.TrimRight(strings.TrimSpace(trimmed[:loc[2]]), " ,")
		if base == "" {
			// The whole name is a legal form, which is not a suffix.
			return "", trimmed
		}
		return form, base
	}
	return "", trimmed
}
