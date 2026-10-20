package lookup

import (
	"context"
	"net/url"
	"sync"
)

// search runs one web search and records problems as notes.
func search(ctx context.Context, c *Client, q, region string, notes *[]string) []Hit {
	hits, _, err := Search(ctx, c, q, region)
	if err != nil {
		*notes = append(*notes, "web search: "+sanitizeErr(err))
	}
	return hits
}

// Find gathers everything about one company in one pass: registers,
// the company's own website (found automatically), official IDs printed
// there, government gazettes and, as a last resort, search snippets.
// site, when not empty, is a website the user typed; it is read instead
// of searching for one.
func Find(ctx context.Context, c *Client, cfg Config, name, site string) Result {
	q := NewQuery(name)
	country := ""
	if len(q.Countries) > 0 {
		country = q.Countries[0]
	}
	region := ddgRegion(country)
	c.configure(cfg)
	useSearch := cfg.Search != "none" && cfg.enabled("search")

	// 1. Registers and the first searches, at the same time.
	var (
		reg          Result
		hits1, hits2 []Hit
		searchNotes  []string
		wg           sync.WaitGroup
		mu           sync.Mutex
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		reg = Registries(ctx, c, cfg, q)
	}()
	if useSearch && site == "" && cfg.enabled("website") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var n []string
			h1 := search(ctx, c, `"`+q.Name+`"`, region, &n)
			var h2 []Hit
			if !hasCompanySite(h1) {
				h2 = search(ctx, c, q.Name+" "+legalWords(country), region, &n)
			}
			mu.Lock()
			hits1, hits2, searchNotes = h1, h2, append(searchNotes, n...)
			mu.Unlock()
		}()
	}
	wg.Wait()
	all := reg
	all.Notes = append(all.Notes, searchNotes...)
	hits := append(append([]Hit(nil), hits1...), hits2...)

	// 2. The company's website.
	if cfg.enabled("website") {
		var cands []string
		if site != "" {
			cands = []string{site}
		} else {
			cands = append(cands, reg.Websites...)
			seen := map[string]bool{}
			for _, w := range reg.Websites {
				if u, ok := NormalizeWebsite(w); ok {
					seen[hostKey(u)] = true
				}
			}
			for _, h := range hits {
				u, err := url.Parse(h.URL)
				if err != nil || notCompanySite(u.Hostname()) || seen[hostKey(u)] {
					continue
				}
				seen[hostKey(u)] = true
				cands = append(cands, u.Scheme+"://"+u.Host+"/")
				if len(cands) >= 6 {
					break
				}
			}
			if cfg.GuessDomains {
				for _, g := range guessDomains(q) {
					u, _ := url.Parse(g)
					if !seen[hostKey(u)] {
						seen[hostKey(u)] = true
						cands = append(cands, g)
					}
				}
			}
		}
		if sr, ok := firstVerified(ctx, c, q, cands, site != ""); ok {
			all.Website = sr.Home
			all.merge(sr.Result)
		} else if site != "" {
			all.merge(sr.Result)
		} else if len(cands) > 0 {
			all.Notes = append(all.Notes, "website: none of "+itoa(len(cands))+" candidate sites names "+q.Name)
		}
	}

	// 3. Official IDs printed on the website.
	var jobs []func(context.Context) (Result, error)
	for _, id := range all.IDs {
		id := id
		switch {
		case id.Kind == "vat" && cfg.enabled("vies"):
			jobs = append(jobs, func(ctx context.Context) (Result, error) { return vies(ctx, c, q, id.Value) })
		case id.Kind == "cnpj" && cfg.enabled("brasil-cnpj"):
			jobs = append(jobs, func(ctx context.Context) (Result, error) { return cnpj(ctx, c, q, id.Value) })
		}
		if len(jobs) >= 3 {
			break
		}
	}
	// 4. Government gazettes of the countries the legal form points to.
	if useSearch && cfg.enabled("official") {
		done := map[string]bool{}
		for i, cc := range q.Countries {
			if i >= 1 {
				break // one gazette search per company, for the likeliest country
			}
			for _, d := range official[cc] {
				if done[d] {
					continue
				}
				done[d] = true
				d := d
				jobs = append(jobs, func(ctx context.Context) (Result, error) { return officialSearch(ctx, c, q, d, region) })
			}
		}
		// Government pages that the general searches already found.
		for _, h := range hits {
			if u, err := url.Parse(h.URL); err == nil && officialHost(u.Hostname()) {
				h := h
				jobs = append(jobs, func(ctx context.Context) (Result, error) { return officialPage(ctx, c, q, h.URL) })
			}
		}
	}
	all.merge(runJobs(ctx, jobs))

	// 5. Last resort for representatives: search result snippets.
	if useSearch && len(all.People) == 0 && cfg.enabled("search") {
		var n []string
		for _, h := range search(ctx, c, `"`+q.Name+`" `+roleWords(country), region, &n) {
			if !Mentions(normWords(h.Title+" "+h.Snippet), q) {
				continue
			}
			u, _ := url.Parse(h.URL)
			host := ""
			if u != nil {
				host = u.Hostname()
			}
			src := Source{ID: "search", Name: "search result from " + host + " (not verified)", URL: h.URL, Kind: KindCommunity}
			ps := ExtractPeople(snippetLines(h.Snippet), src)
			for _, p := range ps {
				all.addPerson(p)
			}
			if len(ps) > 0 {
				all.Sources = append(all.Sources, src)
			}
		}
		all.Notes = append(all.Notes, n...)
	}
	all.sortResult()
	return all
}
