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
