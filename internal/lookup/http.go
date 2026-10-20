// Package lookup gathers candidate values for the form from official
// and public sources. Nothing found here is used without the user
// confirming it.
package lookup

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
)

// maxBody caps every response body held in memory. Nothing is written
// to disk, so lookups never create cache files.
const maxBody = 2 << 20

// UserAgent identifies the tool, as Wikimedia and GLEIF ask clients to do.
const UserAgent = "numb-form/1.0 (+https://github.com/z4dm42f5kp-ux/numb-form)"

// Client is a hardened HTTPS client.
type Client struct {
	hc *http.Client

	mu      sync.Mutex
	engines []string // search engines in order; nil means Engines
	browser *Browser // invisible browser for searches and script-built pages; nil means none
	setting string   // the browser setting browser was found for
}

// configure applies the search engines and browser of cfg.
func (c *Client) configure(cfg Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.engines = nil
	if f := strings.Fields(cfg.Search); len(f) > 0 && f[0] != "none" {
		c.engines = f
	}
	if c.setting != cfg.Browser || (c.browser == nil && cfg.Browser != "" && cfg.Browser != "none") {
		c.browser, c.setting = FindBrowser(cfg.Browser), cfg.Browser
	}
}

// BrowserName names the invisible browser in use, or "".
func (c *Client) BrowserName() string {
	if c.browser == nil {
		return ""
	}
	return c.browser.Name
}

func proxyConfigured() bool {
	for _, k := range []string{"HTTPS_PROXY", "https_proxy", "ALL_PROXY", "all_proxy"} {
		if os.Getenv(k) != "" {
			return true
		}
	}
	return false
}

func publicOnly(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("refusing to dial non-IP %q", host)
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast() || ip.IsInterfaceLocalMulticast() {
		return fmt.Errorf("refusing to connect to non-public address %s", ip)
	}
	return nil
}

// NewClient returns a client that only speaks HTTPS (TLS 1.2+), honours
// the system proxy settings, follows at most five HTTPS redirects and,
// when no proxy is configured, refuses to connect to private addresses.
func NewClient() *Client {
	d := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	if !proxyConfigured() {
		d.Control = publicOnly
	}
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           d.DialContext,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		MaxIdleConns:          8,
		IdleConnTimeout:       30 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	// Cookies are kept in memory for the run only (some search engines
	// set one before answering); they are never written to disk.
	jar, _ := cookiejar.New(nil)
	return &Client{hc: &http.Client{
		Jar:       jar,
		Transport: tr,
		Timeout:   25 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "https" {
				return errors.New("refusing non-HTTPS redirect")
			}
			return nil
		},
	}}
}

// StatusError is returned for HTTP responses other than 200 OK.
type StatusError struct {
	Host string
	Code int
}

func (e *StatusError) Error() string { return fmt.Sprintf("%s: HTTP %d", e.Host, e.Code) }

func (c *Client) get(ctx context.Context, u string, hdr map[string]string) ([]byte, *url.URL, string, error) {
	return c.do(ctx, http.MethodGet, u, "", hdr)
}

// postForm sends an HTML form, as a browser does when a search form is
// submitted.
func (c *Client) postForm(ctx context.Context, u string, form url.Values, hdr map[string]string) ([]byte, *url.URL, string, error) {
	h := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	for k, v := range hdr {
		h[k] = v
	}
	return c.do(ctx, http.MethodPost, u, form.Encode(), h)
}

// do performs one request. The body is returned even for status codes
// other than 200, together with a *StatusError.
func (c *Client) do(ctx context.Context, method, u, body string, hdr map[string]string) ([]byte, *url.URL, string, error) {
	pu, err := url.Parse(u)
	if err != nil {
		return nil, nil, "", err
	}
	if pu.Scheme != "https" {
		return nil, nil, "", fmt.Errorf("refusing non-HTTPS URL %s", u)
	}
	var rd io.Reader
	if body != "" {
		rd = strings.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rd)
	if err != nil {
		return nil, nil, "", err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept-Language", "en, de;q=0.9, es;q=0.8, pt;q=0.7, fr;q=0.6")
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := c.hc.Do(req) // #nosec G107 -- URL is validated above.
	if err != nil {
		return nil, nil, "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, nil, "", err
	}
	if len(b) > maxBody {
		b = b[:maxBody]
	}
	if resp.StatusCode != http.StatusOK {
		return b, resp.Request.URL, "", &StatusError{Host: pu.Host, Code: resp.StatusCode}
	}
	return b, resp.Request.URL, resp.Header.Get("Content-Type"), nil
}

func (c *Client) getJSON(ctx context.Context, u string, hdr map[string]string, v any) error {
	h := map[string]string{"Accept": "application/json"}
	for k, x := range hdr {
		h[k] = x
	}
	b, _, _, err := c.get(ctx, u, h)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("decode %s: %w", hostOf(u), err)
	}
	return nil
}

func hostOf(u string) string {
	if p, err := url.Parse(u); err == nil {
		return p.Host
	}
	return "?"
}

// sanitizeErr removes credentials and query strings from errors.
func sanitizeErr(err error) string {
	if err == nil {
		return ""
	}
	var ue *url.Error
	if errors.As(err, &ue) {
		if p, perr := url.Parse(ue.URL); perr == nil {
			p.User = nil
			p.RawQuery = ""
			return fmt.Sprintf("%s %s: %v", ue.Op, p.String(), ue.Err)
		}
	}
	return strings.ReplaceAll(err.Error(), "\n", " ")
}
