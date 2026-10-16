package textutil

import "testing"

func TestMatchKeys(t *testing.T) {
	eq := func(a, b string) bool {
		for _, x := range MatchKeys(a) {
			for _, y := range MatchKeys(b) {
				if x == y {
					return true
				}
			}
		}
		return false
	}
	for _, p := range [][2]string{
		{"Foo Co., Ltda", "foo co ltda"}, {"Müller GmbH", "Mueller GmbH"}, {"Müller GmbH", "MULLER GMBH"},
		{"Café S.A.", "cafe sa"}, {"Straße AG", "Strasse AG"}, {"Foo S.A.", "Foo S.A"},
	} {
		if !eq(p[0], p[1]) {
			t.Errorf("%q and %q should match", p[0], p[1])
		}
	}
	for _, p := range [][2]string{{"Foo S.A.", "Foo S.A.S."}, {"Foo Ltd", "Foo Ltda"}, {"Foo GmbH", "Foo AG"}} {
		if eq(p[0], p[1]) {
			t.Errorf("%q and %q must not match", p[0], p[1])
		}
	}
}

func TestSafeTerminal(t *testing.T) {
	if got := SafeTerminal("ok\x1b[2J‮gone\x07"); got != "ok[2Jgone" {
		t.Errorf("SafeTerminal = %q", got)
	}
}

func FuzzMatchKeys(f *testing.F) {
	f.Add("Müller GmbH & Co. KG")
	f.Fuzz(func(t *testing.T, s string) {
		for _, k := range MatchKeys(s) {
			if MatchKeys(k)[0] != k {
				t.Fatalf("keys not idempotent for %q", s)
			}
		}
		_ = Similarity(s, "foo")
	})
}
