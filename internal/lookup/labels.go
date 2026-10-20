package lookup

import (
	"sort"
	"strings"
)

// roleLabels maps labels found on legal notices and in registers to an
// English role. Labels are written folded (lower case, no accents, words
// separated by single spaces, hyphens as spaces). An empty role marks a
// label that names people who are NOT authorized representatives, so the
// text after it is skipped.
var roleLabels = map[string]string{
	// German (Germany, Austria, Switzerland).
	"geschaftsfuhrer": "Managing Director", "geschaftsfuhrerin": "Managing Director", "geschaftsfuhrerinnen": "Managing Director",
	"geschaftsfuhrer in": "Managing Director", "geschaftsfuhrer innen": "Managing Director", "geschaeftsfuehrer": "Managing Director",
	"geschaeftsfuehrerin": "Managing Director", "geschaftsfuhrung": "Managing Director", "geschaeftsfuehrung": "Managing Director",
	"geschaftsfuhrende gesellschafter": "Managing Partner", "geschaftsfuhrender gesellschafter": "Managing Partner",
	"geschaftsfuhrende gesellschafterin": "Managing Partner", "geschaftsfuhrende gesellschafter innen": "Managing Partner",
	"kaufmannischer geschaftsfuhrer": "Managing Director (Commercial)", "technischer geschaftsfuhrer": "Managing Director (Technical)",
	"hauptgeschaftsfuhrer": "Chief Executive", "alleiniger geschaftsfuhrer": "Sole Managing Director",
	"vertretungsberechtigter geschaftsfuhrer": "Managing Director", "vertretungsberechtigte geschaftsfuhrer": "Managing Director",
	"vertretungsberechtigte geschaftsfuhrerin": "Managing Director", "vertretungsberechtigte geschaftsfuhrung": "Managing Director",
	"vertretungsberechtigt": "Legal Representative", "vertretungsberechtigte": "Legal Representative", "vertretungsberechtigter": "Legal Representative",
	"vertretungsberechtigte person": "Legal Representative", "vertretungsberechtigte personen": "Legal Representative",
	"vertretungsberechtigter vorstand": "Member of the Management Board", "vertreten durch": "Legal Representative",
	"gesetzlich vertreten durch": "Legal Representative", "gesetzliche vertreter": "Legal Representative", "gesetzlicher vertreter": "Legal Representative",
	"vorstand": "Member of the Management Board", "vorstande": "Member of the Management Board", "der vorstand": "Member of the Management Board",
	"vorstandsmitglied": "Member of the Management Board", "vorstandsmitglieder": "Member of the Management Board",
	"mitglieder des vorstands": "Member of the Management Board", "mitglied des vorstands": "Member of the Management Board",
	"vorstandsvorsitzender": "Chairman of the Management Board", "vorstandsvorsitzende": "Chairwoman of the Management Board",
	"vorsitzender des vorstands": "Chairman of the Management Board", "vorsitzende des vorstands": "Chairwoman of the Management Board",
	"vorsitzender des vorstandes": "Chairman of the Management Board", "vorstandssprecher": "Spokesperson of the Management Board",
	"sprecher des vorstands": "Spokesperson of the Management Board", "prokurist": "Authorized Signatory (Prokurist)",
	"prokuristin": "Authorized Signatory (Prokurist)", "prokuristen": "Authorized Signatory (Prokurist)", "prokura": "Authorized Signatory (Prokurist)",
	"gesamtprokura": "Authorized Signatory (Prokurist)", "einzelprokura": "Authorized Signatory (Prokurist)",
	"inhaber": "Owner", "inhaberin": "Owner", "inhaber in": "Owner", "komplementar": "General Partner", "komplementarin": "General Partner",
	"personlich haftende gesellschafterin": "General Partner", "personlich haftender gesellschafter": "General Partner",
	"liquidator": "Liquidator", "liquidatorin": "Liquidator", "liquidatoren": "Liquidator",
	"geschaftsleitung": "Member of the Executive Management", "geschaftsleiter": "Executive Manager", "geschaftsleiterin": "Executive Manager",
	"mitglied der geschaftsleitung": "Member of the Executive Management", "mitglieder der geschaftsleitung": "Member of the Executive Management",
	"verwaltungsrat": "Member of the Board of Directors", "verwaltungsrate": "Member of the Board of Directors",
	"mitglied des verwaltungsrates": "Member of the Board of Directors", "mitglieder des verwaltungsrates": "Member of the Board of Directors",
	"prasident des verwaltungsrates": "Chairman of the Board of Directors", "prasident des verwaltungsrats": "Chairman of the Board of Directors",
	"verwaltungsratsprasident": "Chairman of the Board of Directors", "verwaltungsratsprasidentin": "Chairwoman of the Board of Directors",
	"delegierter des verwaltungsrates": "Delegate of the Board of Directors", "direktor": "Director", "direktorin": "Director",
	"zeichnungsberechtigt": "Authorized Signatory", "zeichnungsberechtigte": "Authorized Signatory", "zeichnungsberechtigter": "Authorized Signatory",
	// German labels that name other people.
	"aufsichtsrat": "", "aufsichtsrate": "", "aufsichtsratsvorsitzender": "", "aufsichtsratsvorsitzende": "",
	"vorsitzender des aufsichtsrats": "", "vorsitzende des aufsichtsrats": "", "vorsitzender des aufsichtsrates": "",
	"vorsitzende des aufsichtsrates": "", "beirat": "", "datenschutzbeauftragter": "", "datenschutzbeauftragte": "",
	"verantwortlich": "", "verantwortlicher": "", "inhaltlich verantwortlich": "", "redaktion": "", "redaktionell verantwortlich": "",
	"ansprechpartner": "", "ansprechpartnerin": "", "webmaster": "", "konzept": "", "gestaltung": "", "bildnachweis": "", "fotos": "",
	// English.
	"managing director": "Managing Director", "managing directors": "Managing Director", "director": "Director", "directors": "Director",
	"board of directors": "Member of the Board of Directors", "executive director": "Executive Director", "executive directors": "Executive Director",
	"sole director": "Sole Director", "chief executive officer": "Chief Executive Officer", "ceo": "Chief Executive Officer",
	"chief financial officer": "Chief Financial Officer", "cfo": "Chief Financial Officer", "chief operating officer": "Chief Operating Officer",
	"coo": "Chief Operating Officer", "president": "President", "vice president": "Vice President", "chairman": "Chairman",
	"chairwoman": "Chairwoman", "chairperson": "Chairperson", "chairman of the board": "Chairman of the Board",
	"company secretary": "Company Secretary", "secretary": "Secretary", "treasurer": "Treasurer", "general manager": "General Manager",
	"managing partner": "Managing Partner", "managing partners": "Managing Partner", "general partner": "General Partner",
	"owner": "Owner", "owners": "Owner", "proprietor": "Owner", "sole proprietor": "Owner", "legal representative": "Legal Representative",
	"legal representatives": "Legal Representative", "authorized representative": "Authorized Representative",
	"authorized representatives": "Authorized Representative", "authorised representative": "Authorized Representative",
	"authorised representatives": "Authorized Representative", "authorized signatory": "Authorized Signatory",
	"authorised signatory": "Authorized Signatory", "authorized officer": "Authorized Officer", "represented by": "Legal Representative",
	"legally represented by": "Legal Representative", "management": "Management", "management board": "Member of the Management Board",
	"executive board": "Member of the Executive Board", "board of management": "Member of the Management Board",
	"managing board": "Member of the Management Board", "officers": "Officer", "manager": "Manager", "managers": "Manager",
	"managing member": "Managing Member", "member manager": "Managing Member", "trustee": "Trustee", "trustees": "Trustee",
	// English labels that name other people.
	"data protection officer": "", "dpo": "", "editor": "", "responsible for content": "", "responsible for the content": "",
	"contact": "", "contact person": "", "photo credits": "", "registered agent": "", "non executive director": "",
	"non executive directors": "", "supervisory board": "", "chairman of the supervisory board": "", "founder": "", "founders": "",
	// Spanish.
	"representante legal": "Legal Representative", "representantes legales": "Legal Representative", "representante": "Legal Representative",
	"representante legal suplente": "Alternate Legal Representative", "suplente del representante legal": "Alternate Legal Representative",
	"representado por": "Legal Representative", "representada por": "Legal Representative", "representada legalmente por": "Legal Representative",
	"gerente": "Manager", "gerentes": "Manager", "gerente general": "General Manager", "gerente legal": "Legal Manager",
	"administrador": "Administrator", "administradora": "Administrator", "administradores": "Administrator",
	"administrador unico": "Sole Administrator", "administradora unica": "Sole Administrator", "adm unico": "Sole Administrator",
	"administrador solidario": "Joint and Several Administrator", "administradores solidarios": "Joint and Several Administrator",
	"adm solid": "Joint and Several Administrator", "administrador mancomunado": "Joint Administrator",
	"administradores mancomunados": "Joint Administrator", "adm mancom": "Joint Administrator",
	"consejero delegado": "Chief Executive Officer (Managing Director)", "consejera delegada": "Chief Executive Officer (Managing Director)",
	"consejeros delegados": "Chief Executive Officer (Managing Director)", "cons del": "Chief Executive Officer (Managing Director)",
	"consejero": "Member of the Board of Directors", "consejera": "Member of the Board of Directors", "consejeros": "Member of the Board of Directors",
	"consejo de administracion": "Member of the Board of Directors", "presidente": "President", "presidenta": "President",
	"presidente del consejo": "Chairman of the Board", "presidente del consejo de administracion": "Chairman of the Board",
	"presidente del directorio": "Chairman of the Board", "presidente de la junta directiva": "Chairman of the Board",
	"presidente ejecutivo": "Executive President", "vicepresidente": "Vice President", "vicepresidenta": "Vice President",
	"directorio": "Member of the Board of Directors", "director general": "General Director", "directora general": "General Director",
	"director ejecutivo": "Executive Director", "directora ejecutiva": "Executive Director", "director gerente": "Managing Director",
	"apoderado": "Attorney-in-fact", "apoderada": "Attorney-in-fact", "apoderados": "Attorney-in-fact", "apoderado general": "General Attorney-in-fact",
	"apoderado legal": "Legal Attorney-in-fact", "apoderado mancomunado": "Joint Attorney-in-fact", "apoderado solidario": "Joint and Several Attorney-in-fact",
	"apo manc": "Joint Attorney-in-fact", "apo sol": "Joint and Several Attorney-in-fact", "socio gerente": "Managing Partner",
	"socios gerentes": "Managing Partner", "socio administrador": "Managing Partner", "socia administradora": "Managing Partner",
	"vocal": "Board Member", "vocales": "Board Member", "secretario": "Secretary", "secretario del consejo": "Secretary of the Board",
	"tesorero": "Treasurer", "tesorera": "Treasurer", "titular": "Owner", "propietario": "Owner", "propietaria": "Owner",
	"liquidador": "Liquidator", "liquidadora": "Liquidator", "responsable legal": "Legal Representative",
	// Spanish labels that name other people.
	"comisario": "", "comisarios": "", "auditor": "", "auditores": "", "contacto": "", "delegado de proteccion de datos": "",
	// Portuguese.
	"socio administrador s": "Managing Partner", "socios administradores": "Managing Partner", "socio gerente s": "Managing Partner",
	"socia gerente": "Managing Partner", "gerencia": "Management", "diretor": "Director", "diretora": "Director", "diretores": "Director",
	"diretor presidente": "Chief Executive Officer", "diretora presidente": "Chief Executive Officer", "diretor geral": "General Director",
	"diretora geral": "General Director", "diretor executivo": "Executive Director", "diretora executiva": "Executive Director",
	"diretor superintendente": "Superintendent Director", "diretoria": "Member of the Executive Board",
	"presidente do conselho de administracao": "Chairman of the Board", "conselho de administracao": "Member of the Board of Directors",
	"membro do conselho de administracao": "Member of the Board of Directors", "vogal": "Board Member", "vogais": "Board Member",
	"procurador": "Attorney-in-fact", "procuradora": "Attorney-in-fact", "procuradores": "Attorney-in-fact",
	"responsavel legal": "Legal Representative", "empresario": "Owner", "representado legalmente por": "Legal Representative",
	"encarregado de dados": "", "encarregado": "",
	// French.
	"president directeur general": "Chairman and Chief Executive Officer", "pdg": "Chairman and Chief Executive Officer",
	"directeur general": "Chief Executive Officer", "directrice generale": "Chief Executive Officer",
	"directeur general delegue": "Deputy Chief Executive Officer", "directrice generale deleguee": "Deputy Chief Executive Officer",
	"gerant": "Manager", "gerante": "Manager", "gerants": "Manager", "co gerant": "Co-Manager", "cogerant": "Co-Manager",
	"administrateur": "Director", "administrateurs": "Director", "president du conseil d administration": "Chairman of the Board",
	"representant legal": "Legal Representative", "represente par": "Legal Representative", "representee par": "Legal Representative",
	"directeur de la publication": "Publication Director (Legal Representative)", "directrice de la publication": "Publication Director (Legal Representative)",
	"responsable de la publication": "Publication Director (Legal Representative)", "mandataire social": "Corporate Officer",
	"hebergeur": "", "delegue a la protection des donnees": "",
	// Italian.
	"amministratore delegato": "Chief Executive Officer (Managing Director)", "amministratore unico": "Sole Director",
	"amministratori": "Director", "amministratore": "Director", "legale rappresentante": "Legal Representative",
	"rappresentante legale": "Legal Representative", "presidente del consiglio di amministrazione": "Chairman of the Board",
	"consiglio di amministrazione": "Member of the Board of Directors", "direttore generale": "General Director",
	"procuratore": "Attorney-in-fact", "procuratori": "Attorney-in-fact",
	// Dutch.
	"directeur": "Director", "directie": "Management", "bestuurder": "Director", "bestuurders": "Director", "bestuur": "Board Member",
	"algemeen directeur": "General Director", "statutair directeur": "Statutory Director", "statutair bestuurder": "Statutory Director",
	"vertegenwoordigd door": "Legal Representative", "gevolmachtigde": "Authorized Signatory", "eigenaar": "Owner",
	// Polish, Czech, Slovak.
	"prezes zarzadu": "President of the Management Board", "wiceprezes zarzadu": "Vice President of the Management Board",
	"czlonek zarzadu": "Member of the Management Board", "czlonkowie zarzadu": "Member of the Management Board",
	"zarzad": "Member of the Management Board", "prokurent": "Authorized Signatory (Prokurist)", "wlasciciel": "Owner",
	"jednatel": "Managing Director", "jednatele": "Managing Director", "jednatelka": "Managing Director", "konatel": "Managing Director",
	"statutarni organ": "Statutory Body", "predseda predstavenstva": "Chairman of the Board of Directors",
	"clen predstavenstva": "Member of the Board of Directors", "prokurista": "Authorized Signatory (Prokurist)",
	// Nordic.
	"verkstallande direktor": "Chief Executive Officer", "vd": "Chief Executive Officer", "styrelseordforande": "Chairman of the Board",
	"administrerende direktor": "Chief Executive Officer", "daglig leder": "Chief Executive Officer", "toimitusjohtaja": "Chief Executive Officer",
}

// labelIndex maps the first word of each label to the label word lists
// that start with it, longest first.
var labelIndex = map[string][][]string{}

func init() {
	for l := range roleLabels {
		w := strings.Fields(l)
		labelIndex[w[0]] = append(labelIndex[w[0]], w)
	}
	for k := range labelIndex {
		sort.Slice(labelIndex[k], func(i, j int) bool {
			if len(labelIndex[k][i]) != len(labelIndex[k][j]) {
				return len(labelIndex[k][i]) > len(labelIndex[k][j])
			}
			return strings.Join(labelIndex[k][i], " ") < strings.Join(labelIndex[k][j], " ")
		})
	}
}

// weakLabel lists one-word labels that are also ordinary words; they
// count only when followed by a colon or alone on their line.
var weakLabel = map[string]bool{
	"director": true, "directors": true, "manager": true, "managers": true, "president": true, "secretary": true,
	"treasurer": true, "chairman": true, "chairwoman": true, "chairperson": true, "chair": true, "owner": true, "owners": true,
	"officers": true, "management": true, "trustee": true, "trustees": true, "proprietor": true, "ceo": true, "cfo": true, "coo": true,
	"gerente": true, "gerentes": true, "administrador": true, "administradora": true, "administradores": true, "presidente": true,
	"presidenta": true, "representante": true, "titular": true, "propietario": true, "consejero": true, "consejera": true,
	"consejeros": true, "vocal": true, "secretario": true, "tesorero": true, "diretor": true, "diretora": true, "diretores": true,
	"gerencia": true, "directeur": true, "gerant": true, "administrateur": true, "amministratore": true, "bestuur": true,
	"bestuurder": true, "directie": true, "direktor": true, "direktorin": true, "inhaber": true, "inhaberin": true, "eigenaar": true,
	"zarzad": true, "prokura": true, "directorio": true, "diretoria": true, "procurador": true, "apoderado": true, "liquidator": true,
	"liquidador": true,
}

// byLabel reports whether a label ends with "by" in some language, so
// the names follow without a colon.
func byLabel(last string) bool {
	switch last {
	case "durch", "by", "por", "par", "door", "da":
		return true
	}
	return false
}
