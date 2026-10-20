package lookup

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// token is one word of a line with its position in the original text.
type token struct {
	fold       string
	start, end int // byte offsets in the line
}

func tokenize(line string) []token {
	var out []token
	start := -1
	for i, r := range line {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if start < 0 {
				start = i
			}
			continue
		}
		if start >= 0 {
			out = append(out, token{fold: textutil.Fold(line[start:i]), start: start, end: i})
			start = -1
		}
	}
	if start >= 0 {
		out = append(out, token{fold: textutil.Fold(line[start:]), start: start, end: len(line)})
	}
	return out
}

// labelAt returns the length in tokens and the role of the longest
// label starting at token i.
func labelAt(toks []token, i int) (int, string, bool) {
	for _, words := range labelIndex[toks[i].fold] {
		if i+len(words) > len(toks) {
			continue
		}
		ok := true
		for j, w := range words {
			if toks[i+j].fold != w {
				ok = false
				break
			}
		}
		if ok {
			return len(words), roleLabels[strings.Join(words, " ")], true
		}
	}
	return 0, "", false
}

// labelHit is a role label found in a line.
type labelHit struct {
	role       string
	start, end int // byte range of the label
	valueStart int // where the names begin
	colon      bool
}

var reSep = regexp.MustCompile(`^[\s.)]*(?:\((?:in|innen|s|es|a|as|r|e)\)[\s]*)?(?:[:：]|\s[-–—]\s|\s*[-–—]\s)`)

// findLabels finds role labels that are followed by a colon or dash, or
// that start the line.
func findLabels(line string) []labelHit {
	toks := tokenize(line)
	var hits []labelHit
	for i := 0; i < len(toks); i++ {
		n, role, ok := labelAt(toks, i)
		if !ok {
			continue
		}
		// Labels inside words are excluded by tokenization; a label must
		// begin the line or follow a separator or a lower-case word
		// ("vertreten durch", "represented by").
		end := toks[i+n-1].end
		rest := line[end:]
		h := labelHit{role: role, start: toks[i].start, end: end, valueStart: end}
		if m := reSep.FindStringIndex(rest); m != nil {
			h.colon, h.valueStart = true, end+m[1]
		} else if i == 0 && n == 1 && weakLabel[toks[i].fold] && strings.TrimSpace(rest) != "" {
			// "Director of Sales Jane Doe": a common word at the start of
			// a sentence is not a label unless a colon follows.
			continue
		} else if i != 0 && !byLabel(toks[i+n-1].fold) {
			// A role word in running text ("our director said") is not a
			// label; "represented by X" is.
			i += n - 1
			continue
		}
		hits = append(hits, h)
		i += n - 1
	}
	return hits
}

// stopValue ends a list of names: the next field of a legal notice.
var stopValue = regexp.MustCompile(`(?i)\s*(?:\||•|·|\.\s+(?:datos\s+registrales|ceses|dimisiones|revocaciones|nombramientos|reelecciones|cambio|modificaciones|otros\s+conceptos|constituci[oó]n|ampliaci[oó]n|reducci[oó]n|declaraci[oó]n|sociedad\s+unipersonal|disoluci[oó]n|extinci[oó]n|fusi[oó]n|escisi[oó]n|transformaci[oó]n|situaci[oó]n\s+concursal|emisi[oó]n|cancelaciones|traslado|objeto\s+social)\b|\b(?:tel|telefon|telephone|phone|fon|fax|e-?mail|mail|web|www|internet|sitz|registergericht|amtsgericht|handelsregister|registernummer|register|hrb|hra|ust|umsatzsteuer|steuernummer|firmenbuch|fn|uid|domicilio|direcci[oó]n|cif|nif|nit|rfc|rut|ruc|cuit|cnpj|cpf|inscrita|registro|capital|endere[cç]o|sede|si[eè]ge|siret|siren|rcs|tva|vat|company\s+number|registered|address|anschrift|adresse|aufsichtsbeh[oö]rde)\b).*$`)

var splitNames = regexp.MustCompile(`(?i)\s*(?:,|;|/|&|\+|\bund\b|\band\b|\by\b|\be\b|\bet\b|\bsowie\b|\ben\b|\boraz\b)\s*`)

var reNonWord = regexp.MustCompile(`[^\p{L}\p{N}]+`)

var reParen = regexp.MustCompile(`\(([^()]*)\)`)

// parenRoles translates notes such as "(Vorsitzender)" after a name.
var parenRoles = map[string]string{
	"vorsitzender": "Chairman", "vorsitzende": "Chairwoman", "vors": "Chairman", "sprecher": "Spokesperson", "sprecherin": "Spokesperson",
	"chairman": "Chairman", "chair": "Chairperson", "chairwoman": "Chairwoman", "presidente": "President", "president": "President",
	"ceo": "Chief Executive Officer", "cfo": "Chief Financial Officer", "coo": "Chief Operating Officer", "cto": "Chief Technology Officer",
	"stellv": "Deputy", "stellvertretend": "Deputy", "stv": "Deputy", "deputy": "Deputy", "suplente": "Alternate", "vice": "Deputy",
	"einzelvertretungsberechtigt": "Sole Power of Representation", "alleinvertretungsberechtigt": "Sole Power of Representation",
}

// splitPeople turns a value such as "Dr. A B (Vorsitzender), C D, CEO"
// into people with roles.
func splitPeople(value, role string, src Source) []Person {
	var out []Person
	for _, part := range splitNames.Split(value, -1) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		extra := []string{}
		part = reParen.ReplaceAllStringFunc(part, func(p string) string {
			inner := strings.Trim(p, "()")
			if r, ok := parenRoles[strings.Trim(textutil.Fold(strings.TrimSpace(inner)), ". ")]; ok {
				extra = append(extra, r)
			} else if r, ok := TranslateRole(inner); ok {
				extra = append(extra, r)
			}
			return " "
		})
		if n, ok := cleanName(part); ok {
			out = append(out, Person{Name: n, Roles: append([]string{role}, extra...), Sources: []Source{src}})
			continue
		}
		// "X Y, CEO": a role after a name belongs to that name.
		if r, ok := TranslateRole(part); ok && len(out) > 0 {
			out[len(out)-1].Roles = append(out[len(out)-1].Roles, r)
		}
	}
	return out
}

var leadArticles = regexp.MustCompile(`(?i)^\s*(?:(?:den|die|der|des|the|el|la|los|las|o|a|os|as|le|les|il|lo|its|seine[n]?|ihre[n]?|their|su|sus)\s+)+`)

// reResigned marks people who left: "Ceses/Dimisiones. Adm. Unico: X".
var reResigned = regexp.MustCompile(`(?i)\b(?:ceses?|dimisi[oó]n(?:es)?|revocaci[oó]n(?:es)?|renuncias?|bajas?|ehemalige?r?|ausgeschieden|abberufen|resigned|former|removed|retired|sa[ií]da|destitui[cç][aã]o|cessa[cç][aã]o|exonera\w*|cessazione|d[eé]mission)\b`)

// reSection finds gazette section headings ("Nombramientos.", "Ceses/
// Dimisiones.").
var reSection = regexp.MustCompile(`(?i)\b(nombramientos|reelecciones|ceses|dimisiones|revocaciones|cancelaciones|bestellt|bestellungen|ausgeschieden|appointments?|terminations?|resignations?|nomea[cç][oõ]es|cessa[cç][oõ]es)\b`)

// resigned reports whether the text before a label says that the people
// named left: the last gazette section heading is a resignation, or the
// words just before the label say so ("ehemaliger Geschäftsführer").
func resigned(before string) bool {
	if ms := reSection.FindAllStringSubmatch(before, -1); len(ms) > 0 {
		switch strings.ToLower(ms[len(ms)-1][1]) {
		case "nombramientos", "reelecciones", "bestellt", "bestellungen", "appointment", "appointments", "nomeações", "nomeacoes":
			return false
		}
		return true
	}
	if len(before) > 30 {
		before = before[len(before)-30:]
	}
	return reResigned.MatchString(before)
}

// ExtractPeople finds authorized representatives named in the lines of
// a legal notice, register extract or gazette entry.
func ExtractPeople(lines []string, src Source) []Person {
	var out []Person
	for i, line := range lines {
		hits := findLabels(line)
		for k, h := range hits {
			if h.role == "" {
				continue
			}
			end := len(line)
			if k+1 < len(hits) {
				end = hits[k+1].start
			}
			if resigned(line[:h.start]) {
				continue
			}
			role := h.role
			value := line[h.valueStart:end]
			// "Vertreten durch den Geschäftsführer X": the inner role wins.
			if strings.HasPrefix(role, "Legal Representative") {
				v := leadArticles.ReplaceAllString(value, "")
				if inner := findLabels(v); len(inner) > 0 && inner[0].start == 0 && inner[0].role != "" {
					role = inner[0].role
					value = v[inner[0].valueStart:]
					if len(inner) > 1 {
						value = v[inner[0].valueStart:inner[1].start]
					}
				}
			}
			value = stopValue.ReplaceAllString(value, "")
			found := splitPeople(value, role, src)
			// Label alone on its line (tables, definition lists): the
			// names are on the following lines.
			if strings.TrimSpace(strings.Trim(value, ":：-–— ")) == "" {
				for j := i + 1; j < len(lines) && j <= i+4; j++ {
					if len(findLabels(lines[j])) > 0 {
						break
					}
					ps := splitPeople(stopValue.ReplaceAllString(lines[j], ""), role, src)
					if len(ps) == 0 {
						break
					}
					found = append(found, ps...)
				}
			}
			out = append(out, found...)
		}
	}
	return out
}

// TranslateRole returns the English name of a role and whether the role
// was recognised.
func TranslateRole(s string) (string, bool) {
	k := strings.Join(strings.Fields(reNonWord.ReplaceAllString(textutil.Fold(s), " ")), " ")
	if v, ok := roleLabels[k]; ok && v != "" {
		return v, true
	}
	if v, ok := roleMap[roleKey(s)]; ok {
		return v, true
	}
	for _, suf := range []string{"innen", "en", "es", "s"} {
		if v, ok := roleLabels[strings.TrimSuffix(k, suf)]; ok && v != "" && strings.HasSuffix(k, suf) {
			return v, true
		}
		if v, ok := roleMap[strings.TrimSuffix(roleKey(s), suf)]; ok && strings.HasSuffix(roleKey(s), suf) {
			return v, true
		}
	}
	return strings.TrimSpace(s), false
}
