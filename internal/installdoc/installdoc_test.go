package installdoc

import (
	"strings"
	"testing"
)

const sum = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func pins(t *testing.T) Pins {
	t.Helper()
	p, err := ParsePins([]byte("# made up\nversion v9.8.7\nprograms numb-form numb-date\n" + sum + "  numb-form_linux_amd64.tar.gz\n"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParsePins(t *testing.T) {
	p := pins(t)
	if p.Version != "v9.8.7" || len(p.Programs) != 2 || p.Sums["numb-form_linux_amd64.tar.gz"] != sum {
		t.Fatalf("%+v", p)
	}
	again, err := ParsePins(p.Format())
	if err != nil || again.Version != p.Version || again.Sums["numb-form_linux_amd64.tar.gz"] != sum {
		t.Fatalf("Format does not read back: %v %+v", err, again)
	}
	for _, bad := range []string{
		"version 1.2\nprograms numb-form\n" + sum + "  a.tar.gz\n",       // no v
		"version v1.2.3\nprograms rm\n" + sum + "  a.tar.gz\n",           // not a program
		"version v1.2.3\nprograms numb-form\n" + sum + "  ../a.tar.gz\n", // a path
		"version v1.2.3\nprograms numb-form\nabc  a.tar.gz\n",            // not a checksum
		"version v1.2.3\nprograms numb-form\n",                           // no checksums
	} {
		if _, err := ParsePins([]byte(bad)); err == nil {
			t.Errorf("accepted %q", bad)
		}
	}
}

func TestApply(t *testing.T) {
	tpl := map[string]string{
		"get.cmd":  "curl .../@VERSION@/x && echo '@SHA:numb-form_linux_amd64.tar.gz@' && install @PROGRAMS@ && rm @EXES_COMMA@\n",
		"wrap.cmd": "shell --run '@INCLUDE:get@'\n",
		"sums.md":  "| File | SHA-256 |\n@SUMS_TABLE@\n",
	}
	doc := "a\n   <!-- install:get -->\n   old\n   <!-- /install:get -->\nb\n<!-- install:wrap -->\n<!-- /install:wrap -->\n<!-- install:sums -->\n<!-- /install:sums -->\n"
	got, err := Apply(doc, tpl, pins(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"   ```\n   curl .../v9.8.7/x && echo '" + sum + "' && install numb-form numb-date && rm numb-form.exe, numb-date.exe\n   ```\n   <!-- /install:get -->",
		"shell --run 'curl .../v9.8.7/x",
		"| `numb-form_linux_amd64.tar.gz` | `" + sum + "` |\n<!-- /install:sums -->",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("lacks %q:\n%s", want, got)
		}
	}
	if again, _ := Apply(got, tpl, pins(t)); again != got {
		t.Error("Apply is not idempotent")
	}
	for _, bad := range []string{
		"<!-- install:get -->\n",                          // no end marker
		"<!-- install:none -->\n<!-- /install:none -->\n", // no template
	} {
		if _, err := Apply(bad, tpl, pins(t)); err == nil {
			t.Errorf("accepted %q", bad)
		}
	}
	if _, err := Render("@SHA:missing.zip@", pins(t)); err == nil {
		t.Error("rendered a checksum that is not pinned")
	}
	if _, err := Render("@NOPE@", pins(t)); err == nil {
		t.Error("rendered an unknown placeholder")
	}
}
