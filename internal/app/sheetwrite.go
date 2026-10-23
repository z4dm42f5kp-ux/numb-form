package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// cellEdit is one spreadsheet cell to write, with what to show for it.
type cellEdit struct {
	ref, value, label string
}

// writeCells writes edits into one sheet of the spreadsheet at path, which
// held before (SHA-256 have) when it was read. Nothing else may change:
// the new workbook is compared cell by cell first, every step is checked
// with SHA-256, and the original is kept in the log folder until the
// user accepts what is read back from disk, or undoes it. It reports
// whether the user accepted.
func (r *run) writeCells(path string, before []byte, have, sheetName, part string, edits []cellEdit) (bool, error) {
	if err := notOpen(path); err != nil {
		return false, err
	}
	vals := map[string]string{}
	for _, e := range edits {
		vals[e.ref] = e.value
	}
	after, err := sheet.SetText(before, part, vals)
	if err != nil {
		return false, err
	}
	if err := sheet.CheckEdit(before, after, sheetName, vals); err != nil {
		return false, fmt.Errorf("self-check failed, nothing was written: %w", err)
	}
	backup := filepath.Join(r.runDir, "before-"+filepath.Base(path))
	if err := fsutil.WriteNew(backup, before); err != nil {
		return false, fmt.Errorf("could not keep a copy of the original: %w", err)
	}
	if err := fsutil.ReplaceVerified(path, have, after); err != nil {
		_ = os.Remove(backup)
		return false, err
	}
	wrote := fsutil.SHA256Bytes(after)
	r.logf("wrote %d cell(s): sha256 %s -> %s; original kept at %s", len(edits), have, wrote, backup)
	for _, e := range edits {
		r.logf("  %s = %q (%s)", e.ref, e.value, e.label)
	}

	// Read the file back from disk and show what the cells now hold.
	got, err := readSheetFile(path)
	if err != nil {
		return false, err
	}
	cells, err := sheet.AllCells(got)
	if err != nil {
		return false, err
	}
	now := map[string]string{}
	for _, c := range cells[sheetName] {
		now[c.Ref] = c.Val
	}
	r.ui.Printf("\nWritten. %s now reads:\n", filepath.Base(path))
	for _, e := range edits {
		r.ui.Printf("  %-6s %s  (%s)\n", e.ref, ui.Safe(now[e.ref]), ui.Safe(e.label))
	}
	ch, err := r.ui.Choose("Is this correct? [a]ccept, [u]ndo: ", "au", 0)
	if err != nil {
		r.ui.Printf("\nThe changes were written; the original spreadsheet is kept at %s\n", backup)
		return false, err
	}
	if ch == 'u' {
		if err := fsutil.ReplaceVerified(path, wrote, before); err != nil {
			r.ui.Printf("Could not undo: %v\nThe original spreadsheet is kept at %s\n", err, backup)
			return false, err
		}
		_ = os.Remove(backup)
		r.logf("undone: sha256 back to %s", have)
		r.ui.Println("Undone: the spreadsheet is exactly as before (SHA-256 checked).")
		return false, nil
	}
	_ = os.Remove(backup)
	r.logf("accepted")
	return true, nil
}
