package lookup

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

var wdAPI = "https://www.wikidata.org/w/api.php"

type wdClaim struct {
	Rank     string `json:"rank"`
	Mainsnak struct {
		Datavalue struct {
			Value json.RawMessage `json:"value"`
		} `json:"datavalue"`
	} `json:"mainsnak"`
	Qualifiers map[string]json.RawMessage `json:"qualifiers"`
}

type wdEntity struct {
	Labels map[string]struct {
		Value string `json:"value"`
	} `json:"labels"`
	Claims map[string][]wdClaim `json:"claims"`
}

func (e wdEntity) label() string {
	for _, l := range []string{"en", "de", "es", "fr", "pt"} {
		if v, ok := e.Labels[l]; ok {
			return v.Value
		}
	}
	for _, v := range e.Labels {
		return v.Value
	}
	return ""
}

func (c wdClaim) current() bool {
	if c.Rank == "deprecated" {
		return false
	}
	_, ended := c.Qualifiers["P582"] // end time
	return !ended
}

func (c wdClaim) itemID() string {
	var v struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(c.Mainsnak.Datavalue.Value, &v) == nil {
		return v.ID
	}
	return ""
}

func (c wdClaim) str() string {
	var s string
	if json.Unmarshal(c.Mainsnak.Datavalue.Value, &s) == nil {
		return s
	}
	return ""
}

var wdRoles = []struct{ prop, role string }{
	{"P169", "Chief Executive Officer"},
	{"P1037", "Director/Manager"},
	{"P488", "Chairperson"},
}

// wikidata is the lowest-ranked source: community-edited, never treated
// as verified, but useful for finding the official website.
func wikidata(ctx context.Context, c *Client, q Query) (Result, error) {
	var search struct {
		Search []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		} `json:"search"`
	}
	u := wdAPI + "?" + url.Values{
		"action": {"wbsearchentities"}, "format": {"json"}, "language": {"en"}, "uselang": {"en"},
		"type": {"item"}, "limit": {"7"}, "search": {q.Name},
	}.Encode()
	if err := c.getJSON(ctx, u, nil, &search); err != nil {
		return Result{}, err
	}
	// Items whose label is the exact entity may supply people. Items that
	// only share the name without the legal form (often the group or a
	// parent) may only suggest a website: authority is never inferred
	// from a related entity.
	var ids []string
	exact := map[string]bool{}
	baseKey := textutil.MatchKeys(q.Base)[0]
	for _, s := range search.Search {
		if len(ids) >= 2 {
			break
		}
		switch {
		case SameEntity(q, s.Label):
			ids = append(ids, s.ID)
			exact[s.ID] = true
		case len(baseKey) >= 4 && textutil.MatchKeys(s.Label)[0] == baseKey:
			ids = append(ids, s.ID)
		}
	}
	var res Result
	if len(ids) == 0 {
		res.Notes = append(res.Notes, "Wikidata: no item with this exact name")
		return res, nil
	}
	ents, err := wdEntities(ctx, c, ids, "claims|labels")
	if err != nil {
		return res, err
	}
	var refs []string
	for _, e := range ents {
		for _, r := range wdRoles {
			for _, cl := range e.Claims[r.prop] {
				if id := cl.itemID(); id != "" && cl.current() {
					refs = append(refs, id)
				}
			}
		}
		for _, cl := range e.Claims["P159"] {
			if id := cl.itemID(); id != "" && cl.current() {
				refs = append(refs, id)
			}
		}
	}
	labels := map[string]string{}
	if len(refs) > 0 {
		if len(refs) > 40 {
			refs = refs[:40]
		}
		le, err := wdEntities(ctx, c, refs, "labels")
		if err == nil {
			for id, e := range le {
				labels[id] = e.label()
			}
		}
	}
	for _, id := range ids {
		e, ok := ents[id]
		if !ok {
			continue
		}
		src := Source{Name: "Wikidata " + id + " (" + e.label() + ")", URL: "https://www.wikidata.org/wiki/" + id, Kind: KindCommunity}
		res.Sources = append(res.Sources, src)
		for _, cl := range e.Claims["P856"] {
			if w := cl.str(); w != "" && cl.current() {
				res.Websites = append(res.Websites, w)
			}
		}
		if !exact[id] {
			res.Notes = append(res.Notes, "Wikidata "+id+" ("+e.label()+") matches the name without the legal form; used only to suggest a website")
			continue
		}
		for _, cl := range e.Claims["P159"] {
			if l := labels[cl.itemID()]; l != "" && cl.current() {
				res.Addresses = append(res.Addresses, Address{Text: l, Source: Source{Name: src.Name + " headquarters city only", URL: src.URL, Kind: KindCommunity}})
			}
		}
		for _, r := range wdRoles {
			for _, cl := range e.Claims[r.prop] {
				if l := labels[cl.itemID()]; l != "" && cl.current() {
					res.addPerson(Person{Name: l, Roles: []string{r.role}, Sources: []Source{src}})
				}
			}
		}
	}
	return res, nil
}

func wdEntities(ctx context.Context, c *Client, ids []string, props string) (map[string]wdEntity, error) {
	var out struct {
		Entities map[string]wdEntity `json:"entities"`
	}
	u := wdAPI + "?" + url.Values{
		"action": {"wbgetentities"}, "format": {"json"}, "ids": {strings.Join(ids, "|")},
		"props": {props}, "languages": {"en|de|es|fr|pt"},
	}.Encode()
	if err := c.getJSON(ctx, u, nil, &out); err != nil {
		return nil, err
	}
	return out.Entities, nil
}
