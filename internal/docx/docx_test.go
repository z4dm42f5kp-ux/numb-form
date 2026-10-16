package docx

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

const (
	ct   = testfix.ContentTypes
	rels = testfix.Rels
)

func TestFillAndReadBack(t *testing.T) {
	tpl := testfix.Template(t)
	vals := Values{
		FieldName:      "Foo Co., Ltda",
		FieldLegalForm: "Ltda",
		FieldAddress:   "Calle 1 # 2-3, Bogotá & <Co>",
		FieldRegistry:  "SHOULD NOT APPEAR",
		FieldReps:      "Jane Doe (General Manager, Legal Representative)",
		FieldDate:      "16/10/2026",
		FieldPerson:    "Max Muster",
	}
	out, warn, err := Fill(tpl, vals)
	if err != nil {
		t.Fatal(err)
	}
	if len(warn) != 0 {
		t.Fatalf("warnings: %v", warn)
	}
	got, err := ReadBack(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range AllFields {
		want := vals[f]
		if f == FieldRegistry {
			want = ""
		}
		if got[f] != want {
			t.Errorf("%s = %q, want %q", f.Title(), got[f], want)
		}
	}
	if bytes.Contains(out, []byte("SHOULD NOT APPEAR")) {
		t.Fatal("registry number was written")
	}
	// Filling an already filled document must fail.
	if _, _, err := Fill(out, vals); err == nil {
		t.Fatal("refilling a filled document succeeded")
	}
}

func TestEmptyValuesKeepTemplate(t *testing.T) {
	tpl := testfix.Template(t)
	out, _, err := Fill(tpl, Values{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := ReadBack(tpl)
	b, _ := ReadBack(out)
	for _, f := range AllFields {
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
	if _, _, err := Fill(buf.Bytes(), Values{FieldLegalForm: "GmbH"}); err == nil {
		t.Fatal("expected error without section A heading")
	}
}

func FuzzFill(f *testing.F) {
	f.Add("GmbH", "Street 1", "A B (CEO)")
	f.Add("<&>\x00\x1b[31m", "]]>", "‮")
	tpl := testfix.Template(f)
	f.Fuzz(func(t *testing.T, a, b, c string) {
		out, _, err := Fill(tpl, Values{FieldLegalForm: a, FieldAddress: b, FieldReps: c})
		if err != nil {
			return
		}
		if _, err := ReadBack(out); err != nil {
			t.Fatalf("filled output unreadable: %v", err)
		}
	})
}
