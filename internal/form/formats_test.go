package form_test

import (
	"bytes"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/docx"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/odt"
	"github.com/z4dm42f5kp-ux/numb-form/internal/rtf"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

type tc struct {
	name string
	f    form.Filler
	tpl  []byte
}

func cases(t testing.TB) []tc {
	return []tc{
		{"docx", docx.Format{}, testfix.Template(t)},
		{"odt", odt.Format{}, testfix.ODT(t)},
		{"fodt", odt.Format{Flat: true}, testfix.FODT()},
		{"rtf", rtf.Format{}, testfix.RTF()},
	}
}

var vals = form.Values{
	form.FieldName:      "Foo Co., Ltda",
	form.FieldLegalForm: "Ltda",
	form.FieldAddress:   "Calle 1 # 2-3, Bogotá & <Co> {x} \\ y",
	form.FieldRegistry:  "SHOULD NOT APPEAR",
	form.FieldReps:      "Jane Doe (General Manager, Legal Representative), Łukasz Żółć (Director) 😀",
	form.FieldDate:      "16/10/2026",
	form.FieldPerson:    "Max Muster",
}

// Every format must fill every field, read back exactly what was written,
// never write the registry number and refuse to fill twice.
func TestAllFormats(t *testing.T) {
	for _, c := range cases(t) {
		t.Run(c.name, func(t *testing.T) {
			out, warn, err := c.f.Fill(c.tpl, vals)
			if err != nil {
				t.Fatal(err)
			}
			if len(warn) != 0 {
				t.Fatalf("warnings: %v", warn)
			}
			got, err := c.f.ReadBack(out)
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range form.AllFields {
				want := vals[f]
				if f == form.FieldRegistry {
					want = ""
				}
				if got[f] != want {
					t.Errorf("%s = %q, want %q", f.Title(), got[f], want)
				}
			}
			if bytes.Contains(out, []byte("SHOULD NOT APPEAR")) {
				t.Fatal("registry number written")
			}
			if _, _, err := c.f.Fill(out, vals); err == nil {
				t.Fatal("refilling succeeded")
			}
			empty, _, err := c.f.Fill(c.tpl, form.Values{})
			if err != nil {
				t.Fatal(err)
			}
			if c.name != "docx" && c.name != "odt" && !bytes.Equal(empty, c.tpl) {
				t.Fatal("an empty fill changed the file")
			}
		})
	}
}

func FuzzAllFormats(f *testing.F) {
	f.Add("GmbH", "Street 1", "A B (CEO)")
	f.Add("<&>\x00\x1b[31m", "]]>{\\}", "‮\U0001F600")
	cs := cases(f)
	f.Fuzz(func(t *testing.T, a, b, c string) {
		for _, x := range cs {
			out, _, err := x.f.Fill(x.tpl, form.Values{form.FieldLegalForm: a, form.FieldAddress: b, form.FieldReps: c})
			if err != nil {
				continue
			}
			if _, err := x.f.ReadBack(out); err != nil {
				t.Fatalf("%s: filled output unreadable: %v", x.name, err)
			}
		}
	})
}
