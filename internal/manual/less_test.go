package manual

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func lines(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	return b.String()
}

// Keys act at once, as in less: Space pages, arrows move one line, G and
// g jump, /text searches, q quits.
func TestPagerKeys(t *testing.T) {
	p := &pager{lines: strings.Split(strings.TrimSpace(lines(100)), "\n"), rows: 11, width: 79}
	for _, c := range []struct {
		key string
		top int
	}{{" ", 10}, {"\x1b[B", 11}, {"\x1b[A", 10}, {"b", 0}, {"k", 0}, {"G", 90}, {"\x1b[6~", 90}, {"g", 0}, {"d", 5}} {
		if p.key(c.key); p.top != c.top {
			t.Fatalf("after %q top = %d, want %d", c.key, p.top, c.top)
		}
	}
	p.pattern = "line 42"
	if p.search(1); p.top != 41 {
		t.Fatalf("search: top %d", p.top)
	}
	if !p.key("q") || !p.key("Q") {
		t.Fatal("q must quit")
	}
}

func TestLessReadsKeysWithoutEnter(t *testing.T) {
	var out bytes.Buffer
	if err := Less(&out, strings.NewReader("Gg/line 7\rq"), lines(50), 10, 80); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "lines 42-50 of 50  END") || !strings.Contains(out.String(), "lines 7-15 of 50") {
		t.Fatalf("screens:\n%q", out.String())
	}
}

// A line longer than the window becomes several rows, so a page never
// overflows the window.
func TestLongLinesWrap(t *testing.T) {
	got := softWrap([]string{strings.Repeat("x", 25), "short"}, 10)
	if len(got) != 4 || got[2] != "xxxxx" || got[3] != "short" {
		t.Fatalf("%q", got)
	}
	var out bytes.Buffer
	if err := Less(&out, strings.NewReader("q"), strings.Repeat("y", 500)+"\n", 10, 40); err != nil {
		t.Fatal(err)
	}
	for _, row := range strings.Split(out.String(), "\r\n") {
		if n := strings.Count(row, "y"); n > 39 {
			t.Fatalf("row of %d characters in a 40-column window", n)
		}
	}
}
