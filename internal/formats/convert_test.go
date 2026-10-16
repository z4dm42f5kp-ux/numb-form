package formats

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/convert"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

// These tests use a real LibreOffice. They are skipped when it is not
// installed, unless NUMBFORM_TEST_SOFFICE=1 (set in CI) requires it.
func soffice(t *testing.T) string {
	so, err := convert.FindSoffice("")
	if err != nil {
		if os.Getenv("NUMBFORM_TEST_SOFFICE") == "1" {
			t.Fatal(err)
		}
		t.Skip("LibreOffice not installed")
	}
	return so
}

var testVals = form.Values{
	form.FieldName:      "Bär & Söhne GmbH",
	form.FieldLegalForm: "GmbH",
	form.FieldAddress:   "Musterstraße 12, 80333 München, Germany",
	form.FieldReps:      "Erika Mustermann (Managing Director)",
	form.FieldDate:      "16/10/2026",
	form.FieldPerson:    "Max Muster",
}

func roundTrip(t *testing.T, tpl *Template) {
	t.Helper()
	ctx := context.Background()
	if _, err := tpl.Check(); err != nil {
		t.Fatalf("check: %v", err)
	}
	out, _, err := tpl.Fill(ctx, testVals)
	if err != nil {
		t.Fatalf("fill: %v", err)
	}
	got, err := tpl.ReadBack(ctx, out)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	for _, f := range form.AllFields {
		if got[f] != testVals[f] {
			t.Errorf("%s = %q, want %q", f.Title(), got[f], testVals[f])
		}
	}
}

func write(t *testing.T, name string, b []byte) string {
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, b, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// Templates saved by LibreOffice in every format numb-form edits directly.
func TestLibreOfficeTemplates(t *testing.T) {
	so := soffice(t)
	ctx := context.Background()
	src := testfix.Template(t)
	for _, ext := range []string{"odt", "fodt", "rtf", "docx"} {
		t.Run(ext, func(t *testing.T) {
			b, err := convert.Soffice(ctx, so, src, ".docx", ext)
			if err != nil {
				t.Fatal(err)
			}
			tpl, err := LoadTemplate(ctx, write(t, "Appendix A - Risk Assesment."+ext, b), config.Defaults())
			if err != nil {
				t.Fatal(err)
			}
			roundTrip(t, tpl)
		})
	}
}

// Legacy .doc: refused without legacy_formats, then converted, filled and
// written as .docx or, with legacy_output = same, as .doc.
func TestLegacyDoc(t *testing.T) {
	so := soffice(t)
	ctx := context.Background()
	doc, err := convert.Soffice(ctx, so, testfix.Template(t), ".docx", "doc")
	if err != nil {
		t.Fatal(err)
	}
	p := write(t, "Appendix A - Risk Assesment.doc", doc)
	cfg := config.Defaults()
	if _, err := LoadTemplate(ctx, p, cfg); err == nil {
		t.Fatal(".doc accepted without legacy_formats")
	}
	cfg.LegacyFormats = true
	for _, out := range []string{"docx", "same"} {
		cfg.LegacyOutput = out
		tpl, err := LoadTemplate(ctx, p, cfg)
		if err != nil {
			t.Fatal(err)
		}
		if !tpl.Legacy {
			t.Fatal("legacy flag not set")
		}
		roundTrip(t, tpl)
	}
}

// Spreadsheets saved by LibreOffice, including legacy .xls (BIFF8), read
// by numb-form's own parsers.
func TestLibreOfficeSheets(t *testing.T) {
	so := soffice(t)
	ctx := context.Background()
	names := []string{"Foo C.A. de C.V.", "Bär & Söhne GmbH", "Łódź Sp. z o.o."}
	ids := []string{"D1", "D2", "D3"}
	src := testfix.XLSX(t, ids, names)
	for _, ext := range []string{"xls", "ods", "fods", "xlsx"} {
		t.Run(ext, func(t *testing.T) {
			b, err := convert.Soffice(ctx, so, src, ".xlsx", ext)
			if err != nil {
				t.Fatal(err)
			}
			rows, _, err := sheet.Read(b, "."+ext, sheet.Options{Legacy: true})
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != len(names) {
				t.Fatalf("got %d rows %+v", len(rows), rows)
			}
			for i := range names {
				if rows[i].Name != names[i] || rows[i].CustomerID != ids[i] {
					t.Errorf("row %d = %+v", i, rows[i])
				}
			}
		})
	}
}

// Apple Pages: only when a sample is provided (NUMBFORM_TEST_PAGES=file),
// because no CI runner has Pages or a licence to ship sample files.
func TestPages(t *testing.T) {
	p := os.Getenv("NUMBFORM_TEST_PAGES")
	if p == "" {
		t.Skip("set NUMBFORM_TEST_PAGES to a .pages template to run")
	}
	tpl, err := LoadTemplate(context.Background(), p, config.Defaults())
	if err != nil {
		t.Fatal(err)
	}
	roundTrip(t, tpl)
}
