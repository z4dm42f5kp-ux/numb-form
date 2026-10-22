package reportdate

import (
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/docx"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

func TestParse(t *testing.T) {
	for in, want := range map[string]string{
		"16/10/2026": "16-10-2026", "16.10.2026": "16-10-2026", "6-1-2026": "06-01-2026", "2026-10-16": "16-10-2026",
		"10/16/2026": "16-10-2026", "16/10/26": "16-10-2026", "16 October 2026": "16-10-2026", "16. Oktober 2026": "16-10-2026",
		"16 de octubre de 2026": "16-10-2026", "October 16, 2026": "16-10-2026", "on 3 Mar 2025, signed": "03-03-2025",
	} {
		if got, ok := Parse(in); !ok || got != want {
			t.Errorf("Parse(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	for _, in := range []string{"", "Ignore at all costs", "32/01/2026", "29/02/2025", "HRB 12345", "Report 2026"} {
		if got, ok := Parse(in); ok {
			t.Errorf("Parse(%q) = %q", in, got)
		}
	}
}

func TestFromText(t *testing.T) {
	text := "Customer ID: 12/05/2020\nName: Foo GmbH\nDate of the evaluation:\t03.04.2026\nPerson: Max\n"
	if d, ok := FromText(text); !ok || d != "03-04-2026" {
		t.Errorf("got %q %v", d, ok)
	}
	if _, ok := FromText("no dates here"); ok {
		t.Error("found a date in no text")
	}
}

func TestFromForm(t *testing.T) {
	b, _, err := docx.Format{}.Fill(testfix.Template(t), form.Values{form.FieldName: "Foo GmbH", form.FieldDate: "16/10/2026"})
	if err != nil {
		t.Fatal(err)
	}
	if d, how, ok := FromForm(b, "Appendix A - Risk Assesment - Foo GmbH.docx"); !ok || d != "16-10-2026" || how != "Date of the evaluation" {
		t.Errorf("got %q %q %v", d, how, ok)
	}
	if _, _, ok := FromForm(testfix.Template(t), "empty.docx"); ok {
		t.Error("found a date in the empty template")
	}
}
