package lookup

import (
	"net/url"
	"regexp"
	"strings"
)

// Brave Search's result page, asked when DuckDuckGo refuses.
var braveURL = "https://search.brave.com/search"

func bravePage(query, _ string) string {
	return braveURL + "?" + url.Values{"q": {query}, "source": {"web"}}.Encode()
}

var (
	reBraveBlock  = regexp.MustCompile(`(?is)<div[^>]+data-type="web"[^>]*>`)
	reBraveTitle  = regexp.MustCompile(`(?is)class="[^"]*\b(?:title|snippet-title)\b[^"]*"[^>]*>(.*?)</`)
	reBraveSnip   = regexp.MustCompile(`(?is)class="[^"]*(?:snippet-description|generic-snippet|\bcontent\b)[^"]*"[^>]*>(.*?)</div>`)
	reAnchorHref  = regexp.MustCompile(`(?is)<a\b[^>]*?\bhref\s*=\s*["']([^"']+)["']`)
	reEngineHosts = regexp.MustCompile(`(?i)(^|\.)(bing\.com|brave\.com|duckduckgo\.com|google\.com|serper\.dev)$`)
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

// targetURL returns a result's real address; engine-internal and ad
// links give "".
func targetURL(href string) string {
	href = strings.TrimSpace(strings.ReplaceAll(href, "&amp;", "&"))
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
