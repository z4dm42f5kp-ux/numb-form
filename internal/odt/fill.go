package odt

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// Format implements form.Filler for .odt (Flat false) and .fodt (Flat true).
type Format struct{ Flat bool }

var _ form.Filler = Format{}

func escape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(textutil.XMLSafe(s)))
	return b.String()
}

func content(doc []byte, flat bool) ([]byte, error) {
	if flat {
		return doc, nil
	}
	zr, err := zip.NewReader(bytes.NewReader(doc), int64(len(doc)))
	if err != nil {
		return nil, fmt.Errorf("template is not a valid .odt (zip): %w", err)
	}
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "Basic/") || strings.HasPrefix(f.Name, "Scripts/") {
			return nil, errors.New("template contains macros; refusing to use it")
		}
	}
	return form.ZipEntry(zr, "content.xml")
}

// Fill returns a copy of the template with section A filled in.
func (f Format) Fill(template []byte, vals form.Values) ([]byte, []string, error) {
	xmlDoc, err := content(template, f.Flat)
	if err != nil {
		return nil, nil, err
	}
	filled, warnings, err := fillXML(xmlDoc, vals)
	if err != nil {
		return nil, nil, err
	}
	if f.Flat {
		return filled, warnings, nil
	}
	out, err := form.ReplaceZipEntry(template, "content.xml", filled)
	return out, warnings, err
}

// ReadBack returns the values currently in section A.
func (f Format) ReadBack(doc []byte) (form.Values, error) {
	xmlDoc, err := content(doc, f.Flat)
	if err != nil {
		return nil, err
	}
	ps, err := scan(xmlDoc)
	if err != nil {
		return nil, err
	}
	return form.Read(form.Texts(toForm(ps)))
}

func fillXML(doc []byte, vals form.Values) ([]byte, []string, error) {
	ps, err := scan(doc)
	if err != nil {
		return nil, nil, err
	}
	before := toForm(ps)
	jobs, warnings, err := form.Plan(form.Texts(before), vals)
	if err != nil {
		return nil, nil, err
	}
	var edits []form.Edit
	expect := map[int]string{}
	for _, j := range jobs {
		e, sep, err := insertion(ps[j.Index], j.Value)
		if err != nil {
			return nil, nil, fmt.Errorf("field %q: %w", j.Field.Title(), err)
		}
		edits = append(edits, e)
		expect[j.Index] = form.Collapse(before[j.Index].Text + sep + j.Value)
	}
	out := form.Apply(doc, edits)
	aps, err := scan(out)
	if err != nil {
		return nil, nil, fmt.Errorf("self-check failed, output is not well-formed: %w", err)
	}
	if err := form.Verify(doc, out, before, toForm(aps), expect); err != nil {
		return nil, nil, fmt.Errorf("self-check failed, nothing written: %w", err)
	}
	return out, warnings, nil
}

func insertion(p *para, v string) (form.Edit, string, error) {
	if len(p.segs) == 0 {
		return form.Edit{}, "", errors.New("label has no text")
	}
	if p.tabAt >= 0 {
		text := escape(v)
		if p.tabStyle != "" {
			pre := p.prefix
			if pre != "" {
				pre += ":"
			}
			text = "<" + pre + `span ` + pre + `style-name="` + escape(p.tabStyle) + `">` + text + "</" + pre + "span>"
		}
		return form.Edit{At: p.tabAt, Insert: text}, " ", nil
	}
	last := p.segs[len(p.segs)-1]
	sep := " "
	if strings.HasSuffix(last.text, " ") {
		sep = ""
	}
	return form.Edit{At: last.end, Insert: escape(sep + v)}, sep, nil
}
