package rtf

import (
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

func TestEncode(t *testing.T) {
	if got := encode(`a{b}\c ä😀`); got != `a\{b\}\\c \u228?\u-10179?\u-8704?` {
		t.Fatalf("encode = %q", got)
	}
}

// FuzzScan feeds arbitrary templates to the tokenizer and filler.
func FuzzScan(f *testing.F) {
	f.Add(testfix.RTF())
	f.Add([]byte(`{\rtf1{\*\x \par}\bin3 abc\'zz\u-1\par}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		out, _, err := Format{}.Fill(b, form.Values{form.FieldLegalForm: "GmbH"})
		if err != nil {
			return
		}
		if _, err := (Format{}).ReadBack(out); err != nil {
			t.Fatalf("filled output unreadable: %v", err)
		}
	})
}
