// Package textutil holds small text helpers shared by the other packages.
package textutil

import (
	"strings"
	"unicode"
)

// SafeTerminal removes control characters, ANSI escape sequences and
// bidirectional overrides from s so that text fetched from the network
// cannot manipulate the user's terminal.
func SafeTerminal(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\t':
			b.WriteRune(' ')
		case r < 0x20, r >= 0x7f && r <= 0x9f:
			// C0 and C1 controls, including ESC.
		case r >= 0x202a && r <= 0x202e, r >= 0x2066 && r <= 0x2069, r == 0x200e, r == 0x200f:
			// Bidirectional controls.
		case r == unicode.ReplacementChar:
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// OneLine collapses all whitespace runs in s to single spaces and trims it.
func OneLine(s string) string {
	return strings.Join(strings.Fields(SafeTerminal(strings.ReplaceAll(strings.ReplaceAll(s, "\r", " "), "\n", " "))), " ")
}

// XMLSafe drops runes that are not allowed in XML 1.0 documents.
func XMLSafe(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == 0x9, r == 0xA, r == 0xD:
			return r
		case r >= 0x20 && r <= 0xD7FF, r >= 0xE000 && r <= 0xFFFD, r >= 0x10000 && r <= 0x10FFFF:
			return r
		}
		return -1
	}, s)
}

var fold = map[rune]string{
	'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'å': "a", 'ā': "a", 'ă': "a", 'ą': "a",
	'ç': "c", 'ć': "c", 'č': "c",
	'ď': "d", 'đ': "d", 'ð': "d",
	'è': "e", 'é': "e", 'ê': "e", 'ë': "e", 'ē': "e", 'ę': "e", 'ě': "e",
	'ì': "i", 'í': "i", 'î': "i", 'ï': "i", 'ī': "i", 'ı': "i",
	'ł': "l", 'ľ': "l",
	'ñ': "n", 'ń': "n", 'ň': "n",
	'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ø': "o", 'ō': "o", 'ő': "o",
	'ř': "r",
	'ś': "s", 'š': "s", 'ş': "s", 'ß': "ss",
	'ť': "t", 'ţ': "t", 'þ': "th",
	'ù': "u", 'ú': "u", 'û': "u", 'ū': "u", 'ů': "u", 'ű': "u",
	'ý': "y", 'ÿ': "y",
	'ź': "z", 'ż': "z", 'ž': "z",
	'æ': "ae", 'œ': "oe",
}

// umlauts have two common ASCII spellings: ü = u = ue.
var umlautLong = map[rune]string{'ä': "ae", 'ö': "oe", 'ü': "ue"}
var umlautShort = map[rune]string{'ä': "a", 'ö': "o", 'ü': "u"}

// MatchKeys returns the comparison keys for a company or folder name.
// The three variations applied are: case is ignored, accents and umlauts
// are folded (both ü=u and ü=ue spellings are produced), and every
// character that is not a letter or digit is dropped.
func MatchKeys(s string) []string {
	short := foldWith(s, umlautShort)
	long := foldWith(s, umlautLong)
	if short == long {
		return []string{short}
	}
	return []string{short, long}
}

func foldWith(s string, uml map[rune]string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if v, ok := uml[r]; ok {
			b.WriteString(v)
			continue
		}
		if v, ok := fold[r]; ok {
			b.WriteString(v)
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Similarity returns 1 - levenshtein(a,b)/max(len(a),len(b)) over runes.
func Similarity(a, b string) float64 {
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 && len(rb) == 0 {
		return 1
	}
	m := len(ra)
	if len(rb) > m {
		m = len(rb)
	}
	return 1 - float64(levenshtein(ra, rb))/float64(m)
}

func levenshtein(a, b []rune) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}
