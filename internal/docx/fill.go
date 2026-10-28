// Package docx fills section A of a Word (.docx) template.
//
// The document XML is never re-serialised. The decoder is used only to
// find byte offsets, and values are spliced into the original bytes, so
// everything outside the edited text runs stays byte-identical. Values
// are appended to the label's own text run (or a copy of its run
// properties when the label is followed by a tab), so no new formatting
// is introduced. Section B and everything after it is never touched.
//
// In plain words: each value is written right after its label in the
// Word file, and every other byte of the file stays exactly as it was,
// so the layout cannot change.
package docx

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// Format implements form.Filler for .docx files.
type Format struct{}

var _ form.Filler = Format{}

// Fill returns a copy of the template with section A filled in.
func (Format) Fill(template []byte, vals form.Values) ([]byte, []string, error) {
	return rewrite(template, func(doc []byte) ([]byte, []string, error) { return fillXML(doc, vals) })
}

// ReadBack extracts the current values of section A.
func (Format) ReadBack(docx []byte) (form.Values, error) {
	doc, err := mainXML(docx)
	if err != nil {
		return nil, err
	}
	ps, err := scan(doc)
	if err != nil {
		return nil, err
	}
	return form.Read(form.Texts(paras(ps)))
}

func xmlEscape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(textutil.XMLSafe(s)))
	return b.String()
}

func fillXML(doc []byte, vals form.Values) ([]byte, []string, error) {
	ps, err := scan(doc)
	if err != nil {
		return nil, nil, err
	}
	before := paras(ps)
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
	if err := form.Verify(doc, out, before, paras(aps), expect); err != nil {
		return nil, nil, fmt.Errorf("self-check failed, nothing written: %w", err)
	}
	return out, warnings, nil
}

func insertion(p *para, v string) (form.Edit, string, error) {
	last := -1
	for i, r := range p.runs {
		if r.hasText {
			last = i
		}
	}
	if last < 0 {
		return form.Edit{}, "", errors.New("label has no text run")
	}
	lr := p.runs[last]
	// Label followed by a tab (in its own run or in later tab-only runs):
	// put the value after the tab, in a run with the label's properties.
	after := -1
	if lr.tabAfterText {
		after = last
	}
	for i := last + 1; i < len(p.runs); i++ {
		if p.runs[i].tabOnly() {
			after = i
		}
	}
	if after >= 0 {
		pre := p.prefix
		if pre != "" {
			pre += ":"
		}
		ins := "<" + pre + "r>" + string(lr.rPr) + "<" + pre + `t xml:space="preserve">` + xmlEscape(v) + "</" + pre + "t></" + pre + "r>"
		return form.Edit{At: p.runs[after].end, Insert: ins}, " ", nil
	}
	var s seg
	found := false
	for i := len(lr.segs) - 1; i >= 0; i-- {
		if c := p.segs[lr.segs[i]]; c.text != "" {
			s, found = c, true
			break
		}
	}
	if !found || s.selfClose {
		return form.Edit{}, "", errors.New("label text element not found")
	}
	sep := " "
	if strings.HasSuffix(s.text, " ") || strings.HasSuffix(s.text, " ") {
		sep = ""
	}
	return form.Edit{At: s.cEnd, Insert: xmlEscape(sep + v)}, sep, nil
}
