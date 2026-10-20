package lookup

import (
	"regexp"
	"strings"
)

var (
	rePostal    = regexp.MustCompile(`(?:^|[\s,])(?:[A-Z]{1,2}-)?\d{4,5}(?:-\d{3,4})?\s+\p{Lu}|(?i:\b(?:c\.?\s?p\.?|cep|zip|postal\s+code|plz)\s*:?\s*\d{4,5})|\b\d{4}\s?[A-Z]{2}\b|\b[A-Z]{1,2}\d[A-Z\d]?\s?\d[A-Z]{2}\b|\b[A-Z]{2}\s+\d{5}(?:-\d{4})?\b`)
	reStreet    = regexp.MustCompile(`(?i)\p{L}.*\d|\d.*\p{L}|\b(?:stra(?:ss|ß)e|str\.|weg|platz|allee|gasse|ring|damm|calle|avenida|av\.|carrera|cra\.|rua|travessa|rodovia|road|street|st\.|avenue|ave\.|blvd|boulevard|suite|piso|of\.|oficina|col\.|colonia|bairro|via|viale|rue|edificio|torre)\b`)
	reAddrLabel = regexp.MustCompile(`(?i)^\s*(?:anschrift|adresse|address|postanschrift|sitz(?:\s+der\s+gesellschaft)?|gesch[aä]ftssitz|firmensitz|registered\s+office|head\s+office|headquarters|domicilio(?:\s+social|\s+fiscal)?|direcci[oó]n(?:\s+fiscal)?|sede(?:\s+social)?|endere[cç]o|morada|si[eè]ge(?:\s+social)?|indirizzo|sede\s+legale)\s*[:：]\s*(.*)$`)
	reNoAddr    = regexp.MustCompile(`(?i)\b(?:tel|telefon|telephone|phone|fon|fax|e-?mail|www\.|https?://|@|ust|uid|vat|iva|cif|nif|nit|rfc|cnpj|hrb|hra|amtsgericht|registergericht|handelsregister|gesch[aä]ftsf[uü]hrer|managing|director|copyright|©)\b`)
)

func addressLine(l string) bool {
	return len(l) < 90 && !reNoAddr.MatchString(l) && (reStreet.MatchString(l) || rePostal.MatchString(l))
}

// ExtractAddress finds the company's address in a legal notice: after a
// label such as "Anschrift:" or "Domicilio:", or in the lines that follow
// the company name, or as "Street 1" followed by "12345 City".
func ExtractAddress(lines []string, q ...Query) string {
	for i, l := range lines {
		m := reAddrLabel.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		parts := []string{}
		if v := strings.TrimSpace(m[1]); v != "" {
			parts = append(parts, v)
		}
		for j := i + 1; j < len(lines) && j <= i+3 && len(parts) < 3; j++ {
			if !addressLine(lines[j]) {
				break
			}
			parts = append(parts, lines[j])
		}
		a := strings.Join(parts, ", ")
		if rePostal.MatchString(a) || (len(parts) > 0 && reStreet.MatchString(parts[0]) && strings.ContainsAny(a, "0123456789")) {
			return a
		}
	}
	if len(q) > 0 {
		for i, l := range lines {
			if len(l) > 120 || !Mentions(normWords(l), q[0]) {
				continue
			}
			var parts []string
			for j := i + 1; j < len(lines) && j <= i+4 && len(parts) < 3; j++ {
				if !addressLine(lines[j]) {
					break
				}
				parts = append(parts, lines[j])
			}
			if len(parts) >= 1 && rePostal.MatchString(strings.Join(parts, " ")) {
				return strings.Join(parts, ", ")
			}
		}
	}
	for i := 1; i < len(lines) && i < 200; i++ {
		if rePostal.MatchString(lines[i]) && reStreet.MatchString(lines[i-1]) && addressLine(lines[i-1]) && len(lines[i]) < 60 {
			return lines[i-1] + ", " + lines[i]
		}
	}
	return ""
}
