package lookup

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// TestLive looks up the companies listed in the LIVE_COMPANIES secret
// (one per line) against the real services, with the optional keys from
// the environment, and writes every result with its trace to the file
// named by LIVE_OUT. live.yml encrypts that file before anything is
// shown, so company names never appear in logs. It never fails on missing
// data: it measures what the web gives.
func TestLive(t *testing.T) {
	names := strings.TrimSpace(os.Getenv("LIVE_COMPANIES"))
	out := os.Getenv("LIVE_OUT")
	if os.Getenv("NUMB_FORM_LIVE") != "1" || names == "" || out == "" {
		t.Skip("set NUMB_FORM_LIVE=1, LIVE_COMPANIES and LIVE_OUT to query the real services")
	}
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c := NewClient()
	cfg := Config{Browser: "auto", GuessDomains: true,
		SerperKey: os.Getenv("NUMBFORM_SERPER_KEY"), TavilyKey: os.Getenv("NUMBFORM_TAVILY_KEY")}
	enc := json.NewEncoder(f)
	for _, name := range strings.Split(names, "\n") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		var trace bytes.Buffer
		tc := NewTrace(&trace)
		tc.Company(name)
		ctx, cancel := context.WithTimeout(WithTrace(context.Background(), tc), 4*time.Minute)
		start := time.Now()
		res := Find(ctx, c, cfg, name, "")
		cancel()
		type person struct {
			Name     string
			Roles    []string
			Verified bool
			Sources  []string
		}
		var ps []person
		for _, p := range res.People {
			var src []string
			for _, s := range p.Sources {
				src = append(src, s.Name+" "+s.URL)
			}
			ps = append(ps, person{p.Name, p.Roles, p.Verified(), src})
		}
		var as []string
		for _, a := range res.Addresses {
			as = append(as, a.Text+" | "+a.Source.Name+" "+a.Source.URL)
		}
		_ = enc.Encode(map[string]any{
			"name": name, "seconds": int(time.Since(start).Seconds()), "website": res.Website,
			"people": ps, "addresses": as, "notes": res.Notes, "trace": strings.Split(strings.TrimSpace(trace.String()), "\n"),
		})
	}
}
