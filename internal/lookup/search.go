package lookup

import (
	"context"
	"errors"
	"html"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// DuckDuckGo's HTML-only pages, which need no JavaScript and no account.
var (
	ddgHTML = "https://html.duckduckgo.com/html/"
	ddgLite = "https://lite.duckduckgo.com/lite/"
)

// Hit is one search result.
type Hit struct {
	URL     string
	Title   string
	Snippet string
}

// ErrSearchBlocked means the search engine asked for a human check.
var ErrSearchBlocked = errors.New("DuckDuckGo asked for a human check (too many searches); search skipped")

// searchGap spaces out searches, as a person typing would.
var searchGap = 1500 * time.Millisecond

var (
	searchMu   sync.Mutex
	lastSearch time.Time
)

func waitTurn(ctx context.Context) error {
	searchMu.Lock()
	defer searchMu.Unlock()
	if d := searchGap - time.Since(lastSearch); d > 0 {
		t := time.NewTimer(d)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
	lastSearch = time.Now()
	return nil
}

var (
	reResultBlock = regexp.MustCompile(`(?is)<div[^>]+class="[^"]*\bresult\b[^"]*"[^>]*>`)
	reAnchor      = regexp.MustCompile(`(?is)<a\b([^>]*)>(.*?)</a\s*>`)
	reAttrHref    = regexp.MustCompile(`(?is)\bhref\s*=\s*["']([^"']+)["']`)
	reAttrClass   = regexp.MustCompile(`(?is)\bclass\s*=\s*["']([^"']*)["']`)
	reLiteSnippet = regexp.MustCompile(`(?is)<td[^>]+class=["'][^"']*result-snippet[^"']*["'][^>]*>(.*?)</td>`)
)

// resultURL turns a result link into the target address: DuckDuckGo
// links either directly or through /l/?uddg=<target>.
func resultURL(href string) string {
	href = strings.TrimSpace(html.UnescapeString(href))
	if strings.HasPrefix(href, "//") {
		href = "https:" + href
	}
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	if strings.HasSuffix(u.Hostname(), "duckduckgo.com") {
		if t := u.Query().Get("uddg"); t != "" && strings.HasPrefix(u.Path, "/l/") {
			return resultURL(t)
		}
		return "" // ads (y.js) and internal links
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return ""
	}
	return u.String()
}

func plain(s string) string { return strings.Join(HTMLText(s), " ") }

// ParseDDG reads results from DuckDuckGo's HTML or lite result page.
func ParseDDG(page string) ([]Hit, error) {
	low := strings.ToLower(page)
	if strings.Contains(low, "anomaly-modal") || strings.Contains(low, "challenge-form") || strings.Contains(low, "/anomaly.js") {
		return nil, ErrSearchBlocked
	}
	var hits []Hit
	seen := map[string]bool{}
	add := func(h Hit) {
		if h.URL == "" || seen[h.URL] {
			return
		}
		seen[h.URL] = true
		hits = append(hits, h)
	}
	// HTML version: one <div class="result ..."> per result.
	locs := reResultBlock.FindAllStringIndex(page, -1)
	for i, loc := range locs {
		end := len(page)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		head := page[loc[0]:loc[1]]
		if strings.Contains(head, "result--ad") {
			continue
		}
		var h Hit
		for _, a := range reAnchor.FindAllStringSubmatch(page[loc[0]:end], -1) {
			cls := ""
			if m := reAttrClass.FindStringSubmatch(a[1]); m != nil {
				cls = m[1]
			}
			href := ""
			if m := reAttrHref.FindStringSubmatch(a[1]); m != nil {
				href = m[1]
			}
			switch {
			case strings.Contains(cls, "result__a"):
				h.URL, h.Title = resultURL(href), plain(a[2])
			case strings.Contains(cls, "result__snippet"):
				h.Snippet = plain(a[2])
				if h.URL == "" {
					h.URL = resultURL(href)
				}
			}
		}
		add(h)
	}
	if len(hits) > 0 {
		return hits, nil
	}
	// Lite version: <a class="result-link"> followed by a snippet cell.
	anchors := reAnchor.FindAllStringSubmatchIndex(page, -1)
	for i, a := range anchors {
		attrs := page[a[2]:a[3]]
		m := reAttrClass.FindStringSubmatch(attrs)
		if m == nil || !strings.Contains(m[1], "result-link") {
			continue
		}
		hm := reAttrHref.FindStringSubmatch(attrs)
		if hm == nil {
			continue
		}
		h := Hit{URL: resultURL(hm[1]), Title: plain(page[a[4]:a[5]])}
		end := len(page)
		if i+1 < len(anchors) {
			end = anchors[i+1][0]
		}
		if sm := reLiteSnippet.FindStringSubmatch(page[a[1]:end]); sm != nil {
			h.Snippet = plain(sm[1])
		}
		add(h)
	}
	return hits, nil
}

// Search asks DuckDuckGo. region is a DuckDuckGo region such as "de-de"
// or "wt-wt" (no region). It tries the HTML page first and the lite page
// if that fails.
func Search(ctx context.Context, c *Client, query, region string) ([]Hit, error) {
	if region == "" {
		region = "wt-wt"
	}
	var firstErr error
	for _, endpoint := range []string{ddgHTML, ddgLite} {
		if err := waitTurn(ctx); err != nil {
			return nil, err
		}
		body, _, _, err := c.postForm(ctx, endpoint, url.Values{"q": {query}, "kl": {region}, "b": {""}}, map[string]string{"Accept": "text/html", "Referer": endpoint})
		var se *StatusError
		if err != nil && !(errors.As(err, &se) && se.Code == 202) {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		hits, perr := ParseDDG(string(body))
		if perr == nil && err == nil {
			return hits, nil
		}
		if firstErr == nil {
			firstErr = perr
			if firstErr == nil {
				firstErr = ErrSearchBlocked
			}
		}
	}
	return nil, firstErr
}

// ddgRegion picks a DuckDuckGo region for a country code.
func ddgRegion(country string) string {
	r := map[string]string{
		"DE": "de-de", "AT": "at-de", "CH": "ch-de", "ES": "es-es", "PT": "pt-pt", "BR": "br-pt", "MX": "mx-es",
		"AR": "ar-es", "CO": "co-es", "CL": "cl-es", "PE": "pe-es", "VE": "ve-es", "US": "us-en", "GB": "uk-en",
		"IE": "ie-en", "ZA": "za-en", "FR": "fr-fr", "BE": "be-fr", "NL": "nl-nl", "IT": "it-it", "PL": "pl-pl",
		"IN": "in-en", "CA": "ca-en", "AU": "au-en", "NZ": "nz-en", "SG": "sg-en", "MY": "my-en",
	}[country]
	if r == "" {
		return "wt-wt"
	}
	return r
}
