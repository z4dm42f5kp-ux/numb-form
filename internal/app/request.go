package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
)

// RequestName is the file numb-commit writes into the working copy: what
// it will change in the real folder, for people and for numb-commit
// review.
const RequestName = "numb-commit-request.txt"

const requestData = "--- for numb-commit, do not edit below this line ---"

// change is one line of a request.
type change struct {
	kind   string // add, replace, keep, differs, dup (see the kinds below)
	rel    string // slash-separated, relative to both folders
	sha    string // add, replace: of the working copy's file
	old    string // replace: of the real folder's file
	size   int64
	top    string // add, dup: the new folder it belongs to, "" for an existing folder
	dup    string // dup: the real folder's folder of the same company
	origin string // byNumbForm, editedForm or otherFile
	reason string // keep, dup: why it is not copied
}

// Kinds of change. Only add and replace are copied.
const (
	kAdd     = "add"     // a new file
	kReplace = "replace" // the overview spreadsheet, changed by numb-date
	kKeep    = "keep"    // residue or the template: never copied
	kDiffers = "differs" // exists in the real folder with other content: never overwritten
	kDup     = "dup"     // in a new folder of a company the real folder already has
)

// request is what numb-commit will do: copy files from a working copy
// (From) into the real working folder (To). Stage is the copy of exactly
// those files in Downloads that is reviewed and then copied.
type request struct {
	From, To string
	Stage    string
	Made     time.Time
	Status   string // waiting, applied
	Sheet    string // rel path of the overview spreadsheet
	Others   int    // entries at the top of the real folder that do not change
	Changes  []change
}

func (q *request) of(kind string) []change {
	var out []change
	for _, c := range q.Changes {
		if c.kind == kind {
			out = append(out, c)
		}
	}
	return out
}

// residue reports files that are never copied: Office lock files,
// temporary and hidden files, logs, run folders and this request.
func residue(rel string, dir bool) string {
	base := filepath.Base(rel)
	low := strings.ToLower(base)
	switch {
	case low == strings.ToLower(RequestName):
		return "the commit request"
	case strings.HasPrefix(base, "~$") || strings.HasPrefix(base, ".~lock."):
		return "an Office lock file"
	case strings.HasPrefix(base, "."):
		return "a hidden or temporary file"
	case dir && low == "numb-form-runs":
		return "numb-form run folders"
	case !dir && (strings.HasSuffix(low, ".log") || strings.HasSuffix(low, ".tmp") || low == "thumbs.db" || low == "desktop.ini" || low == "manifest.json"):
		return "a log or system file"
	}
	return ""
}

// sameFile compares two files: equal size and time count as equal (a copy
// made with File Explorer keeps both); otherwise their SHA-256 decides.
func sameFile(a, b string, ai, bi fs.FileInfo) (bool, string, string, error) {
	if ai.Size() == bi.Size() && ai.ModTime().Equal(bi.ModTime()) {
		return true, "", "", nil
	}
	ha, _, err := fsutil.SHA256File(a)
	if err != nil {
		return false, "", "", err
	}
	hb, _, err := fsutil.SHA256File(b)
	return ha == hb, ha, hb, err
}

// scanRequest compares the working copy from with the real folder to.
// New files are added and the overview spreadsheet may be replaced;
// nothing is ever deleted or overwritten. The template and residue are
// never copied, and neither is a new folder of a company that the real
// folder already has under another name, such as "D10629 Foo S.A.".
func scanRequest(ctx context.Context, from, to string, rc reviewConfig) (*request, error) {
	q := &request{From: from, To: to, Status: "waiting", Sheet: filepath.ToSlash(rc.sheetRel)}
	err := filepath.WalkDir(from, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if p == from {
			return nil
		}
		rel, _ := filepath.Rel(from, p)
		slash := filepath.ToSlash(rel)
		if strings.ContainsAny(rel, "\t\r\n") {
			return fmt.Errorf("%q has a tab or line break in its name; rename it first", rel)
		}
		if why := residue(rel, d.IsDir()); why != "" {
			if !d.IsDir() || !strings.HasPrefix(d.Name(), ".") {
				q.Changes = append(q.Changes, change{kind: kKeep, rel: slash, reason: "not copied: " + why})
			}
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		if strings.EqualFold(slash, rc.template) {
			q.Changes = append(q.Changes, change{kind: kKeep, rel: slash, reason: "not copied: the template"})
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		target := filepath.Join(to, rel)
		ti, err := os.Lstat(target)
		if errors.Is(err, fs.ErrNotExist) {
			h, _, err := fsutil.SHA256File(p)
			q.Changes = append(q.Changes, change{kind: kAdd, rel: slash, sha: h, size: info.Size(), origin: originOf(p, slash, rc)})
			return err
		}
		if err != nil {
			return err
		}
		same, hs, ht, err := sameFile(p, target, info, ti)
		if err != nil || same {
			return err
		}
		if strings.EqualFold(slash, q.Sheet) {
			q.Changes = append(q.Changes, change{kind: kReplace, rel: slash, sha: hs, old: ht, size: info.Size()})
			return nil
		}
		q.Changes = append(q.Changes, change{kind: kDiffers, rel: slash, origin: originOf(p, slash, rc),
			reason: "differs from the real folder; existing files are never overwritten"})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(q.Changes, func(i, j int) bool { return q.Changes[i].rel < q.Changes[j].rel })
	// Which new folder each file belongs to is fixed now, before copying
	// makes the folders exist.
	for i, c := range q.Changes {
		if c.kind == kAdd {
			q.Changes[i].top = firstMissing(to, c.rel)
		}
	}
	q.skipKnownCompanies(rc)
	q.Others = othersAtTop(to, q)
	return q, nil
}

// skipKnownCompanies turns the files of each new folder whose company the
// real folder already has (under its customer id, for example) into dup:
// numb-form only made the form there because the working copy lacked
// that folder.
func (q *request) skipKnownCompanies(rc reviewConfig) {
	var folders []realFolder
	listed := false
	for i, c := range q.Changes {
		if c.kind != kAdd || c.top == "" {
			continue
		}
		if !listed {
			ov, _ := findOverview(q.To, rc.overviewDir)
			folders, listed = listFolders(q.To, ov), true
		}
		base := path.Base(c.top)
		company, id := base, ""
		if n, ok := rc.names[strings.ToLower(base)]; ok {
			company, id = n, rc.ids[strings.ToLower(base)]
		}
		if f, why, ok := companyFolder(folders, company, id); ok {
			q.Changes[i].kind, q.Changes[i].dup = kDup, f.rel
			q.Changes[i].reason = fmt.Sprintf("not copied: the real folder already has %q for %s (%s)", f.rel, company, why)
		}
	}
}

// othersAtTop counts the entries at the top of the real folder that the
// request does not touch, for the tree.
func othersAtTop(to string, q *request) int {
	entries, err := os.ReadDir(to)
	if err != nil {
		return 0
	}
	touched := map[string]bool{}
	for _, c := range q.Changes {
		if c.kind != kKeep {
			touched[strings.ToLower(strings.SplitN(c.rel, "/", 2)[0])] = true
		}
		if c.dup != "" {
			touched[strings.ToLower(strings.SplitN(c.dup, "/", 2)[0])] = true
		}
	}
	n := 0
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") && !touched[strings.ToLower(e.Name())] {
			n++
		}
	}
	return n
}

// firstMissing returns the first folder of rel that does not exist in
// root, or "" when the file's folder exists.
func firstMissing(root, rel string) string {
	parts := strings.Split(rel, "/")
	for i := 1; i < len(parts); i++ {
		dir := strings.Join(parts[:i], "/")
		if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(dir))); errors.Is(err, fs.ErrNotExist) {
			return dir
		}
	}
	return ""
}

// newFolders groups the added files by the new folder they belong to, as
// decided when the request was made; "" holds files added to existing
// folders.
func (q *request) newFolders() (map[string][]change, []string) {
	groups := map[string][]change{}
	var order []string
	for _, c := range q.of(kAdd) {
		top := c.top
		if _, ok := groups[top]; !ok {
			order = append(order, top)
		}
		groups[top] = append(groups[top], c)
	}
	return groups, order
}
