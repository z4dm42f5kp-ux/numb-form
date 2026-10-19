package lookup

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

// TestLive looks up real companies when NUMB_FORM_LIVE=1 (see
// .github/workflows/live.yml). It reports counts and source names only,
// as GitHub annotations, and never fails on missing data: the services
// change, and this test exists to show what they return.
func TestLive(t *testing.T) {
	if os.Getenv("NUMB_FORM_LIVE") != "1" {
		t.Skip("set NUMB_FORM_LIVE=1 to query the real services")
	}
	companies := []string{
		"Alfred Kärcher SE & Co. KG",
		"Red Bull GmbH",
		"Stadler Rail AG",
		"Mercadona S.A.",
		"Sonae SGPS, S.A.",
		"Grupo Bimbo, S.A.B. de C.V.",
		"Natura Cosméticos S.A.",
		"Arcor S.A.I.C.",
		"Shoprite Holdings Ltd",
		"Publix Super Markets, Inc.",
	}
	c := NewClient()
	for _, name := range companies {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		start := time.Now()
		res := Find(ctx, c, Config{Search: "duckduckgo", GuessDomains: true}, name, "")
		cancel()
		ids := map[string]int{}
		for _, s := range res.Sources {
			ids[s.ID]++
		}
		var src []string
		for k, v := range ids {
			src = append(src, fmt.Sprintf("%s:%d", k, v))
		}
		sort.Strings(src)
		verified, roles := 0, map[string]bool{}
		for _, p := range res.People {
			if p.Verified() {
				verified++
			}
			for _, r := range p.Roles {
				roles[r] = true
			}
		}
		var rl []string
		for r := range roles {
			rl = append(rl, r)
		}
		sort.Strings(rl)
		var idk []string
		for _, id := range res.IDs {
			idk = append(idk, id.Kind)
		}
		notes := strings.Join(res.Notes, " / ")
		if len(notes) > 300 {
			notes = notes[:300] + "…"
		}
		msg := fmt.Sprintf("%.0fs website=%q addresses=%d reps=%d verified=%d roles=[%s] legalnames=%d ids=[%s] register=%d sources=[%s] notes=%s",
			time.Since(start).Seconds(), res.Website, len(res.Addresses), len(res.People), verified, strings.Join(rl, "; "),
			len(res.LegalNames), strings.Join(idk, " "), len(res.Register), strings.Join(src, " "), notes)
		msg = strings.NewReplacer("\n", " ", "\r", " ", "%", "%25").Replace(msg)
		fmt.Printf("::notice title=%s::%s\n", strings.NewReplacer(",", " ", ":", " ").Replace(name), msg)
	}
}
