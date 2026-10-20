package app

import (
	"sort"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/lookup"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// gathered is everything found for one company, filtered by the
// configured sources.
type gathered struct {
	res lookup.Result
	cfg config.Config
}

func union(a, b []string) []string {
	out := append([]string(nil), a...)
	for _, x := range b {
		if !config.Allowed(out, x) {
			out = append(out, x)
		}
	}
	return out
}

// gather looks the company up online in one pass: registers, its own
// website (found automatically unless website = ask or never), official
// numbers printed there, government gazettes and web search.
func (r *run) gather(name string) (gathered, error) {
	g := gathered{cfg: r.cfg}
	if !r.online {
		return g, nil
	}
	enabled := union(r.cfg.AddressSources, r.cfg.RepsSources)
	if r.cfg.Website == "never" {
		var keep []string
		for _, s := range enabled {
			if s != "website" {
				keep = append(keep, s)
			}
		}
		enabled = keep
	}
	if len(enabled) == 0 {
		return g, nil // every source disabled in the configuration
	}
	lc := r.opt.Lookup
	lc.Sources = enabled
	lc.Search = r.cfg.Search
	lc.Browser = r.cfg.Browser
	lc.GuessDomains = r.cfg.GuessDomains
	site := ""
	if r.cfg.Website == "ask" && config.Allowed(enabled, "website") {
		var err error
		if site, err = r.askWebsite(nil); err != nil {
			return g, err
		}
	}
	r.ui.Println("  Searching registers, the company website, official gazettes and the web…")
	g.res = lookup.Find(r.ctx, r.opt.Client, lc, name, site)
	if err := r.ctx.Err(); err != nil {
		return g, err
	}
	if !r.browserNoted {
		r.browserNoted = true
		switch name := r.opt.Client.BrowserName(); {
		case name != "":
			r.ui.Printf("  (%s is used invisibly for websites built with JavaScript and when search engines refuse.)\n", name)
		case r.cfg.Browser != "none":
			r.ui.Println("  (No Google Chrome found: websites built with JavaScript cannot be read.)")
		}
	}
	if g.res.Website != "" {
		r.ui.Printf("  Website: %s\n", ui.Safe(g.res.Website))
	}
	for _, id := range g.res.IDs {
		r.ui.Printf("  Official number on the website: %s %s\n", strings.ToUpper(id.Kind), ui.Safe(id.Value))
	}
	for _, n := range g.res.Register {
		r.ui.Printf("  Register entry on the website: %s\n", ui.Safe(n.Text))
	}
	r.ui.Printf("  Found: %d address(es), %d representative(s) from %d source(s).\n", len(g.addresses()), len(g.people()), len(g.res.Sources))
	for _, n := range g.res.Notes {
		r.logf("lookup note for %q: %s", name, n)
	}
	if len(g.people()) == 0 && len(g.addresses()) == 0 {
		for _, n := range g.res.Notes {
			r.ui.Printf("  · %s\n", ui.Safe(n))
		}
	}
	if r.cfg.Confirm != "none" && len(g.res.Excerpts) > 0 {
		e := g.res.Excerpts[0]
		r.ui.Printf("\n  Excerpt from %s:\n", ui.Safe(e.Source.URL))
		lines := strings.Split(e.Text, "\n")
		if len(lines) > 15 {
			lines = append(lines[:15], "…")
		}
		for _, l := range lines {
			r.ui.Printf("    | %s\n", ui.Safe(l))
		}
	}
	for _, s := range g.res.Sources {
		r.logf("source for %q: %s %s", name, s.Name, s.URL)
	}
	return g, nil
}

func rank(list []string, id string) int {
	for i, s := range list {
		if s == id {
			return i
		}
	}
	return len(list)
}

// addresses returns the candidates from allowed sources, in the
// configured order.
func (g gathered) addresses() []lookup.Address {
	var out []lookup.Address
	for _, a := range g.res.Addresses {
		if config.Allowed(g.cfg.AddressSources, a.Source.ID) {
			out = append(out, a)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return rank(g.cfg.AddressSources, out[i].Source.ID) < rank(g.cfg.AddressSources, out[j].Source.ID)
	})
	return out
}

func (g gathered) trusted(p lookup.Person) bool {
	switch g.cfg.RepsMinTrust {
	case "register":
		return p.Best() <= lookup.KindRegulatory
	case "website":
		return p.Best() <= lookup.KindCompany
	}
	return true
}

// people returns the candidates backed by allowed, trusted sources.
func (g gathered) people() []lookup.Person {
	var out []lookup.Person
	for _, p := range g.res.People {
		var src []lookup.Source
		for _, s := range p.Sources {
			if config.Allowed(g.cfg.RepsSources, s.ID) {
				src = append(src, s)
			}
		}
		if len(src) == 0 {
			continue
		}
		p.Sources = src
		if g.trusted(p) {
			out = append(out, p)
		}
	}
	return out
}

func (g gathered) sourceList() []string {
	var out []string
	for _, s := range g.res.Sources {
		out = append(out, s.String()+" "+s.URL)
	}
	return out
}

func (r *run) askWebsite(suggest []string) (string, error) {
	def := ""
	if len(suggest) > 0 {
		def = suggest[0]
	}
	for {
		p := "  Company website (type an address, Enter = find it automatically): "
		a, err := r.ui.Ask(p)
		if err != nil {
			return "", err
		}
		switch {
		case a == "-":
			return "", nil
		case a == "":
			return def, nil
		}
		if _, ok := lookup.NormalizeWebsite(a); ok {
			return a, nil
		}
		r.ui.Println("  That is not a usable web address.")
	}
}
