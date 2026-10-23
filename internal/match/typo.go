package match

import (
	"sort"
	"strings"
	"unicode"
)

// PlaceFix is a city or country written after a company name that is
// misspelt ("Colomiba") or cut off ("Ciudad de Me").
type PlaceFix struct {
	From    string   // as written
	To      string   // the correction; empty when Choices has several
	How     string   // "typo" or "cut off"
	Choices []string // places it could be, when more than one
	At      int      // byte offset of From in the name
}

// isForm reports whether s on its own is a legal form ("S.A.", "LLC").
func isForm(s string) bool {
	m, ok := findForm("X " + s)
	return ok && m.Base == "X"
}

// placeWord accepts text that may be a place: one to five words of
// letters, dots, hyphens and apostrophes, at most 40 characters.
func placeWord(s string) bool {
	w := strings.Fields(s)
	if len(w) == 0 || len(w) > 5 || len([]rune(s)) > 40 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && r != ' ' && r != '.' && r != '-' && r != '\'' && r != '’' {
			return false
		}
	}
	return true
}

// distance is the edit distance with transpositions (optimal string
// alignment) between two strings, in characters.
func distance(a, b string) int {
	x, y := []rune(a), []rune(b)
	d := make([][]int, len(x)+1)
	for i := range d {
		d[i] = make([]int, len(y)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for i := 1; i <= len(x); i++ {
		for j := 1; j <= len(y); j++ {
			cost := 1
			if x[i-1] == y[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && x[i-1] == y[j-2] && x[i-2] == y[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[len(x)][len(y)]
}

// allowed is how many typing errors a place of n letters may have.
// extra is how many letters the written text has more than the place
// (negative when letters are missing): none below 5 letters, since many
// short names look alike; one wrong, swapped or missing letter from 5;
// one extra letter from 7; two errors from 9.
func allowed(n, extra int) int {
	switch {
	case n < 5, extra > 0 && n < 7:
		return 0
	case n < 9:
		return 1
	}
	return 2
}

func uniq(list []string) []string {
	sort.Strings(list)
	var out []string
	for i, s := range list {
		if i == 0 || s != list[i-1] {
			out = append(out, s)
		}
	}
	return out
}

// placeFix looks at one part. last says it ends the cell, where text is
// cut off when a cell or an export was too short.
func placeFix(p string, last bool) (PlaceFix, bool) {
	f := foldPlace(p)
	if !placeWord(p) || len([]rune(f)) < 4 || isForm(p) {
		return PlaceFix{}, false
	}
	if _, ok := places[f]; ok {
		return PlaceFix{}, false // a known place; accents are left as written
	}
	var found []string
	if last {
		for k, name := range places {
			if len(k) > len(f) && strings.HasPrefix(k, f) {
				found = append(found, name)
			}
		}
		if found = uniq(found); len(found) > 0 {
			return choose(p, "cut off", found), true
		}
	}
	n, best := len([]rune(f)), 3
	for k, name := range places {
		if k[0] != f[0] {
			continue // a first letter is rarely mistyped
		}
		d := distance(f, k)
		if d > allowed(n, n-len([]rune(k))) {
			continue
		}
		switch {
		case d < best:
			best, found = d, []string{name}
		case d == best:
			found = append(found, name)
		}
	}
	if len(found) == 0 {
		return PlaceFix{}, false
	}
	return choose(p, "typo", uniq(found)), true
}

// choose fills To when there is one place, keeping capitals as written.
func choose(p, how string, found []string) PlaceFix {
	if strings.ToUpper(p) == p {
		for i := range found {
			found[i] = strings.ToUpper(found[i])
		}
	}
	fx := PlaceFix{From: p, How: how}
	if len(found) == 1 {
		fx.To = found[0]
	} else {
		fx.Choices = found
	}
	return fx
}

// Bounds on what FixPlaces and PlaceFixed look at: company names have a
// few parts, and a spreadsheet cell may hold anything.
const (
	maxParts = 12
	maxName  = 400
	maxTries = 4096
)

// FixPlaces finds misspelt or cut-off places written after a comma in a
// company name: "Foo S.A. de C.V., Ciudad de Me" and "Foo SAS, Bogta,
// Colomiba". The part before the first comma (the name itself) and legal
// forms are never changed. Apply builds the corrected name.
func FixPlaces(name string) []PlaceFix {
	parts := strings.Split(name, ",")
	if len(parts) > maxParts || len(name) > maxName {
		return nil // not a company name; never spend long on a cell
	}
	var out []PlaceFix
	off := len(parts[0]) + 1
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		subs := []string{part}
		if strings.Contains(part, " - ") || strings.Contains(part, " – ") {
			subs = strings.FieldsFunc(part, func(r rune) bool { return r == '-' || r == '–' })
		}
		from := 0
		for j, sub := range subs {
			p := strings.Trim(strings.TrimSpace(sub), ".;")
			at := strings.Index(part[from:], p) + from
			from = at + len(p)
			last := i == len(parts)-1 && j == len(subs)-1 && strings.TrimSpace(sub) == p
			if fx, ok := placeFix(p, last); ok && p != "" {
				fx.At = off + at
				out = append(out, fx)
			}
		}
		off += len(part) + 1
	}
	return out
}

// Apply writes each fix's To in place of its From.
func Apply(name string, fixes []PlaceFix) string {
	for i := len(fixes) - 1; i >= 0; i-- {
		f := fixes[i]
		if f.To == "" || f.At < 0 || f.At+len(f.From) > len(name) || name[f.At:f.At+len(f.From)] != f.From {
			continue
		}
		name = name[:f.At] + f.To + name[f.At+len(f.From):]
	}
	return name
}

// PlaceFixed reports whether to is name with some of the place fixes
// that FixPlaces offers for it, one choice per fix.
func PlaceFixed(name, to string) bool {
	fixes := FixPlaces(name)
	tries := 1
	for _, f := range fixes {
		if tries *= 2 + len(f.Choices); tries > maxTries {
			return false // too many combinations: a person must look
		}
	}
	set := make([]PlaceFix, len(fixes))
	var try func(int) bool
	try = func(i int) bool {
		if i == len(fixes) {
			return Apply(name, set) == to
		}
		set[i] = fixes[i]
		for _, c := range append([]string{"", fixes[i].To}, fixes[i].Choices...) {
			if set[i].To = c; try(i + 1) {
				return true
			}
		}
		return false
	}
	return name != to && try(0)
}
