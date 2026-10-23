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
	p := &pager{lines: strings.Split(strings.TrimSpace(lines(100)), "\n"), rows: 11}
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
	if err := Less(&out, strings.NewReader("Gg/line 7\rq"), lines(50), 10); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "lines 42-50 of 50  END") || !strings.Contains(out.String(), "lines 7-15 of 50") {
		t.Fatalf("screens:\n%q", out.String())
	}
}
