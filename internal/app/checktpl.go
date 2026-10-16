package app

import (
	"fmt"
	"io"
	"os"

	"github.com/z4dm42f5kp-ux/numb-form/internal/docx"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
)

// CheckTemplate reports whether a template can be filled. With a sample
// path it also writes a filled example there (the path must not exist).
// The template itself is only read.
func CheckTemplate(out io.Writer, tplPath, sample string) int {
	b, err := os.ReadFile(tplPath) // #nosec G304 -- user-supplied template.
	if err != nil {
		fmt.Fprintln(out, "check-template:", err)
		return ExitError
	}
	warns, err := docx.Check(b)
	if err != nil {
		fmt.Fprintln(out, "check-template: NOT fillable:", err)
		return ExitError
	}
	for _, w := range warns {
		fmt.Fprintln(out, "note:", w)
	}
	fmt.Fprintln(out, "Template is fillable: section A and all labels were found; section B is never touched.")
	if sample == "" {
		return ExitOK
	}
	vals := docx.Values{
		docx.FieldName:      "Foo GmbH",
		docx.FieldLegalForm: "GmbH",
		docx.FieldAddress:   "Foo Street 64, 80333 München, Germany",
		docx.FieldReps:      "Jane Doe (Managing Director, Legal Representative), John Roe (Authorized Signatory (Prokurist)) " + CautionText,
		docx.FieldDate:      "26/09/2026",
		docx.FieldPerson:    "Max Muster",
	}
	filled, _, err := docx.Fill(b, vals)
	if err != nil {
		fmt.Fprintln(out, "check-template:", err)
		return ExitError
	}
	if err := fsutil.WriteNew(sample, filled); err != nil {
		fmt.Fprintln(out, "check-template:", err)
		return ExitError
	}
	fmt.Fprintln(out, "Sample written to", sample)
	return ExitOK
}
