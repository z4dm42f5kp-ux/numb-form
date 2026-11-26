// The legal forms that FindLegalForm recognises.

package match

// legalForms lists conventional spellings. Matching tolerates missing or
// extra dots and spaces and any letter case; the text returned keeps the
// punctuation written in the name but uses conventional capitals.
var legalForms = []formDef{
	// German speaking.
	{"GmbH & Co. KGaA", []string{"DE"}}, {"GmbH & Co. KG", dach}, {"GmbH & Co. OHG", []string{"DE"}},
	{"SE & Co. KGaA", []string{"DE"}}, {"SE & Co. KG", []string{"DE"}}, {"AG & Co. KGaA", []string{"DE"}}, {"AG & Co. KG", []string{"DE"}},
	{"UG (haftungsbeschränkt) & Co. KG", []string{"DE"}}, {"UG (haftungsbeschränkt)", []string{"DE"}},
	{"Gesellschaft mit beschränkter Haftung", dach}, {"Aktiengesellschaft", dach}, {"Kommanditgesellschaft", dach},
	{"gGmbH", []string{"DE"}}, {"GmbH", dach}, {"GesmbH", []string{"AT"}}, {"Ges.m.b.H.", []string{"AT"}}, {"mbH", []string{"DE"}},
	{"KGaA", []string{"DE"}}, {"AG", dach}, {"KG", dach}, {"OHG", []string{"DE"}}, {"OG", []string{"AT"}}, {"GbR", []string{"DE"}},
	{"e.K.", []string{"DE"}}, {"e.U.", []string{"AT"}}, {"e.V.", []string{"DE"}}, {"eG", []string{"DE"}}, {"SE", []string{"DE", "AT", "FR", "NL", "LU"}},
	{"Sàrl", []string{"CH", "LU"}}, {"Genossenschaft", []string{"CH", "DE", "AT"}},
	// Spanish and Portuguese speaking.
	{"S.A.P.I. de C.V.", []string{"MX"}}, {"S.A.P.I.B. de C.V.", []string{"MX"}}, {"S.A.B. de C.V.", []string{"MX"}},
	{"S. de R.L. de C.V.", []string{"MX"}}, {"S. de R.L. de C.V.", []string{"MX"}}, {"S. en C. de C.V.", []string{"MX"}},
	{"S.C. de R.L. de C.V.", []string{"MX"}}, {"S.A. de C.V.", []string{"MX", "SV", "HN"}}, {"C.A. de C.V.", []string{"MX", "VE"}},
	{"S. de R.L.", []string{"MX", "HN", "PA", "GT"}}, {"S. en C.S.", []string{"CO"}}, {"S. en C.", []string{"CO", "MX"}},
	{"S.A.E.C.A.", []string{"PY"}}, {"S.A.I.C.A.", []string{"AR"}}, {"S.A.C.I.F.", []string{"AR"}}, {"S.A.I.C.", []string{"AR"}}, {"S.A.C.I.", []string{"AR", "CL", "PY"}},
	{"S.A.C.I.F.I.A.", []string{"AR"}}, {"S.A.C.I.F.I.", []string{"AR"}}, {"S.A.I.C.F.", []string{"AR"}}, {"S.A.C.I.A.", []string{"AR", "PY"}},
	{"S.A.I.C.F.I.A.", []string{"AR"}}, {"S.A.I.C.I.", []string{"AR"}}, {"S.A.C.", []string{"PE"}}, {"C.I.S.A.", []string{"PY"}}, {"S.A.C.I.A.G.", []string{"AR"}},
	{"S.R.L. de C.V.", []string{"MX"}}, {"S.A.S. BIC", []string{"CO"}}, {"S.A. BIC", []string{"CO"}}, {"S.A.S. E.S.P.", []string{"CO"}}, {"S.A. E.S.P.", []string{"CO"}},
	{"E.S.P.", []string{"CO"}}, {"C. Ltda.", []string{"EC"}}, {"Limitada", []string{"CO", "CL", "BR", "EC"}}, {"Compañía Limitada", []string{"EC", "CO"}},
	{"Sociedad por Acciones Simplificada", []string{"CO"}}, {"Sociedad Anónima Cerrada", []string{"PE"}}, {"Sociedad Comercial de Responsabilidad Limitada", []string{"PE"}},
	{"Empresa Individual de Responsabilidad Limitada", []string{"PE", "CL"}}, {"Sociedad en Comandita Simple", hispanic}, {"Sociedad Colectiva", hispanic},
	{"S.A.P.I.", []string{"MX"}}, {"S.A.S.U.", []string{"FR"}}, {"S.A.S.", []string{"CO", "FR", "AR"}}, {"S.A.C.", []string{"PE"}}, {"S.A.A.", []string{"PE"}},
	{"S.A.E.", []string{"EG", "ES"}}, {"S.A.U.", []string{"ES", "AR"}}, {"S.L.U.", []string{"ES"}}, {"S.L.N.E.", []string{"ES"}}, {"S.L.L.", []string{"ES"}}, {"S.L.P.", []string{"ES"}},
	{"S.Coop.", []string{"ES"}}, {"S.C.A.", []string{"ES", "AR", "BE", "LU"}}, {"S.C.S.", []string{"AR", "BE", "LU"}}, {"S.C.R.L.", []string{"PE", "BE"}},
	{"E.I.R.L.", []string{"PE", "CL"}}, {"S.R.L.", []string{"AR", "PE", "BO", "UY", "PY", "IT", "RO", "DO", "CR", "BE"}}, {"SpA", []string{"CL", "IT"}}, {"S.p.A.", []string{"IT", "CL"}},
	{"S.r.l.", []string{"IT"}}, {"S.L.", []string{"ES"}}, {"S.C.", []string{"MX", "PE"}}, {"A.C.", []string{"MX"}}, {"S/A", []string{"BR"}},
	{"Sociedad Anónima de Capital Variable", []string{"MX", "SV"}}, {"Sociedad Anónima", hispanic}, {"Sociedad Limitada", []string{"ES"}},
	{"Sociedad de Responsabilidad Limitada", []string{"MX", "AR", "PE"}}, {"Sociedade Anónima", []string{"PT"}}, {"Sociedade Anônima", []string{"BR"}},
	{"Sociedade Limitada", []string{"BR"}}, {"Compañía Anónima", []string{"VE", "EC"}}, {"Companhia", []string{"BR"}},
	{"S.A.", latamSA}, {"C.A.", []string{"VE", "EC"}}, {"Cía. Ltda.", []string{"EC", "CO"}}, {"Cia. Ltda.", []string{"BR", "EC"}}, {"Ltda.", []string{"BR", "CO", "CL", "EC"}},
	{"Unipessoal Lda.", []string{"PT"}}, {"Lda.", []string{"PT"}}, {"EIRELI", []string{"BR"}}, {"S.G.P.S., S.A.", []string{"PT"}}, {"SGPS", []string{"PT"}},
	// French, Italian, Benelux, Nordic, Central and Eastern Europe.
	{"S.A.R.L.U.", fr}, {"SARLU", fr}, {"S.A.R.L.", fr}, {"G.I.E.", fr}, {"Société Anonyme", fr}, {"Société à Responsabilité Limitée", fr},
	{"Société par Actions Simplifiée", fr}, {"SUARL", []string{"SN", "CI", "CM"}}, {"S.à r.l.", []string{"LU"}}, {"EURL", []string{"FR"}}, {"SNC", fr},
	{"SA/NV", []string{"BE"}}, {"BV/SRL", []string{"BE"}}, {"B.V.", []string{"NL", "BE"}}, {"N.V.", []string{"NL", "BE", "CW"}}, {"BVBA", []string{"BE"}}, {"V.O.F.", []string{"NL"}},
	{"A/S", []string{"DK", "NO"}}, {"ApS", []string{"DK"}}, {"AB (publ)", []string{"SE"}}, {"AB", []string{"SE"}}, {"Oyj", []string{"FI"}}, {"Oy", []string{"FI"}}, {"ASA", []string{"NO"}}, {"AS", []string{"NO", "EE"}},
	{"Sp. z o.o.", []string{"PL"}}, {"S.A.S.", []string{"FR"}}, {"s.r.o.", []string{"CZ", "SK"}}, {"a.s.", []string{"CZ", "SK"}}, {"Kft.", []string{"HU"}}, {"Zrt.", []string{"HU"}}, {"Nyrt.", []string{"HU"}},
	{"d.o.o.", []string{"SI", "HR", "RS", "BA"}}, {"d.d.", []string{"SI", "HR"}}, {"OÜ", []string{"EE"}}, {"UAB", []string{"LT"}}, {"SIA", []string{"LV"}}, {"A.Ş.", []string{"TR"}},
	// English speaking, Africa and Asia.
	{"RF (Pty) Ltd", []string{"ZA"}}, {"(Pty) Ltd", []string{"ZA", "BW", "NA"}}, {"(Proprietary) Limited", []string{"ZA", "BW", "NA"}}, {"Proprietary Limited", []string{"ZA", "AU"}},
	{"Pty Ltd", []string{"AU", "ZA"}}, {"Pty. Ltd.", []string{"AU", "ZA"}}, {"Pvt. Ltd.", []string{"IN", "PK", "LK"}}, {"Private Limited", []string{"IN", "SG"}}, {"Pte. Ltd.", []string{"SG"}},
	{"Sdn. Bhd.", []string{"MY"}}, {"Bhd.", []string{"MY"}}, {"Co., Ltd.", []string{"CN", "JP", "KR", "TH", "VN", "TW"}}, {"K.K.", []string{"JP"}},
	{"Ltd/Gte", []string{"NG"}}, {"NPC", []string{"ZA"}}, {"Ltd. by Guarantee", []string{"NG", "GB"}},
	// Gulf and free zones.
	{"FZ-LLC", []string{"AE"}}, {"FZ LLC", []string{"AE"}}, {"LLC-FZ", []string{"AE"}}, {"L.L.C-FZ", []string{"AE"}}, {"FZCO", []string{"AE"}},
	{"FZE", []string{"AE"}}, {"FZC", []string{"AE"}}, {"DMCC", []string{"AE"}}, {"W.L.L.", []string{"BH", "QA", "KW"}}, {"P.J.S.C.", []string{"AE"}},
	{"P.S.C.", []string{"AE"}}, {"J.S.C.", []string{"AE", "SA", "RU"}}, {"S.A.O.G.", []string{"OM"}}, {"S.A.O.C.", []string{"OM"}}, {"S.P.C.", []string{"OM", "BH"}},
	{"K.S.C.C.", []string{"KW"}}, {"K.S.C.P.", []string{"KW"}}, {"B.S.C. (c)", []string{"BH"}}, {"Q.P.S.C.", []string{"QA"}}, {"Limited", english}, {"Ltd.", english}, {"PLC", english},
	{"PLLC", us}, {"L.L.C.", us}, {"LLC", us}, {"LLLP", us}, {"LLP", []string{"US", "GB"}}, {"L.P.", us}, {"LP", us}, {"P.C.", us}, {"N.A.", us},
	{"Incorporated", us}, {"Inc.", us}, {"Corporation", us}, {"Corp.", us}, {"Co.", us},
}

// ambiguous forms are short words that also occur in ordinary names; they
// are recognised only when written in capitals.
var ambiguous = map[string]bool{"ME": true, "EPP": true, "AS": true, "AB": true, "SE": true, "LP": true, "OG": true, "KG": true, "AG": true, "SA": true, "CA": true, "SC": true, "AC": true, "NA": true, "PC": true, "CO": true, "OY": true, "EG": true, "DD": true}
