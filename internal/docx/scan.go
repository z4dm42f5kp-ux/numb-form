package docx

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
)

const (
	nsW   = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	nsXML = "http://www.w3.org/XML/1998/namespace"
)

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

func paras(ps []*para) []form.Para {
	out := make([]form.Para, len(ps))
	for i, p := range ps {
		out[i] = form.Para{Start: p.start, End: p.end, Text: p.text()}
	}
	return out
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
		off0 := int(dec.InputOffset()) // #nosec G115 -- bounded by len(data).
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("document xml: %w", err)
		}
		off1 := int(dec.InputOffset()) // #nosec G115 -- bounded by len(data).
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
