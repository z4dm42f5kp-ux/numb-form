package match

import (
	"strings"
	"testing"
)

func TestFixPlaces(t *testing.T) {
	for in, want := range map[string]string{
		"Foo S.A de C.V., Ciudad de Me": "Foo S.A de C.V., Ciudad de México",
		"Foo SAS, Bogta, Colomiba":      "Foo SAS, Bogotá, Colombia",
		"Foo S.A., Quito - Ecudor":      "Foo S.A., Quito - Ecuador",
		"Foo LLC, Duabi":                "Foo LLC, Dubái",
		"FOO SA, GUAYAQIL":              "FOO SA, GUAYAQUIL",
		"Foo, Republica Dominic":        "Foo, República Dominicana",
		"Foo Ltd, Nairob":               "Foo Ltd, Nairobi",
		"Foo SRL, Santo Domngo.":        "Foo SRL, Santo Domingo.",
		"Colomiba Trading SAS, Bogta":   "Colomiba Trading SAS, Bogotá", // the name itself never changes
	} {
		if got := Apply(in, FixPlaces(in)); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
	// Known places, legal forms, short words and anything that is not a
	// place stay as written; accents are not added to correct names.
	for _, in := range []string{"Foo SAS, Bogota, Colombia", "Foo y Bar, S.A.", "Foo, Durán", "Foo, Cartago",
		"Foo, Hermanos", "Foo, Planta Sur", "Foo GmbH, Werk 2", "Foo, Lim", "Foo GmbH, Munchen", "Foo, Sucursal", "Foo Bogta"} {
		if fx := FixPlaces(in); len(fx) != 0 {
			t.Errorf("%q: unexpected %+v", in, fx)
		}
	}
	// Several places fit: the user chooses.
	fx := FixPlaces("Foo, Santa")
	if len(fx) != 1 || fx[0].To != "" || !strings.Contains(strings.Join(fx[0].Choices, "|"), "Santa Marta") {
		t.Errorf("Santa: %+v", fx)
	}
}

func TestDistance(t *testing.T) {
	for _, c := range []struct {
		a, b string
		d    int
	}{{"colomiba", "colombia", 1}, {"bogta", "bogota", 1}, {"abc", "abc", 0}, {"kitten", "sitting", 3}} {
		if got := distance(c.a, c.b); got != c.d {
			t.Errorf("distance(%q, %q) = %d, want %d", c.a, c.b, got, c.d)
		}
	}
}

func TestPlaceFixed(t *testing.T) {
	for _, c := range []struct {
		from, to string
		want     bool
	}{
		{"Foo SAS, Bogta, Colomiba", "Foo SAS, Bogotá, Colombia", true},
		{"Foo SAS, Bogta, Colomiba", "Foo SAS, Bogta, Colombia", true}, // a fix left out
		{"Foo, Santa", "Foo, Santa Marta", true},                       // one of several choices
		{"Foo SAS, Bogta, Colomiba", "Bar SAS, Bogotá, Colombia", false},
		{"Foo SAS, Bogta", "Foo SAS, Bogta", false},
		{"Foo GmbH", "Foo AG", false},
		{"Foo, Santa", "Foo, Santa Cruz de la Nada", false},
	} {
		if got := PlaceFixed(c.from, c.to); got != c.want {
			t.Errorf("PlaceFixed(%q, %q) = %v", c.from, c.to, got)
		}
	}
}

// Long or many-part cells are left alone, so a hostile spreadsheet cannot
// make numb-typo or numb-commit's review slow.
func TestFixPlacesBounds(t *testing.T) {
	long := "Foo" + strings.Repeat(", Bogta", maxParts)
	if fx := FixPlaces(long); fx != nil {
		t.Errorf("%d parts: %+v", maxParts+1, fx)
	}
	if PlaceFixed(long, strings.ReplaceAll(long, "Bogta", "Bogotá")) {
		t.Error("a cell beyond the bounds counts as a place correction")
	}
}

func FuzzFixPlaces(f *testing.F) {
	for _, s := range []string{"Foo SAS, Bogta, Colomiba", "Foo S.A de C.V., Ciudad de Me", "Foo, Santa", "Foo, Quito - Ecudor", "FOO SA, GUAYAQIL"} {
		f.Add(s, s)
	}
	f.Fuzz(func(t *testing.T, name, to string) {
		fixes := FixPlaces(name)
		for _, fx := range fixes {
			if fx.At < len(strings.Split(name, ",")[0]) || fx.At+len(fx.From) > len(name) || name[fx.At:fx.At+len(fx.From)] != fx.From {
				t.Fatalf("fix %+v does not point into %q", fx, name)
			}
		}
		if out := Apply(name, fixes); strings.Split(out, ",")[0] != strings.Split(name, ",")[0] {
			t.Fatalf("the name itself changed: %q -> %q", name, out)
		}
		_ = PlaceFixed(name, to)
	})
}
