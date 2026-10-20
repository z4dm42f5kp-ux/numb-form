package lookup

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// reScriptApp recognises pages whose content is built by JavaScript.
var reScriptApp = regexp.MustCompile(`(?i)<div[^>]+id=["'](?:root|app|__next|__nuxt|___gatsby)["'][^>]*>\s*</div>|enable javascript|javascript (?:is )?(?:required|disabled)|javascript aktivieren|habilite (?:o )?javascript|ative o javascript|activez javascript`)

// thin reports whether a page shows too little text to be read without
// running its scripts.
func thin(body []byte) bool {
	text := strings.Join(HTMLText(string(body)), " ")
	return len(text) < 400 || (len(text) < 3000 && reScriptApp.Match(body))
}

// page loads a website page. When the page is built by JavaScript, or the
// site turns away programs (HTTP 403, 429, 503), and an invisible browser
// is configured, the page is loaded in the browser instead, at most
// *budget times per site. Sites that do not exist (DNS or connection
// errors) are never handed to the browser.
func (c *Client) page(ctx context.Context, u string, budget *int) ([]byte, *url.URL, error) {
	body, final, _, err := c.get(ctx, u, map[string]string{"Accept": "text/html"})
	var se *StatusError
	refusedProgram := errors.As(err, &se) && (se.Code == 403 || se.Code == 429 || se.Code == 503)
	if c.browser == nil || *budget <= 0 || (err != nil && !refusedProgram) || (err == nil && !thin(body)) {
		return body, final, err
	}
	*budget--
	target := u
	if final != nil {
		target = final.String()
	}
	rendered, berr := c.browser.Fetch(ctx, target)
	if berr != nil || len(HTMLText(string(rendered))) <= len(HTMLText(string(body))) {
		return body, final, err
	}
	if final == nil {
		final, _ = url.Parse(target)
	}
	return rendered, final, nil
}
