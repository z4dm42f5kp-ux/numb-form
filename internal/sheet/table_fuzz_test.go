package sheet

import (
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/testfix"
)

// FuzzTable feeds arbitrary workbooks to the numb-date reader, cell
// writer and edit check.
func FuzzTable(f *testing.F) {
	f.Add(testfix.XLSX(f, fxIDs, fxNames), "A2", "x")
	f.Add(testfix.ODS(f, fxIDs, fxNames), "B3", "<&>")
	f.Fuzz(func(t *testing.T, b []byte, ref, val string) {
		_, _ = ReadTable(b, "Report created", "Customer Name")
		_, _ = ReadNames(b, "Customer Name", "")
		_, _ = AllCells(b)
		out, err := SetText(b, "xl/worksheets/sheet1.xml", map[string]string{ref: val})
		if err == nil {
			_ = CheckEdit(b, out, "xl/worksheets/sheet1.xml", map[string]string{ref: val})
		}
	})
}
