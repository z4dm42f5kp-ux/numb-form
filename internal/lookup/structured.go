package lookup

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

var (
	reLDJSON = regexp.MustCompile(`(?is)<script[^>]+type\s*=\s*["']application/ld\+json["'][^>]*>(.*?)</script\s*>`)
)

// ldOrg is the part of schema.org Organization data that is used.
type ldOrg struct {
	Type      any               `json:"@type"`
	LegalName string            `json:"legalName"`
	Name      string            `json:"name"`
	Address   any               `json:"address"`
	Graph     []json.RawMessage `json:"@graph"`
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

// structuredData reads schema.org Organization data embedded in a page.
func structuredData(page string) (legalName, address string) {
	var walk func(b []byte, depth int)
	walk = func(b []byte, depth int) {
		if depth > 3 || (legalName != "" && address != "") {
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
		if !strings.Contains(t, "organization") && !strings.Contains(t, "corporation") && !strings.Contains(t, "localbusiness") {
			return
		}
		if legalName == "" {
			legalName = textutil.OneLine(o.LegalName)
		}
		if address == "" {
			address = ldAddress(o.Address)
		}
	}
	for _, m := range reLDJSON.FindAllStringSubmatch(page, 5) {
		walk([]byte(m[1]), 0)
	}
	return legalName, address
}
