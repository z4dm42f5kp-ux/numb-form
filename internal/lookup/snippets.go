package lookup

import (
	"net/url"
	"regexp"
	"strings"
)

// snippetSeps splits a search snippet into the parts directories print
// side by side ("Geschäftsführer: X · Musterstr. 1, 12345 Berlin").
var snippetSeps = strings.NewReplacer(" · ", "\n", " | ", "\n", "; ", "\n", " - ", "\n", " – ", "\n", "... ", "\n", "… ", "\n")

// snippetLines splits a search snippet at sentence ends so that labels
// such as "Geschäftsführer:" start a line.
func snippetLines(s string) []string {
	var out []string
	for _, l := range strings.Split(snippetSeps.Replace(s), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// rePostalCity finds "12345 City" or "C.P. 12345" inside running text.
var rePostalCity = regexp.MustCompile(`(?:\b[A-Z]{1,2}-)?\b\d{4,5}(?:-\d{3})?\s+\p{Lu}[\p{L}.\- ]{1,40}`)

// inlineAddress finds "Street 1, 12345 City" written on one line, as
// search snippets and directory pages do.
func inlineAddress(line string) string {
	parts := strings.Split(line, ",")
	for i, p := range parts {
		p = strings.TrimSpace(p)
		m := rePostalCity.FindString(p)
		if m == "" {
			continue
		}
		var street string
		switch {
		case strings.Index(p, m) > 3 && reStreet.MatchString(p[:strings.Index(p, m)]):
			street = strings.TrimSpace(strings.Trim(p[:strings.Index(p, m)], " ,:"))
		case i > 0 && reStreet.MatchString(parts[i-1]) && len(parts[i-1]) < 70:
			street = strings.TrimSpace(parts[i-1])
		default:
			continue
		}
		if i := strings.LastIndex(street, ":"); i >= 0 {
			street = strings.TrimSpace(street[i+1:])
		}
		if street == "" || reNoAddr.MatchString(street) || !strings.ContainsAny(street, "0123456789") {
			continue
		}
		return street + ", " + strings.TrimSpace(strings.TrimRight(m, " .-"))
	}
	return ""
}

// mineSnippets reads representatives and addresses from the titles and
// snippets of search results that name the company. They are community
// data: shown unverified, with the caution text.
func mineSnippets(hits []Hit, q Query) Result {
	var res Result
	for _, h := range hits {
		text := h.Title + " · " + h.Snippet
		if !Mentions(normWords(text), q) {
			continue
		}
		host := ""
		if u, err := url.Parse(h.URL); err == nil {
			host = strings.TrimPrefix(u.Hostname(), "www.")
		}
		src := Source{ID: "search", Name: "search result from " + host + " (not verified)", URL: h.URL, Kind: KindCommunity}
		lines := snippetLines(h.Snippet)
		found := false
		for _, p := range ExtractPeople(lines, src) {
			res.addPerson(p)
			found = true
		}
		for _, l := range append(lines, h.Snippet) {
			if a := inlineAddress(l); a != "" {
				res.Addresses = append(res.Addresses, Address{Text: a, Source: src})
				found = true
				break
			}
		}
		if found {
			res.Sources = append(res.Sources, src)
		}
	}
	return res
}
