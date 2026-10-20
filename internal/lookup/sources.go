package lookup

import (
	"net/url"
	"strings"
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

// hasCompanySite reports whether search results include a site that may
// be the company's own.
func hasCompanySite(hits []Hit) bool {
	for _, h := range hits {
		if u, err := url.Parse(h.URL); err == nil && !notCompanySite(u.Hostname()) {
			return true
		}
	}
	return false
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
