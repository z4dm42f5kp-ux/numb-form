package match

import "strings"

// countryNames are country names in English, Spanish, Portuguese, French
// and German, folded (see foldPlace).
var countryNames = map[string]bool{
	"colombia": true, "mexico": true, "ecuador": true, "peru": true, "chile": true, "argentina": true, "bolivia": true,
	"paraguay": true, "uruguay": true, "venezuela": true, "brasil": true, "brazil": true, "panama": true, "costa rica": true,
	"guatemala": true, "honduras": true, "el salvador": true, "nicaragua": true, "cuba": true, "republica dominicana": true,
	"dominican republic": true, "rep dominicana": true, "puerto rico": true, "espana": true, "spain": true, "portugal": true,
	"alemania": true, "germany": true, "deutschland": true, "austria": true, "osterreich": true, "suiza": true, "switzerland": true,
	"schweiz": true, "suisse": true, "francia": true, "france": true, "italia": true, "italy": true, "netherlands": true,
	"belgica": true, "belgium": true, "reino unido": true, "united kingdom": true, "uk": true, "england": true,
	"estados unidos": true, "united states": true, "usa": true, "eeuu": true, "canada": true, "emiratos arabes unidos": true,
	"united arab emirates": true, "uae": true, "emirates": true, "cameroon": true, "cameroun": true, "camerun": true,
	"nigeria": true, "ghana": true, "kenya": true, "senegal": true, "cote d ivoire": true, "ivory coast": true,
	"south africa": true, "egypt": true, "morocco": true, "maroc": true, "china": true, "india": true, "turkey": true,
	"saudi arabia": true, "qatar": true, "israel": true, "oman": true, "kuwait": true, "bahrain": true, "jordan": true,
	"lebanon": true, "angola": true, "mozambique": true, "tanzania": true, "uganda": true, "ethiopia": true, "gabon": true,
	"congo": true, "benin": true, "togo": true, "mali": true, "burkina faso": true, "niger": true, "tunisia": true, "algeria": true,
}

// cities are cities that appear after company names, folded.
var cities = map[string]bool{
	"bogota": true, "medellin": true, "cali": true, "barranquilla": true, "cartagena": true, "bucaramanga": true, "pereira": true,
	"ciudad de mexico": true, "cdmx": true, "mexico df": true, "monterrey": true, "guadalajara": true, "puebla": true, "tijuana": true,
	"quito": true, "guayaquil": true, "cuenca": true, "lima": true, "arequipa": true, "santiago": true, "santiago de chile": true,
	"buenos aires": true, "cordoba": true, "rosario": true, "mendoza": true, "la paz": true, "el alto": true, "santa cruz": true,
	"santa cruz de la sierra": true, "cochabamba": true, "asuncion": true, "fernando de la mora": true, "luque": true,
	"san lorenzo": true, "ciudad del este": true, "montevideo": true, "caracas": true, "valencia": true, "maracaibo": true,
	"barquisimeto": true, "sao paulo": true, "rio de janeiro": true, "panama city": true, "ciudad de panama": true,
	"san jose": true, "ciudad de guatemala": true, "tegucigalpa": true, "san pedro sula": true, "san salvador": true,
	"santa tecla": true, "antiguo cuscatlan": true, "managua": true, "santo domingo": true, "santiago de los caballeros": true,
	"douala": true, "yaounde": true, "lagos": true, "abuja": true, "kano": true, "port harcourt": true, "accra": true,
	"nairobi": true, "dakar": true, "abidjan": true, "johannesburg": true, "cape town": true, "cairo": true, "casablanca": true,
	"dubai": true, "abu dhabi": true, "sharjah": true, "ajman": true, "jebel ali": true, "doha": true, "riyadh": true, "jeddah": true,
	"muscat": true, "manama": true, "kuwait city": true, "amman": true, "beirut": true, "miami": true, "doral": true,
	"houston": true, "new york": true, "los angeles": true, "madrid": true, "barcelona": true, "lisboa": true, "lisbon": true,
	"porto": true, "berlin": true, "hamburg": true, "munchen": true, "munich": true, "frankfurt": true, "koln": true,
	"dusseldorf": true, "stuttgart": true, "wien": true, "vienna": true, "zurich": true, "geneve": true, "geneva": true, "basel": true,
}

// foldPlace lowers case and removes accents and punctuation.
func foldPlace(s string) string {
	s = strings.ToLower(fold(strings.TrimSpace(s)))
	s = strings.NewReplacer(".", " ", "'", " ", "’", " ", "-", " ", ",", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

// knownPlace reports whether s is a known city or country.
func knownPlace(s string) bool {
	k := foldPlace(s)
	return countryNames[k] || cities[k]
}

var accents = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e', 'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
	'ó': 'o', 'ò': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u', 'ñ': 'n', 'ç': 'c',
	'Á': 'A', 'À': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'É': 'E', 'È': 'E', 'Ê': 'E', 'Í': 'I', 'Ó': 'O', 'Ô': 'O', 'Õ': 'O',
	'Ö': 'O', 'Ú': 'U', 'Ü': 'U', 'Ñ': 'N', 'Ç': 'C',
}

// fold removes accents: "Bogotá" gives "Bogota".
func fold(s string) string {
	return strings.Map(func(r rune) rune {
		if v, ok := accents[r]; ok {
			return v
		}
		return r
	}, s)
}
