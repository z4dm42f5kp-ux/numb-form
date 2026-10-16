// Package docx fills the "A. Information About the Contracting Party"
// section of the risk assessment template.
//
// The document XML is never re-serialised. The decoder is used only to
// find byte offsets, and values are spliced into the original bytes, so
// everything outside the edited text runs stays byte-identical. Values
// are appended to the label's own text run (or a copy of its run
// properties when the label is followed by a tab), so no new formatting
// is introduced. Section B and everything after it is never touched.
package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

const (
	nsW   = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	nsXML = "http://www.w3.org/XML/1998/namespace"

	maxPart = 64 << 20
)

// Field identifies one line of section A.
type Field int

// Fields of section A, in template order.
const (
	FieldName Field = iota
	FieldLegalForm
	FieldAddress
	FieldRegistry
	FieldReps
	FieldDate
	FieldPerson
)

// AllFields lists the fields in template order.
var AllFields = []Field{FieldName, FieldLegalForm, FieldAddress, FieldRegistry, FieldReps, FieldDate, FieldPerson}

var labels = map[Field]string{
	FieldName:      "name/company name of the contracting party",
	FieldLegalForm: "legal form",
	FieldAddress:   "headquarters/address",
	FieldRegistry:  "commercial registry number",
	FieldReps:      "authorized representatives",
	FieldDate:      "date of the evaluation",
	FieldPerson:    "person processing the form",
}

// Title returns a human label for a field.
func (f Field) Title() string {
	switch f {
	case FieldName:
		return "Name / Company Name of the Contracting Party"
	case FieldLegalForm:
		return "Legal form"
	case FieldAddress:
		return "Headquarters / Address"
	case FieldRegistry:
		return "Commercial Registry Number"
	case FieldReps:
		return "Authorized Representatives"
	case FieldDate:
		return "Date of the evaluation"
	case FieldPerson:
		return "Person Processing the Form"
	}
	return "unknown"
}

// Values maps fields to the text to insert. Missing or empty values leave
// the field blank. FieldRegistry is always left blank.
type Values map[Field]string

type seg struct {
	tagStart, tagEnd int // whole <w:t> element
	cStart, cEnd     int // character content
	selfClose        bool
	preserve         bool
	text             string
}

type run struct {
	start, end   int
	rPr          []byte
	hasText      bool
	hasTab       bool
	tabAfterText bool
	segs         []int // indexes into para.segs
}

type para struct {
	start, end int
	raw        strings.Builder
	segs       []seg
	runs       []run
	prefix     string

	cur     *run // run being scanned
	curSeg  *seg // text element being scanned
	rPrFrom int
}

func (r run) tabOnly() bool { return r.hasTab && !r.hasText }

func (p *para) text() string { return p.raw.String() }

var wsRE = regexp.MustCompile(`\s+`)
var slashRE = regexp.MustCompile(`\s*/\s*`)
var enumRE = regexp.MustCompile(`^(?:[a-z]|[0-9]{1,2}|[ivx]{1,4})[.)]\s+`)

func norm(s string) string {
	s = strings.ReplaceAll(s, " ", " ")
	s = strings.ToLower(s)
	s = wsRE.ReplaceAllString(s, " ")
	s = slashRE.ReplaceAllString(s, "/")
	return strings.TrimSpace(s)
}

func collapse(s string) string {
	return strings.TrimSpace(wsRE.ReplaceAllString(strings.ReplaceAll(s, " ", " "), " "))
}

// scan walks document XML and returns its paragraphs in document order.
// Runs and text elements are tracked per paragraph, so paragraphs nested
// in text boxes do not disturb the enclosing paragraph.
func scan(data []byte) ([]*para, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var (
		paras  []*para
		pstack []*para
		names  []xml.Name
	)
	top := func() *para {
		if len(pstack) == 0 {
			return nil
		}
		return pstack[len(pstack)-1]
	}
	parent := func() xml.Name {
		if len(names) < 2 {
			return xml.Name{}
		}
		return names[len(names)-2]
	}
	for {
		off0 := int(dec.InputOffset())
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("document xml: %w", err)
		}
		off1 := int(dec.InputOffset())
		switch t := tok.(type) {
		case xml.StartElement:
			names = append(names, t.Name)
			if t.Name.Space != nsW {
				continue
			}
			p := top()
			if t.Name.Local == "p" {
				np := &para{start: off0, rPrFrom: -1}
				raw := data[off0:off1]
				if i := bytes.IndexByte(raw, ':'); i > 1 && i < bytes.IndexAny(raw, " />") {
					np.prefix = string(raw[1:i])
				}
				pstack = append(pstack, np)
				continue
			}
			if p == nil {
				continue
			}
			switch t.Name.Local {
			case "r":
				if p.cur == nil {
					p.cur = &run{start: off0}
				}
			case "rPr":
				if p.cur != nil && parent().Local == "r" {
					p.rPrFrom = off0
				}
			case "t":
				if p.cur != nil && parent().Local == "r" {
					p.curSeg = &seg{tagStart: off0, cStart: off1, selfClose: bytes.HasSuffix(data[off0:off1], []byte("/>"))}
					for _, a := range t.Attr {
						if a.Name.Space == nsXML && a.Name.Local == "space" && a.Value == "preserve" {
							p.curSeg.preserve = true
						}
					}
				}
			case "tab", "br", "cr":
				if p.cur != nil && parent().Local == "r" {
					p.raw.WriteString(" ")
					if t.Name.Local == "tab" {
						p.cur.hasTab = true
						if p.cur.hasText {
							p.cur.tabAfterText = true
						}
					}
				}
			}
		case xml.EndElement:
			if len(names) > 0 {
				names = names[:len(names)-1]
			}
			if t.Name.Space != nsW {
				continue
			}
			p := top()
			if p == nil {
				continue
			}
			switch t.Name.Local {
			case "p":
				p.end = off1
				p.cur, p.curSeg = nil, nil
				paras = append(paras, p)
				pstack = pstack[:len(pstack)-1]
			case "r":
				if p.cur != nil {
					p.cur.end = off1
					p.runs = append(p.runs, *p.cur)
				}
				p.cur = nil
			case "rPr":
				if p.cur != nil && p.rPrFrom >= 0 {
					p.cur.rPr = append([]byte(nil), data[p.rPrFrom:off1]...)
					p.rPrFrom = -1
				}
			case "t":
				if s := p.curSeg; s != nil && p.cur != nil {
					if s.selfClose {
						s.cEnd = s.cStart
					} else {
						s.cEnd = off0
					}
					s.tagEnd = off1
					p.segs = append(p.segs, *s)
					if s.text != "" {
						p.cur.hasText = true
						p.cur.tabAfterText = false
					}
					p.cur.segs = append(p.cur.segs, len(p.segs)-1)
					p.curSeg = nil
				}
			}
		case xml.CharData:
			if p := top(); p != nil && p.curSeg != nil {
				p.curSeg.text += string(t)
				p.raw.Write(t)
			}
		}
	}
	if len(pstack) != 0 {
		return nil, errors.New("document xml: unterminated paragraph")
	}
	sort.SliceStable(paras, func(i, j int) bool { return paras[i].start < paras[j].start })
	return paras, nil
}

// layout is the result of locating section A and its labels.
type layout struct {
	paras    []*para
	from, to int           // paragraph index range of section A, exclusive of headings
	targets  map[Field]int // paragraph holding each label
	missing  []Field
}

func locate(paras []*para) (*layout, error) {
	start, end := -1, -1
	for i, p := range paras {
		n := enumRE.ReplaceAllString(norm(p.text()), "")
		if start < 0 && strings.Contains(n, "information about the contracting party") {
			start = i
			continue
		}
		if start >= 0 && strings.Contains(n, "risk assess") {
			end = i
			break
		}
	}
	if start < 0 {
		return nil, errors.New(`heading "A. Information About the Contracting Party" not found`)
	}
	if end < 0 {
		return nil, errors.New(`heading "B. Risk Assessment" not found after section A`)
	}
	for i := start + 1; i < len(paras); i++ {
		if strings.Contains(norm(paras[i].text()), "information about the contracting party") {
			return nil, errors.New("section A heading appears more than once")
		}
	}
	lay := &layout{paras: paras, from: start + 1, to: end, targets: map[Field]int{}}
	for _, f := range AllFields {
		label := labels[f]
		var hits []int
		for i := lay.from; i < lay.to; i++ {
			n := norm(paras[i].text())
			if hasLabel(n, label) {
				hits = append(hits, i)
				continue
			}
			// "Name / Company Name of the" + "Contracting Party:" split
			// over two paragraphs.
			if f == FieldName && hasLabel(n, "contracting party") {
				for j := i - 1; j >= lay.from; j-- {
					pn := strings.TrimSuffix(norm(paras[j].text()), ":")
					if pn == "" {
						continue
					}
					if pn == "name/company name of the" || pn == "name/company name of" {
						hits = append(hits, i)
					}
					break
				}
			}
		}
		switch len(hits) {
		case 0:
			lay.missing = append(lay.missing, f)
		case 1:
			lay.targets[f] = hits[0]
		default:
			return nil, fmt.Errorf("label %q appears %d times in section A", f.Title(), len(hits))
		}
	}
	return lay, nil
}

// hasLabel reports whether paragraph text n starts with label followed by
// nothing but an optional colon and whitespace or further text.
func hasLabel(n, label string) bool {
	if !strings.HasPrefix(n, label) {
		return false
	}
	rest := n[len(label):]
	return rest == "" || strings.HasPrefix(rest, ":") || strings.HasPrefix(rest, " ")
}

// labelRest returns what follows the label (and colon) in paragraph text.
func labelRest(text, label string, f Field) string {
	n := norm(text)
	l := label
	if f == FieldName && !strings.HasPrefix(n, label) {
		l = "contracting party"
	}
	if !strings.HasPrefix(n, l) {
		return ""
	}
	// Map the normalised prefix length back onto the original text by
	// skipping the same number of label characters.
	c := collapse(text)
	idx := strings.Index(strings.ToLower(c), ":")
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(c[idx+1:])
}

type edit struct {
	at     int
	insert string
}

func xmlEscape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(textutil.XMLSafe(s)))
	return b.String()
}

// Fill returns a copy of the template with section A filled in.
func Fill(template []byte, vals Values) ([]byte, []string, error) {
	return rewrite(template, func(doc []byte) ([]byte, []string, error) { return fillXML(doc, vals) })
}

func fillXML(doc []byte, vals Values) ([]byte, []string, error) {
	paras, err := scan(doc)
	if err != nil {
		return nil, nil, err
	}
	lay, err := locate(paras)
	if err != nil {
		return nil, nil, err
	}
	var warnings []string
	for _, f := range lay.missing {
		if f == FieldRegistry {
			warnings = append(warnings, "label \"Commercial Registry Number\" not found (it is left blank anyway)")
			continue
		}
		if strings.TrimSpace(vals[f]) != "" {
			return nil, nil, fmt.Errorf("label %q not found in section A", f.Title())
		}
		warnings = append(warnings, fmt.Sprintf("label %q not found in section A", f.Title()))
	}
	var edits []edit
	expect := map[int]string{}
	for _, f := range AllFields {
		idx, ok := lay.targets[f]
		if !ok {
			continue
		}
		p := paras[idx]
		if rest := labelRest(p.text(), labels[f], f); rest != "" {
			return nil, nil, fmt.Errorf("field %q is not empty in the template (%q)", f.Title(), textutil.OneLine(rest))
		}
		v := textutil.OneLine(vals[f])
		if f == FieldRegistry || v == "" {
			continue
		}
		e, sep, err := insertion(p, v)
		if err != nil {
			return nil, nil, fmt.Errorf("field %q: %w", f.Title(), err)
		}
		edits = append(edits, e)
		expect[idx] = collapse(p.text() + sep + v)
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].at > edits[j].at })
	out := append([]byte(nil), doc...)
	for _, e := range edits {
		out = append(out[:e.at], append([]byte(e.insert), out[e.at:]...)...)
	}
	if err := verify(doc, out, expect); err != nil {
		return nil, nil, fmt.Errorf("self-check failed, nothing written: %w", err)
	}
	return out, warnings, nil
}

func insertion(p *para, v string) (edit, string, error) {
	last := -1
	for i, r := range p.runs {
		if r.hasText {
			last = i
		}
	}
	if last < 0 {
		return edit{}, "", errors.New("label has no text run")
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
		return edit{at: p.runs[after].end, insert: ins}, " ", nil
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
		return edit{}, "", errors.New("label text element not found")
	}
	sep := " "
	if strings.HasSuffix(s.text, " ") || strings.HasSuffix(s.text, " ") {
		sep = ""
	}
	return edit{at: s.cEnd, insert: xmlEscape(sep + v)}, sep, nil
}

func verify(before, after []byte, expect map[int]string) error {
	bp, err := scan(before)
	if err != nil {
		return err
	}
	ap, err := scan(after)
	if err != nil {
		return fmt.Errorf("output is not well-formed: %w", err)
	}
	if len(bp) != len(ap) {
		return fmt.Errorf("paragraph count changed (%d -> %d)", len(bp), len(ap))
	}
	for i := range bp {
		want := bp[i].text()
		if e, ok := expect[i]; ok {
			if collapse(ap[i].text()) != e {
				return fmt.Errorf("paragraph %d reads %q, want %q", i, ap[i].text(), e)
			}
			continue
		}
		if ap[i].text() != want {
			return fmt.Errorf("paragraph %d outside the edited fields changed", i)
		}
	}
	// Bytes before the first and after the last edited paragraph must be
	// identical, which proves section B and everything else is untouched.
	if len(expect) > 0 {
		first, lastIdx := len(bp), -1
		for i := range expect {
			first = min(first, i)
			lastIdx = max(lastIdx, i)
		}
		if !bytes.Equal(before[:bp[first].start], after[:ap[first].start]) {
			return errors.New("bytes before section A changed")
		}
		if !bytes.Equal(before[bp[lastIdx].end:], after[ap[lastIdx].end:]) {
			return errors.New("bytes after the last edited field changed")
		}
	}
	return nil
}

// ReadBack extracts the current values of section A from a .docx file.
func ReadBack(docx []byte) (Values, error) {
	zr, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		return nil, err
	}
	main, err := mainPart(zr)
	if err != nil {
		return nil, err
	}
	doc, err := readEntry(zr, main)
	if err != nil {
		return nil, err
	}
	paras, err := scan(doc)
	if err != nil {
		return nil, err
	}
	lay, err := locate(paras)
	if err != nil {
		return nil, err
	}
	out := Values{}
	for f, idx := range lay.targets {
		out[f] = labelRest(paras[idx].text(), labels[f], f)
	}
	return out, nil
}

// Check validates that a template has a fillable section A.
func Check(template []byte) ([]string, error) {
	_, w, err := Fill(template, Values{})
	return w, err
}

func readEntry(zr *zip.Reader, name string) ([]byte, error) {
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		b, err := io.ReadAll(io.LimitReader(rc, maxPart+1))
		if err != nil {
			return nil, err
		}
		if len(b) > maxPart {
			return nil, fmt.Errorf("%s is too large", name)
		}
		return b, nil
	}
	return nil, fmt.Errorf("%s not found", name)
}

func mainPart(zr *zip.Reader) (string, error) {
	rels, err := readEntry(zr, "_rels/.rels")
	if err != nil {
		return "", errors.New("not a .docx file (_rels/.rels missing)")
	}
	var r struct {
		Rel []struct {
			Type   string `xml:"Type,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if err := xml.Unmarshal(rels, &r); err != nil {
		return "", fmt.Errorf("_rels/.rels: %w", err)
	}
	for _, x := range r.Rel {
		if strings.HasSuffix(x.Type, "/officeDocument") {
			return strings.TrimPrefix(path.Clean("/"+x.Target), "/"), nil
		}
	}
	return "", errors.New("main document part not found")
}

// rewrite copies every zip entry byte-for-byte except the main document
// part, which is replaced by fn's output. It then re-reads the result and
// checks that all other entries are identical.
func rewrite(src []byte, fn func([]byte) ([]byte, []string, error)) ([]byte, []string, error) {
	zr, err := zip.NewReader(bytes.NewReader(src), int64(len(src)))
	if err != nil {
		return nil, nil, fmt.Errorf("template is not a valid .docx (zip): %w", err)
	}
	main, err := mainPart(zr)
	if err != nil {
		return nil, nil, err
	}
	ct, err := readEntry(zr, "[Content_Types].xml")
	if err != nil {
		return nil, nil, err
	}
	if bytes.Contains(ct, []byte("macroEnabled")) {
		return nil, nil, errors.New("template contains macros; refusing to use it")
	}
	doc, err := readEntry(zr, main)
	if err != nil {
		return nil, nil, err
	}
	newDoc, warnings, err := fn(doc)
	if err != nil {
		return nil, nil, err
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range zr.File {
		if f.Name == main {
			w, err := zw.CreateHeader(&zip.FileHeader{Name: f.Name, Method: zip.Deflate, Modified: f.Modified})
			if err != nil {
				return nil, nil, err
			}
			if _, err := w.Write(newDoc); err != nil {
				return nil, nil, err
			}
			continue
		}
		if err := zw.Copy(f); err != nil {
			return nil, nil, fmt.Errorf("copy %s: %w", f.Name, err)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, nil, err
	}
	out := buf.Bytes()
	zo, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	if err != nil {
		return nil, nil, fmt.Errorf("self-check: %w", err)
	}
	if len(zo.File) != len(zr.File) {
		return nil, nil, errors.New("self-check: entry count changed")
	}
	for i, f := range zr.File {
		g := zo.File[i]
		if g.Name != f.Name {
			return nil, nil, errors.New("self-check: entry order changed")
		}
		if f.Name == main {
			continue
		}
		if g.CRC32 != f.CRC32 || g.UncompressedSize64 != f.UncompressedSize64 {
			return nil, nil, fmt.Errorf("self-check: %s changed", f.Name)
		}
	}
	got, err := readEntry(zo, main)
	if err != nil || !bytes.Equal(got, newDoc) {
		return nil, nil, errors.New("self-check: document part did not round-trip")
	}
	return out, warnings, nil
}
