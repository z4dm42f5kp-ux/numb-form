// Places written after a company name.

package match

import (
	"strings"
	"unicode"
)

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
