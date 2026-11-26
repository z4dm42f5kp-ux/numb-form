package manual

import (
	"strings"
	"testing"
)

// Comments, such as the markers around generated install commands, are
// not shown.
func TestCommentsHidden(t *testing.T) {
	md := "# TITLE\n\nText.\n\n<!-- install:linux -->\n```\nmake install\n```\n<!-- /install:linux -->\n"
	var b strings.Builder
	roffText(md, &b)
	for name, out := range map[string]string{"text": Text(md, 80), "roff": b.String()} {
		if strings.Contains(out, "<!--") || strings.Contains(out, "install:linux") || !strings.Contains(out, "make install") {
			t.Errorf("%s:\n%s", name, out)
		}
	}
}

// A diagram is replaced by its description in the terminal.
func TestDiagramInTerminal(t *testing.T) {
	got := Text("![The whole procedure](diagrams/pipeline.svg)\n", 80)
	if !strings.Contains(got, "[Diagram: The whole procedure.]") || strings.Contains(got, "svg") {
		t.Fatalf("got %q", got)
	}
}
