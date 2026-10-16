package odt

import (
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

// FuzzScan feeds arbitrary flat templates to the scanner and filler.
func FuzzScan(f *testing.F) {
	f.Add(testfix.FODT())
	f.Add([]byte(`<office:document xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"><text:p>x<text:tab/></text:p></office:document>`))
	f.Fuzz(func(t *testing.T, b []byte) {
		out, _, err := Format{Flat: true}.Fill(b, form.Values{form.FieldLegalForm: "GmbH"})
		if err != nil {
			return
		}
		if _, err := (Format{Flat: true}).ReadBack(out); err != nil {
			t.Fatalf("filled output unreadable: %v", err)
		}
	})
}
