// Package rtf fills section A of Rich Text Format templates, which every
// office suite can save.
//
// The file is tokenised only to find where each paragraph's visible text
// ends. Values are inserted right after the label's last character, inside
// the same group, so they take the label's formatting; the inserted text
// is wrapped in a plain group that only sets the Unicode fallback count.
//
// In plain words: as for Word files, values are written after their
// labels in .rtf files, and nothing else changes.
package rtf

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
)

// skipDest are destinations whose text is not part of the body.
var skipDest = map[string]bool{
	"fonttbl": true, "colortbl": true, "stylesheet": true, "info": true, "pict": true,
	"object": true, "header": true, "headerl": true, "headerr": true, "headerf": true,
	"footer": true, "footerl": true, "footerr": true, "footerf": true, "footnote": true,
	"annotation": true, "fldinst": true, "themedata": true, "colorschememapping": true,
	"latentstyles": true, "datastore": true, "xmlnstbl": true, "listtable": true,
	"listoverridetable": true, "rsidtbl": true, "generator": true, "mmathPr": true,
	"pgdsctbl": true, "filetbl": true, "revtbl": true, "userprops": true, "background": true,
	"shp": true, "shpinst": true, "nonshppict": true, "bkmkstart": true, "bkmkend": true,
}

type para struct {
	start, end int
	text       strings.Builder
	lastEnd    int // offset just after the token of the last visible character
	tabEnd     int // offset after a tab that follows the last character, or -1
}

type state struct {
	skip bool
	uc   int
}

// scan returns the body paragraphs, in order.
func scan(data []byte) ([]*para, error) {
	if !strings.HasPrefix(strings.TrimLeft(string(data[:min(len(data), 16)]), " \r\n\t"), `{\rtf`) {
		return nil, errors.New("not an RTF file")
	}
	stack := []state{{uc: 1}}
	cur := &para{start: 0, lastEnd: -1, tabEnd: -1}
	var paras []*para
	pendingSkip := 0 // fallback characters still to skip after \uN
	var high rune    // pending high surrogate of a \uN pair
	newGroup := false
	finish := func(at int) {
		cur.end = at
		paras = append(paras, cur)
		cur = &para{start: at, lastEnd: -1, tabEnd: -1}
	}
	visible := func(r rune, end int) {
		st := stack[len(stack)-1]
		if st.skip {
			return
		}
		cur.text.WriteRune(r)
		cur.lastEnd = end
		cur.tabEnd = -1
	}
	for i := 0; i < len(data); {
		c := data[i]
		switch c {
		case '{':
			stack = append(stack, stack[len(stack)-1])
			newGroup = true
			i++
			continue
		case '}':
			if len(stack) == 1 {
				return nil, errors.New("rtf: unbalanced braces")
			}
			stack = stack[:len(stack)-1]
			i++
			newGroup = false
			continue
		case '\r', '\n':
			i++
			continue
		case '\\':
			word, param, hasParam, end := control(data, i)
			first := newGroup
			newGroup = false
			st := &stack[len(stack)-1]
			switch {
			case word == "*" && first:
				st.skip = true
			case first && skipDest[word]:
				st.skip = true
			case word == "bin" && hasParam:
				end += max(0, min(param, len(data)-end))
			case word == "uc" && hasParam:
				st.uc = max(0, min(param, 8))
			case word == "u" && hasParam:
				// \uN is a signed 16-bit value; anything else is ignored.
				if param < -32768 || param > 65535 {
					break
				}
				if param < 0 {
					param += 65536
				}
				r := rune(param)
				switch {
				case utf16.IsSurrogate(r) && r < 0xDC00:
					high = r
				case utf16.IsSurrogate(r) && high != 0:
					visible(utf16.DecodeRune(high, r), end)
					high = 0
				default:
					visible(r, end)
					high = 0
				}
				pendingSkip = st.uc
			case word == "'":
				if pendingSkip > 0 {
					pendingSkip--
				} else if end <= len(data) {
					v, _ := strconv.ParseUint(string(data[i+2:end]), 16, 8)
					visible(cp1252(byte(v)), end)
				}
			case word == "par" || word == "cell" || word == "row" || word == "sect" || word == "page" || word == "\n" || word == "\r":
				if !st.skip {
					finish(end)
				}
			case word == "line":
				visible(' ', end)
			case word == "tab":
				if !st.skip {
					cur.text.WriteByte(' ')
					if cur.lastEnd >= 0 {
						cur.tabEnd = end
					}
				}
			case word == "~":
				visible(' ', end)
			case word == "\\" || word == "{" || word == "}":
				visible(rune(word[0]), end)
			case word == "emdash":
				visible('—', end)
			case word == "endash":
				visible('–', end)
			case word == "lquote" || word == "rquote":
				visible('\'', end)
			case word == "ldblquote" || word == "rdblquote":
				visible('"', end)
			}
			i = end
			continue
		}
		newGroup = false
		if pendingSkip > 0 {
			pendingSkip--
			i++
			continue
		}
		visible(rune(c), i+1)
		i++
	}
	if len(stack) != 1 {
		return nil, errors.New("rtf: unbalanced braces")
	}
	finish(len(data))
	return paras, nil
}

// control parses the control word or symbol at data[i] == '\\' and
// returns it, its numeric parameter and the offset after it (including
// the single delimiting space).
func control(data []byte, i int) (string, int, bool, int) {
	j := i + 1
	if j >= len(data) {
		return "", 0, false, j
	}
	c := data[j]
	if !isLetter(c) {
		if c == '\'' {
			return "'", 0, false, min(j+3, len(data))
		}
		return string(c), 0, false, j + 1
	}
	k := j
	for k < len(data) && isLetter(data[k]) && k-j < 32 {
		k++
	}
	word := string(data[j:k])
	p := k
	if p < len(data) && (data[p] == '-' || isDigit(data[p])) {
		p++
		for p < len(data) && isDigit(data[p]) && p-k < 11 {
			p++
		}
	}
	param, err := strconv.Atoi(string(data[k:p]))
	has := err == nil && p > k
	if p < len(data) && data[p] == ' ' {
		p++
	}
	return word, param, has, p
}

func isLetter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
func isDigit(c byte) bool  { return c >= '0' && c <= '9' }

func cp1252(b byte) rune {
	table := [32]rune{'€', 0x81, '‚', 'ƒ', '„', '…', '†', '‡', 'ˆ', '‰', 'Š', '‹', 'Œ', 0x8d, 'Ž', 0x8f,
		0x90, '‘', '’', '“', '”', '•', '–', '—', '˜', '™', 'š', '›', 'œ', 0x9d, 'ž', 'Ÿ'}
	if b >= 0x80 && b < 0xa0 {
		return table[b-0x80]
	}
	return rune(b)
}

func toForm(ps []*para) []form.Para {
	out := make([]form.Para, len(ps))
	for i, p := range ps {
		out[i] = form.Para{Start: p.start, End: p.end, Text: p.text.String()}
	}
	return out
}
