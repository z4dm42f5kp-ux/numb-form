package lookup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Tavily (tavily.com) searches the web and writes a short answer from
// the pages it found, with their addresses. It needs a free key
// (NUMBFORM_TAVILY_KEY; 1,000 a month without a credit card). It is the
// last resort: what it says is never taken as verified.
var tavilyAPI = "https://api.tavily.com/search"

// AIPrompt is the question asked; {name} and {place} are replaced. It can
// be replaced with the text of ai-prompt.txt (Tavily accepts up to 400
// characters).
var AIPrompt = `Who are the current legal representatives (full names and roles) and what is the registered headquarters address of the company "{name}"{place}? Answer only as: Representatives: Name (Role), Name (Role). Address: street, city, country. Write Unknown when not found.`

type tavilyResp struct {
	Answer  string `json:"answer"`
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

func aiQuestion(q Query) string {
	place := ""
	if len(q.Places) > 0 {
		place = " (" + strings.Join(q.Places, ", ") + ")"
	} else if len(q.Countries) > 0 {
		place = " (" + CountryName(q.Countries[0]) + ")"
	}
	s := strings.NewReplacer("{name}", q.Name, "{place}", place).Replace(AIPrompt)
	if len(s) > 400 {
		s = s[:400]
	}
	return s
}

// tavily asks for the representatives and address, reads the answer and
// the pages it was based on.
func tavily(ctx context.Context, c *Client, q Query) (Result, error) {
	var res Result
	if c.tavilyKey == "" {
		return res, nil
	}
	body, _ := json.Marshal(map[string]any{
		"query": aiQuestion(q), "search_depth": "advanced", "include_answer": "advanced", "max_results": 8,
	})
	raw, _, _, err := c.do(ctx, http.MethodPost, tavilyAPI, string(body), map[string]string{
		"Authorization": "Bearer " + c.tavilyKey, "Content-Type": "application/json", "Accept": "application/json",
	})
	var r tavilyResp
	if err == nil {
		err = json.Unmarshal(raw, &r)
	}
	tr(ctx).Ev("ai", "q", aiQuestion(q), "answer", r.Answer, "n", len(r.Results), "err", err)
	if err != nil {
		return res, err
	}
	var hits []Hit
	var cited []string
	for _, x := range r.Results {
		hits = append(hits, Hit{URL: x.URL, Title: x.Title, Snippet: x.Content})
		if u, err := url.Parse(x.URL); err == nil && len(cited) < 3 {
			cited = append(cited, u.Hostname())
		}
	}
	res.merge(mineSnippets(hits, q))
	if a := strings.TrimSpace(r.Answer); a != "" {
		src := Source{ID: "ai", Name: fmt.Sprintf("AI answer (Tavily) from %s (not verified)", strings.Join(cited, ", ")), URL: "https://tavily.com", Kind: KindCommunity}
		lines := snippetLines(strings.NewReplacer("Representatives:", "\nRepresentatives:", "Address:", "\nAddress:").Replace(a))
		for _, l := range lines {
			low := strings.ToLower(l)
			switch {
			case strings.HasPrefix(low, "representatives:") && !strings.Contains(low, "unknown"):
				for _, p := range splitPeople(strings.TrimSpace(l[len("representatives:"):]), "", src) {
					p.Roles = dropEmpty(p.Roles)
					if len(p.Roles) == 0 {
						p.Roles = []string{"Legal Representative"}
					}
					res.addPerson(p)
				}
			case strings.HasPrefix(low, "address:") && !strings.Contains(low, "unknown"):
				res.Addresses = append(res.Addresses, Address{Text: strings.Trim(strings.TrimSpace(l[len("address:"):]), ". "), Source: src})
			}
		}
		res.Sources = append(res.Sources, src)
	}
	return res, nil
}

func dropEmpty(s []string) []string {
	var out []string
	for _, x := range s {
		if strings.TrimSpace(x) != "" {
			out = append(out, x)
		}
	}
	return out
}
