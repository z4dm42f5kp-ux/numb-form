package lookup

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync"
	"time"
)

// notCompanySites are directories, social networks and data brokers.
// They are never taken for the company's own website.
var notCompanySites = []string{
	"linkedin.", "facebook.", "instagram.", "twitter.", "x.com", "youtube.", "tiktok.", "pinterest.", "wikipedia.org",
	"wikidata.org", "wikimedia.org", "northdata.", "dnb.com", "bloomberg.", "crunchbase.", "zoominfo.", "kompass.", "yelp.",
	"opencorporates.", "companyhouse.de", "firmenwissen.", "moneyhouse.", "unternehmensregister.", "handelsregister.",
	"google.", "bing.", "duckduckgo.", "amazon.", "ebay.", "glassdoor.", "indeed.", "xing.", "kununu.", "infobel.",
	"cylex", "gelbeseiten.", "dasoertliche.", "11880.", "wlw.", "europages.", "tripadvisor.", "empresite.", "einforma.",
	"axesor.", "infoempresa.", "cnpj.", "econodata.", "casadosdados.", "dateas.", "universidadperu.", "datosperu.",
	"bizapedia.", "opengovus.", "corporationwiki.", "buzzfile.", "manta.", "dandb.", "emis.", "owler.", "rocketreach.",
	"apollo.io", "signalhire.", "lusha.", "cbinsights.", "pitchbook.", "craft.co", "companiesmarketcap.", "marketscreener.",
	"reuters.", "ft.com", "wsj.com", "forbes.", "github.", "medium.", "reddit.", "quora.", "trustpilot.", "glassdoor",
	"firmania.", "brownbook.", "hotfrog.", "yellowpages.", "paginasamarillas.", "seccionamarilla.", "guiamais.",
	"empresas.", "informa.", "creditreform.", "implisense.", "firmenabc.", "herold.at", "wko.at", "zefix.", "sogc.",
	"boe.es", "sunbiz.org", "sec.gov", "boletinoficial.", "gov.", "gob.", ".gov", ".gob", "admin.ch", "europa.eu",
	"companieshouse", "company-information.service.gov.uk", "b2bhint.", "companiesdb.", "allbiz.", "cybo.", "waze.", "maps.",
}

func notCompanySite(host string) bool {
	h := "." + strings.ToLower(host)
	for _, s := range notCompanySites {
		if strings.Contains(h, s) {
			return true
		}
	}
	return false
}

// official lists government websites whose pages are searched for the
// company, by country. Their pages are read like a register entry.
var official = map[string][]string{
	"ES": {"boe.es"},                // Boletín Oficial del Registro Mercantil (BORME)
	"PT": {"publicacoes.mj.pt"},     // Portal da Justiça, company publications
	"AR": {"boletinoficial.gob.ar"}, // Boletín Oficial, second section (companies)
	"US": {"sunbiz.org"},            // Florida Division of Corporations
	"GB": {"find-and-update.company-information.service.gov.uk"},
}

// officialHost reports whether host belongs to a government register or
// gazette listed above.
func officialHost(host string) bool {
	h := strings.TrimPrefix(strings.ToLower(host), "www.")
	for _, list := range official {
		for _, d := range list {
			if h == d || strings.HasSuffix(h, "."+d) {
				return true
			}
		}
	}
	return false
}

// legalWords is the word people add to find a company's legal notice,
// by country.
func legalWords(country string) string {
	switch country {
	case "DE", "AT", "CH":
		return "impressum"
	case "ES", "MX", "AR", "CO", "PE", "CL", "EC", "VE", "UY", "PY", "BO", "GT", "CR", "PA", "DO", "HN", "SV", "NI":
		return "aviso legal"
	case "BR", "PT":
		return "cnpj"
	case "FR", "BE", "LU", "MA", "SN", "CI", "CM":
		return "mentions légales"
	}
	return "legal notice"
}

func roleWords(country string) string {
	switch country {
	case "DE", "AT":
		return "Geschäftsführer"
	case "CH":
		return "Verwaltungsrat Geschäftsführer"
	case "ES", "MX", "AR", "CO", "PE", "CL", "EC", "VE", "UY", "PY", "BO", "GT", "CR", "PA", "DO", "HN", "SV", "NI":
		return "representante legal gerente general"
	case "BR":
		return "sócio administrador"
	case "PT":
		return "gerente administrador"
	case "FR", "BE", "LU":
		return "président gérant"
	}
	return "managing director CEO"
}

// searchState stops searching for a while after DuckDuckGo asks for a
// human check, so the rest of the run is not slowed down.
var (
	searchStateMu sync.Mutex
	searchPaused  time.Time
)

func search(ctx context.Context, c *Client, q, region string, notes *[]string) []Hit {
	searchStateMu.Lock()
	paused := time.Now().Before(searchPaused)
	searchStateMu.Unlock()
	if paused {
		return nil
	}
	hits, err := Search(ctx, c, q, region)
	if errors.Is(err, ErrSearchBlocked) {
		searchStateMu.Lock()
		searchPaused = time.Now().Add(3 * time.Minute)
		searchStateMu.Unlock()
	}
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
			h2 := search(ctx, c, q.Name+" "+legalWords(country), region, &n)
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
			if i >= 3 {
				break
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

// snippetLines splits a search snippet at sentence ends so that labels
// such as "Geschäftsführer:" start a line.
func snippetLines(s string) []string {
	s = strings.NewReplacer(" · ", "\n", " | ", "\n", "; ", "\n", " - ", "\n", "... ", "\n", "… ", "\n").Replace(s)
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// firstVerified reads candidate websites, three at a time, and returns
// the first (in candidate order) that names the company. A site the user
// typed is accepted even when unverified.
func firstVerified(ctx context.Context, c *Client, q Query, cands []string, typed bool) (SiteResult, bool) {
	if len(cands) == 0 {
		return SiteResult{}, false
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	results := make([]*SiteResult, len(cands))
	done := make([]chan struct{}, len(cands))
	sem := make(chan struct{}, 3)
	for i := range cands {
		done[i] = make(chan struct{})
		go func(i int) {
			defer close(done[i])
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			r := ReadSite(ctx, c, cands[i], q)
			results[i] = &r
		}(i)
	}
	var notes []string
	for i := range cands {
		<-done[i]
		r := results[i]
		if r == nil {
			continue
		}
		if r.Verified || typed {
			cancel()
			r.Notes = append(notes, r.Notes...)
			return *r, r.Verified || typed
		}
		notes = append(notes, r.Notes...)
	}
	return SiteResult{Result: Result{Notes: notes}}, false
}

func runJobs(ctx context.Context, jobs []func(context.Context) (Result, error)) Result {
	results := make([]Result, len(jobs))
	var wg sync.WaitGroup
	for i := range jobs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			jctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			r, err := jobs[i](jctx)
			if err != nil {
				r.Notes = append(r.Notes, sanitizeErr(err))
			}
			results[i] = r
		}(i)
	}
	wg.Wait()
	var all Result
	for _, r := range results {
		all.merge(r)
	}
	return all
}
