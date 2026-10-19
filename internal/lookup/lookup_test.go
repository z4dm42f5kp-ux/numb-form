package lookup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testClient(srv *httptest.Server) *Client { return &Client{hc: srv.Client()} }

func TestSameEntity(t *testing.T) {
	cases := []struct {
		q, name string
		want    bool
	}{
		{"Foo GmbH", "FOO GMBH", true},
		{"Foo GmbH", "Foo Gesellschaft mit beschränkter Haftung", true},
		{"Foo Ltd", "FOO LIMITED", true},
		{"Foo Ltd", "FOO LTDA", false},
		{"Foo GmbH", "Foo AG", false},
		{"Foo GmbH", "Foo", false},
		{"Foo S.A.", "Foo SA", true},
		{"Foo S.A.", "Foo S.A.S.", false},
	}
	for _, c := range cases {
		if got := SameEntity(NewQuery(c.q), c.name); got != c.want {
			t.Errorf("SameEntity(%q, %q) = %v", c.q, c.name, got)
		}
	}
}

func TestTranslateRole(t *testing.T) {
	for in, want := range map[string]string{
		"Geschäftsführer": "Managing Director", "Gerente General": "General Manager",
		"Representante Legal": "Legal Representative", "Prokuristin": "Authorized Signatory (Prokurist)",
		"Administrador Único": "Sole Administrator", "Geschäftsführerinnen": "Managing Director",
	} {
		if got, ok := TranslateRole(in); !ok || got != want {
			t.Errorf("TranslateRole(%q) = %q, %v", in, got, ok)
		}
	}
	if _, ok := TranslateRole("Schatzmeister der Welt"); ok {
		t.Error("unknown role reported as translated")
	}
}

func TestExtractPeople(t *testing.T) {
	page := `<html><body><h1>Impressum</h1><p>Foo GmbH<br>Musterstraße 12<br>80333 München</p>
<p>Vertreten durch die Geschäftsführer: Dr. Erika Mustermann und Max Mustermann</p>
<p>Prokurist: Hans-Peter von Beispiel</p><p>Registergericht: Amtsgericht München, HRB 12345</p>
<script>var x = "Geschäftsführer: Evil Script";</script></body></html>`
	lines := HTMLText(page)
	ps := ExtractPeople(lines, Source{Name: "x", Kind: KindCompany})
	var r Result
	for _, p := range ps {
		r.addPerson(p)
	}
	got := map[string]string{}
	for _, p := range r.People {
		got[p.Name] = strings.Join(p.Roles, ",")
	}
	want := map[string]string{
		"Dr. Erika Mustermann":    "Managing Director",
		"Max Mustermann":          "Managing Director",
		"Hans-Peter von Beispiel": "Authorized Signatory (Prokurist)",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: %q, want %q (all: %v)", k, got[k], v, got)
		}
	}
	if a := ExtractAddress(lines); a != "Musterstraße 12, 80333 München" {
		t.Errorf("address %q", a)
	}
}

func TestChName(t *testing.T) {
	if got := chName("O'NEIL-SMITH, Jane Mary"); got != "Jane Mary O'Neil-Smith" {
		t.Errorf("chName = %q", got)
	}
}

func TestRegistriesWithMockServers(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/gleif/lei-records", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("filter[fulltext]") != "Foo" {
			t.Errorf("gleif query %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"data":[
 {"id":"529900T8BM49AURSDO55","attributes":{"lei":"529900T8BM49AURSDO55","entity":{"legalName":{"name":"Foo GmbH"},
  "legalAddress":{"addressLines":["Musterstraße 12"],"city":"München","postalCode":"80333","country":"DE"},
  "headquartersAddress":{"addressLines":["Musterstraße 12"],"city":"München","postalCode":"80333","country":"DE"},
  "legalForm":{"id":"2HBR","other":null},"status":"ACTIVE"}}},
 {"id":"OTHER","attributes":{"lei":"OTHER","entity":{"legalName":{"name":"Foo Holding AG"},"legalAddress":{"city":"Zug","country":"CH"},"headquartersAddress":{},"legalForm":{},"status":"ACTIVE"}}}]}`))
	})
	mux.HandleFunc("/wd", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("action") {
		case "wbsearchentities":
			_, _ = w.Write([]byte(`{"search":[{"id":"Q1","label":"Foo GmbH"},{"id":"Q2","label":"Foo"}]}`))
		case "wbgetentities":
			if strings.Contains(r.URL.Query().Get("props"), "claims") {
				_, _ = w.Write([]byte(`{"entities":{
 "Q1":{"labels":{"en":{"value":"Foo GmbH"}},"claims":{
  "P856":[{"rank":"normal","mainsnak":{"datavalue":{"value":"https://foo.example"}}}],
  "P169":[{"rank":"normal","mainsnak":{"datavalue":{"value":{"id":"Q10"}}}},
          {"rank":"normal","mainsnak":{"datavalue":{"value":{"id":"Q11"}}},"qualifiers":{"P582":[]}}]}},
 "Q2":{"labels":{"en":{"value":"Foo"}},"claims":{"P169":[{"rank":"normal","mainsnak":{"datavalue":{"value":{"id":"Q12"}}}}]}}}}`))
				return
			}
			_, _ = w.Write([]byte(`{"entities":{"Q10":{"labels":{"en":{"value":"Erika Mustermann"}}},"Q11":{"labels":{"en":{"value":"Former Boss"}}},"Q12":{"labels":{"en":{"value":"Parent Person"}}}}}`))
		}
	})
	mux.HandleFunc("/ch/search/companies", func(w http.ResponseWriter, r *http.Request) {
		if u, _, ok := r.BasicAuth(); !ok || u != "free-key" {
			t.Error("missing basic auth")
		}
		_, _ = w.Write([]byte(`{"items":[{"title":"FOO GMBH","company_number":"FC012345","company_status":"active","address_snippet":"1 Road, London, EC1A 1AA"}]}`))
	})
	mux.HandleFunc("/ch/company/FC012345/officers", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"name":"MUSTERMANN, Erika","officer_role":"director"},{"name":"GONE, Old","officer_role":"director","resigned_on":"2020-01-01"}]}`))
	})
	srv := httptest.NewTLSServer(mux)
	defer srv.Close()
	gleifAPI, wdAPI, chAPI = srv.URL+"/gleif", srv.URL+"/wd", srv.URL+"/ch"

	res := Registries(context.Background(), testClient(srv), Config{CompaniesHouseKey: "free-key"}, NewQuery("Foo GmbH"))
	if len(res.Addresses) == 0 || !strings.HasPrefix(res.Addresses[0].Text, "1 Road") || res.Addresses[0].Source.Kind != KindGovernment {
		t.Errorf("addresses: %+v", res.Addresses)
	}
	var gleifAddr bool
	for _, a := range res.Addresses {
		if a.Text == "Musterstraße 12, 80333 München, Germany" {
			gleifAddr = true
		}
		if strings.Contains(a.Text, "Zug") {
			t.Error("address of a different entity (holding) was used")
		}
	}
	if !gleifAddr {
		t.Errorf("GLEIF address missing: %+v", res.Addresses)
	}
	if len(res.People) != 1 || res.People[0].Name != "Erika Mustermann" || !res.People[0].Verified() {
		t.Fatalf("people: %+v", res.People)
	}
	if len(res.People[0].Sources) != 2 {
		t.Errorf("sources not merged: %+v", res.People[0].Sources)
	}
	if len(res.Websites) != 1 || res.Websites[0] != "https://foo.example" {
		t.Errorf("websites %v", res.Websites)
	}
}

func TestClientRefusesHTTPAndPrivate(t *testing.T) {
	c := NewClient()
	if _, _, _, err := c.get(context.Background(), "http://example.com", nil); err == nil {
		t.Error("plain HTTP accepted")
	}
	if err := publicOnly("tcp", "127.0.0.1:443", nil); err == nil {
		t.Error("loopback accepted")
	}
	if err := publicOnly("tcp", "10.1.2.3:443", nil); err == nil {
		t.Error("private address accepted")
	}
	if err := publicOnly("tcp", "93.184.216.34:443", nil); err != nil {
		t.Errorf("public address refused: %v", err)
	}
}
