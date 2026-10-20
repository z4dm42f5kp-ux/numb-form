package lookup

import (
	"context"
	"sync"
	"time"
)

// Config holds optional free API keys and the enabled sources.
type Config struct {
	CompaniesHouseKey string
	// Sources lists the enabled sources ("companies-house", "vies",
	// "brasil-cnpj", "official", "gleif", "website", "search",
	// "wikidata"); nil enables all.
	Sources []string
	// Search lists the search engines in order of preference
	// ("duckduckgo mojeek yahoo brave", the default when empty), or
	// "none".
	Search string
	// Browser is "auto" (Google Chrome, else Chromium or Edge), "none",
	// or the path of a Chromium-based browser, started invisibly.
	Browser string
	// GuessDomains also tries addresses such as companyname.de.
	GuessDomains bool
}

func (c Config) enabled(s string) bool {
	if c.Sources == nil {
		return true
	}
	for _, x := range c.Sources {
		if x == s {
			return true
		}
	}
	return false
}

// Registries queries every configured register concurrently.
func Registries(ctx context.Context, c *Client, cfg Config, q Query) Result {
	type job struct {
		name string
		fn   func(context.Context) (Result, error)
	}
	var jobs []job
	if cfg.enabled("gleif") {
		jobs = append(jobs, job{"GLEIF", func(ctx context.Context) (Result, error) { return gleif(ctx, c, q) }})
	}
	if cfg.enabled("wikidata") {
		jobs = append(jobs, job{"Wikidata", func(ctx context.Context) (Result, error) { return wikidata(ctx, c, q) }})
	}
	if cfg.CompaniesHouseKey != "" && cfg.enabled("companies-house") {
		jobs = append(jobs, job{"Companies House", func(ctx context.Context) (Result, error) {
			return companiesHouse(ctx, c, cfg.CompaniesHouseKey, q)
		}})
	}
	results := make([]Result, len(jobs))
	var wg sync.WaitGroup
	for i := range jobs {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			jctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			res, err := jobs[i].fn(jctx)
			if err != nil {
				res.Notes = append(res.Notes, jobs[i].name+": "+sanitizeErr(err))
			}
			results[i] = res
		}()
	}
	wg.Wait()
	var all Result
	for _, r := range results {
		all.merge(r)
	}
	all.sortResult()
	return all
}
