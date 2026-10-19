package lookup

import (
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// roleLabels maps labels found on legal notices and in registers to an
// English role. Labels are written folded (lower case, no accents, words
// separated by single spaces, hyphens as spaces). An empty role marks a
// label that names people who are NOT authorized representatives, so the
// text after it is skipped.
var roleLabels = map[string]string{
	// German (Germany, Austria, Switzerland).
	"geschaftsfuhrer": "Managing Director", "geschaftsfuhrerin": "Managing Director", "geschaftsfuhrerinnen": "Managing Director",
	"geschaftsfuhrer in": "Managing Director", "geschaftsfuhrer innen": "Managing Director", "geschaeftsfuehrer": "Managing Director",
	"geschaeftsfuehrerin": "Managing Director", "geschaftsfuhrung": "Managing Director", "geschaeftsfuehrung": "Managing Director",
	"geschaftsfuhrende gesellschafter": "Managing Partner", "geschaftsfuhrender gesellschafter": "Managing Partner",
	"geschaftsfuhrende gesellschafterin": "Managing Partner", "geschaftsfuhrende gesellschafter innen": "Managing Partner",
	"kaufmannischer geschaftsfuhrer": "Managing Director (Commercial)", "technischer geschaftsfuhrer": "Managing Director (Technical)",
	"hauptgeschaftsfuhrer": "Chief Executive", "alleiniger geschaftsfuhrer": "Sole Managing Director",
	"vertretungsberechtigter geschaftsfuhrer": "Managing Director", "vertretungsberechtigte geschaftsfuhrer": "Managing Director",
	"vertretungsberechtigte geschaftsfuhrerin": "Managing Director", "vertretungsberechtigte geschaftsfuhrung": "Managing Director",
	"vertretungsberechtigt": "Legal Representative", "vertretungsberechtigte": "Legal Representative", "vertretungsberechtigter": "Legal Representative",
	"vertretungsberechtigte person": "Legal Representative", "vertretungsberechtigte personen": "Legal Representative",
	"vertretungsberechtigter vorstand": "Member of the Management Board", "vertreten durch": "Legal Representative",
	"gesetzlich vertreten durch": "Legal Representative", "gesetzliche vertreter": "Legal Representative", "gesetzlicher vertreter": "Legal Representative",
	"vorstand": "Member of the Management Board", "vorstande": "Member of the Management Board", "der vorstand": "Member of the Management Board",
	"vorstandsmitglied": "Member of the Management Board", "vorstandsmitglieder": "Member of the Management Board",
	"mitglieder des vorstands": "Member of the Management Board", "mitglied des vorstands": "Member of the Management Board",
	"vorstandsvorsitzender": "Chairman of the Management Board", "vorstandsvorsitzende": "Chairwoman of the Management Board",
	"vorsitzender des vorstands": "Chairman of the Management Board", "vorsitzende des vorstands": "Chairwoman of the Management Board",
	"vorsitzender des vorstandes": "Chairman of the Management Board", "vorstandssprecher": "Spokesperson of the Management Board",
	"sprecher des vorstands": "Spokesperson of the Management Board", "prokurist": "Authorized Signatory (Prokurist)",
	"prokuristin": "Authorized Signatory (Prokurist)", "prokuristen": "Authorized Signatory (Prokurist)", "prokura": "Authorized Signatory (Prokurist)",
	"gesamtprokura": "Authorized Signatory (Prokurist)", "einzelprokura": "Authorized Signatory (Prokurist)",
	"inhaber": "Owner", "inhaberin": "Owner", "inhaber in": "Owner", "komplementar": "General Partner", "komplementarin": "General Partner",
	"personlich haftende gesellschafterin": "General Partner", "personlich haftender gesellschafter": "General Partner",
	"liquidator": "Liquidator", "liquidatorin": "Liquidator", "liquidatoren": "Liquidator",
	"geschaftsleitung": "Member of the Executive Management", "geschaftsleiter": "Executive Manager", "geschaftsleiterin": "Executive Manager",
	"mitglied der geschaftsleitung": "Member of the Executive Management", "mitglieder der geschaftsleitung": "Member of the Executive Management",
	"verwaltungsrat": "Member of the Board of Directors", "verwaltungsrate": "Member of the Board of Directors",
	"mitglied des verwaltungsrates": "Member of the Board of Directors", "mitglieder des verwaltungsrates": "Member of the Board of Directors",
	"prasident des verwaltungsrates": "Chairman of the Board of Directors", "prasident des verwaltungsrats": "Chairman of the Board of Directors",
	"verwaltungsratsprasident": "Chairman of the Board of Directors", "verwaltungsratsprasidentin": "Chairwoman of the Board of Directors",
	"delegierter des verwaltungsrates": "Delegate of the Board of Directors", "direktor": "Director", "direktorin": "Director",
	"zeichnungsberechtigt": "Authorized Signatory", "zeichnungsberechtigte": "Authorized Signatory", "zeichnungsberechtigter": "Authorized Signatory",
	// German labels that name other people.
	"aufsichtsrat": "", "aufsichtsrate": "", "aufsichtsratsvorsitzender": "", "aufsichtsratsvorsitzende": "",
	"vorsitzender des aufsichtsrats": "", "vorsitzende des aufsichtsrats": "", "vorsitzender des aufsichtsrates": "",
	"vorsitzende des aufsichtsrates": "", "beirat": "", "datenschutzbeauftragter": "", "datenschutzbeauftragte": "",
	"verantwortlich": "", "verantwortlicher": "", "inhaltlich verantwortlich": "", "redaktion": "", "redaktionell verantwortlich": "",
	"ansprechpartner": "", "ansprechpartnerin": "", "webmaster": "", "konzept": "", "gestaltung": "", "bildnachweis": "", "fotos": "",
	// English.
	"managing director": "Managing Director", "managing directors": "Managing Director", "director": "Director", "directors": "Director",
	"board of directors": "Member of the Board of Directors", "executive director": "Executive Director", "executive directors": "Executive Director",
	"sole director": "Sole Director", "chief executive officer": "Chief Executive Officer", "ceo": "Chief Executive Officer",
	"chief financial officer": "Chief Financial Officer", "cfo": "Chief Financial Officer", "chief operating officer": "Chief Operating Officer",
	"coo": "Chief Operating Officer", "president": "President", "vice president": "Vice President", "chairman": "Chairman",
	"chairwoman": "Chairwoman", "chairperson": "Chairperson", "chair": "Chairperson", "chairman of the board": "Chairman of the Board",
	"company secretary": "Company Secretary", "secretary": "Secretary", "treasurer": "Treasurer", "general manager": "General Manager",
	"managing partner": "Managing Partner", "managing partners": "Managing Partner", "partner": "Partner", "partners": "Partner",
	"owner": "Owner", "owners": "Owner", "proprietor": "Owner", "sole proprietor": "Owner", "legal representative": "Legal Representative",
	"legal representatives": "Legal Representative", "authorized representative": "Authorized Representative",
	"authorized representatives": "Authorized Representative", "authorised representative": "Authorized Representative",
	"authorised representatives": "Authorized Representative", "authorized signatory": "Authorized Signatory",
	"authorised signatory": "Authorized Signatory", "authorized officer": "Authorized Officer", "represented by": "Legal Representative",
	"legally represented by": "Legal Representative", "management": "Management", "management board": "Member of the Management Board",
	"executive board": "Member of the Executive Board", "board of management": "Member of the Management Board",
	"managing board": "Member of the Management Board", "officers": "Officer", "manager": "Manager", "managers": "Manager",
	"managing member": "Managing Member", "member manager": "Managing Member", "trustee": "Trustee", "trustees": "Trustee",
	// English labels that name other people.
	"data protection officer": "", "dpo": "", "editor": "", "responsible for content": "", "responsible for the content": "",
	"contact": "", "contact person": "", "photo credits": "", "registered agent": "", "non executive director": "",
	"non executive directors": "", "supervisory board": "", "chairman of the supervisory board": "", "founder": "", "founders": "",
	// Spanish.
	"representante legal": "Legal Representative", "representantes legales": "Legal Representative", "representante": "Legal Representative",
	"representante legal suplente": "Alternate Legal Representative", "suplente del representante legal": "Alternate Legal Representative",
	"representado por": "Legal Representative", "representada por": "Legal Representative", "representada legalmente por": "Legal Representative",
	"gerente": "Manager", "gerentes": "Manager", "gerente general": "General Manager", "gerente legal": "Legal Manager",
	"administrador": "Administrator", "administradora": "Administrator", "administradores": "Administrator",
	"administrador unico": "Sole Administrator", "administradora unica": "Sole Administrator", "adm unico": "Sole Administrator",
	"administrador solidario": "Joint and Several Administrator", "administradores solidarios": "Joint and Several Administrator",
	"adm solid": "Joint and Several Administrator", "administrador mancomunado": "Joint Administrator",
	"administradores mancomunados": "Joint Administrator", "adm mancom": "Joint Administrator",
	"consejero delegado": "Chief Executive Officer (Managing Director)", "consejera delegada": "Chief Executive Officer (Managing Director)",
	"consejeros delegados": "Chief Executive Officer (Managing Director)", "cons del": "Chief Executive Officer (Managing Director)",
	"consejero": "Member of the Board of Directors", "consejera": "Member of the Board of Directors", "consejeros": "Member of the Board of Directors",
	"consejo de administracion": "Member of the Board of Directors", "presidente": "President", "presidenta": "President",
	"presidente del consejo": "Chairman of the Board", "presidente del consejo de administracion": "Chairman of the Board",
	"presidente del directorio": "Chairman of the Board", "presidente de la junta directiva": "Chairman of the Board",
	"presidente ejecutivo": "Executive President", "vicepresidente": "Vice President", "vicepresidenta": "Vice President",
	"directorio": "Member of the Board of Directors", "director general": "General Director", "directora general": "General Director",
	"director ejecutivo": "Executive Director", "directora ejecutiva": "Executive Director", "director gerente": "Managing Director",
	"apoderado": "Attorney-in-fact", "apoderada": "Attorney-in-fact", "apoderados": "Attorney-in-fact", "apoderado general": "General Attorney-in-fact",
	"apoderado legal": "Legal Attorney-in-fact", "apoderado mancomunado": "Joint Attorney-in-fact", "apoderado solidario": "Joint and Several Attorney-in-fact",
	"apo manc": "Joint Attorney-in-fact", "apo sol": "Joint and Several Attorney-in-fact", "socio gerente": "Managing Partner",
	"socios gerentes": "Managing Partner", "socio administrador": "Managing Partner", "socia administradora": "Managing Partner",
	"vocal": "Board Member", "vocales": "Board Member", "secretario": "Secretary", "secretario del consejo": "Secretary of the Board",
	"tesorero": "Treasurer", "tesorera": "Treasurer", "titular": "Owner", "propietario": "Owner", "propietaria": "Owner",
	"liquidador": "Liquidator", "liquidadora": "Liquidator", "responsable legal": "Legal Representative",
	// Spanish labels that name other people.
	"comisario": "", "comisarios": "", "auditor": "", "auditores": "", "contacto": "", "delegado de proteccion de datos": "",
	// Portuguese.
	"socio administrador s": "Managing Partner", "socios administradores": "Managing Partner", "socio gerente s": "Managing Partner",
	"socia gerente": "Managing Partner", "gerencia": "Management", "diretor": "Director", "diretora": "Director", "diretores": "Director",
	"diretor presidente": "Chief Executive Officer", "diretora presidente": "Chief Executive Officer", "diretor geral": "General Director",
	"diretora geral": "General Director", "diretor executivo": "Executive Director", "diretora executiva": "Executive Director",
	"diretor superintendente": "Superintendent Director", "diretoria": "Member of the Executive Board",
	"presidente do conselho de administracao": "Chairman of the Board", "conselho de administracao": "Member of the Board of Directors",
	"membro do conselho de administracao": "Member of the Board of Directors", "vogal": "Board Member", "vogais": "Board Member",
	"procurador": "Attorney-in-fact", "procuradora": "Attorney-in-fact", "procuradores": "Attorney-in-fact",
	"responsavel legal": "Legal Representative", "empresario": "Owner", "representado legalmente por": "Legal Representative",
	"encarregado de dados": "", "encarregado": "",
	// French.
	"president directeur general": "Chairman and Chief Executive Officer", "pdg": "Chairman and Chief Executive Officer",
	"directeur general": "Chief Executive Officer", "directrice generale": "Chief Executive Officer",
	"directeur general delegue": "Deputy Chief Executive Officer", "directrice generale deleguee": "Deputy Chief Executive Officer",
	"gerant": "Manager", "gerante": "Manager", "gerants": "Manager", "co gerant": "Co-Manager", "cogerant": "Co-Manager",
	"administrateur": "Director", "administrateurs": "Director", "president du conseil d administration": "Chairman of the Board",
	"representant legal": "Legal Representative", "represente par": "Legal Representative", "representee par": "Legal Representative",
	"directeur de la publication": "Publication Director (Legal Representative)", "directrice de la publication": "Publication Director (Legal Representative)",
	"responsable de la publication": "Publication Director (Legal Representative)", "mandataire social": "Corporate Officer",
	"hebergeur": "", "delegue a la protection des donnees": "",
	// Italian.
	"amministratore delegato": "Chief Executive Officer (Managing Director)", "amministratore unico": "Sole Director",
	"amministratori": "Director", "amministratore": "Director", "legale rappresentante": "Legal Representative",
	"rappresentante legale": "Legal Representative", "presidente del consiglio di amministrazione": "Chairman of the Board",
	"consiglio di amministrazione": "Member of the Board of Directors", "direttore generale": "General Director",
	"procuratore": "Attorney-in-fact", "procuratori": "Attorney-in-fact",
	// Dutch.
	"directeur": "Director", "directie": "Management", "bestuurder": "Director", "bestuurders": "Director", "bestuur": "Board Member",
	"algemeen directeur": "General Director", "statutair directeur": "Statutory Director", "statutair bestuurder": "Statutory Director",
	"vertegenwoordigd door": "Legal Representative", "gevolmachtigde": "Authorized Signatory", "eigenaar": "Owner",
	// Polish, Czech, Slovak.
	"prezes zarzadu": "President of the Management Board", "wiceprezes zarzadu": "Vice President of the Management Board",
	"czlonek zarzadu": "Member of the Management Board", "czlonkowie zarzadu": "Member of the Management Board",
	"zarzad": "Member of the Management Board", "prokurent": "Authorized Signatory (Prokurist)", "wlasciciel": "Owner",
	"jednatel": "Managing Director", "jednatele": "Managing Director", "jednatelka": "Managing Director", "konatel": "Managing Director",
	"statutarni organ": "Statutory Body", "predseda predstavenstva": "Chairman of the Board of Directors",
	"clen predstavenstva": "Member of the Board of Directors", "prokurista": "Authorized Signatory (Prokurist)",
	// Nordic.
	"verkstallande direktor": "Chief Executive Officer", "vd": "Chief Executive Officer", "styrelseordforande": "Chairman of the Board",
	"administrerende direktor": "Chief Executive Officer", "daglig leder": "Chief Executive Officer", "toimitusjohtaja": "Chief Executive Officer",
}

// labelIndex maps the first word of each label to the label word lists
// that start with it, longest first.
var labelIndex = map[string][][]string{}

func init() {
	for l := range roleLabels {
		w := strings.Fields(l)
		labelIndex[w[0]] = append(labelIndex[w[0]], w)
	}
	for k := range labelIndex {
		sort.Slice(labelIndex[k], func(i, j int) bool {
			if len(labelIndex[k][i]) != len(labelIndex[k][j]) {
				return len(labelIndex[k][i]) > len(labelIndex[k][j])
			}
			return strings.Join(labelIndex[k][i], " ") < strings.Join(labelIndex[k][j], " ")
		})
	}
}

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

// byLabel reports whether a label ends with "by" in some language, so
// the names follow without a colon.
func byLabel(last string) bool {
	switch last {
	case "durch", "by", "por", "par", "door", "da":
		return true
	}
	return false
}

// stopValue ends a list of names: the next field of a legal notice.
var stopValue = regexp.MustCompile(`(?i)\s*(?:\||•|·|\.\s+(?:datos\s+registrales|ceses|dimisiones|revocaciones|nombramientos|reelecciones|cambio|modificaciones|otros\s+conceptos|constituci[oó]n|ampliaci[oó]n|reducci[oó]n|declaraci[oó]n|sociedad\s+unipersonal|disoluci[oó]n|extinci[oó]n|fusi[oó]n|escisi[oó]n|transformaci[oó]n|situaci[oó]n\s+concursal|emisi[oó]n|cancelaciones|traslado|objeto\s+social)\b|\b(?:tel|telefon|telephone|phone|fon|fax|e-?mail|mail|web|www|internet|sitz|registergericht|amtsgericht|handelsregister|registernummer|register|hrb|hra|ust|umsatzsteuer|steuernummer|firmenbuch|fn|uid|domicilio|direcci[oó]n|cif|nif|nit|rfc|rut|ruc|cuit|cnpj|cpf|inscrita|registro|capital|endere[cç]o|sede|si[eè]ge|siret|siren|rcs|tva|vat|company\s+number|registered|address|anschrift|adresse|aufsichtsbeh[oö]rde)\b).*$`)

var splitNames = regexp.MustCompile(`(?i)\s*(?:,|;|/|&|\+|\bund\b|\band\b|\by\b|\be\b|\bet\b|\bsowie\b|\ben\b|\boraz\b)\s*`)

var honorific = regexp.MustCompile(`(?i)^(?:(?:herrn?|frau|mr|mrs|ms|miss|mx|sr|sra|srta|don|dona|doña|d|dña|dna|m|mme|mlle|sig|sig\.ra|dott|dott\.ssa|ing|dipl\.?-?\s*(?:ing|kfm|kff|vw|oec|inf|wirt\.?-?ing)\.?(?:\s*\(fh\))?|mag|lic|licda|lcdo|lcda|arq|abog|eng|ir|drs|mba|ll\.?m|m\.?sc|b\.?sc|b\.?a|m\.?a)\.?\s+)+`)

var titleKeep = regexp.MustCompile(`^(?:(?:Dr|Prof|Dra|PhD)\.?\s+)+`)

var nameStop = map[string]bool{
	"gmbh": true, "ag": true, "kg": true, "ltd": true, "inc": true, "llc": true, "sa": true, "ltda": true, "mbh": true, "co": true,
	"sl": true, "slu": true, "sau": true, "sas": true, "srl": true, "sociedad": true, "limited": true, "corp": true, "bv": true,
	"nv": true, "lda": true, "spa": true, "sarl": true, "plc": true, "eireli": true, "cia": true, "compania": true, "companhia": true,
	"strasse": true, "str": true, "street": true, "calle": true, "avenida": true, "av": true, "rua": true, "road": true,
	"gesellschaft": true, "company": true, "holding": true, "group": true, "gruppe": true, "grupo": true, "www": true,
	"the": true, "der": true, "die": true, "das": true, "el": true, "los": true, "las": true, "le": true, "les": true,
	"und": true, "and": true, "oder": true, "or": true, "with": true, "mit": true, "con": true, "com": true, "de": false,
	"management": true, "board": true, "team": true, "directors": true, "vorstand": true, "geschaftsfuhrung": true,
	"impressum": true, "imprint": true, "kontakt": true, "contact": true, "datenschutz": true, "privacy": true,
	"deutschland": true, "germany": true, "austria": true, "osterreich": true, "schweiz": true, "switzerland": true, "mexico": true,
	"brasil": true, "brazil": true, "espana": true, "spain": true, "portugal": true, "colombia": true, "argentina": true,
	"herr": true, "frau": true, "seit": true, "since": true, "desde": true, "bis": true, "until": true,
}

var particles = map[string]bool{"von": true, "van": true, "de": true, "del": true, "der": true, "da": true, "das": true, "dos": true, "di": true, "zu": true, "la": true, "le": true, "y": true, "e": true, "ten": true, "ter": true, "den": true, "vom": true, "zum": true}

// titleCase turns "PEREZ GARCIA JUAN" into "Perez Garcia Juan".
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		lw := strings.ToLower(w)
		if particles[textutil.Fold(lw)] && i > 0 {
			words[i] = lw
			continue
		}
		parts := strings.Split(lw, "-")
		for j, p := range parts {
			r := []rune(p)
			if len(r) > 0 {
				r[0] = unicode.ToUpper(r[0])
			}
			// O'Neil, D'Angelo.
			for k := 1; k+1 < len(r); k++ {
				if r[k] == '\'' || r[k] == '’' {
					r[k+1] = unicode.ToUpper(r[k+1])
				}
			}
			parts[j] = string(r)
		}
		words[i] = strings.Join(parts, "-")
	}
	return strings.Join(words, " ")
}

func allUpper(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

// cleanName strips honorifics and checks that s looks like a person's
// name: 2-6 words, capitalised, letters, hyphens, apostrophes and dots.
func cleanName(s string) (string, bool) {
	s = strings.Trim(textutil.OneLine(s), " .:,;-–—*")
	title := ""
	if m := titleKeep.FindString(s); m != "" {
		title, s = strings.Join(strings.Fields(m), " ")+" ", strings.TrimSpace(s[len(m):])
	}
	s = strings.TrimSpace(honorific.ReplaceAllString(s, ""))
	if m := titleKeep.FindString(s); m != "" && title == "" {
		title, s = strings.Join(strings.Fields(m), " ")+" ", strings.TrimSpace(s[len(m):])
	}
	if len(s) > 70 {
		return "", false
	}
	if allUpper(s) {
		s = titleCase(s)
	}
	words := strings.Fields(s)
	if len(words) < 2 || len(words) > 6 {
		return "", false
	}
	long := 0
	for _, w := range words {
		fw := textutil.Fold(strings.Trim(w, "."))
		if nameStop[fw] {
			return "", false
		}
		r := []rune(w)
		if !unicode.IsUpper(r[0]) {
			if particles[fw] {
				continue
			}
			return "", false
		}
		letters := 0
		for _, c := range r {
			switch {
			case unicode.IsLetter(c):
				letters++
			case c == '-' || c == '\'' || c == '’' || c == '.':
			default:
				return "", false
			}
		}
		if letters >= 2 {
			long++
		}
	}
	if long < 2 {
		return "", false
	}
	return title + strings.Join(words, " "), true
}

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
