package lookup

import (
	"context"
	"html"
	"regexp"
	"sort"
	"strings"
)

// SiteResult is what a company website says about the company.
type SiteResult struct {
	Result
	Home     string // the home page that was read
	Verified bool   // the site belongs to the company (see operator)
	Pages    int

	mentioned bool // some page names the company, perhaps only in passing

	legal, other []Person // people from legal notices and the home page, and from other pages
}

type pageLink struct {
	url  string
	prio int
}

// ReadSite reads a company website: the home page and up to four pages
// that carry its legal identity (legal notice, about, contact, privacy).
// Verified is set when the company runs the site. typed marks a site the
// user typed: its findings are kept even when unverified.
func ReadSite(ctx context.Context, c *Client, site string, q Query, typed bool) SiteResult {
	var out SiteResult
	base, ok := NormalizeWebsite(site)
	if !ok {
		out.Notes = append(out.Notes, "website: not a usable address: "+site)
		return out
	}
	budget := 3 // pages the invisible browser may load for this site
	body, final, err := c.page(ctx, base.String(), &budget)
	if err != nil {
		out.Notes = append(out.Notes, "website "+base.Host+": "+sanitizeErr(err))
		return out
	}
	if final == nil {
		final = base
	}
	out.Home = final.Scheme + "://" + final.Host + "/"
	links := map[string]int{}
	for _, m := range reLink.FindAllStringSubmatch(string(body), -1) {
		href := html.UnescapeString(m[1])
		text := strings.Join(HTMLText(m[2]), " ")
		u, err := final.Parse(href)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || !sameSite(u, final) {
			continue
		}
		u.Scheme, u.Fragment = "https", ""
		for p, re := range legalPages {
			if re.MatchString(u.Path) || re.MatchString(text) {
				if old, ok := links[u.String()]; !ok || p < old {
					links[u.String()] = p
				}
				break
			}
		}
	}
	var list []pageLink
	for u, p := range links {
		list = append(list, pageLink{u, p})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].prio != list[j].prio {
			return list[i].prio < list[j].prio
		}
		return len(list[i].url) < len(list[j].url)
	})
	var pages []string
	perPrio := map[int]int{}
	for _, l := range list {
		if len(pages) >= 4 || perPrio[l.prio] >= 2 {
			continue
		}
		perPrio[l.prio]++
		pages = append(pages, l.url)
	}
	if len(links) == 0 || perPrio[0] == 0 {
		for _, p := range legalPaths {
			if len(pages) >= 4 {
				break
			}
			u, _ := final.Parse(p)
			pages = append(pages, u.String())
		}
	}
	out.read(q, final.String(), string(body), "company website", true)
	for _, p := range pages {
		if ctx.Err() != nil {
			break
		}
		b, fu, err := c.page(ctx, p, &budget)
		if err != nil || fu == nil || !sameSite(fu, final) {
			continue
		}
		legal := links[p] == 0 || !inLinks(links, p)
		what := "company legal notice"
		if !legal {
			what = "company website"
		}
		out.read(q, fu.String(), string(b), what, legal)
	}
	// Representatives from the legal notice win; other pages (about,
	// team, contact) are used only when the legal notice names nobody.
	people := out.legal
	if len(people) == 0 {
		people = out.other
	}
	for _, p := range people {
		out.addPerson(p)
	}
	if !out.Verified && out.mentioned && domainIsName(final, q) {
		out.Verified = true
	}
	if !out.Verified && !typed {
		if out.mentioned {
			out.Notes = append(out.Notes, "website "+final.Host+": names "+q.Name+" only in passing; not taken as its website")
		} else {
			out.Notes = append(out.Notes, "website "+final.Host+": no page names "+q.Name)
		}
		out.Result = Result{Notes: out.Notes}
		out.legal, out.other = nil, nil
	}
	return out
}

func inLinks(links map[string]int, p string) bool {
	_, ok := links[p]
	return ok
}

// read extracts everything from one page of the site. legal is set for
// the home page and legal-notice pages.
func (s *SiteResult) read(q Query, pageURL, body, what string, legal bool) {
	s.Pages++
	all := HTMLText(body)
	named := Mentions(normWords(strings.Join(all, "\n")), q)
	if named {
		s.mentioned = true
		if legal && operator(all, q) {
			s.Verified = true
		}
	}
	lines := companyLines(all, q)
	joined := strings.Join(lines, "\n")
	src := Source{ID: "website", Name: what + " " + pageURL, URL: pageURL, Kind: KindCompany}
	before := len(s.legal) + len(s.other) + len(s.Addresses)
	found := ExtractPeople(lines, src)
	if legal {
		s.legal = append(s.legal, found...)
	} else {
		s.other = append(s.other, found...)
	}
	if !named {
		// IDs, addresses and register entries only from pages that name
		// the company.
		return
	}
	if a := ExtractAddress(lines, q); a != "" {
		s.Addresses = append(s.Addresses, Address{Text: a, Source: src})
	}
	ln, la := structuredData(body)
	if ln != "" && (Mentions(normWords(ln), q) || SameEntity(q, ln)) {
		s.LegalNames = append(s.LegalNames, LegalName{Name: ln, Source: src})
		if la != "" {
			s.Addresses = append(s.Addresses, Address{Text: la, Source: src})
		}
	}
	for _, l := range lines {
		if len(l) < 120 && Mentions(normWords(l), Query{Name: q.Base, Base: q.Base}) {
			if f, ok := legalNameIn(l, q); ok {
				s.LegalNames = append(s.LegalNames, LegalName{Name: f, Source: src})
				break
			}
		}
	}
	s.IDs = appendIDs(s.IDs, ExtractIDs(joined))
	for _, r := range ExtractRegister(lines) {
		s.Register = append(s.Register, RegisterNote{Text: r, Source: src})
	}
	if len(s.legal)+len(s.other)+len(s.Addresses) > before || len(s.Excerpts) == 0 {
		s.Sources = append(s.Sources, src)
		s.Excerpts = append(s.Excerpts, Excerpt{Source: src, Text: excerpt(lines)})
	}
}

func appendIDs(a, b []ID) []ID {
	for _, x := range b {
		dup := false
		for _, y := range a {
			if x == y {
				dup = true
				break
			}
		}
		if !dup {
			a = append(a, x)
		}
	}
	return a
}

// Website reads the legal notice of a site the user typed.
func Website(ctx context.Context, c *Client, site string, q ...Query) Result {
	qq := Query{}
	if len(q) > 0 {
		qq = q[0]
	}
	r := ReadSite(ctx, c, site, qq, true)
	return r.Result
}

var (
	// Pages that carry the legal identity of a company, by priority.
	legalPages = []*regexp.Regexp{
		regexp.MustCompile(`(?i)impressum|imprint|legal[\s_-]*notice|aviso[\s_-]*legal|nota[\s_-]*legal|informaci[oó]n[\s_-]*legal|mentions[\s_-]*l[eé]gales|note[\s_-]*legali|colofon|offenlegung|legal[\s_-]*info|site[\s_-]*notice|corporate[\s_-]*information|aviso[\s_-]*de[\s_-]*privacidad|dados[\s_-]*da[\s_-]*empresa|/legal\b`),
		regexp.MustCompile(`(?i)about[\s_-]*us|[uü]ber[\s_-]*uns|ueber[\s_-]*uns|qui[eé]nes[\s_-]*somos|quem[\s_-]*somos|sobre[\s_-]*n[oó]s|sobre[\s_-]*nosotros|nosotros|qui[\s_-]*sommes|chi[\s_-]*siamo|unternehmen|empresa|company|management|leadership|f[uü]hrung|directorio|diretoria|board|team`),
		regexp.MustCompile(`(?i)kontakt|contact|contacto|contato|contatti`),
		regexp.MustCompile(`(?i)datenschutz|privacy|privacidad|privacidade|confidentialit[eé]|terms|t[eé]rminos|termos|condiciones|agb|disclaimer`),
	}
	// Paths tried when the home page links to no legal page.
	legalPaths = []string{"/impressum", "/imprint", "/aviso-legal", "/legal-notice", "/mentions-legales", "/legal"}
)
