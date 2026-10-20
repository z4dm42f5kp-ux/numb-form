package lookup

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
)

// legalNameIn returns the company's full legal name when a line writes
// the base name followed by a legal form ("Foo Bar GmbH & Co. KG").
func legalNameIn(line string, q Query) (string, bool) {
	words := strings.Fields(q.Base)
	if len(words) == 0 {
		return "", false
	}
	var p []string
	for _, w := range words {
		p = append(p, regexp.QuoteMeta(w))
	}
	re, err := regexp.Compile(`(?i)` + strings.Join(p, `[\s\-.,&]+`) + `[^|•·;\n]{0,40}`)
	if err != nil {
		return "", false
	}
	m := re.FindString(line)
	if m == "" {
		return "", false
	}
	// Cut at the first comma or dash that follows the legal form.
	for cut := len(m); cut > len(q.Base); cut-- {
		cand := strings.TrimSpace(m[:cut])
		if f, _ := match.LegalForm(cand); f != "" {
			return strings.Trim(cand, " ,;:"), true
		}
	}
	return "", false
}

// entityHeading reports whether a line names a legal entity on its own,
// such as "Foo Bar GmbH" or "© 2025 Foo S.A.": short, no label, no colon.
func entityHeading(l string) (string, bool) {
	if len(l) > 100 || strings.ContainsAny(l, ":：") || len(strings.Fields(l)) > 12 {
		return "", false
	}
	m, ok := match.FindLegalForm(l)
	if !ok || len(strings.Fields(m.Base)) > 8 {
		return "", false
	}
	return strings.TrimSpace(normWords(m.Base)), true
}

// companyLines returns the lines of a page that are about the company.
// A page naming several legal entities (a group's legal notice) is cut
// into blocks at each entity name, and only the blocks of the company
// are kept. A page naming no other entity is kept whole.
func companyLines(lines []string, q Query) []string {
	type head struct {
		i    int
		ours bool
	}
	var heads []head
	others := false
	for i, l := range lines {
		base, ok := entityHeading(l)
		if !ok || base == "" {
			continue
		}
		ours := Mentions(normWords(l), q)
		if !ours {
			// The base name alone also counts as ours ("Foo" in "© Foo").
			ours = strings.Contains(normWords(q.Base), " "+base+" ") || strings.Contains(" "+base+" ", normWords(q.Base))
		}
		heads = append(heads, head{i, ours})
		if !ours {
			others = true
		}
	}
	if !others {
		return lines
	}
	var out []string
	for k, h := range heads {
		if !h.ours {
			continue
		}
		end := len(lines)
		for _, n := range heads[k+1:] {
			if !n.ours {
				end = n.i
				break
			}
		}
		out = append(out, lines[h.i:end]...)
	}
	return out
}

// domainIsName reports whether the site's domain is the company's name,
// such as foobar.de for "Foo Bar GmbH".
func domainIsName(u *url.URL, q Query) bool {
	label := strings.SplitN(hostKey(u), ".", 2)[0]
	for _, s := range slugs(q) {
		if label == s {
			return true
		}
	}
	return false
}

// operator reports whether the company runs the site: on a legal notice
// or home page, it is the first legal entity named, or it appears in the
// copyright line.
func operator(lines []string, q Query) bool {
	first := true
	for _, l := range lines {
		if strings.Contains(l, "©") || strings.Contains(strings.ToLower(l), "copyright") {
			if Mentions(normWords(l), q) {
				return true
			}
			continue
		}
		if _, ok := entityHeading(l); ok && first {
			first = false
			if Mentions(normWords(l), q) {
				return true
			}
		}
	}
	return false
}
