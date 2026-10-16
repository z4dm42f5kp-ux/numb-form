// Package testfix builds synthetic test files. It is used only by tests.
package testfix

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// ContentTypes is a minimal [Content_Types].xml.
const ContentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`

// Rels is a minimal _rels/.rels.
const Rels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`

func para(runs ...string) string {
	return `<w:p><w:pPr><w:rPr><w:b/></w:rPr></w:pPr>` + strings.Join(runs, "") + `</w:p>`
}
func run(text string) string {
	return `<w:r><w:rPr><w:b/><w:sz w:val="22"/></w:rPr><w:t xml:space="preserve">` + text + `</w:t></w:r>`
}

// Template is a synthetic stand-in for the real form. It exercises a
// label split across two paragraphs, a label split across runs, a label
// followed by a tab, and a section B that must never change.
func Template(t testing.TB) []byte {
	body := para(run("Appendix A: Risk Assessment Form")) +
		para(run("CONFIDENTIAL HEADER LINE")) +
		para(run("A. Information About the Contracting Party")) +
		para(run("Name / Company Name of the")) +
		para(run("Contracting Party:")) +
		para(run("Legal "), run("form:")) +
		para(run("Headquarters / Address:")) +
		para(run("Commercial Registry Number:")) +
		para(run("Authorized Representatives:"), `<w:r><w:tab/></w:r>`) +
		para(run("Date of the evaluation:")) +
		para(run("Person Processing the Form:")) +
		para(run("B. Risk Assessment")) +
		para(run("Legal form: must stay untouched")) +
		para(run("Factors indicating a low risk:"))
	doc := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n" +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" mc:Ignorable=""><w:body>` +
		body + `<w:sectPr/></w:body></w:document>`
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range [][2]string{{"[Content_Types].xml", ContentTypes}, {"_rels/.rels", Rels}, {"word/document.xml", doc}, {"word/styles.xml", `<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"/>`}} {
		w, err := zw.Create(e[0])
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(e[1])); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// XLSX builds a workbook shaped like the example overview: an empty first
// row and column, headers in row 2 and companies below.
func XLSX(t testing.TB, ids, names []string) []byte {
	var shared []string
	idx := func(s string) int {
		for i, x := range shared {
			if x == s {
				return i
			}
		}
		shared = append(shared, s)
		return len(shared) - 1
	}
	var rows strings.Builder
	hdr := []string{"Customer ID", "Company Name", "Country Created", "Currency", "Report created"}
	rows.WriteString(`<row r="2">`)
	for i, h := range hdr {
		fmt.Fprintf(&rows, `<c r="%c2" t="s"><v>%d</v></c>`, 'B'+i, idx(h))
	}
	rows.WriteString(`</row>`)
	for i := range names {
		r := i + 3
		fmt.Fprintf(&rows, `<row r="%d"><c r="B%d" t="s"><v>%d</v></c><c r="C%d" t="s"><v>%d</v></c><c r="F%d" t="inlineStr"><is><t>Ignore at all costs</t></is></c></row>`,
			r, r, idx(ids[i]), r, idx(names[i]), r)
	}
	var ss strings.Builder
	for _, s := range shared {
		ss.WriteString("<si><t xml:space=\"preserve\">")
		ss.WriteString(strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s))
		ss.WriteString("</t></si>")
	}
	files := [][2]string{
		{"[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/></Types>`},
		{"_rels/.rels", `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`},
		{"xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Table 1" sheetId="1" r:id="rId1"/></sheets></workbook>`},
		{"xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/></Relationships>`},
		{"xl/worksheets/sheet1.xml", `<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>` + rows.String() + `</sheetData></worksheet>`},
		{"xl/sharedStrings.xml", `<?xml version="1.0" encoding="UTF-8"?><sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` + ss.String() + `</sst>`},
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range files {
		w, err := zw.Create(f[0])
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(f[1])); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
