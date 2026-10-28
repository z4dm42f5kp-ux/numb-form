package textutil

import (
	"testing"
	"unicode"
)

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
	if got := SafeTerminal("ok\x1b[2J\u202egone\x07"); got != "ok[2Jgone" {
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

func TestMatchKeysDecomposed(t *testing.T) {
	// "Müller" as macOS writes it: u followed by the combining diaeresis.
	nfd := "Müller GmbH"
	for _, other := range []string{"Müller GmbH", "Mueller GmbH", "Muller GmbH"} {
		if !shareKey(MatchKeys(nfd), MatchKeys(other)) {
			t.Errorf("%q and %q should match: %q %q", nfd, other, MatchKeys(nfd), MatchKeys(other))
		}
	}
}

func TestMatchKeysFullUnicode(t *testing.T) {
	pairs := [][2]string{
		{"Ștefănescu S.R.L.", "Stefanescu SRL"},
		{"Doğan A.Ş.", "Dogan AS"},
		{"Ｆｏｏ ＧｍｂＨ", "Foo GmbH"},
		{"Ǎrbol", "Arbol"},
	}
	for _, p := range pairs {
		if shareKey(MatchKeys(p[0]), MatchKeys(p[1])) {
			t.Errorf("basic mode should not fold %q", p[0])
		}
	}
	SetFullUnicode(true)
	defer SetFullUnicode(false)
	for _, p := range pairs {
		if !shareKey(MatchKeys(p[0]), MatchKeys(p[1])) {
			t.Errorf("full mode: %q and %q should match: %q %q", p[0], p[1], MatchKeys(p[0]), MatchKeys(p[1]))
		}
	}
	// The umlaut spellings stay as in the basic mode.
	if k := MatchKeys("Müller"); len(k) != 2 || k[0] != "muller" || k[1] != "mueller" {
		t.Errorf("full mode umlauts: %q", k)
	}
}

func TestUnifoldTable(t *testing.T) {
	m := fullFold()
	if len(m) < 1000 {
		t.Fatalf("table has %d entries", len(m))
	}
	for r, v := range m {
		if v == "" {
			t.Errorf("%U maps to nothing", r)
		}
		for _, c := range v {
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				t.Errorf("%U maps to %q, which is not only letters and digits", r, v)
			}
		}
	}
}

func shareKey(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if x == y {
				return true
			}
		}
	}
	return false
}
