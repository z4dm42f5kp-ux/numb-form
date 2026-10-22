package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// company handles one spreadsheet row. A returned error stops the run;
// problems specific to this company are recorded in its status instead.
func (r *run) company(row sheet.Row) (runlog.Company, error) {
	c := runlog.Company{Name: row.Name, Cell: row.Sheet + "!" + row.Cell}
	done, err := r.existing(row, &c)
	if done || err != nil {
		return c, err
	}
	folderName, changed := fsutil.SafeName(row.Name, false)
	if changed {
		r.ui.Printf("  note: folder will be named %q (characters not allowed on Windows were adjusted)\n", folderName)
		r.logf("folder name %q adjusted to %q", row.Name, folderName)
	}
	folder := filepath.Join(r.workRoot, folderName)
	if _, err := os.Lstat(folder); err == nil {
		c.Status, c.Reason = "skipped", "a file or folder named "+folderName+" already exists"
		r.ui.Printf("  skipped: %s\n", c.Reason)
		return c, nil
	}
	fileBase, _ := fsutil.SafeName(row.Name, true)
	fileName := strings.TrimSuffix(r.tpl.Name, filepath.Ext(r.tpl.Name)) + " - " + fileBase + r.tpl.OutExt

	vals, err := r.askValues(row.Name)
	if err != nil {
		return c, err
	}

	if err := os.Mkdir(folder, 0o755); err != nil { // #nosec G301 -- company folders live on a team share.
		c.Status, c.Reason = "error", err.Error()
		return c, nil
	}
	r.logf("created folder %s", folderName)
	accepted := false
	defer func() {
		if accepted {
			return
		}
		if err := removeCreated(r.workRoot, folder); err != nil {
			r.logf("ERROR could not remove %s: %v", folder, err)
		} else {
			r.logf("removed folder %s", folderName)
		}
	}()
	target := filepath.Join(folder, fileName)
	shown, ok, err := r.confirmLoop(&c, target, vals)
	if err != nil || !ok {
		return c, err
	}
	h, n, err := fsutil.SHA256File(target)
	if err != nil {
		return c, err
	}
	accepted = true
	c.Status, c.Folder = "accepted", folderName
	c.Files = []runlog.File{{Rel: filepath.ToSlash(filepath.Join(folderName, fileName)), SHA256: h, Size: n}}
	c.Fields = map[string]string{}
	for _, f := range form.AllFields {
		c.Fields[f.Title()] = shown[f]
	}
	r.ix.Add(folder, folderName, folderName)
	r.logf("accepted %q -> %s sha256=%s fields=%q", row.Name, c.Files[0].Rel, h, c.Fields)
	return c, nil
}

// existing reports (done=true) when the company already has a folder,
// asking about near matches.
func (r *run) existing(row sheet.Row, c *runlog.Company) (bool, error) {
	exact, fuzzy := r.ix.Find(row.Name)
	if len(exact) > 0 {
		c.Status, c.Folder, c.Reason = "exists", exact[0].Rel, "folder name matches"
		r.ui.Printf("  already has a folder: %s — nothing to do\n", ui.Safe(exact[0].Rel))
		r.logf("exists %q -> %s", row.Name, exact[0].Rel)
		return true, nil
	}
	for _, cand := range fuzzy {
		ok, err := r.ui.YesNo(fmt.Sprintf("  Is the existing folder %q this company's folder? (%s)", ui.Safe(cand.Folder.Rel), cand.Reason), false)
		if err != nil {
			return true, err
		}
		r.logf("fuzzy %q vs %s (%.2f): user said %v", row.Name, cand.Folder.Rel, cand.Score, ok)
		if ok {
			c.Status, c.Folder, c.Reason = "exists", cand.Folder.Rel, "confirmed by user: "+cand.Reason
			return true, nil
		}
	}
	return false, nil
}

// confirmLoop writes the document, shows what it now reads and lets the
// user accept, edit or reject it.
func (r *run) confirmLoop(c *runlog.Company, target string, vals form.Values) (form.Values, bool, error) {
	for {
		if err := r.write(target, vals); err != nil {
			c.Status, c.Reason = "error", err.Error()
			r.ui.Printf("  error: %v\n", err)
			r.logf("ERROR %q: %v", c.Name, err)
			return nil, false, nil
		}
		shown, err := r.show(target)
		if err != nil {
			c.Status, c.Reason = "error", err.Error()
			return nil, false, nil
		}
		ch, err := r.ui.Choose("  Is this correct? [a]ccept, [e]dit a field, [r]eject: ", "aer", 0)
		if err != nil {
			return nil, false, err
		}
		switch ch {
		case 'a':
			return shown, true, nil
		case 'r':
			c.Status, c.Reason = "rejected", "rejected by user"
			r.logf("rejected %q", c.Name)
			return nil, false, nil
		case 'e':
			if err := r.editField(vals); err != nil {
				return nil, false, err
			}
			if err := os.Remove(target); err != nil {
				return nil, false, err
			}
		}
	}
}

// removeCreated deletes a folder this run created, after checking that it
// lies inside the working root.
func removeCreated(root, folder string) error {
	rr, err := filepath.Rel(root, folder)
	if err != nil || rr == "." || strings.HasPrefix(rr, "..") || filepath.IsAbs(rr) || strings.ContainsRune(rr, filepath.Separator) {
		return errors.New("refusing to remove a folder outside the working folder")
	}
	return os.RemoveAll(folder)
}
