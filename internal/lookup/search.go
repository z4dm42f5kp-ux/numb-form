package lookup

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Hit is one search result.
type Hit struct {
	URL     string
	Title   string
	Snippet string
}

// ErrSearchBlocked means a search engine refused to answer (a human
// check, "unusual traffic", or a rate limit).
var ErrSearchBlocked = errors.New("the search engine asked for a human check")

// Engines lists the supported search engines in their default order.
var Engines = []string{"serper", "duckduckgo", "brave"}

// engine is one keyless search engine.
type engine struct {
	// fetch asks the engine and returns its result page.
	fetch func(ctx context.Context, c *Client, query, region string) ([]byte, error)
	// parse reads the results from a page.
	parse func(page string) ([]Hit, error)
	// page is the address of a result page, for the headless browser.
	page func(query, region string) string
}

var engines = map[string]engine{
	"serper":     {fetchSerper, parseSerper, serperPage},
	"duckduckgo": {fetchDDG, ParseDDG, ddgPage},
	"brave":      {fetchGet(bravePage), ParseBrave, bravePage},
}

// searchGap spaces out searches on one engine, as a person typing would;
// up to searchJitter is added at random. pauseAfterBlock rests an engine
// that refused.
var (
	searchGap       = 3 * time.Second
	searchJitter    = 2 * time.Second
	pauseAfterBlock = time.Minute
)

var (
	searchMu    sync.Mutex
	lastSearch  = map[string]time.Time{}
	pausedUntil = map[string]time.Time{}
)

func paused(name string) bool {
	searchMu.Lock()
	defer searchMu.Unlock()
	return time.Now().Before(pausedUntil[name])
}

func pause(name string) {
	searchMu.Lock()
	pausedUntil[name] = time.Now().Add(pauseAfterBlock)
	searchMu.Unlock()
}

// waitTurn waits until the engine may be asked again.
func waitTurn(ctx context.Context, name string) error {
	searchMu.Lock()
	gap := searchGap
	if searchJitter > 0 {
		gap += time.Duration(rand.Int63n(int64(searchJitter))) // #nosec G404 -- pacing, not security.
	}
	d := gap - time.Since(lastSearch[name])
	lastSearch[name] = time.Now().Add(max(d, 0))
	searchMu.Unlock()
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// refused reports whether an engine's answer means "not now".
func refused(err error) bool {
	var se *StatusError
	if errors.As(err, &se) {
		switch se.Code {
		case 202, 403, 429, 503, 999:
			return true
		}
	}
	return errors.Is(err, ErrSearchBlocked)
}

// blockedPage recognises human-check pages that come with status 200.
func blockedPage(page string) bool {
	low := strings.ToLower(page)
	for _, m := range []string{"captcha", "unusual traffic", "automated queries", "are you a robot", "not a robot", "anomaly-modal", "challenge-form", "/anomaly.js"} {
		if strings.Contains(low, m) {
			return true
		}
	}
	return false
}

// Search asks the configured engines in order (nil means Engines) and
// returns the results of the first one that answers with results. An
// engine that refuses is rested for a minute and the next one is asked.
// When every engine refuses and a browser is set on the client, the
// engines are asked once more through the invisible browser. It also
// returns the name of the engine that answered.
func Search(ctx context.Context, c *Client, query, region string) ([]Hit, string, error) {
	if region == "" {
		region = "wt-wt"
	}
	names := c.engines
	if names == nil {
		names = Engines
	}
	var errs []string
	allRefused := true
	for _, name := range names {
		e, ok := engines[name]
		if name == "serper" && c.serperKey == "" {
			continue // only with a key
		}
		if !ok || paused(name) {
			if ok {
				errs = append(errs, name+": resting after a refusal")
			}
			continue
		}
		if err := waitTurn(ctx, name); err != nil {
			return nil, "", err
		}
		page, err := e.fetch(ctx, c, query, region)
		var hits []Hit
		if err == nil {
			hits, err = e.parse(string(page))
			if err == nil && len(hits) == 0 && blockedPage(string(page)) {
				err = ErrSearchBlocked
			}
		}
		if err != nil {
			if refused(err) {
				pause(name)
			} else {
				allRefused = false
			}
			errs = append(errs, name+": "+sanitizeErr(err))
			continue
		}
		allRefused = false
		if len(hits) > 0 {
			return hits, name, nil
		}
	}
	if allRefused && c.browser != nil {
		for _, name := range names {
			e, ok := engines[name]
			if !ok {
				continue
			}
			target := e.page(query, region)
			if target == "" {
				continue
			}
			page, err := c.browser.Fetch(ctx, target)
			if err != nil {
				errs = append(errs, name+" in "+c.browser.Name+": "+sanitizeErr(err))
				continue
			}
			if hits, err := e.parse(string(page)); err == nil && len(hits) > 0 {
				return hits, name + " (through " + c.browser.Name + ")", nil
			}
		}
	}
	if len(errs) == 0 {
		return nil, "", nil // every engine answered: no results
	}
	return nil, "", fmt.Errorf("%s", strings.Join(errs, "; "))
}

// fetchGet asks an engine whose result page is a plain GET request.
func fetchGet(page func(query, region string) string) func(context.Context, *Client, string, string) ([]byte, error) {
	return func(ctx context.Context, c *Client, query, region string) ([]byte, error) {
		b, _, _, err := c.get(ctx, page(query, region), map[string]string{"Accept": "text/html"})
		return b, err
	}
}
