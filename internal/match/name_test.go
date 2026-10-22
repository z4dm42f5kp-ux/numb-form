package match

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseName(t *testing.T) {
	cases := []struct {
		in, core, form string
		places         []string
		cc             string
		confident      bool
	}{
		{"Acme SAS, Bogota, Colombia", "Acme", "SAS", []string{"Bogota", "Colombia"}, "CO", true},
		{"ETS Zorbu, Douala.", "Zorbu", "Ets", []string{"Douala"}, "CM", true},
		{"Zorbu Chile SPA,", "Zorbu Chile", "SpA", nil, "CL", true},
		{"Zorbu, S.A.", "Zorbu", "S.A.", nil, "ES", false},
		{"Zorbu Lab, SRL, Santo Domingo.", "Zorbu Lab", "SRL", []string{"Santo Domingo"}, "DO", true},
		{"Zorbu C.I.S.A, San Lorenzo", "Zorbu", "C.I.S.A", []string{"San Lorenzo"}, "PY", true},
		{"Zorbu Freight LLC, Dubai", "Zorbu Freight", "LLC", []string{"Dubai"}, "AE", true},
		{"Zorbu Ltda, Springfield", "Zorbu", "Ltda", []string{"Springfield"}, "BR", false},
		{"Zorbu", "Zorbu", "", nil, "", false},
	}
	for _, c := range cases {
		n := ParseName(c.in)
		if n.Core != c.core || n.Form != c.form || strings.Join(n.Places, "|") != strings.Join(c.places, "|") || n.Confident != c.confident {
			t.Errorf("ParseName(%q) = core %q form %q places %q confident %v; want %q %q %q %v",
				c.in, n.Core, n.Form, n.Places, n.Confident, c.core, c.form, c.places, c.confident)
		}
		if c.cc != "" && (len(n.Countries) == 0 || n.Countries[0] != c.cc) {
			t.Errorf("ParseName(%q) countries = %v, want %s first", c.in, n.Countries, c.cc)
		}
	}
}

func TestQueries(t *testing.T) {
	got := ParseName("Acme SAS, Bogota, Colombia").Queries()
	want := []string{"Acme SAS, Bogota, Colombia", `"Acme SAS"`, "Acme SAS Bogota Colombia", `"Acme" Colombia`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Queries = %q, want %q", got, want)
	}
	// An ambiguous form gives no country search.
	for _, q := range ParseName("Zorbu, S.A.").Queries() {
		if q == `"Zorbu" Spain` {
			t.Errorf("searched a guessed country: %q", q)
		}
	}
	if got := JoinInitials("X Y Z Q Trading LLC"); got != "XYZQ Trading LLC" {
		t.Errorf("JoinInitials = %q", got)
	}
	if got := ASCII("Químicas Señal"); got != "Quimicas Senal" {
		t.Errorf("ASCII = %q", got)
	}
}
