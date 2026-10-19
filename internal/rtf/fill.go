package rtf

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// Format implements form.Filler for .rtf files.
type Format struct{}

var _ form.Filler = Format{}

// encode writes s as RTF text: ASCII as is, specials escaped, everything
// else as \uN with a "?" fallback.
func encode(s string) string {
	var b strings.Builder
	for _, r := range textutil.XMLSafe(s) {
		switch {
		case r == '\\' || r == '{' || r == '}':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\t' || r == '\n' || r == '\r':
			b.WriteByte(' ')
		case r >= 0x20 && r < 0x7f:
			b.WriteRune(r)
		default:
			for _, u := range utf16.Encode([]rune{r}) {
				b.WriteString(`\u` + strconv.Itoa(int(int16(u))) + `?`)
			}
		}
	}
	return b.String()
}

// Fill returns a copy of the template with section A filled in.
func (Format) Fill(template []byte, vals form.Values) ([]byte, []string, error) {
	ps, err := scan(template)
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
		p := ps[j.Index]
		if p.lastEnd < 0 {
			return nil, nil, fmt.Errorf("field %q: label has no text", j.Field.Title())
		}
		at := p.lastEnd
		if p.tabEnd >= 0 {
			at = p.tabEnd
		}
		sep := " "
		if strings.HasSuffix(before[j.Index].Text, " ") && p.tabEnd < 0 {
			sep = ""
		}
		// "\uc1" plus its delimiting space, then the text.
		edits = append(edits, form.Edit{At: at, Insert: `{\uc1 ` + encode(sep+j.Value) + `}`})
		expect[j.Index] = form.Collapse(before[j.Index].Text + sep + j.Value)
	}
	out := form.Apply(template, edits)
	aps, err := scan(out)
	if err != nil {
		return nil, nil, fmt.Errorf("self-check failed: %w", err)
	}
	if err := form.Verify(template, out, before, toForm(aps), expect); err != nil {
		return nil, nil, fmt.Errorf("self-check failed, nothing written: %w", err)
	}
	return out, warnings, nil
}

// ReadBack returns the values currently in section A.
func (Format) ReadBack(doc []byte) (form.Values, error) {
	ps, err := scan(doc)
	if err != nil {
		return nil, err
	}
	if len(ps) == 0 {
		return nil, errors.New("rtf: no paragraphs")
	}
	return form.Read(form.Texts(toForm(ps)))
}
