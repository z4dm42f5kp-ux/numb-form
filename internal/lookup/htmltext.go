package lookup

import (
	"net/url"
	"strings"
)

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
