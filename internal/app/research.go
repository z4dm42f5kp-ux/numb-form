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

// gather runs the online lookups allowed by the configuration.
func (r *run) gather(name string) (gathered, error) {
	g := gathered{cfg: r.cfg}
	if !r.online {
		return g, nil
	}
	enabled := union(r.cfg.AddressSources, r.cfg.RepsSources)
	if len(enabled) == 0 {
		return g, nil // every source disabled in the configuration
	}
	lc := r.opt.Lookup
	lc.Sources = enabled
	r.ui.Println("  Looking up official sources…")
	g.res = lookup.Registries(r.ctx, r.opt.Client, lc, lookup.NewQuery(name))
	if err := r.ctx.Err(); err != nil {
		return g, err
	}
	for _, n := range g.res.Notes {
		r.ui.Printf("  · %s\n", ui.Safe(n))
	}
	if config.Allowed(enabled, "website") && r.cfg.Website != "never" {
		site := ""
		if r.cfg.Website == "ask" {
			var err error
			if site, err = r.askWebsite(g.res.Websites); err != nil {
				return g, err
			}
		} else if len(g.res.Websites) > 0 {
			site = g.res.Websites[0]
		}
		if site != "" {
			r.ui.Println("  Reading the company's legal notice…")
			w := lookup.Website(r.ctx, r.opt.Client, site)
			for _, n := range w.Notes {
				r.ui.Printf("  · %s\n", ui.Safe(n))
			}
			g.res.Merge(w)
			for _, e := range w.Excerpts {
				r.ui.Printf("\n  Excerpt from %s:\n", ui.Safe(e.Source.URL))
				for _, l := range strings.Split(e.Text, "\n") {
					r.ui.Printf("    | %s\n", ui.Safe(l))
				}
			}
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
		r.ui.Printf("  Website suggested by Wikidata: %s\n", ui.Safe(def))
	}
	for {
		p := "  Official company website to read its legal notice (Enter = suggestion, '-' = skip, or type an address): "
		if def == "" {
			p = "  Official company website to read its legal notice (type an address, Enter = skip): "
		}
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
