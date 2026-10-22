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
	"dateas.com", "cuitonline.com", "einforma.co", "informacolombia.com", "empresas.mx", "datacreditoempresas.com.co",
	"directorio-empresas.einforma.co", "indicadores.ar", "bidsfactory.com", "ampliz.com", "dealroom.co", "panjiva.com",
	"emis.com", "dunsguide.com", "connectamericas.com", "infobel.com", "kompass.com", "empresite.com.co",
	"informacionempresarial.com.co", "verifiquese.com", "boliviaempresas.com", "paraguayempresas.com",
	"rdempresas.com", "empresasdeecuador.com", "ecuadorempresas.com", "guiaempresas.com.ve", "elsalvadorempresas.com",
	"cameroun-entreprises.com", "nigeriagalleria.com", "vconnect.com", "ngcheckers.com", "dubaicompanies.ae",
	"yello.ae", "connect.ae", "opencorporates.com.co",
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

// governmentHost reports whether a host belongs to a government, such as
// a procurement portal or an official gazette (.gov, .gob, .gouv).
func governmentHost(host string) bool {
	h := "." + strings.ToLower(host) + "."
	for _, m := range []string{".gov.", ".gob.", ".gouv.", ".gub.", ".go.", ".admin.ch.", ".europa.eu."} {
		if strings.Contains(h, m) {
			return true
		}
	}
	return false
}

// socialHost reports whether a host is a social network, whose pages
// need a login and are not read (their search snippets still are).
func socialHost(host string) bool {
	h := "." + strings.ToLower(host)
	for _, s := range []string{".linkedin.com", ".facebook.com", ".instagram.com", ".twitter.com", ".x.com", ".youtube.com", ".tiktok.com", ".pinterest."} {
		if strings.Contains(h, s) {
			return true
		}
	}
	return false
}

// hitPage reads the part of a search result page about the company:
// a directory entry, a procurement record, a gazette or a news article.
func hitPage(ctx context.Context, c *Client, q Query, pageURL string) (Result, error) {
	var res Result
	budget := 1
	body, final, err := c.page(ctx, pageURL, &budget)
	if err != nil {
		tr(ctx).Ev("page", "u", pageURL, "err", err)
		return res, err
	}
	if final == nil {
		final, _ = url.Parse(pageURL)
	}
	all := HTMLText(string(body))
	if !Mentions(normWords(strings.Join(all, "\n")), q) {
		tr(ctx).Ev("page", "u", final.String(), "named", false, "lines", len(all))
		return res, nil
	}
	lines := companyLines(all, q)
	host := strings.TrimPrefix(final.Hostname(), "www.")
	src := Source{ID: "search", Name: "web page " + host + " (not verified)", URL: final.String(), Kind: KindCommunity}
	switch {
	case directoryHost(host):
		src = Source{ID: "directory", Name: "company data directory " + host + " (not official)", URL: final.String(), Kind: KindCommunity}
	case governmentHost(host):
		src = Source{ID: "search", Name: "government website " + host, URL: final.String(), Kind: KindGovernment}
	}
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
	res.IDs = appendIDs(res.IDs, ExtractIDs(strings.Join(lines, "\n")))
	if len(res.People)+len(res.Addresses) > 0 {
		res.Sources = append(res.Sources, src)
	}
	tr(ctx).Ev("page", "u", final.String(), "named", true, "kind", src.Kind.String(), "lines", len(lines), "people", len(res.People), "addr", a)
	return res, nil
}
