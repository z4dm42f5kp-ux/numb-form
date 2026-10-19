package lookup

import (
	"regexp"
	"strings"
	"unicode"
)

// ID is an official identifier found on a page, such as a VAT number.
type ID struct {
	Kind  string // "vat", "che", "cnpj"
	Value string // normalised: "ATU13585627", "CHE123456789", "19131243000197"
}

// RegisterNote is a register entry printed on a legal notice, such as
// "Amtsgericht München HRB 12345". It is shown to the user, who may type
// it into the Commercial Registry Number field.
type RegisterNote struct {
	Text   string
	Source Source
}

var (
	// EU VAT numbers with their country prefix. Spaces, dots and dashes
	// inside are allowed. Greece uses EL.
	reVAT  = regexp.MustCompile(`\b(AT\s?U(?:[\s.-]?\d){8}|BE\s?[01](?:[\s.-]?\d){9}|BG(?:[\s.-]?\d){9,10}|CY(?:[\s.-]?\d){8}[\s.-]?[A-Z]|CZ(?:[\s.-]?\d){8,10}|DE(?:[\s.-]?\d){9}|DK(?:[\s.-]?\d){8}|EE(?:[\s.-]?\d){9}|EL(?:[\s.-]?\d){9}|ES[\s.-]?[A-Z0-9](?:[\s.-]?\d){7}[\s.-]?[A-Z0-9]|FI(?:[\s.-]?\d){8}|FR[\s.-]?[A-HJ-NP-Z0-9]{2}(?:[\s.-]?\d){9}|HR(?:[\s.-]?\d){11}|HU(?:[\s.-]?\d){8}|IE\d[A-Z0-9+*]\d{5}[A-Z]{1,2}|IT(?:[\s.-]?\d){11}|LT(?:\d{9}|\d{12})|LU(?:[\s.-]?\d){8}|LV(?:[\s.-]?\d){11}|MT(?:[\s.-]?\d){8}|NL(?:[\s.-]?\d){9}[\s.-]?B[\s.-]?\d{2}|PL(?:[\s.-]?\d){10}|PT(?:[\s.-]?\d){9}|RO(?:[\s.-]?\d){2,10}|SE(?:[\s.-]?\d){12}|SI(?:[\s.-]?\d){8}|SK(?:[\s.-]?\d){10})\b`)
	reCHE  = regexp.MustCompile(`\bCHE[\s-]?(\d{3})[.\s]?(\d{3})[.\s]?(\d{3})\b`)
	reCNPJ = regexp.MustCompile(`\b(\d{2})\.?(\d{3})\.?(\d{3})/?(\d{4})-?(\d{2})\b`)

	// Register entries as printed on legal notices.
	reRegister = regexp.MustCompile(`(?i)(?:\b(?:amtsgericht|registergericht|handelsregister|firmenbuch(?:gericht)?|handelsregisteramt|registro\s+mercantil(?:\s+de\s+[\p{L} ]+?)?|registro\s+p[uú]blico\s+de\s+comercio|rcs\s+[\p{L}-]+|company\s+(?:registration\s+)?(?:no|number)|registered\s+in\s+england(?:\s+and\s+wales)?)\b[^\n]{0,60}?)?\b(?:HR[AB]\s?\d{1,6}(?:\s?[A-Z]{1,2})?|FN\s?\d{1,6}\s?[a-z]|CHE[\s-]?\d{3}[.\s]?\d{3}[.\s]?\d{3}|(?:tomo|folio|hoja|inscripci[oó]n)\s*[\w-]+(?:[,;]\s*(?:tomo|folio|hoja|libro|secci[oó]n|inscripci[oó]n)\s*[\w-]+){1,5}|NIRE\s*\d{8,11})`)
)

func digitsOnly(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, s)
}

func compact(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '+' || r == '*' {
			return unicode.ToUpper(r)
		}
		return -1
	}, s)
}

// validCNPJ checks the two check digits of a Brazilian CNPJ.
func validCNPJ(d string) bool {
	if len(d) != 14 || strings.Count(d, d[:1]) == 14 {
		return false
	}
	calc := func(n int) byte {
		w := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}[13-n:]
		sum := 0
		for i := 0; i < n; i++ {
			sum += int(d[i]-'0') * w[i]
		}
		r := sum % 11
		if r < 2 {
			return '0'
		}
		return byte('0' + 11 - r)
	}
	return d[12] == calc(12) && d[13] == calc(13)
}

// validCHE checks the check digit of a Swiss UID (CHE-123.456.789).
func validCHE(d string) bool {
	if len(d) != 9 {
		return false
	}
	w := []int{5, 4, 3, 2, 7, 6, 5, 4}
	sum := 0
	for i := 0; i < 8; i++ {
		sum += int(d[i]-'0') * w[i]
	}
	c := 11 - sum%11
	if c == 11 {
		c = 0
	}
	return c != 10 && int(d[8]-'0') == c
}

// ExtractIDs finds VAT numbers, Swiss UIDs and CNPJs in text.
func ExtractIDs(text string) []ID {
	var out []ID
	seen := map[string]bool{}
	add := func(id ID) {
		if !seen[id.Kind+id.Value] {
			seen[id.Kind+id.Value] = true
			out = append(out, id)
		}
	}
	for _, m := range reVAT.FindAllString(text, -1) {
		add(ID{Kind: "vat", Value: compact(m)})
	}
	for _, m := range reCHE.FindAllStringSubmatch(text, -1) {
		if d := m[1] + m[2] + m[3]; validCHE(d) {
			add(ID{Kind: "che", Value: "CHE" + d})
		}
	}
	for _, m := range reCNPJ.FindAllStringSubmatch(text, -1) {
		d := m[1] + m[2] + m[3] + m[4] + m[5]
		// Without punctuation, 14 digits are only taken after "CNPJ".
		if !strings.ContainsAny(m[0], "./-") && !strings.Contains(strings.ToUpper(text), "CNPJ") {
			continue
		}
		if validCNPJ(d) {
			add(ID{Kind: "cnpj", Value: d})
		}
	}
	return out
}

// ExtractRegister finds register entries printed in lines.
func ExtractRegister(lines []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, l := range lines {
		for _, m := range reRegister.FindAllString(l, 3) {
			m = strings.Trim(strings.Join(strings.Fields(m), " "), " ,;:")
			if m != "" && !seen[strings.ToLower(m)] {
				seen[strings.ToLower(m)] = true
				out = append(out, m)
			}
		}
	}
	return out
}
