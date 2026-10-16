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
	}
	for in, want := range cases {
		f, b := LegalForm(in)
		if f != want[0] || b != want[1] {
			t.Errorf("LegalForm(%q) = %q, %q; want %q, %q", in, f, b, want[0], want[1])
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
