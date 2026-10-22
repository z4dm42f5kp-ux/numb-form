// Package reportdate finds the date a risk assessment form was made:
// the "Date of the evaluation" in section A, read with the form's own
// reader, or, for forms laid out differently, the date next to a date
// label in the document's text.
package reportdate

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/formats"
)

// Layout is the written form of a date: DD-MM-YYYY.
const Layout = "DD-MM-YYYY"

// Format writes day, month and year as DD-MM-YYYY.
func Format(d, m, y int) string { return fmt.Sprintf("%02d-%02d-%04d", d, m, y) }

var months = map[string]int{
	"january": 1, "february": 2, "march": 3, "april": 4, "may": 5, "june": 6, "july": 7, "august": 8,
	"september": 9, "october": 10, "november": 11, "december": 12,
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "jun": 6, "jul": 7, "aug": 8, "sep": 9, "sept": 9, "oct": 10, "nov": 11, "dec": 12,
	"januar": 1, "februar": 2, "märz": 3, "maerz": 3, "mai": 5, "juni": 6, "juli": 7, "oktober": 10, "dezember": 12,
	"enero": 1, "febrero": 2, "marzo": 3, "abril": 4, "mayo": 5, "junio": 6, "julio": 7, "agosto": 8,
	"septiembre": 9, "setiembre": 9, "octubre": 10, "noviembre": 11, "diciembre": 12,
}

var (
	reNumeric = regexp.MustCompile(`\b(\d{1,2})\s*[./-]\s*(\d{1,2})\s*[./-]\s*(\d{4}|\d{2})\b`)
	reISO     = regexp.MustCompile(`\b(\d{4})-(\d{1,2})-(\d{1,2})\b`)
	reDayName = regexp.MustCompile(`(?i)\b(\d{1,2})\.?\s+(?:de\s+)?([a-zäéíóú]{3,10})\.?,?\s+(?:de\s+)?(\d{4})\b`)
	reNameDay = regexp.MustCompile(`(?i)\b([a-zäéíóú]{3,10})\.?\s+(\d{1,2}),?\s+(\d{4})\b`)
)

func valid(d, m, y int) bool {
	if y < 1990 || y > 2100 || m < 1 || m > 12 || d < 1 {
		return false
	}
	days := [13]int{0, 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if m == 2 && !(y%4 == 0 && (y%100 != 0 || y%400 == 0)) {
		return d <= 28
	}
	return d <= days[m]
}

func atoi(s string) int { n, _ := strconv.Atoi(s); return n }

// Parse returns the first date in s as DD-MM-YYYY. Numeric dates are read
// day first (16/10/2026, 16.10.2026); when the first number cannot be a
// month day pair that way (10/16/2026) they are read month first.
func Parse(s string) (string, bool) {
	type cand struct {
		at   int
		date string
	}
	var best *cand
	keep := func(at, d, m, y int) {
		if valid(d, m, y) && (best == nil || at < best.at) {
			best = &cand{at, Format(d, m, y)}
		}
	}
	for _, m := range reISO.FindAllStringSubmatchIndex(s, -1) {
		keep(m[0], atoi(s[m[6]:m[7]]), atoi(s[m[4]:m[5]]), atoi(s[m[2]:m[3]]))
	}
	for _, m := range reNumeric.FindAllStringSubmatchIndex(s, -1) {
		a, b, y := atoi(s[m[2]:m[3]]), atoi(s[m[4]:m[5]]), atoi(s[m[6]:m[7]])
		if y < 100 {
			y += 2000
		}
		if b > 12 && a <= 12 {
			a, b = b, a
		}
		keep(m[0], a, b, y)
	}
	for _, m := range reDayName.FindAllStringSubmatchIndex(s, -1) {
		if mo, ok := months[strings.ToLower(s[m[4]:m[5]])]; ok {
			keep(m[0], atoi(s[m[2]:m[3]]), mo, atoi(s[m[6]:m[7]]))
		}
	}
	for _, m := range reNameDay.FindAllStringSubmatchIndex(s, -1) {
		if mo, ok := months[strings.ToLower(s[m[2]:m[3]])]; ok {
			keep(m[0], atoi(s[m[4]:m[5]]), mo, atoi(s[m[6]:m[7]]))
		}
	}
	if best == nil {
		return "", false
	}
	return best.date, true
}

// labels are searched in the text in this order; the date must follow
// within a short distance.
var labels = []*regexp.Regexp{
	regexp.MustCompile(`(?i)date\s+of\s+(the\s+)?evaluation`),
	regexp.MustCompile(`(?i)evaluation\s+date`),
	regexp.MustCompile(`(?i)fecha\s+de\s+(la\s+)?evaluaci[oó]n`),
	regexp.MustCompile(`(?i)datum\s+der\s+(bewertung|prüfung|evaluierung)`),
	regexp.MustCompile(`(?i)report\s+(created|date)`),
	regexp.MustCompile(`(?i)\b(date|datum|fecha)\b`),
}

// FromText finds the date written after a date label.
func FromText(text string) (string, bool) {
	for _, l := range labels {
		for _, m := range l.FindAllStringIndex(text, -1) {
			after := text[m[1]:]
			if len(after) > 60 {
				after = after[:60]
			}
			if nl := strings.Index(after, "\n"); nl >= 0 && strings.TrimSpace(after[:nl]) != "" {
				after = after[:nl] // the date is on the label's line
			}
			if d, ok := Parse(after); ok {
				return d, true
			}
		}
	}
	return "", false
}

var (
	reParaEnd = regexp.MustCompile(`</(w:p|text:p|text:h)>|<w:br/>|<text:line-break/>`)
	reTab     = regexp.MustCompile(`<w:tab/>|<text:tab/>|<text:s/>`)
	reTag     = regexp.MustCompile(`<[^>]*>`)
)

// Text returns the plain text of a .docx, .odt or .fodt document, one
// paragraph per line.
func Text(b []byte, ext string) (string, error) {
	var x []byte
	switch strings.ToLower(ext) {
	case ".docx", ".docm":
		var err error
		if x, err = zipPart(b, "word/document.xml"); err != nil {
			return "", err
		}
	case ".odt":
		var err error
		if x, err = zipPart(b, "content.xml"); err != nil {
			return "", err
		}
	case ".fodt":
		x = b
	default:
		return "", fmt.Errorf("%s text cannot be read", ext)
	}
	s := reParaEnd.ReplaceAllString(string(x), "\n")
	s = reTab.ReplaceAllString(s, " ")
	s = reTag.ReplaceAllString(s, "")
	return html.UnescapeString(s), nil
}

func zipPart(b []byte, name string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		out, err := io.ReadAll(io.LimitReader(rc, 64<<20+1))
		if err != nil {
			return nil, err
		}
		if len(out) > 64<<20 {
			return nil, errors.New(name + " is too large")
		}
		return out, nil
	}
	return nil, errors.New(name + " missing")
}

// FromForm returns the report date of a form file's bytes, and how it was
// found.
func FromForm(b []byte, name string) (date, how string, ok bool) {
	ext := strings.ToLower(filepath.Ext(name))
	if v, err := formats.ReadFilled(b, ext); err == nil {
		if d, ok := Parse(v[form.FieldDate]); ok {
			return d, "Date of the evaluation", true
		}
	}
	if t, err := Text(b, ext); err == nil {
		if d, ok := FromText(t); ok {
			return d, "date in the document text", true
		}
	}
	return "", "", false
}
