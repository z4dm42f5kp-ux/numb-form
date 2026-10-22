package lookup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestInlineAddressAndSnippets(t *testing.T) {
	for in, want := range map[string]string{
		"Foo Maschinen GmbH, Industriestraße 5, 80333 München": "Industriestraße 5, 80333 München",
		"Sitz: Musterweg 12a, 1010 Wien":                       "Musterweg 12a, 1010 Wien",
		"Calle Mayor 7, 28013 Madrid, España":                  "Calle Mayor 7, 28013 Madrid",
		"Gegründet 1990 in München":                            "",
	} {
		if got := inlineAddress(in); got != want {
			t.Errorf("inlineAddress(%q) = %q, want %q", in, got, want)
		}
	}
	hits := []Hit{
		{URL: "https://www.northdata.de/Foo+Maschinen+GmbH", Title: "Foo Maschinen GmbH, München", Snippet: "Foo Maschinen GmbH · Geschäftsführer: Erika Mustermann · Industriestraße 5, 80333 München"},
		{URL: "https://www.other.de/", Title: "Foo Maschinenbau AG", Snippet: "Vorstand: Otto Other"},
	}
	r := mineSnippets(hits, NewQuery("Foo Maschinen GmbH"))
	if len(r.People) != 1 || r.People[0].Name != "Erika Mustermann" || r.People[0].Verified() {
		t.Errorf("people %+v", r.People)
	}
	if len(r.Addresses) != 1 || r.Addresses[0].Text != "Industriestraße 5, 80333 München" || r.Addresses[0].Source.Kind != KindCommunity {
		t.Errorf("addresses %+v", r.Addresses)
	}
}

func TestNameThenRole(t *testing.T) {
	got := people([]string{"Max Mustermann – Geschäftsführer", "Jane Doe (CEO)", "Juan Pérez, Gerente General", "Anna Team – Head of Sales", "Ehemaliger: Fritz Früher (Geschäftsführer)"})
	want := map[string]string{"Max Mustermann": "Managing Director", "Jane Doe": "Chief Executive Officer", "Juan Pérez": "General Manager"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
}

func TestTargetedSearches(t *testing.T) {
	var queries []string
	mux := http.NewServeMux()
	var srv *httptest.Server
	mux.HandleFunc("/ddg/", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		queries = append(queries, r.PostForm.Get("q"))
		_, _ = w.Write([]byte(`<div class="result results_links web-result"><div class="result__body"><h2 class="result__title">
<a class="result__a" href="` + srv.URL + `/page">Foo Maschinen GmbH</a></h2>
<a class="result__snippet" href="` + srv.URL + `/page">Foo Maschinen GmbH · Hafenweg 3, 20457 Hamburg</a></div></div>`))
	})
	mux.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<h1>Foo Maschinen GmbH</h1><p>Geschäftsführer: Erika Mustermann</p>`))
	})
	srv = httptest.NewTLSServer(mux)
	defer srv.Close()
	ddgHTML, ddgLite = srv.URL+"/ddg/", srv.URL+"/ddg/"
	searchGap, searchJitter = time.Millisecond, 0
	c := testClient(srv)
	c.engines = []string{"duckduckgo"}
	s := &state{c: c, q: NewQuery("Foo Maschinen GmbH"), country: "DE", lang: "DE", region: "de-de", search: true, read: map[string]bool{}}
	targetedStep(context.Background(), s)
	if len(queries) != 1 || !strings.Contains(queries[0], "Geschäftsführer") {
		t.Errorf("queries %q", queries)
	}
	if len(s.all.People) != 1 || s.all.People[0].Name != "Erika Mustermann" {
		t.Errorf("people %+v", s.all.People)
	}
	if len(s.all.Addresses) != 1 || s.all.Addresses[0].Text != "Hafenweg 3, 20457 Hamburg" {
		t.Errorf("addresses %+v", s.all.Addresses)
	}
}

func TestDirectoryPage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<h1>Foo Maschinen GmbH</h1><p>Industriestraße 5, 80333 München</p>
<h2>Management</h2><ul><li>Erika Mustermann – Geschäftsführer</li><li>Ehemaliger Geschäftsführer: Fritz Früher</li></ul>`))
	})
	srv := httptest.NewTLSServer(mux)
	defer srv.Close()
	r, err := hitPage(context.Background(), testClient(srv), NewQuery("Foo Maschinen GmbH"), srv.URL+"/foo")
	if err != nil {
		t.Fatal(err)
	}
	if len(r.People) != 1 || r.People[0].Name != "Erika Mustermann" || r.People[0].Sources[0].ID != "search" {
		t.Errorf("people %+v", r.People)
	}
	if len(r.Addresses) != 1 || r.Addresses[0].Text != "Industriestraße 5, 80333 München" {
		t.Errorf("addresses %+v", r.Addresses)
	}
}
