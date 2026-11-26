package reportdate

import (
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

// FuzzFromForm feeds arbitrary form files to the report date reader of
// numb-date.
func FuzzFromForm(f *testing.F) {
	f.Add(testfix.Template(f), "a.docx")
	f.Add(testfix.ODT(f), "a.odt")
	f.Add(testfix.FODT(), "a.fodt")
	f.Add(testfix.RTF(), "a.rtf")
	f.Fuzz(func(t *testing.T, b []byte, name string) {
		_, _, _ = FromForm(b, name)
		_, _ = Parse(string(b))
	})
}
