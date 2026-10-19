package docx

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

const (
	ct   = testfix.ContentTypes
	rels = testfix.Rels
)

func TestFillAndReadBack(t *testing.T) {
	tpl := testfix.Template(t)
	vals := form.Values{
		form.FieldName:      "Foo Co., Ltda",
		form.FieldLegalForm: "Ltda",
		form.FieldAddress:   "Calle 1 # 2-3, Bogotá & <Co>",
		form.FieldRegistry:  "No commercial registry number was found online consult internal sources for further information",
		form.FieldReps:      "Jane Doe (General Manager, Legal Representative)",
		form.FieldDate:      "16/10/2026",
		form.FieldPerson:    "Max Muster",
	}
	out, warn, err := fmtx.Fill(tpl, vals)
	if err != nil {
		t.Fatal(err)
	}
	if len(warn) != 0 {
		t.Fatalf("warnings: %v", warn)
	}
	got, err := fmtx.ReadBack(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range form.AllFields {
		want := vals[f]
		if got[f] != want {
			t.Errorf("%s = %q, want %q", f.Title(), got[f], want)
		}
	}
	// Filling an already filled document must fail.
	if _, _, err := fmtx.Fill(out, vals); err == nil {
		t.Fatal("refilling a filled document succeeded")
	}
}

func TestEmptyValuesKeepTemplate(t *testing.T) {
	tpl := testfix.Template(t)
	out, _, err := fmtx.Fill(tpl, form.Values{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := fmtx.ReadBack(tpl)
	b, _ := fmtx.ReadBack(out)
	for _, f := range form.AllFields {
		if a[f] != b[f] {
			t.Fatalf("%s changed", f.Title())
		}
	}
}

func TestMissingSection(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range [][2]string{{"[Content_Types].xml", ct}, {"_rels/.rels", rels}, {"word/document.xml", `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Legal form:</w:t></w:r></w:p></w:body></w:document>`}} {
		w, _ := zw.Create(e[0])
		_, _ = w.Write([]byte(e[1]))
	}
	_ = zw.Close()
	if _, _, err := fmtx.Fill(buf.Bytes(), form.Values{form.FieldLegalForm: "GmbH"}); err == nil {
		t.Fatal("expected error without section A heading")
	}
}

func FuzzFill(f *testing.F) {
	f.Add("GmbH", "Street 1", "A B (CEO)")
	f.Add("<&>\x00\x1b[31m", "]]>", "\u202e")
	tpl := testfix.Template(f)
	f.Fuzz(func(t *testing.T, a, b, c string) {
		out, _, err := fmtx.Fill(tpl, form.Values{form.FieldLegalForm: a, form.FieldAddress: b, form.FieldReps: c})
		if err != nil {
			return
		}
		if _, err := fmtx.ReadBack(out); err != nil {
			t.Fatalf("filled output unreadable: %v", err)
		}
	})
}

var fmtx = Format{}
