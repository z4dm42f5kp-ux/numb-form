package lookup

import (
	"context"
	"net/url"
	"strings"
	"time"
)

// state is what one lookup has gathered so far.
type state struct {
	c       *Client
	cfg     Config
	q       Query
	site    string // a website the user typed
	region  string // search region, such as "co-es"
	country string // the company's country, when known
	lang    string // the country whose words to search with, even a guess
	search  bool
	hits    []Hit           // every search result so far
	read    map[string]bool // pages already read
	all     Result
}

// provider is one source in the waterfall.
type provider struct {
	name string
	run  func(ctx context.Context, s *state)
}

// providers are tried in this order until representatives and the
// address are settled (see done). Cheap, official sources come first;
// the AI answer comes last.
var providers = []provider{
	{"registers", registersStep}, // GLEIF, Colombia SECOP II, Companies House
	{"search", searchStep},       // name variants until a result names the company
	{"website", websiteStep},     // the company's own site
	{"pages", pagesStep},         // directories, procurement and news pages found by search
	{"ids", idsStep},             // VAT number -> EU VIES, CNPJ -> Receita Federal
	{"targeted", targetedStep},   // searches for the representatives and the address
	{"places", placesStep},       // Google Maps through Serper, for the address
	{"ai", aiStep},               // Tavily's sourced answer, never verified
}

// needs reports what is still missing: a representative from a register
// or confirmed by two kinds of source, and an address from a register or
// the company's own website.
func needs(r *Result) (reps, addr bool) {
	reps = true
	for _, p := range r.People {
		if p.Verified() {
			reps = false
			break
		}
	}
	addr = true
	for _, a := range r.Addresses {
		if a.Source.Kind <= KindCompany {
			addr = false
			break
		}
	}
	return reps, addr
}

// Find looks a company up through the providers in order and stops as
// soon as its representatives and address are settled. Everything it
// does is written to the trace in ctx (output.txt). site, when not
// empty, is a website the user typed.
func Find(ctx context.Context, c *Client, cfg Config, name, site string) Result {
	q := NewQuery(name)
	c.configure(cfg)
	s := &state{c: c, cfg: cfg, q: q, site: site, read: map[string]bool{}}
	if len(q.Countries) > 0 && q.Confident {
		s.country = q.Countries[0] // otherwise search worldwide
	}
	if len(q.Countries) > 0 {
		s.lang = q.Countries[0]
	}
	s.region = ddgRegion(s.country)
	s.search = cfg.Search != "none" && cfg.enabled("search")
	t := tr(ctx)
	t.Ev("query", "clean", q.Name, "core", q.Base, "form", q.Form, "places", q.Places, "cc", q.Countries, "searches", q.Searches, "browser", c.BrowserName(), "serper", c.serperKey != "", "tavily", c.tavilyKey != "")
	for _, p := range providers {
		if ctx.Err() != nil {
			t.Ev("stop", "err", ctx.Err())
			break
		}
		if reps, addr := needs(&s.all); !reps && !addr {
			t.Ev("settled", "before", p.name)
			break
		}
		start := time.Now()
		np, na := len(s.all.People), len(s.all.Addresses)
		p.run(ctx, s)
		t.Ev("provider", "p", p.name, "ms", time.Since(start).Milliseconds(), "people", len(s.all.People)-np, "addr", len(s.all.Addresses)-na)
	}
	s.all.sortResult()
	var ps, as []string
	for _, p := range s.all.People {
		ps = append(ps, p.Name+" ["+strings.Join(p.Roles, ",")+"] "+p.Best().String())
	}
	for _, a := range s.all.Addresses {
		as = append(as, a.Text+" ["+a.Source.ID+"]")
	}
	t.Ev("result", "site", s.all.Website, "people", ps, "addr", as, "notes", s.all.Notes)
	return s.all
}

func registersStep(ctx context.Context, s *state) {
	s.all.merge(Registries(ctx, s.c, s.cfg, s.q))
}

// search runs one web search and records problems as notes.
func (s *state) searchOnce(ctx context.Context, query string) []Hit {
	hits, engine, err := Search(ctx, s.c, query, s.region)
	named := 0
	for _, h := range hits {
		if Mentions(normWords(h.Title+" "+h.Snippet), s.q) {
			named++
		}
	}
	tr(ctx).Ev("search", "q", query, "eng", engine, "hits", len(hits), "named", named, "err", err)
	if err != nil {
		s.all.Notes = append(s.all.Notes, "web search: "+sanitizeErr(err))
	}
	s.hits = append(s.hits, hits...)
	return hits
}

// searchStep tries the name the way a person would (literal, clean in
// quotes, with its city, core name with country, without accents) and
// stops at the first search whose results name the company.
func searchStep(ctx context.Context, s *state) {
	if !s.search {
		return
	}
	for i, query := range s.q.Searches {
		if i >= 5 || ctx.Err() != nil {
			break
		}
		named := false
		for _, h := range s.searchOnce(ctx, query) {
			if Mentions(normWords(h.Title+" "+h.Snippet), s.q) {
				named = true
				break
			}
		}
		if named {
			break
		}
	}
	s.all.merge(mineSnippets(s.hits, s.q))
}

// websiteStep reads the company's own website: the one the user typed,
// or sites among the search results and guessed from the name.
func websiteStep(ctx context.Context, s *state) {
	if !s.cfg.enabled("website") {
		return
	}
	var cands []string
	seen := map[string]bool{}
	add := func(raw string) {
		u, ok := NormalizeWebsite(raw)
		if !ok || seen[hostKey(u)] || notCompanySite(u.Hostname()) || directoryHost(u.Hostname()) {
			return
		}
		seen[hostKey(u)] = true
		cands = append(cands, u.Scheme+"://"+u.Host+"/")
	}
	if s.site != "" {
		add(s.site)
	} else {
		add(s.all.Website)
		for _, w := range s.all.Websites {
			add(w)
		}
		// Results that name the company first, then the others.
		for pass := 0; pass < 2; pass++ {
			for _, h := range s.hits {
				if (pass == 0) == Mentions(normWords(h.Title+" "+h.Snippet), s.q) && len(cands) < 6 {
					add(h.URL)
				}
			}
		}
		if s.cfg.GuessDomains {
			for _, g := range guessDomains(s.q) {
				add(g)
			}
		}
	}
	tr(ctx).Ev("site-candidates", "u", cands)
	sr, ok := firstVerified(ctx, s.c, s.q, cands, s.site != "")
	tr(ctx).Ev("site", "home", sr.Home, "ok", ok, "pages", sr.Pages, "people", len(sr.People), "addr", len(sr.Addresses), "notes", sr.Notes)
	if ok || s.site != "" {
		s.all.Website = sr.Home
		s.all.merge(sr.Result)
	}
}

// pagesStep reads the pages among the search results that name the
// company: directories, procurement records, news. At most four.
func pagesStep(ctx context.Context, s *state) {
	s.readHits(ctx, s.hits, 4)
}

func (s *state) readHits(ctx context.Context, hits []Hit, max int) {
	var jobs []func(context.Context) (Result, error)
	for _, h := range hits {
		u, err := url.Parse(h.URL)
		if err != nil || s.read[h.URL] || strings.HasSuffix(strings.ToLower(u.Path), ".pdf") || len(jobs) >= max {
			continue
		}
		dir := directoryHost(u.Hostname())
		if !dir && (!Mentions(normWords(h.Title+" "+h.Snippet), s.q) || socialHost(u.Hostname())) {
			continue
		}
		if !s.cfg.enabled("directory") && dir {
			continue
		}
		s.read[h.URL] = true
		link := h.URL
		jobs = append(jobs, func(ctx context.Context) (Result, error) { return hitPage(ctx, s.c, s.q, link) })
	}
	s.all.merge(runJobs(ctx, jobs))
}

func idsStep(ctx context.Context, s *state) {
	var jobs []func(context.Context) (Result, error)
	for _, id := range s.all.IDs {
		id := id
		switch {
		case id.Kind == "vat" && s.cfg.enabled("vies"):
			jobs = append(jobs, func(ctx context.Context) (Result, error) { return vies(ctx, s.c, s.q, id.Value) })
		case id.Kind == "cnpj" && s.cfg.enabled("brasil-cnpj"):
			jobs = append(jobs, func(ctx context.Context) (Result, error) { return cnpj(ctx, s.c, s.q, id.Value) })
		}
		if len(jobs) >= 3 {
			break
		}
	}
	s.all.merge(runJobs(ctx, jobs))
}

// targetedStep searches for what is still missing, with the local words
// for representatives or address, and reads the pages that name it.
func targetedStep(ctx context.Context, s *state) {
	if !s.search {
		return
	}
	where := ""
	if len(s.q.Places) > 0 {
		where = " " + s.q.Places[len(s.q.Places)-1]
	}
	if reps, _ := needs(&s.all); reps {
		h := s.searchOnce(ctx, `"`+s.q.Name+`" `+roleWords(s.lang)+where)
		s.all.merge(mineSnippets(h, s.q))
		s.readHits(ctx, h, 2)
	}
	if _, addr := needs(&s.all); addr && len(s.all.Addresses) == 0 {
		h := s.searchOnce(ctx, `"`+s.q.Name+`" `+addressWords(s.lang)+where)
		s.all.merge(mineSnippets(h, s.q))
		s.readHits(ctx, h, 2)
	}
}

func placesStep(ctx context.Context, s *state) {
	if s.c.serperKey == "" || len(s.all.Addresses) > 0 {
		return
	}
	r, err := serperPlaces(ctx, s.c, s.q, s.region)
	if err != nil {
		r.Notes = append(r.Notes, "Google Maps: "+sanitizeErr(err))
	}
	s.all.merge(r)
}

func aiStep(ctx context.Context, s *state) {
	if s.c.tavilyKey == "" || !s.cfg.enabled("ai") {
		return
	}
	r, err := tavily(ctx, s.c, s.q)
	if err != nil {
		r.Notes = append(r.Notes, "AI answer: "+sanitizeErr(err))
	}
	s.all.merge(r)
}

// addressWords is what people add to a search for a company's address.
func addressWords(country string) string {
	switch country {
	case "DE", "AT", "CH":
		return "Adresse Sitz"
	case "BR", "PT":
		return "endereço sede"
	case "FR", "BE", "LU", "MA", "SN", "CI", "CM":
		return "adresse siège"
	case "US", "GB", "IE", "ZA", "NG", "KE", "GH", "IN", "AU", "CA", "NZ", "SG", "AE":
		return "address headquarters"
	}
	return "dirección domicilio"
}
