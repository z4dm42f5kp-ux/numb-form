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

// scanner holds the state of scan.
type scanner struct {
	data        []byte
	stack       []state
	cur         *para
	paras       []*para
	pendingSkip int  // fallback characters still to skip after \uN
	high        rune // pending high surrogate of a \uN pair
	newGroup    bool
}

func (sc *scanner) finish(at int) {
	sc.cur.end = at
	sc.paras = append(sc.paras, sc.cur)
	sc.cur = &para{start: at, lastEnd: -1, tabEnd: -1}
}

func (sc *scanner) visible(r rune, end int) {
	if sc.stack[len(sc.stack)-1].skip {
		return
	}
	sc.cur.text.WriteRune(r)
	sc.cur.lastEnd = end
	sc.cur.tabEnd = -1
}

// scan returns the body paragraphs, in order.
func scan(data []byte) ([]*para, error) {
	if !strings.HasPrefix(strings.TrimLeft(string(data[:min(len(data), 16)]), " \r\n\t"), `{\rtf`) {
		return nil, errors.New("not an RTF file")
	}
	sc := &scanner{data: data, stack: []state{{uc: 1}}, cur: &para{start: 0, lastEnd: -1, tabEnd: -1}}
	for i := 0; i < len(data); {
		c := data[i]
		switch c {
		case '{':
			sc.stack = append(sc.stack, sc.stack[len(sc.stack)-1])
			sc.newGroup = true
			i++
			continue
		case '}':
			if len(sc.stack) == 1 {
				return nil, errors.New("rtf: unbalanced braces")
			}
			sc.stack = sc.stack[:len(sc.stack)-1]
			i++
			sc.newGroup = false
			continue
		case '\r', '\n':
			i++
			continue
		case '\\':
			i = sc.controlWord(i)
			continue
		}
		sc.newGroup = false
		if sc.pendingSkip > 0 {
			sc.pendingSkip--
			i++
			continue
		}
		sc.visible(rune(c), i+1)
		i++
	}
	if len(sc.stack) != 1 {
		return nil, errors.New("rtf: unbalanced braces")
	}
	sc.finish(len(data))
	return sc.paras, nil
}

// controlWord handles the control word at data[i] and returns the offset
// after it.
func (sc *scanner) controlWord(i int) int {
	data := sc.data
	word, param, hasParam, end := control(data, i)
	first := sc.newGroup
	sc.newGroup = false
	st := &sc.stack[len(sc.stack)-1]
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
		sc.unicode(param, end)
		sc.pendingSkip = st.uc
	case word == "'":
		if sc.pendingSkip > 0 {
			sc.pendingSkip--
		} else if end <= len(data) {
			v, _ := strconv.ParseUint(string(data[i+2:end]), 16, 8)
			sc.visible(cp1252(byte(v)), end)
		}
	case word == "par" || word == "cell" || word == "row" || word == "sect" || word == "page" || word == "\n" || word == "\r":
		if !st.skip {
			sc.finish(end)
		}
	case word == "tab":
		if !st.skip {
			sc.cur.text.WriteByte(' ')
			if sc.cur.lastEnd >= 0 {
				sc.cur.tabEnd = end
			}
		}
	case word == "line" || word == "~":
		sc.visible(' ', end)
	case word == "\\" || word == "{" || word == "}":
		sc.visible(rune(word[0]), end)
	case word == "emdash":
		sc.visible('—', end)
	case word == "endash":
		sc.visible('–', end)
	case word == "lquote" || word == "rquote":
		sc.visible('\'', end)
	case word == "ldblquote" || word == "rdblquote":
		sc.visible('"', end)
	}
	return end
}

// unicode handles \uN, a signed 16-bit value; anything else is ignored.
func (sc *scanner) unicode(param, end int) {
	if param < -32768 || param > 65535 {
		return
	}
	if param < 0 {
		param += 65536
	}
	r := rune(param)
	switch {
	case utf16.IsSurrogate(r) && r < 0xDC00:
		sc.high = r
	case utf16.IsSurrogate(r) && sc.high != 0:
		sc.visible(utf16.DecodeRune(sc.high, r), end)
		sc.high = 0
	default:
		sc.visible(r, end)
		sc.high = 0
	}
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
