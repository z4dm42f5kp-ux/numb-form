package lookup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const braveFixture = `<div class="snippet svelte-1" data-pos="0" data-type="web"><a href="https://www.foo.pt/" class="h"><div class="title svelte-x">Foo, Lda.</div></a>
<div class="snippet-description">Foo, Lda. — gerente: João Silva</div></div>
<div class="snippet" data-type="web"><a href="https://search.brave.com/x"><div class="title">internal</div></a></div>`

func TestEngineParsers(t *testing.T) {
	cases := []struct {
		name  string
		parse func(string) ([]Hit, error)
		page  string
		url   string
		title string
		snip  string
	}{
		{"brave", ParseBrave, braveFixture, "https://www.foo.pt/", "Foo, Lda.", "Foo, Lda. — gerente: João Silva"},
	}
	for _, c := range cases {
		hits, err := c.parse(c.page)
		if err != nil || len(hits) != 1 {
			t.Errorf("%s: %+v %v", c.name, hits, err)
			continue
		}
		if hits[0].URL != c.url || hits[0].Title != c.title || hits[0].Snippet != c.snip {
			t.Errorf("%s: %+v", c.name, hits[0])
		}
	}
}

func TestSearchFallsBack(t *testing.T) {
	ddgCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/ddg/", func(w http.ResponseWriter, r *http.Request) {
		ddgCalls++
		w.WriteHeader(202)
		_, _ = w.Write([]byte(`<div class="anomaly-modal">bots</div>`))
	})
	mux.HandleFunc("/brave", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != `"Foo GmbH"` {
			t.Errorf("brave query %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(braveFixture))
	})
	srv := httptest.NewTLSServer(mux)
	defer srv.Close()
	ddgHTML, ddgLite, braveURL = srv.URL+"/ddg/", srv.URL+"/ddg/", srv.URL+"/brave"
	searchGap, searchJitter = time.Millisecond, 0
	c := testClient(srv)
	c.engines = []string{"duckduckgo", "brave"}
	hits, engine, err := Search(context.Background(), c, `"Foo GmbH"`, "")
	if err != nil || engine != "brave" || len(hits) != 1 {
		t.Fatalf("hits %+v engine %q err %v", hits, engine, err)
	}
	if !paused("duckduckgo") {
		t.Error("DuckDuckGo not rested after refusing")
	}
	before := ddgCalls
	if _, engine, _ = Search(context.Background(), c, `"Foo GmbH"`, ""); engine != "brave" || ddgCalls != before {
		t.Errorf("a resting engine was asked again (%d calls) or wrong engine %q", ddgCalls-before, engine)
	}
	pausedUntil = map[string]time.Time{}
}

// fakeBrowser writes a script that behaves like "chrome --dump-dom URL":
// it prints a page and records the profile folder it was given.
func fakeBrowser(t *testing.T, page string) (*Browser, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in browser is a shell script")
	}
	dir := t.TempDir()
	rec := filepath.Join(dir, "profile.txt")
	script := "#!/bin/sh\nfor a in \"$@\"; do case \"$a\" in --user-data-dir=*) echo \"${a#--user-data-dir=}\" > " + rec + ";; esac; done\ncat <<'HTML'\n" + page + "\nHTML\n"
	p := filepath.Join(dir, "chrome")
	if err := os.WriteFile(p, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return &Browser{Path: p, Name: "Test Chrome"}, rec
}

func TestBrowserFetchAndCleanup(t *testing.T) {
	b, rec := fakeBrowser(t, `<html><body><h1>Impressum</h1><p>Foo GmbH</p></body></html>`)
	out, err := b.Fetch(context.Background(), "https://www.foo.de/")
	if err != nil || !strings.Contains(string(out), "Foo GmbH") {
		t.Fatalf("%q %v", out, err)
	}
	profile, _ := os.ReadFile(rec)
	if p := strings.TrimSpace(string(profile)); p == "" {
		t.Error("no profile folder given to the browser")
	} else if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("profile folder %s left behind", p)
	}
	if _, err := b.Fetch(context.Background(), "http://www.foo.de/"); err == nil {
		t.Error("plain HTTP given to the browser")
	}
	if FindBrowser("none") != nil || FindBrowser("/does/not/exist") != nil {
		t.Error("browser setting none or a missing path must give no browser")
	}
}

func TestScriptBuiltPageUsesBrowser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><body><div id="root"></div><script src="app.js"></script></body></html>`))
	})
	srv := httptest.NewTLSServer(mux)
	defer srv.Close()
	b, _ := fakeBrowser(t, `<html><body><div id="root"><h1>Impressum</h1><p>Foo Maschinen GmbH</p><p>Geschäftsführer: Erika Mustermann</p></div></body></html>`)
	c := testClient(srv)
	c.browser = b
	budget := 1
	body, _, err := c.page(context.Background(), srv.URL+"/", &budget)
	if err != nil || !strings.Contains(string(body), "Erika Mustermann") || budget != 0 {
		t.Fatalf("body %q err %v budget %d", body, err, budget)
	}
	body, _, _ = c.page(context.Background(), srv.URL+"/", &budget)
	if strings.Contains(string(body), "Erika") {
		t.Error("browser used beyond its budget")
	}
}
