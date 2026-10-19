package match

import "testing"

func TestLegalForm(t *testing.T) {
	cases := map[string][2]string{
		"Foo C.A. de C.V.":       {"C.A. de C.V.", "Foo"},
		"Foo Co., Ltda":          {"Ltda", "Foo Co."},
		"Foo GmbH":               {"GmbH", "Foo"},
		"Foo S.A.S.":             {"S.A.S.", "Foo"},
		"Foo Corp.":              {"Corp.", "Foo"},
		"Foo S.A.E":              {"S.A.E", "Foo"},
		"Foo C.A.":               {"C.A.", "Foo"},
		"Foo S.A.":               {"S.A.", "Foo"},
		"Foo Inc":                {"Inc", "Foo"},
		"Foo Ltd":                {"Ltd", "Foo"},
		"Foo Ltda":               {"Ltda", "Foo"},
		"Müller GmbH & Co. KG":   {"GmbH & Co. KG", "Müller"},
		"Foo S. de R.L. de C.V.": {"S. de R.L. de C.V.", "Foo"},
		"Sound of Me":            {"", "Sound of Me"},
		"Atlas":                  {"", "Atlas"},
		"GmbH":                   {"", "GmbH"},
		"Foo SA":                 {"SA", "Foo"},
		// Spaces in abbreviations, capitals, tails after the form.
		"FOO S. A.":                       {"S. A.", "FOO"},
		"Foo S A":                         {"S A", "Foo"},
		"FOO, S. A. DE C. V.":             {"S. A. de C. V.", "FOO"},
		"FOO GMBH":                        {"GmbH", "FOO"},
		"Foo gmbh":                        {"GmbH", "Foo"},
		"FOO LTDA.":                       {"Ltda.", "FOO"},
		"Foo GmbH (Niederlassung Wien)":   {"GmbH", "Foo"},
		"Foo AG - Zweigniederlassung Zug": {"AG", "Foo"},
		"Foo S.A. (Panamá)":               {"S.A.", "Foo"},
		"Foo C.A. en liquidación":         {"C.A.", "Foo"},
		"Foo (Pty) Ltd":                   {"(Pty) Ltd", "Foo"},
		"FOO (PTY) LTD":                   {"(Pty) Ltd", "FOO"},
		"Foo Holdings LLC":                {"LLC", "Foo Holdings"},
		"Foo, Inc.":                       {"Inc.", "Foo"},
		"Foo Nigeria Ltd/Gte":             {"Ltd/Gte", "Foo Nigeria"},
		"Foo, Lda.":                       {"Lda.", "Foo"},
		"Foo Unipessoal Lda":              {"Unipessoal Lda", "Foo"},
		"Foo SpA":                         {"SpA", "Foo"},
		"Foo Cía. Ltda.":                  {"Cía. Ltda.", "Foo"},
		"Foo S.L.":                        {"S.L.", "Foo"},
		"Foo Gesellschaft mit beschränkter Haftung": {"Gesellschaft mit beschränkter Haftung", "Foo"},
		"Foo SARL":    {"SARL", "Foo"},
		"Foo GesmbH":  {"GesmbH", "Foo"},
		"Foo e.U.":    {"e.U.", "Foo"},
		"Agua Sana":   {"", "Agua Sana"},
		"Casa Grande": {"", "Casa Grande"},
	}
	for in, want := range cases {
		f, b := LegalForm(in)
		if f != want[0] || b != want[1] {
			t.Errorf("LegalForm(%q) = %q, %q; want %q, %q", in, f, b, want[0], want[1])
		}
	}
}

func TestLegalFormCountries(t *testing.T) {
	for name, want := range map[string]string{
		"Foo GmbH": "DE", "Foo S.A. de C.V.": "MX", "Foo (Pty) Ltd": "ZA", "Foo LLC": "US", "Foo Lda.": "PT", "Foo Ltda": "BR", "Foo S.L.": "ES",
	} {
		m, ok := FindLegalForm(name)
		if !ok || len(m.Countries) == 0 || m.Countries[0] != want {
			t.Errorf("FindLegalForm(%q) countries = %v, want %s first", name, m.Countries, want)
		}
	}
}

func TestFind(t *testing.T) {
	ix := &Index{}
	for _, n := range []string{"Foo GmbH", "FOO S.A.S", "Fooo Inc", "Müller AG", "Foo", "D00001 - Bar Ltd"} {
		ix.Add("/x/"+n, n, n)
	}
	exact := func(name, want string) {
		t.Helper()
		ex, _ := ix.Find(name)
		if len(ex) == 0 || ex[0].Name != want {
			t.Errorf("Find(%q) exact = %v, want %q", name, ex, want)
		}
	}
	exact("foo gmbh", "Foo GmbH")
	exact("Foo S.A.S.", "FOO S.A.S")
	exact("Mueller AG", "Müller AG")
	exact("Muller A.G.", "Müller AG")

	// Different legal forms are never offered.
	_, fz := ix.Find("Foo S.A.")
	for _, c := range fz {
		if c.Folder.Name == "Foo GmbH" || c.Folder.Name == "FOO S.A.S" {
			t.Errorf("Foo S.A. offered %q", c.Folder.Name)
		}
	}
	// Close spellings and missing forms are offered for confirmation.
	want := map[string]string{"Foo Inc": "Fooo Inc", "Foo Ltda": "Foo", "Bar Ltd": "D00001 - Bar Ltd"}
	for name, folder := range want {
		ex, fz := ix.Find(name)
		found := false
		for _, c := range fz {
			if c.Folder.Name == folder {
				found = true
			}
		}
		if len(ex) > 0 || !found {
			t.Errorf("Find(%q): exact %v fuzzy %v; want fuzzy %q", name, ex, fz, folder)
		}
	}
}
