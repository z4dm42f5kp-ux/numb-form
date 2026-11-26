// Casing of a legal form as it is written in a name.

package match

import (
	"strings"
	"unicode"
)

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
