package lookup

import (
	"strings"
	"unicode"
)

// roleMap translates common German, Spanish, Portuguese, French and
// English role names to English. Keys are lower case, without accents.
var roleMap = map[string]string{
	// German.
	"geschaftsfuhrer": "Managing Director", "geschaftsfuhrerin": "Managing Director",
	"geschaftsfuhrende gesellschafterin": "Managing Partner", "geschaftsfuhrender gesellschafter": "Managing Partner",
	"geschaftsfuhrung": "Managing Director",
	"vorstand":         "Member of the Management Board", "vorstandsmitglied": "Member of the Management Board",
	"vorstandsvorsitzender": "Chairman of the Management Board", "vorstandsvorsitzende": "Chairwoman of the Management Board",
	"vorsitzender des vorstands": "Chairman of the Management Board", "vorsitzende des vorstands": "Chairwoman of the Management Board",
	"prokurist": "Authorized Signatory (Prokurist)", "prokuristin": "Authorized Signatory (Prokurist)",
	"inhaber": "Owner", "inhaberin": "Owner",
	"personlich haftender gesellschafter": "General Partner", "komplementar": "General Partner",
	"liquidator": "Liquidator", "liquidatorin": "Liquidator",
	"vertreten durch": "Legal Representative", "vertretungsberechtigt": "Legal Representative",
	// Spanish.
	"gerente general": "General Manager", "gerente": "Manager", "representante legal": "Legal Representative",
	"suplente del representante legal": "Alternate Legal Representative", "representante legal suplente": "Alternate Legal Representative",
	"administrador unico": "Sole Administrator", "administradora unica": "Sole Administrator",
	"administrador": "Administrator", "administradora": "Administrator",
	"administrador solidario": "Joint and Several Administrator", "administrador mancomunado": "Joint Administrator",
	"consejero delegado": "Chief Executive Officer (Managing Director)", "consejera delegada": "Chief Executive Officer (Managing Director)",
	"director general": "General Director", "directora general": "General Director",
	"director ejecutivo": "Executive Director", "directora ejecutiva": "Executive Director",
	"presidente": "President", "presidenta": "President",
	"presidente del directorio": "Chairman of the Board", "presidente de la junta directiva": "Chairman of the Board",
	"apoderado": "Attorney-in-fact", "apoderada": "Attorney-in-fact", "apoderado general": "General Attorney-in-fact",
	"vicepresidente": "Vice President", "secretario": "Secretary", "tesorero": "Treasurer",
	// Portuguese.
	"diretor": "Director", "diretora": "Director", "diretor presidente": "Chief Executive Officer",
	"socio administrador": "Managing Partner", "socia administradora": "Managing Partner",
	// French.
	"president": "President", "directeur general": "Chief Executive Officer", "gerant": "Manager", "gerante": "Manager",
	// English.
	"managing director": "Managing Director", "director": "Director", "general manager": "General Manager",
	"legal representative": "Legal Representative", "chief executive officer": "Chief Executive Officer",
	"ceo": "Chief Executive Officer", "chairman": "Chairman", "chairperson": "Chairperson", "secretary": "Secretary",
	"company secretary": "Company Secretary", "represented by": "Legal Representative",
	"authorized representative": "Authorized Representative", "authorised representative": "Authorized Representative",
	"owner": "Owner", "partner": "Partner", "managing partner": "Managing Partner",
}

var roleFold = map[rune]string{'ä': "a", 'ö': "o", 'ü': "u", 'ß': "ss", 'á': "a", 'é': "e", 'í': "i", 'ó': "o", 'ú': "u", 'ñ': "n", 'ç': "c", 'â': "a", 'ê': "e", 'ô': "o", 'ã': "a", 'õ': "o", 'è': "e", 'à': "a"}

func roleKey(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if v, ok := roleFold[r]; ok {
			b.WriteString(v)
			continue
		}
		if unicode.IsLetter(r) || r == ' ' {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
