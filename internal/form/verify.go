package form

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
)

// Para is a paragraph's byte range in the source and its plain text.
type Para struct {
	Start, End int
	Text       string
}

// Texts returns the paragraph texts.
func Texts(ps []Para) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.Text
	}
	return out
}

// Edit inserts Insert at byte offset At.
type Edit struct {
	At     int
	Insert string
}

// Apply returns a copy of doc with the edits applied.
func Apply(doc []byte, edits []Edit) []byte {
	es := append([]Edit(nil), edits...)
	sort.SliceStable(es, func(i, j int) bool { return es[i].At > es[j].At })
	out := append([]byte(nil), doc...)
	for _, e := range es {
		out = append(out[:e.At], append([]byte(e.Insert), out[e.At:]...)...)
	}
	return out
}

// Verify checks a fill: the same paragraphs exist, only the expected
// paragraphs changed and read as expected, and every byte before the
// first and after the last edited paragraph is unchanged (so section B
// and everything else is untouched).
func Verify(before, after []byte, bp, ap []Para, expect map[int]string) error {
	if len(bp) != len(ap) {
		return fmt.Errorf("paragraph count changed (%d -> %d)", len(bp), len(ap))
	}
	for i := range bp {
		if e, ok := expect[i]; ok {
			if Collapse(ap[i].Text) != e {
				return fmt.Errorf("paragraph %d reads %q, want %q", i, ap[i].Text, e)
			}
			continue
		}
		if ap[i].Text != bp[i].Text {
			return fmt.Errorf("paragraph %d outside the edited fields changed", i)
		}
	}
	if len(expect) == 0 {
		if !bytes.Equal(before, after) {
			return errors.New("document changed although nothing was filled")
		}
		return nil
	}
	first, last := len(bp), -1
	for i := range expect {
		first = min(first, i)
		last = max(last, i)
	}
	if !bytes.Equal(before[:bp[first].Start], after[:ap[first].Start]) {
		return errors.New("bytes before section A changed")
	}
	if !bytes.Equal(before[bp[last].End:], after[ap[last].End:]) {
		return errors.New("bytes after the last edited field changed")
	}
	return nil
}
