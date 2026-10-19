// Package form holds what every template format shares: the fields of
// section A, how section A and its labels are found in a list of
// paragraphs, and the self-check run after every fill.
package form

// Field identifies one line of section A.
type Field int

// Fields of section A, in template order.
const (
	FieldName Field = iota
	FieldLegalForm
	FieldAddress
	FieldRegistry
	FieldReps
	FieldDate
	FieldPerson
)

// AllFields lists the fields in template order.
var AllFields = []Field{FieldName, FieldLegalForm, FieldAddress, FieldRegistry, FieldReps, FieldDate, FieldPerson}

// labels are the normalised label texts (see Norm).
var labels = map[Field]string{
	FieldName:      "name/company name of the contracting party",
	FieldLegalForm: "legal form",
	FieldAddress:   "headquarters/address",
	FieldRegistry:  "commercial registry number",
	FieldReps:      "authorized representatives",
	FieldDate:      "date of the evaluation",
	FieldPerson:    "person processing the form",
}

var titles = map[Field]string{
	FieldName:      "Name / Company Name of the Contracting Party",
	FieldLegalForm: "Legal form",
	FieldAddress:   "Headquarters / Address",
	FieldRegistry:  "Commercial Registry Number",
	FieldReps:      "Authorized Representatives",
	FieldDate:      "Date of the evaluation",
	FieldPerson:    "Person Processing the Form",
}

// Title returns the label as printed on the form.
func (f Field) Title() string {
	if t, ok := titles[f]; ok {
		return t
	}
	return "unknown"
}

// Values maps fields to the text to insert. Missing or empty values leave
// the field blank. FieldRegistry is always left blank.
type Values map[Field]string

// Filler fills one template format.
type Filler interface {
	// Fill returns a copy of the template with section A filled in.
	Fill(template []byte, v Values) (out []byte, warnings []string, err error)
	// ReadBack returns the values currently in section A.
	ReadBack(doc []byte) (Values, error)
}

// Check validates that a template has a fillable section A.
func Check(f Filler, template []byte) ([]string, error) {
	_, w, err := f.Fill(template, Values{})
	return w, err
}
