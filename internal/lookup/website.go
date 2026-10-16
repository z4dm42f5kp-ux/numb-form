package lookup

import (
	"context"
	"html"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

var (
	reScript  = regexp.MustCompile(`(?is)<(script|style|noscript|template|svg)\b.*?</(script|style|noscript|template|svg)\s*>`)
	reComment = regexp.MustCompile(`(?s)<!--.*?-->`)
	reBreak   = regexp.MustCompile(`(?i)<(br|/p|/div|/li|/h[1-6]|/tr|/td|/th|/dd|/dt|/address|/section|/article|hr)\b[^>]*>`)
	reTag     = regexp.MustCompile(`(?s)<[^>]*>`)
	reLink    = regexp.MustCompile(`(?is)<a\b[^>]*?href\s*=\s*["']([^"'#]+)["'][^>]*>(.*?)</a\s*>`)
	reLegal   = regexp.MustCompile(`(?i)impressum|imprint|legal[\s_-]*notice|aviso[\s_-]*legal|nota[\s_-]*legal|informaci[oó]n[\s_-]*legal|mentions[\s_-]*l[eé]gales|aviso[\s_-]*de[\s_-]*privacidad|legal[\s_-]*information|/legal\b`)
	rePostal  = regexp.MustCompile(`^(?:[A-Z]{1,2}-)?\d{4,5}\s+\p{Lu}`)
	reStreet  = regexp.MustCompile(`\p{L}.*\d|\d.*\p{L}`)
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

type roleRule struct {
	re   *regexp.Regexp
	role string
}

var roleRules = []roleRule{
	{regexp.MustCompile(`(?i)^(?:vertretungsberechtigte[rn]?\s+)?geschäftsführer(?:in|innen)?\b`), "Managing Director"},
	{regexp.MustCompile(`(?i)^geschäftsführung\b`), "Managing Director"},
	{regexp.MustCompile(`(?i)^vorstandsvorsitzende[rn]?\b|^vorsitzende[rn]?\s+des\s+vorstand(?:e)?s\b`), "Chairman of the Management Board"},
	{regexp.MustCompile(`(?i)^vorstand\b`), "Member of the Management Board"},
	{regexp.MustCompile(`(?i)^prokurist(?:in|en)?\b`), "Authorized Signatory (Prokurist)"},
	{regexp.MustCompile(`(?i)^inhaber(?:in)?\b`), "Owner"},
	{regexp.MustCompile(`(?i)^(?:gesetzlich\s+)?vertreten\s+durch\b`), "Legal Representative"},
	{regexp.MustCompile(`(?i)^managing\s+directors?\b`), "Managing Director"},
	{regexp.MustCompile(`(?i)^(?:legally\s+)?represented\s+by\b`), "Legal Representative"},
	{regexp.MustCompile(`(?i)^authori[sz]ed\s+representatives?\b`), "Authorized Representative"},
	{regexp.MustCompile(`(?i)^(?:ceo|chief\s+executive\s+officer)\b`), "Chief Executive Officer"},
	{regexp.MustCompile(`(?i)^representante\s+legal\b`), "Legal Representative"},
	{regexp.MustCompile(`(?i)^gerente\s+general\b`), "General Manager"},
	{regexp.MustCompile(`(?i)^administrador(?:a)?\s+[úu]nic[oa]\b`), "Sole Administrator"},
	{regexp.MustCompile(`(?i)^consejer[oa]\s+delegad[oa]\b`), "Chief Executive Officer (Managing Director)"},
	{regexp.MustCompile(`(?i)^director(?:a)?\s+general\b`), "General Director"},
	{regexp.MustCompile(`(?i)^gérant(?:e)?\b|^directeur\s+général\b|^président\b`), ""},
}

var leadWords = regexp.MustCompile(`(?i)^[\s:–—\-]*(?:(?:der|die|den|des|durch|by|the|el|la|los|las|sr\.?|sra\.?|herrn?|frau|mr\.?|mrs\.?|ms\.?)\s+)*`)
var splitNames = regexp.MustCompile(`(?i)\s*(?:,|;|/|&|\bund\b|\band\b|\by\b|\bsowie\b|\bet\b)\s*`)

var nameStop = map[string]bool{
	"gmbh": true, "ag": true, "kg": true, "ltd": true, "inc": true, "llc": true, "sa": true, "s.a.": true, "ltda": true,
	"straße": true, "strasse": true, "str.": true, "tel": true, "tel.": true, "telefon": true, "fax": true, "e-mail": true, "email": true,
	"registergericht": true, "amtsgericht": true, "handelsregister": true, "hrb": true, "ust-id": true, "ust-idnr.": true,
	"gesellschaft": true, "company": true, "street": true, "calle": true, "avenida": true, "www": true,
}

// looksLikeName accepts 2-5 capitalised words made of letters, hyphens,
// apostrophes and dots, such as "Dr. Erika Mustermann".
func looksLikeName(s string) bool {
	if len(s) > 60 {
		return false
	}
	words := strings.Fields(s)
	if len(words) < 2 || len(words) > 5 {
		return false
	}
	for _, w := range words {
		if nameStop[strings.ToLower(w)] {
			return false
		}
		r := []rune(w)
		if !unicode.IsUpper(r[0]) {
			// Allow particles such as "von", "de", "van", "del".
			switch strings.ToLower(w) {
			case "von", "van", "de", "del", "der", "da", "dos", "di", "zu", "la":
				continue
			}
			return false
		}
		for _, c := range r {
			if !(unicode.IsLetter(c) || c == '-' || c == '\'' || c == '.') {
				return false
			}
		}
	}
	return true
}

// ExtractPeople finds representatives named in legal-notice text.
func ExtractPeople(lines []string, src Source) []Person {
	var out []Person
	for i, l := range lines {
		for _, rule := range roleRules {
			loc := rule.re.FindStringIndex(l)
			if loc == nil {
				continue
			}
			role := rule.role
			matched := l[loc[0]:loc[1]]
			if role == "" {
				role, _ = TranslateRole(matched)
			}
			rest := l[loc[1]:]
			// "Vertreten durch den Geschäftsführer X": use the inner role.
			rest = leadWords.ReplaceAllString(rest, "")
			for _, inner := range roleRules {
				if m := inner.re.FindStringIndex(rest); m != nil && inner.role != "" {
					role = inner.role
					rest = leadWords.ReplaceAllString(rest[m[1]:], "")
					break
				}
			}
			rest = strings.Trim(rest, " :.-–—")
			if rest == "" && i+1 < len(lines) {
				rest = lines[i+1]
			}
			for _, n := range splitNames.Split(rest, -1) {
				n = strings.Trim(n, " .:")
				if looksLikeName(n) {
					out = append(out, Person{Name: n, Roles: []string{role}, Sources: []Source{src}})
				}
			}
			break
		}
	}
	return out
}

// ExtractAddress finds "Street 1" followed by "12345 City" in a legal notice.
func ExtractAddress(lines []string) string {
	for i := 1; i < len(lines) && i < 80; i++ {
		if rePostal.MatchString(lines[i]) && reStreet.MatchString(lines[i-1]) && len(lines[i-1]) < 80 && len(lines[i]) < 60 {
			return lines[i-1] + ", " + lines[i]
		}
	}
	return ""
}

func excerpt(lines []string) string {
	start := 0
	for i, l := range lines {
		if reLegal.MatchString(l) {
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

func sameSite(a, b *url.URL) bool {
	h := func(u *url.URL) string { return strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.") }
	x, y := h(a), h(b)
	return x == y || strings.HasSuffix(x, "."+y) || strings.HasSuffix(y, "."+x)
}

// Website reads the company's own legal notice: it loads the home page,
// follows links to Impressum / Aviso legal / Legal notice pages on the
// same site (at most three) and extracts representatives and an address.
func Website(ctx context.Context, c *Client, site string) Result {
	var res Result
	base, ok := NormalizeWebsite(site)
	if !ok {
		res.Notes = append(res.Notes, "website: not a usable HTTPS address")
		return res
	}
	body, final, _, err := c.get(ctx, base.String(), map[string]string{"Accept": "text/html"})
	if err != nil {
		res.Notes = append(res.Notes, "website: "+sanitizeErr(err))
		return res
	}
	if final == nil {
		final = base
	}
	var pages []string
	seen := map[string]bool{}
	for _, m := range reLink.FindAllStringSubmatch(string(body), -1) {
		href := html.UnescapeString(m[1])
		text := strings.Join(HTMLText(m[2]), " ")
		if !reLegal.MatchString(href) && !reLegal.MatchString(text) {
			continue
		}
		u, err := final.Parse(href)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || !sameSite(u, final) {
			continue
		}
		u.Scheme, u.Fragment = "https", ""
		if !seen[u.String()] && len(pages) < 3 {
			seen[u.String()] = true
			pages = append(pages, u.String())
		}
	}
	if len(pages) == 0 {
		for _, p := range []string{"/impressum", "/imprint", "/aviso-legal", "/legal-notice"} {
			u, _ := final.Parse(p)
			pages = append(pages, u.String())
		}
	}
	found := 0
	for _, p := range pages {
		if found >= 2 {
			break
		}
		b, fu, _, err := c.get(ctx, p, map[string]string{"Accept": "text/html"})
		if err != nil {
			continue
		}
		lines := HTMLText(string(b))
		joined := strings.Join(lines, "\n")
		if !reLegal.MatchString(joined) {
			continue
		}
		found++
		page := p
		if fu != nil {
			page = fu.String()
		}
		src := Source{Name: "company legal notice " + page, URL: page, Kind: KindCompany}
		res.Sources = append(res.Sources, src)
		for _, person := range ExtractPeople(lines, src) {
			res.addPerson(person)
		}
		if a := ExtractAddress(lines); a != "" {
			res.Addresses = append(res.Addresses, Address{Text: a, Source: src})
		}
		res.Excerpts = append(res.Excerpts, Excerpt{Source: src, Text: excerpt(lines)})
	}
	if found == 0 {
		res.Notes = append(res.Notes, "website: no legal notice page found on "+final.Host)
	}
	return res
}
