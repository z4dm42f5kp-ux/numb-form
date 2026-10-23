package form_test

import (
	"archive/zip"
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
	form.FieldRegistry:  "No commercial registry number was found online consult internal sources for further information",
	form.FieldReps:      "Jane Doe (General Manager, Legal Representative), Łukasz Żółć (Director) 😀",
	form.FieldDate:      "16/10/2026",
	form.FieldPerson:    "Max Muster",
}

// Every format must fill every field, read back exactly what was written,
// and refuse to fill twice.
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
				if got[f] != want {
					t.Errorf("%s = %q, want %q", f.Title(), got[f], want)
				}
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

// A template without one of the labels must fail the check up front.
func TestCheckNeedsEveryLabel(t *testing.T) {
	for _, c := range cases(t) {
		if _, err := form.Check(c.f, c.tpl); err != nil {
			t.Fatalf("%s: complete template rejected: %v", c.name, err)
		}
	}
	broken := bytes.Replace(testfix.RTF(), []byte("Date of the evaluation:"), []byte("Datum:"), 1)
	if _, err := form.Check(rtf.Format{}, broken); err == nil {
		t.Fatal("template without the date label passed the check")
	}
}

func FuzzAllFormats(f *testing.F) {
	f.Add("GmbH", "Street 1", "A B (CEO)")
	f.Add("<&>\x00\x1b[31m", "]]>{\\}", "\u202e\U0001F600")
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

// A filled form is recognised as numb-form's own until anything in it
// changes, in section A or elsewhere.
func TestAsFilled(t *testing.T) {
	members := map[string]string{"docx": "word/document.xml", "odt": "content.xml"}
	for _, c := range cases(t) {
		t.Run(c.name, func(t *testing.T) {
			out, _, err := c.f.Fill(c.tpl, vals)
			if err != nil {
				t.Fatal(err)
			}
			if ok, err := form.AsFilled(c.f, c.tpl, out); err != nil || !ok {
				t.Fatalf("own form not recognised: %v %v", ok, err)
			}
			// The smallest edit outside section A: a line break at the end.
			edited := append(append([]byte(nil), out...), '\n')
			if m := members[c.name]; m != "" {
				zr, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
				if err != nil {
					t.Fatal(err)
				}
				x, err := form.ZipEntry(zr, m)
				if err != nil {
					t.Fatal(err)
				}
				if edited, err = form.ReplaceZipEntry(out, m, append(x, '\n')); err != nil {
					t.Fatal(err)
				}
			}
			if ok, _ := form.AsFilled(c.f, c.tpl, edited); ok {
				t.Fatal("an edited form counts as unchanged")
			}
			other := form.Values{}
			for k, v := range vals {
				other[k] = v
			}
			other[form.FieldAddress] = "Elsewhere 2"
			again, _, err := c.f.Fill(c.tpl, other)
			if err != nil {
				t.Fatal(err)
			}
			if ok, _ := form.AsFilled(c.f, c.tpl, again); !ok {
				t.Fatal("a form filled with other values is not recognised")
			}
		})
	}
}
