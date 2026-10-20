package lookup

import (
	"context"
	"net/url"
	"strings"
)

// directories are company-data websites that republish register data:
// managing directors, officers, addresses. They are not official, so
// what they say is shown unverified with the caution text, unless an
// official source or the company's own website says the same.
var directories = []string{
	// Germany, Austria, Switzerland.
	"northdata.de", "northdata.com", "firmenwissen.de", "companyhouse.de", "online-handelsregister.de", "unternehmensverzeichnis.org",
	"moneyhouse.ch", "easymonitoring.ch", "firmenabc.at", "firmen.wko.at", "compnet.at",
	// Spain and Portugal.
	"librebor.me", "empresite.eleconomista.es", "infoempresa.com", "einforma.com", "infocif.es", "axesor.es", "racius.com", "einforma.pt",
	// Latin America.
	"cnpj.biz", "casadosdados.com.br", "cnpja.com", "econodata.com.br", "datosperu.org", "universidadperu.com", "genealog.cl",
	"dateas.com", "cuitonline.com", "einforma.co", "informacolombia.com", "empresas.mx",
	// United States, Africa and worldwide.
	"opencorporates.com", "bizapedia.com", "opengovus.com", "corporationwiki.com", "dnb.com", "b2bhint.com",
}

func directoryHost(host string) bool {
	h := strings.TrimPrefix(strings.ToLower(host), "www.")
	for _, d := range directories {
		if h == d || strings.HasSuffix(h, "."+d) {
			return true
		}
	}
	return false
}

// directoryPage reads the part of a directory page about the company.
func directoryPage(ctx context.Context, c *Client, q Query, pageURL string) (Result, error) {
	var res Result
	budget := 1
	body, final, err := c.page(ctx, pageURL, &budget)
	if err != nil {
		return res, err
	}
	if final == nil {
		final, _ = url.Parse(pageURL)
	}
	all := HTMLText(string(body))
	if !Mentions(normWords(strings.Join(all, "\n")), q) {
		return res, nil
	}
	lines := companyLines(all, q)
	host := strings.TrimPrefix(final.Hostname(), "www.")
	src := Source{ID: "directory", Name: "company data directory " + host + " (not official)", URL: final.String(), Kind: KindCommunity}
	for _, p := range ExtractPeople(lines, src) {
		res.addPerson(p)
	}
	a := ExtractAddress(lines, q)
	if a == "" {
		for _, l := range lines {
			if a = inlineAddress(l); a != "" {
				break
			}
		}
	}
	if a != "" {
		res.Addresses = append(res.Addresses, Address{Text: a, Source: src})
	}
	if len(res.People)+len(res.Addresses) > 0 {
		res.Sources = append(res.Sources, src)
	}
	return res, nil
}

// directoryJobs reads up to max directory pages found among hits.
func directoryJobs(c *Client, q Query, hits []Hit, max int) []func(context.Context) (Result, error) {
	var jobs []func(context.Context) (Result, error)
	seen := map[string]bool{}
	for _, h := range hits {
		u, err := url.Parse(h.URL)
		if err != nil || !directoryHost(u.Hostname()) || seen[hostKey(u)] {
			continue
		}
		seen[hostKey(u)] = true
		link := h.URL
		jobs = append(jobs, func(ctx context.Context) (Result, error) { return directoryPage(ctx, c, q, link) })
		if len(jobs) >= max {
			break
		}
	}
	return jobs
}
