package sheet

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"unicode/utf8"
)

// cp1252 maps the bytes 0x80-0x9F of Windows-1252, which Excel uses for
// "CSV (comma delimited)" exports on Western European systems. The other
// high bytes equal their Latin-1 code points.
var cp1252 = [32]rune{
	'€', 0x81, '‚', 'ƒ', '„', '…', '†', '‡', 'ˆ', '‰', 'Š', '‹', 'Œ', 0x8d, 'Ž', 0x8f,
	0x90, '‘', '’', '“', '”', '•', '–', '—', '˜', '™', 'š', '›', 'œ', 0x9d, 'ž', 'Ÿ',
}

func decodeText(b []byte) string {
	b = bytes.TrimPrefix(b, []byte("\xef\xbb\xbf"))
	if utf8.Valid(b) {
		return string(b)
	}
	var s strings.Builder
	s.Grow(len(b))
	for _, c := range b {
		switch {
		case c < 0x80:
			s.WriteByte(c)
		case c < 0xa0:
			s.WriteRune(cp1252[c-0x80])
		default:
			s.WriteRune(rune(c))
		}
	}
	return s.String()
}

// sniff picks the delimiter that occurs most often in the first line.
func sniff(text string) rune {
	first, _, _ := strings.Cut(text, "\n")
	best, n := ',', -1
	for _, d := range []rune{';', ',', '\t', '|'} {
		if c := strings.Count(first, string(d)); c > n {
			best, n = d, c
		}
	}
	return best
}

// readDelimited reads CSV or TSV. comma 0 means detect it.
func readDelimited(b []byte, comma rune, opt Options) ([]Row, []string, error) {
	text := decodeText(b)
	var warnings []string
	if !utf8.Valid(bytes.TrimPrefix(b, []byte("\xef\xbb\xbf"))) {
		warnings = append(warnings, "spreadsheet is not UTF-8; read as Windows-1252")
	}
	if comma == 0 {
		comma = sniff(text)
	}
	r := csv.NewReader(strings.NewReader(text))
	r.Comma = comma
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	r.ReuseRecord = false
	recs, err := r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("read delimited text: %w", err)
	}
	var cells []Cell
	for i, rec := range recs {
		for j, v := range rec {
			if strings.TrimSpace(v) == "" {
				continue
			}
			cells = append(cells, Cell{Col: j + 1, Row: i + 1, Ref: Ref(j+1, i+1), Val: v})
		}
	}
	rows, err := fromSheets([]string{"sheet"}, [][]Cell{cells}, opt)
	return rows, warnings, err
}
