package lookup

import (
	"context"
	"net/url"
	"regexp"
	"strings"
)

// officialSearch searches one government website for the company and
// reads the first two pages found there.
func officialSearch(ctx context.Context, c *Client, q Query, domain, region string) (Result, error) {
	var notes []string
	hits := search(ctx, c, `"`+q.Name+`" site:`+domain, region, &notes)
	var res Result
	res.Notes = notes
	n := 0
	for _, h := range hits {
		u, err := url.Parse(h.URL)
		if err != nil || !officialHost(u.Hostname()) {
			continue
		}
		r, err := officialPage(ctx, c, q, h.URL)
		if err != nil {
			res.Notes = append(res.Notes, sanitizeErr(err))
		}
		res.merge(r)
		if n++; n >= 2 {
			break
		}
	}
	return res, nil
}

var (
	reEntryStart  = regexp.MustCompile(`^\d{3,7}\s*-\s*\S`) // BORME: "123456 - FOO SL."
	reSunbizTitle = regexp.MustCompile(`^Title\s+([A-Z][A-Z ,/&]{0,40})$`)
	sunbizCodes   = map[string]string{
		"P": "President", "PRES": "President", "PRESIDENT": "President", "VP": "Vice President", "V": "Vice President",
		"S": "Secretary", "SEC": "Secretary", "SECRETARY": "Secretary", "T": "Treasurer", "TREAS": "Treasurer", "TREASURER": "Treasurer",
		"D": "Director", "DIR": "Director", "DIRECTOR": "Director", "C": "Chairman", "CHAIRMAN": "Chairman", "CEO": "Chief Executive Officer",
		"CFO": "Chief Financial Officer", "COO": "Chief Operating Officer", "MGR": "Manager", "MANAGER": "Manager",
		"MGRM": "Managing Member", "MMBR": "Managing Member", "AMBR": "Authorized Member", "AP": "Authorized Person",
		"AR": "Authorized Representative", "GP": "General Partner", "M": "Member", "MBR": "Member",
	}
)

// officialPage reads a government register or gazette page. Only the
// part about the company is used: on gazettes that list many companies,
// the entry that names it.
func officialPage(ctx context.Context, c *Client, q Query, pageURL string) (Result, error) {
	var res Result
	budget := 1
	body, final, err := c.page(ctx, pageURL, &budget)
	if err != nil {
		return res, err
	}
	if final == nil {
		final, _ = url.Parse(pageURL)
	}
	lines := HTMLText(string(body))
	host := strings.TrimPrefix(final.Hostname(), "www.")
	src := Source{ID: "official", Name: "government register " + host + final.EscapedPath(), URL: final.String(), Kind: KindGovernment}
	var window []string
	for i, l := range lines {
		if !Mentions(normWords(l), q) {
			continue
		}
		window = append(window, l)
		for j := i + 1; j < len(lines) && j <= i+25; j++ {
			if reEntryStart.MatchString(lines[j]) {
				break
			}
			window = append(window, lines[j])
		}
		break
	}
	if len(window) == 0 {
		res.Notes = append(res.Notes, host+": the page does not name "+q.Name)
		return res, nil
	}
	people := ExtractPeople(window, src)
	if strings.HasSuffix(host, "sunbiz.org") {
		people = append(people, sunbizOfficers(lines, src)...)
		if a := labelledBlock(lines, "Principal Address"); a != "" {
			res.Addresses = append(res.Addresses, Address{Text: a, Source: src})
		}
	}
	for _, p := range people {
		res.addPerson(p)
	}
	if a := ExtractAddress(window, q); a != "" {
		res.Addresses = append(res.Addresses, Address{Text: a, Source: src})
	}
	if len(res.People)+len(res.Addresses) > 0 {
		res.Sources = append(res.Sources, src)
		res.Excerpts = append(res.Excerpts, Excerpt{Source: src, Text: strings.Join(window[:min(len(window), 12)], "\n")})
	}
	return res, nil
}

// sunbizOfficers reads the "Officer/Director Detail" list of a Florida
// Division of Corporations record: "Title P" followed by "DOE, JOHN".
func sunbizOfficers(lines []string, src Source) []Person {
	var out []Person
	for i, l := range lines {
		m := reSunbizTitle.FindStringSubmatch(strings.TrimSpace(l))
		if m == nil || i+1 >= len(lines) {
			continue
		}
		var roles []string
		for _, code := range strings.FieldsFunc(m[1], func(r rune) bool { return r == ',' || r == ' ' || r == '/' || r == '&' }) {
			if r, ok := sunbizCodes[code]; ok {
				roles = append(roles, r)
			}
		}
		if len(roles) == 0 {
			continue
		}
		if n, ok := cleanName(chName(lines[i+1])); ok {
			out = append(out, Person{Name: n, Roles: roles, Sources: []Source{src}})
		}
	}
	return out
}

// labelledBlock returns up to three address lines after a heading line.
func labelledBlock(lines []string, heading string) string {
	for i, l := range lines {
		if !strings.EqualFold(strings.TrimSpace(l), heading) {
			continue
		}
		var parts []string
		for j := i + 1; j < len(lines) && len(parts) < 3; j++ {
			if !addressLine(lines[j]) {
				break
			}
			parts = append(parts, lines[j])
		}
		return strings.Join(parts, ", ")
	}
	return ""
}
