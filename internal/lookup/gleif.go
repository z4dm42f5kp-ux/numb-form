package lookup

import (
	"context"
	"net/url"
	"strings"
)

var gleifAPI = "https://api.gleif.org/api/v1"

type gleifAddr struct {
	AddressLines []string `json:"addressLines"`
	City         string   `json:"city"`
	Region       string   `json:"region"`
	Country      string   `json:"country"`
	PostalCode   string   `json:"postalCode"`
}

type gleifResp struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			LEI    string `json:"lei"`
			Entity struct {
				LegalName struct {
					Name string `json:"name"`
				} `json:"legalName"`
				OtherNames []struct {
					Name string `json:"name"`
				} `json:"otherNames"`
				LegalAddress        gleifAddr `json:"legalAddress"`
				HeadquartersAddress gleifAddr `json:"headquartersAddress"`
				LegalForm           struct {
					ID    string  `json:"id"`
					Other *string `json:"other"`
				} `json:"legalForm"`
				Status string `json:"status"`
			} `json:"entity"`
		} `json:"attributes"`
	} `json:"data"`
}

// FormatAddress joins address parts into one line:
// "Street 1, 80333 München, Germany".
func FormatAddress(lines []string, postal, city, country string) string {
	var parts []string
	for _, l := range lines {
		if l = strings.TrimSpace(l); l != "" {
			parts = append(parts, l)
		}
	}
	pc := strings.TrimSpace(strings.TrimSpace(postal) + " " + strings.TrimSpace(city))
	if pc != "" {
		parts = append(parts, pc)
	}
	if c := CountryName(country); c != "" {
		parts = append(parts, c)
	}
	return strings.Join(parts, ", ")
}

// gleif searches the Global LEI Index (https://api.gleif.org), a free,
// keyless register overseen by the Regulatory Oversight Committee.
func gleif(ctx context.Context, c *Client, q Query) (Result, error) {
	// Searching for the core name finds the record however the register
	// writes the legal form ("GRUPO BIMBO, S.A.B. DE C.V."); SameEntity
	// then keeps only exact matches.
	term := q.Base
	if len(strings.TrimSpace(term)) < 3 {
		term = q.Name
	}
	u := gleifAPI + "/lei-records?" + url.Values{
		"filter[fulltext]": {term},
		"page[size]":       {"20"},
	}.Encode()
	var resp gleifResp
	if err := c.getJSON(ctx, u, map[string]string{"Accept": "application/vnd.api+json"}, &resp); err != nil {
		return Result{}, err
	}
	var res Result
	for _, d := range resp.Data {
		e := d.Attributes.Entity
		names := []string{e.LegalName.Name}
		for _, o := range e.OtherNames {
			names = append(names, o.Name)
		}
		ok := false
		for _, n := range names {
			if SameEntity(q, n) {
				ok = true
				break
			}
		}
		if !ok {
			continue
		}
		lei := d.Attributes.LEI
		if lei == "" {
			lei = d.ID
		}
		name := "GLEIF LEI " + lei + " (" + e.LegalName.Name + ")"
		if e.Status != "" && e.Status != "ACTIVE" {
			name += " [status " + e.Status + "]"
		}
		src := Source{ID: "gleif", Name: name, URL: "https://search.gleif.org/#/record/" + url.PathEscape(lei), Kind: KindRegulatory}
		res.Sources = append(res.Sources, src)
		if a := FormatAddress(e.HeadquartersAddress.AddressLines, e.HeadquartersAddress.PostalCode, e.HeadquartersAddress.City, e.HeadquartersAddress.Country); a != "" {
			res.Addresses = append(res.Addresses, Address{Text: a, Source: Source{ID: "gleif", Name: src.Name + " headquarters address", URL: src.URL, Kind: src.Kind}})
		}
		if a := FormatAddress(e.LegalAddress.AddressLines, e.LegalAddress.PostalCode, e.LegalAddress.City, e.LegalAddress.Country); a != "" {
			res.Addresses = append(res.Addresses, Address{Text: a, Source: Source{ID: "gleif", Name: src.Name + " legal address", URL: src.URL, Kind: src.Kind}})
		}
		if e.LegalForm.Other != nil && strings.TrimSpace(*e.LegalForm.Other) != "" {
			res.FormHints = append(res.FormHints, strings.TrimSpace(*e.LegalForm.Other))
		}
	}
	if len(res.Sources) == 0 {
		res.Notes = append(res.Notes, "GLEIF: no LEI record with this exact name")
	}
	return res, nil
}
