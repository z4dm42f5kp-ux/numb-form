package docx

import (
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

// FuzzTemplate feeds arbitrary .docx templates to the filler.
func FuzzTemplate(f *testing.F) {
	f.Add(testfix.Template(f))
	f.Fuzz(func(t *testing.T, b []byte) {
		out, _, err := fmtx.Fill(b, form.Values{form.FieldLegalForm: "GmbH"})
		if err == nil {
			if _, err := fmtx.ReadBack(out); err != nil {
				t.Fatalf("filled output unreadable: %v", err)
			}
		}
		_, _ = fmtx.ReadBack(b)
	})
}
