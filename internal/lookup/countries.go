package lookup

import "strings"

var countries = map[string]string{
	"AD": "Andorra", "AE": "United Arab Emirates", "AR": "Argentina", "AT": "Austria", "AU": "Australia",
	"BE": "Belgium", "BG": "Bulgaria", "BO": "Bolivia", "BR": "Brazil", "BS": "Bahamas", "BZ": "Belize",
	"CA": "Canada", "CH": "Switzerland", "CL": "Chile", "CN": "China", "CO": "Colombia", "CR": "Costa Rica",
	"CU": "Cuba", "CW": "Curaçao", "CY": "Cyprus", "CZ": "Czechia", "DE": "Germany", "DK": "Denmark",
	"DO": "Dominican Republic", "EC": "Ecuador", "EE": "Estonia", "EG": "Egypt", "ES": "Spain", "FI": "Finland",
	"FR": "France", "GB": "United Kingdom", "GG": "Guernsey", "GI": "Gibraltar", "GR": "Greece", "GT": "Guatemala",
	"HK": "Hong Kong", "HN": "Honduras", "HR": "Croatia", "HU": "Hungary", "ID": "Indonesia", "IE": "Ireland",
	"IL": "Israel", "IM": "Isle of Man", "IN": "India", "IS": "Iceland", "IT": "Italy", "JE": "Jersey",
	"JM": "Jamaica", "JP": "Japan", "KR": "South Korea", "KY": "Cayman Islands", "LI": "Liechtenstein",
	"LT": "Lithuania", "LU": "Luxembourg", "LV": "Latvia", "MA": "Morocco", "MC": "Monaco", "MT": "Malta",
	"MX": "Mexico", "MY": "Malaysia", "NI": "Nicaragua", "NL": "Netherlands", "NO": "Norway", "NZ": "New Zealand",
	"PA": "Panama", "PE": "Peru", "PH": "Philippines", "PL": "Poland", "PR": "Puerto Rico", "PT": "Portugal",
	"PY": "Paraguay", "QA": "Qatar", "RO": "Romania", "RS": "Serbia", "SA": "Saudi Arabia", "SE": "Sweden",
	"SG": "Singapore", "SI": "Slovenia", "SK": "Slovakia", "SV": "El Salvador", "TH": "Thailand", "TR": "Türkiye",
	"TT": "Trinidad and Tobago", "TW": "Taiwan", "UA": "Ukraine", "US": "USA", "UY": "Uruguay", "VE": "Venezuela",
	"VG": "British Virgin Islands", "VN": "Vietnam", "ZA": "South Africa",
}

// CountryName returns an English name for an ISO 3166-1 alpha-2 code, or
// the code itself when unknown.
func CountryName(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if n, ok := countries[code]; ok {
		return n
	}
	return code
}
