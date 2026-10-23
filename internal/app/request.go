package app

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	kind   string // add, replace, keep
	rel    string // slash-separated, relative to the working folder
	sha    string // add, replace: of the working copy's file
	old    string // replace: of the real folder's file
	size   int64
	reason string // keep
	top    string // add: the new folder it belongs to, "" for an existing folder
}

// request is what numb-commit will do: copy files from a working copy
// (From) into the real working folder (To).
type request struct {
	From, To string
	Made     time.Time
	Status   string // waiting, applied
	Sheet    string // rel path of the overview spreadsheet
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
// New files are added, the overview spreadsheet (sheetRel) may be
// replaced, and nothing is ever deleted; the template and residue are
// never copied.
func scanRequest(ctx context.Context, from, to, sheetRel, templateRel string) (*request, error) {
	q := &request{From: from, To: to, Status: "waiting", Sheet: filepath.ToSlash(sheetRel)}
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
				q.Changes = append(q.Changes, change{kind: "keep", rel: slash, reason: "not copied: " + why})
			}
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		if strings.EqualFold(rel, templateRel) {
			q.Changes = append(q.Changes, change{kind: "keep", rel: slash, reason: "not copied: the template"})
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
			q.Changes = append(q.Changes, change{kind: "add", rel: slash, sha: h, size: info.Size()})
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
			q.Changes = append(q.Changes, change{kind: "replace", rel: slash, sha: hs, old: ht, size: info.Size()})
			return nil
		}
		q.Changes = append(q.Changes, change{kind: "keep", rel: slash, reason: "differs from the real folder; numb-commit replaces only the spreadsheet"})
		return nil
	})
	sort.SliceStable(q.Changes, func(i, j int) bool { return q.Changes[i].rel < q.Changes[j].rel })
	// Which new folder each file belongs to is fixed now, before copying
	// makes the folders exist.
	for i, c := range q.Changes {
		if c.kind == "add" {
			q.Changes[i].top = firstMissing(to, c.rel)
		}
	}
	return q, err
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
	for _, c := range q.of("add") {
		top := c.top
		if _, ok := groups[top]; !ok {
			order = append(order, top)
		}
		groups[top] = append(groups[top], c)
	}
	return groups, order
}

func kb(n int64) string { return strconv.FormatInt((n+1023)/1024, 10) + " KB" }

// write saves the request as text people can read, followed by the same
// data in lines for numb-commit.
func (q *request) write(path string) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "numb-commit request: what numb-commit copies into the real folder\n\n")
	fmt.Fprintf(&b, "From:   %s\nTo:     %s\nMade:   %s\nStatus: %s\n\n", q.From, q.To, q.Made.Format("02/01/2006 15:04"), q.Status)
	groups, order := q.newFolders()
	for _, top := range order {
		if top == "" {
			continue
		}
		fmt.Fprintf(&b, "New folder  %s\n", filepath.FromSlash(top))
		for _, c := range groups[top] {
			fmt.Fprintf(&b, "  + %s  (%s, sha256 %s)\n", filepath.FromSlash(c.rel), kb(c.size), c.sha)
		}
		b.WriteString("\n")
	}
	if fs := groups[""]; len(fs) > 0 {
		b.WriteString("New files in existing folders\n")
		for _, c := range fs {
			fmt.Fprintf(&b, "  + %s  (%s, sha256 %s)\n", filepath.FromSlash(c.rel), kb(c.size), c.sha)
		}
		b.WriteString("\n")
	}
	for _, c := range q.of("replace") {
		fmt.Fprintf(&b, "Replaced  %s\n  sha256 %s in the real folder now\n  sha256 %s afterwards\n  the replaced version is kept in the log folder\n\n", filepath.FromSlash(c.rel), c.old, c.sha)
	}
	if ks := q.of("keep"); len(ks) > 0 {
		b.WriteString("Left alone\n")
		for _, c := range ks {
			fmt.Fprintf(&b, "  - %s  (%s)\n", filepath.FromSlash(c.rel), c.reason)
		}
		b.WriteString("\n")
	}
	b.WriteString("Nothing in the real folder is deleted, and no file there is overwritten except the ones under Replaced.\n\n")
	b.WriteString(requestData + "\n")
	fmt.Fprintf(&b, "from\t%s\nto\t%s\nmade\t%s\nstatus\t%s\nsheet\t%s\n", q.From, q.To, q.Made.Format(time.RFC3339), q.Status, q.Sheet)
	for _, c := range q.Changes {
		extra := c.reason
		if c.kind == "add" {
			extra = "folder:" + c.top
		}
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%d\t%s\n", c.kind, c.rel, c.sha, c.old, c.size, extra)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b.Bytes(), 0o644); err != nil { // #nosec G306 -- read by the team.
		return err
	}
	return os.Rename(tmp, path)
}

// readRequest parses the data lines of a request file.
func readRequest(path string) (*request, error) {
	b, err := os.ReadFile(path) // #nosec G304 -- the request in the working folder.
	if err != nil {
		return nil, err
	}
	i := bytes.Index(b, []byte(requestData))
	if i < 0 {
		return nil, fmt.Errorf("%s has no data section", filepath.Base(path))
	}
	q := &request{}
	sc := bufio.NewScanner(bytes.NewReader(b[i+len(requestData):]))
	for sc.Scan() {
		f := strings.Split(sc.Text(), "\t")
		switch {
		case len(f) == 2 && f[0] == "from":
			q.From = f[1]
		case len(f) == 2 && f[0] == "to":
			q.To = f[1]
		case len(f) == 2 && f[0] == "made":
			q.Made, _ = time.Parse(time.RFC3339, f[1])
		case len(f) == 2 && f[0] == "status":
			q.Status = f[1]
		case len(f) == 2 && f[0] == "sheet":
			q.Sheet = f[1]
		case len(f) == 6 && (f[0] == "add" || f[0] == "replace" || f[0] == "keep"):
			n, _ := strconv.ParseInt(f[4], 10, 64)
			rel := f[1]
			if rel == "" || strings.HasPrefix(rel, "/") || strings.Contains("/"+rel+"/", "/../") || filepath.IsAbs(filepath.FromSlash(rel)) {
				return nil, fmt.Errorf("unsafe path %q in the request", rel)
			}
			c := change{kind: f[0], rel: rel, sha: f[2], old: f[3], size: n, reason: f[5]}
			if c.kind == "add" {
				c.reason, c.top = "", strings.TrimPrefix(f[5], "folder:")
				if c.top != "" && !strings.HasPrefix(c.rel, c.top+"/") {
					return nil, fmt.Errorf("unsafe folder %q in the request", c.top)
				}
			}
			q.Changes = append(q.Changes, c)
		case strings.TrimSpace(sc.Text()) == "":
		default:
			return nil, fmt.Errorf("%s: unreadable line %q", filepath.Base(path), sc.Text())
		}
	}
	if q.From == "" || q.To == "" {
		return nil, fmt.Errorf("%s does not say where to copy from and to", filepath.Base(path))
	}
	return q, sc.Err()
}
