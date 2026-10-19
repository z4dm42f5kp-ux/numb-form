package form

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

var (
	wsRE    = regexp.MustCompile(`\s+`)
	slashRE = regexp.MustCompile(`\s*/\s*`)
	enumRE  = regexp.MustCompile(`^(?:[a-z]|[0-9]{1,2}|[ivx]{1,4})[.)]\s+`)
)

// Norm lower-cases s, collapses white space and removes spaces around
// slashes, so "Headquarters / Address:" becomes "headquarters/address:".
func Norm(s string) string {
	s = strings.ReplaceAll(s, " ", " ")
	s = strings.ToLower(s)
	s = wsRE.ReplaceAllString(s, " ")
	s = slashRE.ReplaceAllString(s, "/")
	return strings.TrimSpace(s)
}

// Collapse collapses white space runs to single spaces.
func Collapse(s string) string {
	return strings.TrimSpace(wsRE.ReplaceAllString(strings.ReplaceAll(s, " ", " "), " "))
}

// Layout tells where section A and its labels are.
type Layout struct {
	From, To int           // paragraph range of section A, headings excluded
	Targets  map[Field]int // paragraph holding each label
	Missing  []Field
}

// Locate finds section A (from the "Information About the Contracting
// Party" heading to the "Risk Assessment" heading) and each label in it.
func Locate(texts []string) (*Layout, error) {
	start, end := -1, -1
	for i, t := range texts {
		n := enumRE.ReplaceAllString(Norm(t), "")
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
	for i := start + 1; i < len(texts); i++ {
		if strings.Contains(Norm(texts[i]), "information about the contracting party") {
			return nil, errors.New("section A heading appears more than once")
		}
	}
	lay := &Layout{From: start + 1, To: end, Targets: map[Field]int{}}
	for _, f := range AllFields {
		var hits []int
		for i := lay.From; i < lay.To; i++ {
			if isLabel(texts, i, lay.From, f) {
				hits = append(hits, i)
			}
		}
		switch len(hits) {
		case 0:
			lay.Missing = append(lay.Missing, f)
		case 1:
			lay.Targets[f] = hits[0]
		default:
			return nil, fmt.Errorf("label %q appears %d times in section A", f.Title(), len(hits))
		}
	}
	return lay, nil
}

func isLabel(texts []string, i, from int, f Field) bool {
	n := Norm(texts[i])
	if hasLabel(n, labels[f]) {
		return true
	}
	if f != FieldName || !hasLabel(n, "contracting party") {
		return false
	}
	// "Name / Company Name of the" + "Contracting Party:" split over two
	// paragraphs.
	for j := i - 1; j >= from; j-- {
		pn := strings.TrimSuffix(Norm(texts[j]), ":")
		if pn == "" {
			continue
		}
		return pn == "name/company name of the" || pn == "name/company name of"
	}
	return false
}

func hasLabel(n, label string) bool {
	if !strings.HasPrefix(n, label) {
		return false
	}
	rest := n[len(label):]
	return rest == "" || strings.HasPrefix(rest, ":") || strings.HasPrefix(rest, " ")
}

// LabelRest returns what follows the label's colon in paragraph text.
func LabelRest(text string) string {
	c := Collapse(text)
	idx := strings.Index(c, ":")
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(c[idx+1:])
}

// Job is one field to write.
type Job struct {
	Field Field
	Index int    // paragraph index
	Value string // one line, already cleaned
}

// Plan checks the template's section A against vals and returns the
// fields to write. Labels must be present for every non-empty value and
// every field must still be blank in the template.
func Plan(texts []string, vals Values) ([]Job, []string, error) {
	lay, err := Locate(texts)
	if err != nil {
		return nil, nil, err
	}
	var warnings []string
	for _, f := range lay.Missing {
		if f == FieldRegistry {
			warnings = append(warnings, `label "Commercial Registry Number" not found (it is left blank anyway)`)
			continue
		}
		if strings.TrimSpace(vals[f]) != "" {
			return nil, nil, fmt.Errorf("label %q not found in section A", f.Title())
		}
		warnings = append(warnings, fmt.Sprintf("label %q not found in section A", f.Title()))
	}
	var jobs []Job
	for _, f := range AllFields {
		idx, ok := lay.Targets[f]
		if !ok {
			continue
		}
		if rest := LabelRest(texts[idx]); rest != "" {
			return nil, nil, fmt.Errorf("field %q is not empty in the template (%q)", f.Title(), textutil.OneLine(rest))
		}
		v := textutil.OneLine(vals[f])
		if f == FieldRegistry || v == "" {
			continue
		}
		jobs = append(jobs, Job{Field: f, Index: idx, Value: v})
	}
	return jobs, warnings, nil
}

// Read returns the values of section A from paragraph texts.
func Read(texts []string) (Values, error) {
	lay, err := Locate(texts)
	if err != nil {
		return nil, err
	}
	out := Values{}
	for f, idx := range lay.Targets {
		out[f] = LabelRest(texts[idx])
	}
	return out, nil
}
