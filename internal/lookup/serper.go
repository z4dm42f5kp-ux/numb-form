package lookup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Serper (serper.dev) returns Google's results as JSON. It needs a free
// key (NUMBFORM_SERPER_KEY; 2,500 searches without a credit card).
var serperAPI = "https://google.serper.dev"

type serperResp struct {
	Organic []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"organic"`
	KnowledgeGraph struct {
		Title       string            `json:"title"`
		Website     string            `json:"website"`
		Description string            `json:"description"`
		Attributes  map[string]string `json:"attributes"`
	} `json:"knowledgeGraph"`
	Places []struct {
		Title   string `json:"title"`
		Address string `json:"address"`
		Website string `json:"website"`
		Cid     string `json:"cid"`
	} `json:"places"`
}

func (c *Client) serperCall(ctx context.Context, endpoint, query, region string) (serperResp, error) {
	var r serperResp
	if c.serperKey == "" {
		return r, fmt.Errorf("serper: no key")
	}
	body := map[string]any{"q": query, "num": 10}
	if cc, lang, ok := strings.Cut(region, "-"); ok && cc != "wt" {
		if cc == "uk" {
			cc = "gb"
		}
		body["gl"], body["hl"] = cc, lang
	}
	b, _ := json.Marshal(body)
	raw, _, _, err := c.do(ctx, http.MethodPost, serperAPI+"/"+endpoint, string(b), map[string]string{
		"X-API-KEY": c.serperKey, "Content-Type": "application/json", "Accept": "application/json",
	})
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(raw, &r)
	return r, err
}

// fetchSerper asks Google through Serper and returns the results as a
// page that parseSerper reads. The knowledge panel, when Google shows
// one, becomes a first result whose snippet lists its facts
// ("Headquarters: ...", "CEO: ...").
func fetchSerper(ctx context.Context, c *Client, query, region string) ([]byte, error) {
	r, err := c.serperCall(ctx, "search", query, region)
	if err != nil {
		return nil, err
	}
	return json.Marshal(r)
}

func parseSerper(page string) ([]Hit, error) {
	var r serperResp
	if err := json.Unmarshal([]byte(page), &r); err != nil {
		return nil, err
	}
	var hits []Hit
	if kg := r.KnowledgeGraph; kg.Title != "" {
		var facts []string
		for k, v := range kg.Attributes {
			facts = append(facts, k+": "+v)
		}
		u := kg.Website
		if u == "" {
			u = "https://www.google.com/search?kgmid=" + strings.ReplaceAll(kg.Title, " ", "+")
		}
		hits = append(hits, Hit{URL: u, Title: kg.Title, Snippet: strings.Join(append(facts, kg.Description), " · ")})
	}
	for _, o := range r.Organic {
		if o.Link != "" {
			hits = append(hits, Hit{URL: o.Link, Title: o.Title, Snippet: o.Snippet})
		}
	}
	return hits, nil
}

func serperPage(string, string) string { return "" }

// serperPlaces asks Google Maps (through Serper) for the company and
// returns the address of the place that names it.
func serperPlaces(ctx context.Context, c *Client, q Query, region string) (Result, error) {
	var res Result
	query := q.Name
	if len(q.Places) > 0 {
		query += " " + strings.Join(q.Places, " ")
	}
	r, err := c.serperCall(ctx, "places", query, region)
	tr(ctx).Ev("places", "q", query, "n", len(r.Places), "err", err)
	if err != nil {
		return res, err
	}
	for _, p := range r.Places {
		if p.Address == "" || !Mentions(normWords(p.Title+" "+p.Address), q) {
			continue
		}
		src := Source{ID: "search", Name: "Google Maps place " + p.Title + " (through Serper, not verified)", URL: "https://www.google.com/maps?cid=" + p.Cid, Kind: KindCommunity}
		res.Addresses = append(res.Addresses, Address{Text: p.Address, Source: src})
		res.Sources = append(res.Sources, src)
		if p.Website != "" {
			res.Websites = append(res.Websites, p.Website)
		}
		break
	}
	return res, nil
}
