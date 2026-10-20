package lookup

import (
	"net/url"
	"regexp"
	"strings"
)

// Result pages of the other keyless engines, used in turn when
// DuckDuckGo refuses.
var (
	mojeekURL = "https://www.mojeek.com/search"
	yahooURL  = "https://search.yahoo.com/search"
	braveURL  = "https://search.brave.com/search"
)

func mojeekPage(query, _ string) string { return mojeekURL + "?" + url.Values{"q": {query}}.Encode() }
func yahooPage(query, _ string) string  { return yahooURL + "?" + url.Values{"p": {query}}.Encode() }
func bravePage(query, _ string) string {
	return braveURL + "?" + url.Values{"q": {query}, "source": {"web"}}.Encode()
}

var (
	reMojeekList  = regexp.MustCompile(`(?is)<ul[^>]+class="[^"]*\bresults[^"]*"[^>]*>(.*?)</ul>`)
	reH2Anchor    = regexp.MustCompile(`(?is)<h2[^>]*>.*?<a\b([^>]*)>(.*?)</a\s*>`)
	reMojeekSnip  = regexp.MustCompile(`(?is)<p[^>]+class="s"[^>]*>(.*?)</p>`)
	reYahooBlock  = regexp.MustCompile(`(?is)<div[^>]+class="[^"]*\b(?:relsrch|algo)\b[^"]*"[^>]*>`)
	reH3          = regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>`)
	reYahooSnip   = regexp.MustCompile(`(?is)<(?:div|p|span)[^>]+class="[^"]*(?:compText|\bfc-\w+ lh-\d+|s-desc)[^"]*"[^>]*>(.*?)</(?:div|p|span)>`)
	reBraveBlock  = regexp.MustCompile(`(?is)<div[^>]+data-type="web"[^>]*>`)
	reBraveTitle  = regexp.MustCompile(`(?is)class="[^"]*\b(?:title|snippet-title)\b[^"]*"[^>]*>(.*?)</`)
	reBraveSnip   = regexp.MustCompile(`(?is)class="[^"]*(?:snippet-description|generic-snippet|\bcontent\b)[^"]*"[^>]*>(.*?)</div>`)
	reAnchorHref  = regexp.MustCompile(`(?is)<a\b[^>]*?\bhref\s*=\s*["']([^"']+)["']`)
	reEngineHosts = regexp.MustCompile(`(?i)(^|\.)(mojeek\.com|yahoo\.com|yimg\.com|bing\.com|brave\.com|duckduckgo\.com)$`)
)

// blocks cuts a page at each match of start.
func blocks(page string, start *regexp.Regexp) []string {
	locs := start.FindAllStringIndex(page, -1)
	out := make([]string, 0, len(locs))
	for i, l := range locs {
		end := len(page)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		out = append(out, page[l[0]:end])
	}
	return out
}

// targetURL returns a result's real address: Yahoo wraps it as
// .../RU=<address>/RK=...; engine-internal and ad links give "".
func targetURL(href string) string {
	href = strings.TrimSpace(strings.ReplaceAll(href, "&amp;", "&"))
	if i := strings.Index(href, "/RU="); i >= 0 {
		t := href[i+4:]
		for _, cut := range []string{"/RK=", "/RS="} {
			if j := strings.Index(t, cut); j >= 0 {
				t = t[:j]
			}
		}
		if u, err := url.QueryUnescape(t); err == nil {
			href = u
		}
	}
	if strings.HasPrefix(href, "//") {
		href = "https:" + href
	}
	u, err := url.Parse(href)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || reEngineHosts.MatchString(u.Hostname()) {
		return ""
	}
	return u.String()
}

func firstSub(re *regexp.Regexp, s string) string {
	if m := re.FindStringSubmatch(s); m != nil {
		return plain(m[1])
	}
	return ""
}

func addHit(hits []Hit, seen map[string]bool, h Hit) []Hit {
	if h.URL == "" || seen[h.URL] {
		return hits
	}
	seen[h.URL] = true
	return append(hits, h)
}

// ParseMojeek reads Mojeek's result page.
func ParseMojeek(page string) ([]Hit, error) {
	var hits []Hit
	seen := map[string]bool{}
	for _, list := range reMojeekList.FindAllStringSubmatch(page, -1) {
		for _, li := range strings.Split(list[1], "<li")[1:] {
			m := reH2Anchor.FindStringSubmatch(li)
			if m == nil {
				continue
			}
			href := reAttrHref.FindStringSubmatch(m[1])
			if href == nil {
				continue
			}
			hits = addHit(hits, seen, Hit{URL: targetURL(href[1]), Title: plain(m[2]), Snippet: firstSub(reMojeekSnip, li)})
		}
	}
	return hits, nil
}

// ParseYahoo reads Yahoo's result page (results from Bing).
func ParseYahoo(page string) ([]Hit, error) {
	var hits []Hit
	seen := map[string]bool{}
	for _, b := range blocks(page, reYahooBlock) {
		var link string
		for _, m := range reAnchorHref.FindAllStringSubmatch(b, 6) {
			if link = targetURL(m[1]); link != "" {
				break
			}
		}
		hits = addHit(hits, seen, Hit{URL: link, Title: firstSub(reH3, b), Snippet: firstSub(reYahooSnip, b)})
	}
	return hits, nil
}

// ParseBrave reads Brave Search's result page.
func ParseBrave(page string) ([]Hit, error) {
	var hits []Hit
	seen := map[string]bool{}
	for _, b := range blocks(page, reBraveBlock) {
		var link string
		for _, m := range reAnchorHref.FindAllStringSubmatch(b, 6) {
			if link = targetURL(m[1]); link != "" {
				break
			}
		}
		hits = addHit(hits, seen, Hit{URL: link, Title: firstSub(reBraveTitle, b), Snippet: firstSub(reBraveSnip, b)})
	}
	return hits, nil
}
