package i18n

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var verb = regexp.MustCompile(`%[-+# 0]*[0-9]*(\.[0-9]+)?[a-zA-Z%]`)

// TestCatalogVerbs checks that every translation has the format verbs of
// its English text, in the same order.
func TestCatalogVerbs(t *testing.T) {
	for lang, c := range catalogs {
		for k, v := range c {
			if a, b := verb.FindAllString(k, -1), verb.FindAllString(v, -1); strings.Join(a, " ") != strings.Join(b, " ") {
				t.Errorf("%s: %q has verbs %v, its translation %v", lang, k, a, b)
			}
		}
	}
}

// TestCatalogCurrent checks that every English text in the catalogs is
// still printed by the programs, so a changed message is not left
// untranslated without notice, and that both languages have the same
// entries.
func TestCatalogCurrent(t *testing.T) {
	files, _ := filepath.Glob("../app/*.go")
	more, _ := filepath.Glob("../ui/*.go")
	var src strings.Builder
	for _, f := range append(files, more...) {
		b, err := os.ReadFile(f) // #nosec G304 -- the source files of this module.
		if err != nil {
			t.Fatal(err)
		}
		src.Write(b)
	}
	for lang, c := range catalogs {
		for k := range c {
			if !strings.Contains(src.String(), strconv.Quote(k)) {
				t.Errorf("%s: %q is not printed anywhere", lang, k)
			}
		}
	}
	for k := range es {
		if _, ok := de[k]; !ok {
			t.Errorf("de: missing %q", k)
		}
	}
	for k := range de {
		if _, ok := es[k]; !ok {
			t.Errorf("es: missing %q", k)
		}
	}
}

func TestParse(t *testing.T) {
	for in, want := range map[string]string{"1": "en", "2": "es", "3": "de", "ES": "es", "Deutsch": "de", "español": "es", " en ": "en"} {
		if got, ok := Parse(in); !ok || got != want {
			t.Errorf("Parse(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	for _, in := range []string{"", "4", "fr", "x"} {
		if _, ok := Parse(in); ok {
			t.Errorf("Parse(%q) accepted", in)
		}
	}
	if T("de", "Copied to %s\n") != "Kopiert nach %s\n" || T("en", "Copied to %s\n") != "Copied to %s\n" || T("es", "no such text") != "no such text" {
		t.Error("T does not translate")
	}
}
