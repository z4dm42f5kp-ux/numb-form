package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/odt"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

// The language is the first question when the configuration has none.
// The questions follow in that language; the form stays in English.
func TestLanguageQuestion(t *testing.T) {
	for _, tc := range []struct{ lang, answer, want string }{
		{"", "3", "Sitz / Adresse"},
		{"", "es", "Sede / Dirección"},
		{"", "", "Headquarters / Address"}, // Enter: the system language (English here)
		{"de", "", "Sitz / Adresse"},       // from the configuration: not asked
	} {
		t.Setenv("LC_ALL", "C")
		root := filepath.Join(t.TempDir(), "Clients")
		if err := os.MkdirAll(filepath.Join(root, "Overview"), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "Overview", "list.csv"), testfix.CSV([]string{"D1"}, []string{"Foo Trading"}, ";", false), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "Appendix A - Risk Assessment.odt"), testfix.ODT(t), 0o600); err != nil {
			t.Fatal(err)
		}
		cfg := direct(t)
		cfg.OverviewDir = "overview"
		cfg.Language = tc.lang
		in := "P\n\n\n\n\na\n" // person, use it, legal form, address, representatives, accept
		if tc.lang == "" {
			in = tc.answer + "\n" + in
		}
		var out bytes.Buffer
		if code := Run(Options{Parent: root, AllowNonTTY: true, Config: cfg, In: strings.NewReader(in), Out: &out, Now: fixedNow}); code != ExitOK {
			t.Fatalf("%+v: exit %d\n%s", tc, code, out.String())
		}
		asked := strings.Contains(out.String(), "Language / Idioma / Sprache:")
		if asked != (tc.lang == "") || !strings.Contains(out.String(), tc.want) {
			t.Errorf("%+v: asked %v, output:\n%s", tc, asked, out.String())
		}
		b, err := os.ReadFile(filepath.Join(root, "Foo Trading", "Appendix A - Risk Assessment - Foo Trading.odt"))
		if err != nil {
			t.Fatal(err)
		}
		if got, _ := (odt.Format{}).ReadBack(b); got[form.FieldAddress] != NoAddressText {
			t.Errorf("%+v: address %q, want the English text", tc, got[form.FieldAddress])
		}
	}
}
