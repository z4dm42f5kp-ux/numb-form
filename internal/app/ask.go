package app

import (
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// askValues fills the obvious fields from the spreadsheet, the run and
// the name, and asks the person for the address and the representatives.
// Every value can still be changed with [e]dit before accepting.
func (r *run) askValues(name string) (form.Values, error) {
	vals := form.Values{form.FieldName: name, form.FieldRegistry: RegistryText, form.FieldDate: r.date, form.FieldPerson: r.person}
	r.ui.Println()
	lf, _ := match.LegalForm(name)
	if lf != "" {
		r.ui.Printf("  Legal form: %s (from the name)\n", ui.Safe(lf))
	} else {
		v, err := r.ui.Ask("  Legal form (the name has none; Enter = blank): ")
		if err != nil {
			return nil, err
		}
		lf = v
	}
	vals[form.FieldLegalForm] = textutil.OneLine(lf)

	addr, err := r.ui.Ask("  Headquarters / Address (one line, Enter = blank): ")
	if err != nil {
		return nil, err
	}
	vals[form.FieldAddress] = textutil.OneLine(addr)

	r.ui.Println("  Authorized Representatives: Full Name (Role 1, Role 2), Full Name (Role) — roles in English.")
	reps, err := r.ui.Ask("  Representatives (Enter = \"No representatives were found…\", '-' = blank): ")
	if err != nil {
		return nil, err
	}
	switch reps = textutil.OneLine(reps); {
	case reps == "":
		reps = NoRepsText
	case strings.TrimSpace(reps) == "-":
		reps = ""
	}
	vals[form.FieldReps] = reps
	return vals, nil
}
