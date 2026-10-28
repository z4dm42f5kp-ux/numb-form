// Package odt fills section A of OpenDocument text templates (.odt and
// flat .fodt), as written by LibreOffice, Apache OpenOffice and others.
//
// As with .docx, the XML is only scanned for byte offsets and values are
// spliced into the original bytes. A value joins the label's own text (so
// it inherits the label's span style); when the label is followed by a
// tab, the value is wrapped in the label's existing span style.
//
// In plain words: as for Word files, values are written after their
// labels in LibreOffice and OpenOffice files, and nothing else changes.
package odt

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
)

const (
	nsText   = "urn:oasis:names:tc:opendocument:xmlns:text:1.0"
	nsOffice = "urn:oasis:names:tc:opendocument:xmlns:office:1.0"
)

type seg struct {
	end   int    // byte offset just after the character data
	text  string // decoded text
	style string // style of the innermost enclosing span, if any
}

type para struct {
	start, end int
	prefix     string // namespace prefix of the text namespace, e.g. "text"
	raw        strings.Builder
	segs       []seg
	tabAt      int    // offset after the last tab that follows the last text
	tabStyle   string // span style in effect at that tab
	spans      []string
	skip       int // depth inside annotations and notes
}

func (p *para) style() string {
	for i := len(p.spans) - 1; i >= 0; i-- {
		if p.spans[i] != "" {
			return p.spans[i]
		}
	}
	return ""
}

func attr(se xml.StartElement, space, local string) string {
	for _, a := range se.Attr {
		if a.Name.Space == space && a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

// scan returns the paragraphs (text:p and text:h) in document order.
func scan(data []byte) ([]*para, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var paras, stack []*para
	top := func() *para {
		if len(stack) == 0 {
			return nil
		}
		return stack[len(stack)-1]
	}
	for {
		off0 := int(dec.InputOffset()) // #nosec G115 -- bounded by len(data).
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("document xml: %w", err)
		}
		off1 := int(dec.InputOffset()) // #nosec G115 -- bounded by len(data).
		p := top()
		switch t := tok.(type) {
		case xml.StartElement:
			if p != nil && p.skip > 0 {
				p.skip++
				continue
			}
			if t.Name.Space == nsText && (t.Name.Local == "p" || t.Name.Local == "h") {
				np := &para{start: off0, tabAt: -1}
				raw := data[off0:off1]
				if i := bytes.IndexByte(raw, ':'); i > 1 && i < bytes.IndexAny(raw, " />") {
					np.prefix = string(raw[1:i])
				}
				stack = append(stack, np)
				continue
			}
			if p == nil {
				continue
			}
			switch {
			case t.Name.Space == nsOffice && t.Name.Local == "annotation",
				t.Name.Space == nsText && (t.Name.Local == "note" || t.Name.Local == "tracked-changes"):
				p.skip = 1
			case t.Name.Space == nsText && (t.Name.Local == "span" || t.Name.Local == "a"):
				p.spans = append(p.spans, attr(t, nsText, "style-name"))
			case t.Name.Space == nsText && t.Name.Local == "s":
				n, _ := strconv.Atoi(attr(t, nsText, "c"))
				p.raw.WriteString(strings.Repeat(" ", max(1, min(n, 100))))
			case t.Name.Space == nsText && (t.Name.Local == "tab" || t.Name.Local == "line-break"):
				p.raw.WriteString(" ")
				if t.Name.Local == "tab" && len(p.segs) > 0 {
					// offset after the (self-closing) tab element
					p.tabAt, p.tabStyle = off1, p.style()
				}
			}
		case xml.EndElement:
			if p == nil {
				continue
			}
			if p.skip > 0 {
				p.skip--
				continue
			}
			switch {
			case t.Name.Space == nsText && (t.Name.Local == "p" || t.Name.Local == "h"):
				p.end = off1
				paras = append(paras, p)
				stack = stack[:len(stack)-1]
			case t.Name.Space == nsText && (t.Name.Local == "span" || t.Name.Local == "a"):
				if len(p.spans) > 0 {
					p.spans = p.spans[:len(p.spans)-1]
				}
			case t.Name.Space == nsText && t.Name.Local == "tab":
				if p.tabAt >= 0 && p.tabAt < off1 && len(p.segs) > 0 {
					p.tabAt = off1
				}
			}
		case xml.CharData:
			if p == nil || p.skip > 0 {
				continue
			}
			s := string(t)
			p.raw.WriteString(s)
			if strings.TrimSpace(s) != "" {
				p.segs = append(p.segs, seg{end: off1, text: s, style: p.style()})
				p.tabAt = -1 // text after a tab: the tab is not trailing
			}
		}
	}
	if len(stack) != 0 {
		return nil, errors.New("document xml: unterminated paragraph")
	}
	sort.SliceStable(paras, func(i, j int) bool { return paras[i].start < paras[j].start })
	return paras, nil
}

func toForm(ps []*para) []form.Para {
	out := make([]form.Para, len(ps))
	for i, p := range ps {
		out[i] = form.Para{Start: p.start, End: p.end, Text: p.raw.String()}
	}
	return out
}
