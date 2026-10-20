package lookup

import (
	"context"
	"net/url"
	"strings"
)

var nominatimAPI = "https://nominatim.openstreetmap.org/search"

type osmPlace struct {
	Name     string            `json:"name"`
	Display  string            `json:"display_name"`
	Category string            `json:"category"`
	Address  map[string]string `json:"address"`
}

// osmAddress asks OpenStreetMap's Nominatim (free, keyless, one request
// per company) for a place named like the company, such as an office or
// a factory, and returns its address. OpenStreetMap is community data.
func osmAddress(ctx context.Context, c *Client, q Query) (Result, error) {
	var res Result
	base := strings.TrimSpace(q.Base)
	if len(strings.Fields(normWords(base))) < 2 && len(strings.TrimSpace(normWords(base))) < 6 {
		return res, nil // too short to tell the company apart from other places
	}
	v := url.Values{"q": {base}, "format": {"jsonv2"}, "addressdetails": {"1"}, "limit": {"5"}}
	var cc []string
	for i, c := range q.Countries {
		if i < 5 {
			cc = append(cc, strings.ToLower(c))
		}
	}
	if len(cc) > 0 {
		v.Set("countrycodes", strings.Join(cc, ","))
	}
	var places []osmPlace
	if err := c.getJSON(ctx, nominatimAPI+"?"+v.Encode(), nil, &places); err != nil {
		return res, err
	}
	want := normWords(base)
	for _, p := range places {
		name := normWords(p.Name)
		if p.Name == "" || !(strings.Contains(name, want) || (q.Form != "" && Mentions(name, q))) {
			continue
		}
		a := p.Address
		street := strings.TrimSpace(a["road"] + " " + a["house_number"])
		city := a["city"]
		for _, k := range []string{"town", "village", "municipality", "suburb"} {
			if city == "" {
				city = a[k]
			}
		}
		if a["road"] == "" || city == "" {
			continue
		}
		text := street + ", " + strings.TrimSpace(a["postcode"]+" "+city)
		if a["country"] != "" {
			text += ", " + CountryName(strings.ToUpper(a["country_code"]))
		}
		src := Source{ID: "osm", Name: "OpenStreetMap place " + p.Name + " (community data)", URL: "https://www.openstreetmap.org/", Kind: KindCommunity}
		res.Addresses = append(res.Addresses, Address{Text: text, Source: src})
		res.Sources = append(res.Sources, src)
		break
	}
	return res, nil
}
