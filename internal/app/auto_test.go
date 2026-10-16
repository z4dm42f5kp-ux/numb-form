package app

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/lookup"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

var (
	srcCH   = lookup.Source{ID: "companies-house", Name: "CH", Kind: lookup.KindGovernment}
	srcLEI  = lookup.Source{ID: "gleif", Name: "GLEIF", Kind: lookup.KindRegulatory}
	srcSite = lookup.Source{ID: "website", Name: "site", Kind: lookup.KindCompany}
	srcWiki = lookup.Source{ID: "wikidata", Name: "wd", Kind: lookup.KindCommunity}
)

func sample() lookup.Result {
	return lookup.Result{
		Addresses: []lookup.Address{
			{Text: "Wiki City", Source: srcWiki},
			{Text: "Site Street 1", Source: srcSite},
			{Text: "LEI Street 2", Source: srcLEI},
		},
		People: []lookup.Person{
			{Name: "Reg Person", Roles: []string{"Director"}, Sources: []lookup.Source{srcCH}},
			{Name: "Site Person", Roles: []string{"Managing Director"}, Sources: []lookup.Source{srcSite}},
			{Name: "Wiki Person", Roles: []string{"Chief Executive Officer"}, Sources: []lookup.Source{srcWiki}},
		},
	}
}

func autoRun(t *testing.T, mut func(*config.Config)) form.Values {
	t.Helper()
	cfg := config.Defaults()
	cfg.Confirm, cfg.Person = "none", "P"
	mut(&cfg)
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	log, err := runlog.Open(t.TempDir() + "/log")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	var out bytes.Buffer
	r := &run{cfg: cfg, ui: ui.New(context.Background(), strings.NewReader(""), &out), log: log, date: "d", person: "P"}
	return r.autoValues("Foo GmbH", gathered{res: sample(), cfg: cfg})
}

func TestAutoSourcesAndTrust(t *testing.T) {
	v := autoRun(t, func(*config.Config) {})
	// Default order: companies-house gleif website wikidata.
	if v[form.FieldAddress] != "LEI Street 2" {
		t.Errorf("address %q", v[form.FieldAddress])
	}
	if !strings.HasPrefix(v[form.FieldReps], "Reg Person (Director), Site Person (Managing Director), Wiki Person") ||
		!strings.HasSuffix(v[form.FieldReps], CautionText) {
		t.Errorf("reps %q", v[form.FieldReps])
	}
	if v[form.FieldLegalForm] != "GmbH" {
		t.Errorf("legal form %q", v[form.FieldLegalForm])
	}

	v = autoRun(t, func(c *config.Config) { c.RepsMinTrust = "register" })
	if v[form.FieldReps] != "Reg Person (Director)" {
		t.Errorf("register-only reps %q", v[form.FieldReps])
	}

	v = autoRun(t, func(c *config.Config) {
		c.RepsSources = []string{"wikidata"}
		c.RepsMinTrust = "website"
	})
	if v[form.FieldReps] != NoRepsText {
		t.Errorf("untrusted-only reps %q", v[form.FieldReps])
	}

	v = autoRun(t, func(c *config.Config) { c.AddressSources = []string{"website", "gleif"} })
	if v[form.FieldAddress] != "Site Street 1 "+CautionText {
		t.Errorf("website-first address %q", v[form.FieldAddress])
	}

	v = autoRun(t, func(c *config.Config) {
		c.AddressSources = []string{"wikidata"}
		c.Caution = "never"
	})
	if v[form.FieldAddress] != "Wiki City" {
		t.Errorf("caution never %q", v[form.FieldAddress])
	}

	v = autoRun(t, func(c *config.Config) {
		c.RepsSources = nil
		c.RepsFallback = "blank"
	})
	if v[form.FieldReps] != "" {
		t.Errorf("blank fallback %q", v[form.FieldReps])
	}
}
