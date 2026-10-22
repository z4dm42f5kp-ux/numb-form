package lookup

import "strings"

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
	case "FR", "BE", "LU", "CM", "SN", "CI", "MA":
		return "gérant directeur général"
	case "US", "GB", "IE", "ZA", "NG", "KE", "GH", "IN", "AE":
		return "managing director CEO owner"
	}
	return "managing director CEO"
}
