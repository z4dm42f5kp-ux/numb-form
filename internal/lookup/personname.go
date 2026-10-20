package lookup

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

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
