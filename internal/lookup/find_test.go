package lookup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"
)

const ddgFixture = `<html><body>
<div class="result results_links results_links_deep result--ad">
 <div class="links_main links_deep result__body"><h2 class="result__title">
 <a rel="nofollow" class="result__a" href="https://duckduckgo.com/y.js?ad_domain=ads.example">Ad</a></h2></div></div>
<div class="result results_links results_links_deep web-result ">
 <div class="links_main links_deep result__body">
  <h2 class="result__title"><a rel="nofollow" class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fwww.foo-maschinen.de%2Fimpressum&amp;rut=abc">Impressum - <b>Foo Maschinen GmbH</b></a></h2>
  <a class="result__snippet" href="https://www.foo-maschinen.de/impressum">Foo Maschinen GmbH, Geschäftsführer: Erika Mustermann</a>
 </div></div>
<div class="result results_links results_links_deep web-result ">
 <div class="links_main links_deep result__body">
  <h2 class="result__title"><a class="result__a" rel="nofollow" href="https://www.northdata.de/Foo+Maschinen+GmbH">Foo Maschinen GmbH, München - North Data</a></h2>
  <a class="result__snippet" href="https://www.northdata.de/Foo+Maschinen+GmbH">Geschäftsführer: Erika Mustermann</a>
 </div></div>
</body></html>`

func TestParseDDG(t *testing.T) {
	hits, err := ParseDDG(ddgFixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("hits %+v", hits)
	}
	if hits[0].URL != "https://www.foo-maschinen.de/impressum" || hits[0].Title != "Impressum - Foo Maschinen GmbH" || !strings.Contains(hits[0].Snippet, "Erika") {
		t.Errorf("first hit %+v", hits[0])
	}
	lite := `<table><tr><td><a rel="nofollow" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Ffoo.example%2F" class='result-link'>Foo</a></td></tr>
<tr><td class='result-snippet'>Foo S.A. de C.V.</td></tr></table>`
	if hits, err = ParseDDG(lite); err != nil || len(hits) != 1 || hits[0].URL != "https://foo.example/" || hits[0].Snippet != "Foo S.A. de C.V." {
		t.Errorf("lite: %+v %v", hits, err)
	}
	if _, err := ParseDDG(`<div class="anomaly-modal__title">Unfortunately, bots use DuckDuckGo too.</div>`); err != ErrSearchBlocked {
		t.Errorf("anomaly page: %v", err)
	}
}

func people(lines []string) map[string]string {
	var r Result
	for _, p := range ExtractPeople(lines, Source{ID: "website", Kind: KindCompany}) {
		r.addPerson(p)
	}
	got := map[string]string{}
	for _, p := range r.People {
		got[p.Name] = strings.Join(p.Roles, ", ")
	}
	return got
}

func TestExtractPeopleLanguages(t *testing.T) {
	cases := []struct {
		name string
		html string
		want map[string]string
	}{
		{"de footer line", `<footer>© 2025 Foo GmbH | Geschäftsführer: Max Mustermann, Dr. Erika Musterfrau | Amtsgericht Köln HRB 1234</footer>`,
			map[string]string{"Max Mustermann": "Managing Director", "Dr. Erika Musterfrau": "Managing Director"}},
		{"de table", `<table><tr><td>Vorstand</td></tr><tr><td>Dipl.-Ing. Hans Meier (Vorsitzender)</td></tr><tr><td>Anna Schmidt</td></tr><tr><td>Aufsichtsrat</td></tr><tr><td>Otto Ober</td></tr></table>`,
			map[string]string{"Hans Meier": "Member of the Management Board, Chairman", "Anna Schmidt": "Member of the Management Board"}},
		{"de not reps", `<p>Verantwortlich für den Inhalt nach § 18 Abs. 2 MStV: Peter Presse</p><p>Datenschutzbeauftragter: Daniel Daten</p><p>Vorsitzender des Aufsichtsrats: Albert Aufsicht</p>`,
			map[string]string{}},
		{"es", `<p>Representante Legal: JUAN PÉREZ GARCÍA</p><p>Gerente General: María José López</p><p>Director de Marketing: Pablo Publicidad</p>`,
			map[string]string{"Juan Pérez García": "Legal Representative", "María José López": "General Manager"}},
		{"pt", `<p>Sócio-Administrador: João da Silva</p><p>Diretora Presidente: Ana Souza</p>`,
			map[string]string{"João da Silva": "Managing Partner", "Ana Souza": "Chief Executive Officer"}},
		{"fr", `<p>Directeur de la publication : Jean Dupont, Président</p>`,
			map[string]string{"Jean Dupont": "Publication Director (Legal Representative), President"}},
		{"en", `<p>Directors: John Smith and Jane Doe</p><p>Registered agent: Agent Person</p><p>Managing Director(s): Tom Brown</p>`,
			map[string]string{"John Smith": "Director", "Jane Doe": "Director", "Tom Brown": "Managing Director"}},
		{"represented by inner role", `<p>Represented by the Managing Director Peter Parker</p><p>Vertreten durch den Vorstand: Clara Klein</p>`,
			map[string]string{"Peter Parker": "Managing Director", "Clara Klein": "Member of the Management Board"}},
		{"borme", `<p>123456 - FOO SL.</p><p>Ceses/Dimisiones. Adm. Unico: GOMEZ RUIZ LUIS. Nombramientos. Adm. Unico: PEREZ GARCIA JUAN. Datos registrales. T 1, F 2, S 8, H M 3, I/A 4 (1.02.24).</p>`,
			map[string]string{"Perez Garcia Juan": "Sole Administrator"}},
		{"former", `<p>Ehemaliger Geschäftsführer: Fritz Früher</p>`, map[string]string{}},
	}
	for _, c := range cases {
		got := people(HTMLText(c.html))
		if len(got) != len(c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
			continue
		}
		for k, v := range c.want {
			if got[k] != v {
				t.Errorf("%s: %s = %q, want %q (all %v)", c.name, k, got[k], v, got)
			}
		}
	}
}

func TestIDs(t *testing.T) {
	text := "USt-IdNr.: DE 123 456 789 · UID: ATU13585627 · CHE-116.281.710 MWST · CNPJ: 19.131.243/0001-97 · NIF ESB12345678 · tel 11.222.333/4444-55"
	got := map[string]bool{}
	for _, id := range ExtractIDs(text) {
		got[id.Kind+":"+id.Value] = true
	}
	for _, want := range []string{"vat:DE123456789", "vat:ATU13585627", "che:CHE116281710", "cnpj:19131243000197", "vat:ESB12345678"} {
		if !got[want] {
			t.Errorf("missing %s in %v", want, got)
		}
	}
	if len(got) != 5 {
		t.Errorf("extra ids: %v", got)
	}
	reg := ExtractRegister([]string{"Registergericht: Amtsgericht München HRB 123456", "Firmenbuch FN 123456 a"})
	if len(reg) != 2 || !strings.Contains(reg[0], "HRB 123456") {
		t.Errorf("register %v", reg)
	}
}

func TestMentionsAndSlugs(t *testing.T) {
	q := NewQuery("Foo Bar S.A. de C.V.")
	for text, want := range map[string]bool{
		"© FOO BAR, S. A. DE C. V. Todos los derechos": true,
		"Foo Bar Sociedad Anónima de Capital Variable": true,
		"Foo Bar Industries":                           false,
		"Foo Barra S.A. de C.V.":                       false,
	} {
		if Mentions(normWords(text), q) != want {
			t.Errorf("Mentions(%q) != %v", text, want)
		}
	}
	if Mentions(normWords("Welcome to Atlas"), NewQuery("Atlas")) {
		t.Error("a one-word name without legal form must not match on its own")
	}
	if g := guessDomains(NewQuery("Foo Bar GmbH")); len(g) == 0 || g[0] != "https://www.foobar.de/" {
		t.Errorf("guesses %v", g)
	}
	if g := guessDomains(NewQuery("Comercial Foo Ltda")); len(g) == 0 || g[0] != "https://www.comercialfoo.com.br/" {
		t.Errorf("guesses %v", g)
	}
}

func TestFindOnePass(t *testing.T) {
	mux := http.NewServeMux()
	var srv *httptest.Server
	mux.HandleFunc("/ddg/", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		q := r.PostForm.Get("q")
		switch {
		case strings.Contains(q, "site:"):
			_, _ = w.Write([]byte(`<div class="result results_links"><div class="result__body"></div></div>`))
		default:
			_, _ = w.Write([]byte(`<div class="result results_links web-result"><div class="links_main result__body">
<h2 class="result__title"><a class="result__a" href="` + srv.URL + `/">Foo Maschinen</a></h2></div></div>`))
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			_, _ = w.Write([]byte(`<html><a href="/impressum">Impressum</a><a href="/kontakt">Kontakt</a></html>`))
		case "/impressum":
			_, _ = w.Write([]byte(`<h1>Impressum</h1><p>Foo Maschinen GmbH<br>Industriestraße 5<br>80333 München</p>
<p>Geschäftsführer: Erika Mustermann</p><p>Amtsgericht München HRB 98765</p><p>USt-IdNr. ATU13585627</p>`))
		default:
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/vies/ms/AT/vat/U13585627", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"isValid":true,"name":"FOO MASCHINEN GMBH","address":"INDUSTRIESTRASSE 5\n80333 MUENCHEN"}`))
	})
	mux.HandleFunc("/gleif/lei-records", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"data":[]}`)) })
	srv = httptest.NewTLSServer(mux)
	defer srv.Close()
	ddgHTML, ddgLite, viesAPI, gleifAPI = srv.URL+"/ddg/", srv.URL+"/ddg/", srv.URL+"/vies", srv.URL+"/gleif"
	searchGap, searchJitter = time.Millisecond, 0

	res := Find(context.Background(), testClient(srv), Config{Search: "duckduckgo"}, "Foo Maschinen GmbH", "")
	if res.Website != srv.URL+"/" && !strings.HasPrefix(res.Website, srv.URL) {
		t.Errorf("website %q", res.Website)
	}
	if len(res.People) != 1 || res.People[0].Name != "Erika Mustermann" || res.People[0].Roles[0] != "Managing Director" {
		t.Errorf("people %+v", res.People)
	}
	var kinds []string
	for _, a := range res.Addresses {
		kinds = append(kinds, a.Source.ID+"="+a.Text)
	}
	sort.Strings(kinds)
	if len(res.Addresses) < 2 || res.Addresses[0].Source.ID != "vies" {
		t.Errorf("addresses %v", kinds)
	}
	if len(res.Register) == 0 || !strings.Contains(res.Register[0].Text, "HRB 98765") {
		t.Errorf("register %+v", res.Register)
	}
	if len(res.IDs) != 1 || res.IDs[0].Value != "ATU13585627" {
		t.Errorf("ids %+v", res.IDs)
	}
}
