package testfix

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

const odfHead = `<?xml version="1.0" encoding="UTF-8"?>` +
	`<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" ` +
	`xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" office:version="1.3">` +
	`<office:automatic-styles><style:style style:name="B" style:family="text"/></office:automatic-styles><office:body><office:text>`

func odp(inner string) string { return `<text:p text:style-name="Standard">` + inner + `</text:p>` }
func ods(t string) string     { return `<text:span text:style-name="B">` + t + `</text:span>` }

func odtBody() string {
	return odp(ods("Appendix A: Risk Assessment Form")) +
		odp(ods("CONFIDENTIAL HEADER LINE")) +
		`<text:h text:outline-level="1">A. Information About the Contracting Party</text:h>` +
		odp(ods("Name / Company Name of the")) +
		odp(ods("Contracting Party:")) +
		odp(ods("Legal ")+ods("form:")) +
		odp(ods("Headquarters / Address:")+`<office:annotation><text:p>Legal form: a comment</text:p></office:annotation>`) +
		odp(ods("Commercial Registry Number:")) +
		odp(ods("Authorized Representatives:")+`<text:tab/>`) +
		odp(ods("Date of the evaluation:")) +
		odp(ods("Person Processing the Form:")) +
		`<text:h text:outline-level="1">B. Risk Assessment</text:h>` +
		odp("Legal form: must stay untouched") +
		odp(ods("Factors indicating a low risk:"))
}

// ODT builds an OpenDocument text template like the real form.
func ODT(t testing.TB) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte("application/vnd.oasis.opendocument.text"))
	w, _ = zw.Create("content.xml")
	_, _ = w.Write([]byte(odfHead + odtBody() + `</office:text></office:body></office:document-content>`))
	w, _ = zw.Create("styles.xml")
	_, _ = w.Write([]byte(`<office:document-styles xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"/>`))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// FODT builds a flat OpenDocument text template.
func FODT() []byte {
	h := strings.Replace(odfHead, "office:document-content", "office:document", 1)
	return []byte(h + odtBody() + `</office:text></office:body></office:document>`)
}

// RTF builds a Rich Text template like the real form.
func RTF() []byte {
	var b strings.Builder
	b.WriteString(`{\rtf1\ansi\ansicpg1252\deff0{\fonttbl{\f0\fswiss Arial;}}{\colortbl;\red0\green0\blue0;}` + "\n")
	b.WriteString(`{\*\generator numb-form test;}{\info{\title Legal form: not body}}` + "\n")
	b.WriteString(`\pard\plain\f0\fs22{\b Appendix A: Risk Assessment Form}\par` + "\n")
	b.WriteString(`{\b CONFIDENTIAL HEADER LINE}\par` + "\n")
	b.WriteString(`{\b\fs28 A.\tab Information About the Contracting Party}\par` + "\n")
	b.WriteString(`{\b Name / Company Name of the}\par{\b Contracting Party:}\par` + "\n")
	b.WriteString(`{\b Legal }{\b form:}\par` + "\n")
	b.WriteString(`{\b Headquarters / Address:}{\*\annotation Legal form: comment}\par` + "\n")
	b.WriteString(`{\b Commercial Registry Number:}\par` + "\n")
	b.WriteString(`{\b Authorized Representatives:\tab}\par` + "\n")
	b.WriteString(`{\b Date of the evaluation:}\par` + "\n")
	b.WriteString(`{\b Person Processing the Form:}\par` + "\n")
	b.WriteString(`{\b\fs28 B. Risk Assessment}\par` + "\n")
	b.WriteString(`Legal form: must stay untouched \'e4荤?\par` + "\n")
	b.WriteString(`Factors indicating a low risk:\par}` + "\n")
	return []byte(b.String())
}
