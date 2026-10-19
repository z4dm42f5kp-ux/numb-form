package lookup

import (
	"context"
	"encoding/json"
	"html"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

var (
	reScript  = regexp.MustCompile(`(?is)<(script|style|noscript|template|svg)\b.*?</(script|style|noscript|template|svg)\s*>`)
	reLDJSON  = regexp.MustCompile(`(?is)<script[^>]+type\s*=\s*["']application/ld\+json["'][^>]*>(.*?)</script\s*>`)
	reComment = regexp.MustCompile(`(?s)<!--.*?-->`)
	reBreak   = regexp.MustCompile(`(?i)<(br|/p|/div|/li|/h[1-6]|/tr|/td|/th|/dd|/dt|/address|/section|/article|/footer|/header|hr|p|div|li|h[1-6]|tr|td|th|dd|dt|address|section|article|footer)\b[^>]*>`)
	reTag     = regexp.MustCompile(`(?s)<[^>]*>`)
	reLink    = regexp.MustCompile(`(?is)<a\b[^>]*?href\s*=\s*["']([^"'#]+)["'][^>]*>(.*?)</a\s*>`)

	// Pages that carry the legal identity of a company, by priority.
	legalPages = []*regexp.Regexp{
		regexp.MustCompile(`(?i)impressum|imprint|legal[\s_-]*notice|aviso[\s_-]*legal|nota[\s_-]*legal|informaci[oó]n[\s_-]*legal|mentions[\s_-]*l[eé]gales|note[\s_-]*legali|colofon|offenlegung|legal[\s_-]*info|site[\s_-]*notice|corporate[\s_-]*information|aviso[\s_-]*de[\s_-]*privacidad|dados[\s_-]*da[\s_-]*empresa|/legal\b`),
		regexp.MustCompile(`(?i)about[\s_-]*us|[uü]ber[\s_-]*uns|ueber[\s_-]*uns|qui[eé]nes[\s_-]*somos|quem[\s_-]*somos|sobre[\s_-]*n[oó]s|sobre[\s_-]*nosotros|nosotros|qui[\s_-]*sommes|chi[\s_-]*siamo|unternehmen|empresa|company|management|leadership|f[uü]hrung|directorio|diretoria|board|team`),
		regexp.MustCompile(`(?i)kontakt|contact|contacto|contato|contatti`),
		regexp.MustCompile(`(?i)datenschutz|privacy|privacidad|privacidade|confidentialit[eé]|terms|t[eé]rminos|termos|condiciones|agb|disclaimer`),
	}
	// Paths tried when the home page links to no legal page.
	legalPaths = []string{"/impressum", "/imprint", "/aviso-legal", "/legal-notice", "/mentions-legales", "/legal"}

	rePostal    = regexp.MustCompile(`(?:^|[\s,])(?:[A-Z]{1,2}-)?\d{4,5}(?:-\d{3,4})?\s+\p{Lu}|(?i:\b(?:c\.?\s?p\.?|cep|zip|postal\s+code|plz)\s*:?\s*\d{4,5})|\b\d{4}\s?[A-Z]{2}\b|\b[A-Z]{1,2}\d[A-Z\d]?\s?\d[A-Z]{2}\b|\b[A-Z]{2}\s+\d{5}(?:-\d{4})?\b`)
	reStreet    = regexp.MustCompile(`(?i)\p{L}.*\d|\d.*\p{L}|\b(?:stra(?:ss|ß)e|str\.|weg|platz|allee|gasse|ring|damm|calle|avenida|av\.|carrera|cra\.|rua|travessa|rodovia|road|street|st\.|avenue|ave\.|blvd|boulevard|suite|piso|of\.|oficina|col\.|colonia|bairro|via|viale|rue|edificio|torre)\b`)
	reAddrLabel = regexp.MustCompile(`(?i)^\s*(?:anschrift|adresse|address|postanschrift|sitz(?:\s+der\s+gesellschaft)?|gesch[aä]ftssitz|firmensitz|registered\s+office|head\s+office|headquarters|domicilio(?:\s+social|\s+fiscal)?|direcci[oó]n(?:\s+fiscal)?|sede(?:\s+social)?|endere[cç]o|morada|si[eè]ge(?:\s+social)?|indirizzo|sede\s+legale)\s*[:：]\s*(.*)$`)
	reNoAddr    = regexp.MustCompile(`(?i)\b(?:tel|telefon|telephone|phone|fon|fax|e-?mail|www\.|https?://|@|ust|uid|vat|iva|cif|nif|nit|rfc|cnpj|hrb|hra|amtsgericht|registergericht|handelsregister|gesch[aä]ftsf[uü]hrer|managing|director|copyright|©)\b`)
)

// HTMLText converts HTML to plain lines.
func HTMLText(s string) []string {
	s = reScript.ReplaceAllString(s, " ")
	s = reComment.ReplaceAllString(s, " ")
	s = reBreak.ReplaceAllString(s, "\n")
	s = reTag.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	var out []string
	for _, l := range strings.Split(s, "\n") {
		l = textutil.OneLine(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

func addressLine(l string) bool {
	return len(l) < 90 && !reNoAddr.MatchString(l) && (reStreet.MatchString(l) || rePostal.MatchString(l))
}

// ExtractAddress finds the company's address in a legal notice: after a
// label such as "Anschrift:" or "Domicilio:", or in the lines that follow
// the company name, or as "Street 1" followed by "12345 City".
func ExtractAddress(lines []string, q ...Query) string {
	for i, l := range lines {
		m := reAddrLabel.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		parts := []string{}
		if v := strings.TrimSpace(m[1]); v != "" {
			parts = append(parts, v)
		}
		for j := i + 1; j < len(lines) && j <= i+3 && len(parts) < 3; j++ {
			if !addressLine(lines[j]) {
				break
			}
			parts = append(parts, lines[j])
		}
		a := strings.Join(parts, ", ")
		if rePostal.MatchString(a) || (len(parts) > 0 && reStreet.MatchString(parts[0]) && strings.ContainsAny(a, "0123456789")) {
			return a
		}
	}
	if len(q) > 0 {
		for i, l := range lines {
			if len(l) > 120 || !Mentions(normWords(l), q[0]) {
				continue
			}
			var parts []string
			for j := i + 1; j < len(lines) && j <= i+4 && len(parts) < 3; j++ {
				if !addressLine(lines[j]) {
					break
				}
				parts = append(parts, lines[j])
			}
			if len(parts) >= 1 && rePostal.MatchString(strings.Join(parts, " ")) {
				return strings.Join(parts, ", ")
			}
		}
	}
	for i := 1; i < len(lines) && i < 200; i++ {
		if rePostal.MatchString(lines[i]) && reStreet.MatchString(lines[i-1]) && addressLine(lines[i-1]) && len(lines[i]) < 60 {
			return lines[i-1] + ", " + lines[i]
		}
	}
	return ""
}

// ldOrg is the part of schema.org Organization data that is used.
type ldOrg struct {
	Type      any               `json:"@type"`
	LegalName string            `json:"legalName"`
	Name      string            `json:"name"`
	Address   any               `json:"address"`
	Graph     []json.RawMessage `json:"@graph"`
}

func ldAddress(a any) string {
	switch v := a.(type) {
	case string:
		return textutil.OneLine(v)
	case map[string]any:
		get := func(k string) string {
			switch x := v[k].(type) {
			case string:
				return textutil.OneLine(x)
			case map[string]any:
				if n, ok := x["name"].(string); ok {
					return textutil.OneLine(n)
				}
			}
			return ""
		}
		var parts []string
		if s := get("streetAddress"); s != "" {
			parts = append(parts, s)
		}
		if pc := strings.TrimSpace(get("postalCode") + " " + get("addressLocality")); pc != "" {
			parts = append(parts, pc)
		}
		if c := get("addressCountry"); c != "" {
			parts = append(parts, CountryName(c))
		}
		if len(parts) >= 2 {
			return strings.Join(parts, ", ")
		}
	case []any:
		if len(v) > 0 {
			return ldAddress(v[0])
		}
	}
	return ""
}

func typeString(t any) string {
	switch v := t.(type) {
	case string:
		return v
	case []any:
		var parts []string
		for _, x := range v {
			if s, ok := x.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, " ")
	}
	return ""
}

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

// structuredData reads schema.org Organization data embedded in a page.
func structuredData(page string) (legalName, address string) {
	var walk func(b []byte, depth int)
	walk = func(b []byte, depth int) {
		if depth > 3 || (legalName != "" && address != "") {
			return
		}
		var list []json.RawMessage
		if json.Unmarshal(b, &list) == nil {
			for _, x := range list {
				walk(x, depth+1)
			}
			return
		}
		var o ldOrg
		if json.Unmarshal(b, &o) != nil {
			return
		}
		for _, g := range o.Graph {
			walk(g, depth+1)
		}
		t := strings.ToLower(typeString(o.Type))
		if !strings.Contains(t, "organization") && !strings.Contains(t, "corporation") && !strings.Contains(t, "localbusiness") {
			return
		}
		if legalName == "" {
			legalName = textutil.OneLine(o.LegalName)
		}
		if address == "" {
			address = ldAddress(o.Address)
		}
	}
	for _, m := range reLDJSON.FindAllStringSubmatch(page, 5) {
		walk([]byte(m[1]), 0)
	}
	return legalName, address
}

// SiteResult is what a company website says about the company.
type SiteResult struct {
	Result
	Home     string // the home page that was read
	Verified bool   // a page names the company
	Pages    int
}

type pageLink struct {
	url  string
	prio int
}

// ReadSite reads a company website: the home page and up to four pages
// that carry its legal identity (legal notice, about, contact, privacy).
// Verified is set when one of them names the company.
func ReadSite(ctx context.Context, c *Client, site string, q Query) SiteResult {
	var out SiteResult
	base, ok := NormalizeWebsite(site)
	if !ok {
		out.Notes = append(out.Notes, "website: not a usable address: "+site)
		return out
	}
	body, final, _, err := c.get(ctx, base.String(), map[string]string{"Accept": "text/html"})
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
	out.read(q, final.String(), string(body), "company website")
	for _, p := range pages {
		if ctx.Err() != nil {
			break
		}
		b, fu, _, err := c.get(ctx, p, map[string]string{"Accept": "text/html"})
		if err != nil || fu == nil || !sameSite(fu, final) {
			continue
		}
		out.read(q, fu.String(), string(b), "company legal notice")
	}
	if !out.Verified {
		out.Notes = append(out.Notes, "website "+final.Host+": no page names "+q.Name)
	}
	return out
}

// read extracts everything from one page of the site.
func (s *SiteResult) read(q Query, pageURL, body, what string) {
	s.Pages++
	lines := HTMLText(body)
	joined := strings.Join(lines, "\n")
	norm := normWords(joined)
	named := Mentions(norm, q)
	if named {
		s.Verified = true
	}
	src := Source{ID: "website", Name: what + " " + pageURL, URL: pageURL, Kind: KindCompany}
	before := len(s.People) + len(s.Addresses)
	for _, p := range ExtractPeople(lines, src) {
		s.addPerson(p)
	}
	if a := ExtractAddress(lines, q); a != "" && named {
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
	if len(s.People)+len(s.Addresses) > before || (named && len(s.Excerpts) == 0) {
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

func excerpt(lines []string) string {
	start := 0
	for i, l := range lines {
		if legalPages[0].MatchString(l) {
			start = i
			break
		}
	}
	end := min(len(lines), start+30)
	text := strings.Join(lines[start:end], "\n")
	if len(text) > 2000 {
		text = text[:2000] + " …"
	}
	return text
}

// NormalizeWebsite turns user input such as "example.com" into an HTTPS URL.
func NormalizeWebsite(s string) (*url.URL, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return nil, false
	}
	if u.Scheme == "http" {
		u.Scheme = "https"
	}
	if u.Scheme != "https" {
		return nil, false
	}
	u.User = nil
	return u, true
}

func hostKey(u *url.URL) string { return strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.") }

func sameSite(a, b *url.URL) bool {
	x, y := hostKey(a), hostKey(b)
	return x == y || strings.HasSuffix(x, "."+y) || strings.HasSuffix(y, "."+x)
}

// Website reads the legal notice of a site the user typed.
func Website(ctx context.Context, c *Client, site string, q ...Query) Result {
	qq := Query{}
	if len(q) > 0 {
		qq = q[0]
	}
	r := ReadSite(ctx, c, site, qq)
	return r.Result
}
