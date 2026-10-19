package lookup

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"
	"unicode"
)

var chAPI = "https://api.company-information.service.gov.uk"

func chAuth(key string) map[string]string {
	// HTTP basic authentication: the key is the user name, no password.
	return map[string]string{"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(key+":"))}
}

// companiesHouse queries the UK government register (free API key from
// https://developer.company-information.service.gov.uk/).
func companiesHouse(ctx context.Context, c *Client, key string, q Query) (Result, error) {
	var search struct {
		Items []struct {
			Title          string `json:"title"`
			CompanyNumber  string `json:"company_number"`
			CompanyStatus  string `json:"company_status"`
			AddressSnippet string `json:"address_snippet"`
		} `json:"items"`
	}
	u := chAPI + "/search/companies?" + url.Values{"q": {q.Name}, "items_per_page": {"10"}}.Encode()
	if err := c.getJSON(ctx, u, chAuth(key), &search); err != nil {
		return Result{}, err
	}
	var res Result
	for _, it := range search.Items {
		if !SameEntity(q, it.Title) || it.CompanyNumber == "" {
			continue
		}
		src := Source{
			ID:   "companies-house",
			Name: "Companies House " + it.CompanyNumber + " (" + it.Title + ", " + it.CompanyStatus + ")",
			URL:  "https://find-and-update.company-information.service.gov.uk/company/" + url.PathEscape(it.CompanyNumber),
			Kind: KindGovernment,
		}
		res.Sources = append(res.Sources, src)
		if it.AddressSnippet != "" {
			res.Addresses = append(res.Addresses, Address{Text: it.AddressSnippet + ", United Kingdom", Source: src})
		}
		var off struct {
			Items []struct {
				Name        string `json:"name"`
				OfficerRole string `json:"officer_role"`
				ResignedOn  string `json:"resigned_on"`
			} `json:"items"`
		}
		ou := chAPI + "/company/" + url.PathEscape(it.CompanyNumber) + "/officers?items_per_page=50"
		if err := c.getJSON(ctx, ou, chAuth(key), &off); err != nil {
			res.Notes = append(res.Notes, "Companies House officers: "+sanitizeErr(err))
			continue
		}
		for _, o := range off.Items {
			if o.ResignedOn != "" {
				continue
			}
			res.addPerson(Person{Name: chName(o.Name), Roles: []string{chRole(o.OfficerRole)}, Sources: []Source{src}})
		}
	}
	if len(res.Sources) == 0 {
		res.Notes = append(res.Notes, "Companies House: no company with this exact name")
	}
	return res, nil
}

// chName turns "DOE, Jane Mary" into "Jane Mary Doe".
func chName(s string) string {
	last, first, ok := strings.Cut(s, ",")
	if !ok {
		return titleCase(s)
	}
	return strings.TrimSpace(strings.TrimSpace(first) + " " + titleCase(strings.TrimSpace(last)))
}

func titleCase(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		for j := 1; j < len(r); j++ {
			if r[j-1] == '-' || r[j-1] == '\'' {
				r[j] = unicode.ToUpper(r[j])
			}
		}
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

func chRole(r string) string {
	switch r {
	case "director":
		return "Director"
	case "secretary":
		return "Company Secretary"
	case "corporate-director":
		return "Corporate Director"
	case "corporate-secretary":
		return "Corporate Secretary"
	case "llp-member":
		return "LLP Member"
	case "llp-designated-member":
		return "Designated LLP Member"
	case "nominee-director":
		return "Nominee Director"
	}
	return titleCase(strings.ReplaceAll(r, "-", " "))
}
