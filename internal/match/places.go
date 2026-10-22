package match

import "strings"

// countryNames maps country names in English, Spanish, Portuguese,
// French and German (folded to lower case without accents) to ISO codes.
var countryNames = map[string]string{
	"colombia": "CO", "mexico": "MX", "ecuador": "EC", "peru": "PE", "chile": "CL", "argentina": "AR", "bolivia": "BO",
	"paraguay": "PY", "uruguay": "UY", "venezuela": "VE", "brasil": "BR", "brazil": "BR", "panama": "PA", "costa rica": "CR",
	"guatemala": "GT", "honduras": "HN", "el salvador": "SV", "salvador": "SV", "nicaragua": "NI", "cuba": "CU",
	"republica dominicana": "DO", "dominican republic": "DO", "rep dominicana": "DO", "puerto rico": "PR",
	"espana": "ES", "spain": "ES", "portugal": "PT", "alemania": "DE", "germany": "DE", "deutschland": "DE",
	"austria": "AT", "osterreich": "AT", "suiza": "CH", "switzerland": "CH", "schweiz": "CH", "suisse": "CH",
	"francia": "FR", "france": "FR", "italia": "IT", "italy": "IT", "paises bajos": "NL", "netherlands": "NL",
	"belgica": "BE", "belgium": "BE", "reino unido": "GB", "united kingdom": "GB", "uk": "GB", "england": "GB",
	"estados unidos": "US", "united states": "US", "usa": "US", "eeuu": "US", "canada": "CA",
	"emiratos arabes unidos": "AE", "united arab emirates": "AE", "uae": "AE", "emirates": "AE",
	"cameroon": "CM", "cameroun": "CM", "camerun": "CM", "nigeria": "NG", "ghana": "GH", "kenya": "KE",
	"senegal": "SN", "cote d ivoire": "CI", "ivory coast": "CI", "south africa": "ZA", "sudafrica": "ZA",
	"egypt": "EG", "egipto": "EG", "morocco": "MA", "marruecos": "MA", "maroc": "MA", "china": "CN", "india": "IN",
	"turkey": "TR", "turquia": "TR", "saudi arabia": "SA", "arabia saudita": "SA", "qatar": "QA", "israel": "IL",
}

// cities maps cities that appear after company names to ISO codes.
var cities = map[string]string{
	"bogota": "CO", "medellin": "CO", "cali": "CO", "barranquilla": "CO", "cartagena": "CO", "bucaramanga": "CO",
	"ciudad de mexico": "MX", "cdmx": "MX", "mexico df": "MX", "monterrey": "MX", "guadalajara": "MX", "puebla": "MX",
	"quito": "EC", "guayaquil": "EC", "cuenca": "EC", "lima": "PE", "arequipa": "PE", "santiago": "CL", "santiago de chile": "CL",
	"buenos aires": "AR", "cordoba": "AR", "rosario": "AR", "la paz": "BO", "el alto": "BO", "santa cruz": "BO",
	"santa cruz de la sierra": "BO", "cochabamba": "BO", "asuncion": "PY", "fernando de la mora": "PY", "luque": "PY",
	"san lorenzo": "PY", "ciudad del este": "PY", "montevideo": "UY", "caracas": "VE", "valencia": "VE", "maracaibo": "VE",
	"barquisimeto": "VE", "sao paulo": "BR", "rio de janeiro": "BR", "panama city": "PA", "ciudad de panama": "PA",
	"san jose": "CR", "ciudad de guatemala": "GT", "tegucigalpa": "HN", "san pedro sula": "HN", "san salvador": "SV",
	"santa tecla": "SV", "antiguo cuscatlan": "SV", "managua": "NI", "santo domingo": "DO", "santiago de los caballeros": "DO",
	"douala": "CM", "yaounde": "CM", "lagos": "NG", "abuja": "NG", "kano": "NG", "port harcourt": "NG", "accra": "GH",
	"nairobi": "KE", "dakar": "SN", "abidjan": "CI", "johannesburg": "ZA", "cape town": "ZA", "cairo": "EG",
	"casablanca": "MA", "dubai": "AE", "abu dhabi": "AE", "sharjah": "AE", "ajman": "AE", "doha": "QA", "riyadh": "SA",
	"miami": "US", "doral": "US", "houston": "US", "new york": "US", "los angeles": "US", "madrid": "ES", "barcelona": "ES",
	"lisboa": "PT", "lisbon": "PT", "porto": "PT", "berlin": "DE", "hamburg": "DE", "munchen": "DE", "munich": "DE",
	"frankfurt": "DE", "wien": "AT", "vienna": "AT", "zurich": "CH", "geneve": "CH", "geneva": "CH", "basel": "CH",
}

// countryEnglish gives the English name used in searches.
var countryEnglish = map[string]string{
	"CO": "Colombia", "MX": "Mexico", "EC": "Ecuador", "PE": "Peru", "CL": "Chile", "AR": "Argentina", "BO": "Bolivia",
	"PY": "Paraguay", "UY": "Uruguay", "VE": "Venezuela", "BR": "Brazil", "PA": "Panama", "CR": "Costa Rica",
	"GT": "Guatemala", "HN": "Honduras", "SV": "El Salvador", "NI": "Nicaragua", "DO": "Dominican Republic",
	"ES": "Spain", "PT": "Portugal", "DE": "Germany", "AT": "Austria", "CH": "Switzerland", "FR": "France", "US": "USA",
	"GB": "United Kingdom", "AE": "Dubai", "CM": "Cameroon", "NG": "Nigeria", "GH": "Ghana", "KE": "Kenya", "ZA": "South Africa",
	"SN": "Senegal", "CI": "Cote d'Ivoire", "EG": "Egypt", "MA": "Morocco", "IT": "Italy", "NL": "Netherlands", "BE": "Belgium",
}

// CountryName returns the English name of an ISO code for searches.
func CountryName(cc string) string {
	if n, ok := countryEnglish[cc]; ok {
		return n
	}
	return cc
}

func foldPlace(s string) string {
	s = strings.ToLower(ASCII(strings.TrimSpace(s)))
	s = strings.NewReplacer(".", " ", "'", " ", "’", " ", "-", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

// Place reports whether s is a known city or country, with its country.
func Place(s string) (string, bool) {
	k := foldPlace(s)
	if cc, ok := countryNames[k]; ok {
		return cc, true
	}
	if cc, ok := cities[k]; ok {
		return cc, true
	}
	return "", false
}

func countryWord(w string) (string, bool) {
	cc, ok := countryNames[foldPlace(w)]
	return cc, ok
}

// CountryWords returns the names of a country in the languages above.
func CountryWords(cc string) []string {
	var out []string
	for name, c := range countryNames {
		if c == cc && len(name) > 3 {
			out = append(out, name)
		}
	}
	return out
}
