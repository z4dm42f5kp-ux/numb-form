package lookup

import (
	"encoding/json"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// ldOrg is the part of schema.org Organization data that is used.
type ldOrg struct {
	Type      any               `json:"@type"`
	LegalName string            `json:"legalName"`
	Name      string            `json:"name"`
	Address   any               `json:"address"`
	Graph     []json.RawMessage `json:"@graph"`
	Founder   json.RawMessage   `json:"founder"`
	Employee  json.RawMessage   `json:"employee"`
	Member    json.RawMessage   `json:"member"`
}

// ldPerson is a schema.org Person with a job title.
type ldPerson struct {
	Name     string `json:"name"`
	JobTitle any    `json:"jobTitle"`
}

// ldPeople reads persons whose job title is a representative role.
func ldPeople(raw json.RawMessage, src Source) []Person {
	if len(raw) == 0 {
		return nil
	}
	var list []ldPerson
	if json.Unmarshal(raw, &list) != nil {
		var one ldPerson
		if json.Unmarshal(raw, &one) != nil {
			return nil
		}
		list = []ldPerson{one}
	}
	var out []Person
	for _, p := range list {
		title := ""
		switch v := p.JobTitle.(type) {
		case string:
			title = v
		case []any:
			if len(v) > 0 {
				title, _ = v[0].(string)
			}
		}
		role, ok := TranslateRole(title)
		n, okName := cleanName(p.Name)
		if ok && okName {
			out = append(out, Person{Name: n, Roles: []string{role}, Sources: []Source{src}})
		}
	}
	return out
}

func ldAddress(a any) string {
	switch v := a.(type) {
	case string:
		return textutil.OneLine(v)
	case map[string]any:
		get := func(k string) string {
			switch x := v[k].(type) {
			case string:
				return textutil.OneLine(x)
			case map[string]any:
				if n, ok := x["name"].(string); ok {
					return textutil.OneLine(n)
				}
			}
			return ""
		}
		var parts []string
		if s := get("streetAddress"); s != "" {
			parts = append(parts, s)
		}
		if pc := strings.TrimSpace(get("postalCode") + " " + get("addressLocality")); pc != "" {
			parts = append(parts, pc)
		}
		if c := get("addressCountry"); c != "" {
			parts = append(parts, CountryName(c))
		}
		if len(parts) >= 2 {
			return strings.Join(parts, ", ")
		}
	case []any:
		if len(v) > 0 {
			return ldAddress(v[0])
		}
	}
	return ""
}

func typeString(t any) string {
	switch v := t.(type) {
	case string:
		return v
	case []any:
		var parts []string
		for _, x := range v {
			if s, ok := x.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, " ")
	}
	return ""
}

// structured is what a page's schema.org data says about an organization.
type structured struct {
	legalName, address string
	people             []Person
}

// structuredData reads schema.org Organization data (JSON-LD) of a page.
func structuredData(p Page, src Source) structured {
	var out structured
	var walk func(b []byte, depth int)
	walk = func(b []byte, depth int) {
		if depth > 3 {
			return
		}
		var list []json.RawMessage
		if json.Unmarshal(b, &list) == nil {
			for _, x := range list {
				walk(x, depth+1)
			}
			return
		}
		var o ldOrg
		if json.Unmarshal(b, &o) != nil {
			return
		}
		for _, g := range o.Graph {
			walk(g, depth+1)
		}
		t := strings.ToLower(typeString(o.Type))
		if !strings.Contains(t, "organization") && !strings.Contains(t, "corporation") && !strings.Contains(t, "localbusiness") && !strings.Contains(t, "store") {
			return
		}
		if out.legalName == "" {
			out.legalName = textutil.OneLine(o.LegalName)
			if out.legalName == "" {
				out.legalName = textutil.OneLine(o.Name)
			}
		}
		if out.address == "" {
			out.address = ldAddress(o.Address)
		}
		for _, raw := range []json.RawMessage{o.Founder, o.Employee, o.Member} {
			out.people = append(out.people, ldPeople(raw, src)...)
		}
	}
	for _, b := range p.JSONLD {
		walk(b, 0)
	}
	if out.address == "" {
		out.address = p.metaAddress()
	}
	if out.legalName == "" {
		out.legalName = p.Micro["legalName"]
	}
	return out
}
