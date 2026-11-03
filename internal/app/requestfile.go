package app

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const requestData = "--- for numb-commit, do not edit below this line ---"

// maxRequest bounds the request file numb-commit reads.
const maxRequest = 16 << 20

func kb(n int64) string { return strconv.FormatInt((n+1023)/1024, 10) + " KB" }

// text is the part of the request for people: what changes, as a list
// and as the real folder's tree before and after.
func (q *request) text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "numb-commit request: what numb-commit copies into the real folder\n\n")
	fmt.Fprintf(&b, "From:   %s\nTo:     %s\n", q.From, q.To)
	if q.Stage != "" {
		fmt.Fprintf(&b, "Staged: %s\n", q.Stage)
	}
	fmt.Fprintf(&b, "Made:   %s\nStatus: %s\n\n", q.Made.Format("02/01/2006 15:04"), q.Status)
	line := func(c change) {
		note := ""
		if l := originLabel(c.origin); l != "" {
			note = ", " + l
		}
		fmt.Fprintf(&b, "  + %s  (%s%s, sha256 %s)\n", filepath.FromSlash(c.rel), kb(c.size), note, c.sha)
	}
	groups, order := q.newFolders()
	for _, top := range order {
		if top == "" {
			continue
		}
		fmt.Fprintf(&b, "New folder  %s\n", filepath.FromSlash(top))
		for _, c := range groups[top] {
			line(c)
		}
		b.WriteString("\n")
	}
	if fs := groups[""]; len(fs) > 0 {
		b.WriteString("New files in existing folders\n")
		for _, c := range fs {
			line(c)
		}
		b.WriteString("\n")
	}
	for _, c := range q.of(kReplace) {
		fmt.Fprintf(&b, "Replaced  %s\n  sha256 %s in the real folder now\n  sha256 %s afterwards\n  the replaced version is kept in the log folder\n\n", filepath.FromSlash(c.rel), c.old, c.sha)
	}
	var left []change
	for _, c := range q.Changes {
		if c.kind == kKeep || c.kind == kDiffers || c.kind == kDup {
			left = append(left, c)
		}
	}
	if len(left) > 0 {
		b.WriteString("Left alone\n")
		for _, c := range left {
			note := c.reason
			if l := originLabel(c.origin); l != "" {
				note += "; " + l
			}
			fmt.Fprintf(&b, "  - %s  (%s)\n", filepath.FromSlash(c.rel), note)
		}
		b.WriteString("\n")
	}
	b.WriteString("Nothing in the real folder is deleted, and no file there is overwritten except the ones under Replaced.\n\n")
	b.WriteString(q.tree())
	return b.String()
}

// write saves the request as text people can read, followed by the same
// data in lines for numb-commit.
func (q *request) write(path string) error {
	var b bytes.Buffer
	// The byte order mark makes Windows PowerShell 5.1 (Get-Content) and
	// old Notepad read the file as UTF-8, so the tree stays readable.
	b.WriteString("\ufeff")
	b.WriteString(q.text())
	b.WriteString("\n" + requestData + "\n")
	fmt.Fprintf(&b, "from\t%s\nto\t%s\nstage\t%s\nmade\t%s\nstatus\t%s\nsheet\t%s\nothers\t%d\n",
		q.From, q.To, q.Stage, q.Made.Format(time.RFC3339), q.Status, q.Sheet, q.Others)
	for _, c := range q.Changes {
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\n", c.kind, c.rel, c.sha, c.old, c.size, c.top, c.dup, c.origin, c.reason)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b.Bytes(), 0o644); err != nil { // #nosec G306 -- read by the team.
		return err
	}
	return os.Rename(tmp, path)
}

// safeRel accepts a slash-separated path inside a folder: no empty, "."
// or ".." parts, and no backslash, which Windows reads as a separator.
func safeRel(rel string) bool {
	if rel == "" || filepath.IsAbs(filepath.FromSlash(rel)) {
		return false
	}
	for _, part := range strings.Split(rel, "/") {
		if part == "" || part == "." || part == ".." || strings.ContainsRune(part, '\\') {
			return false
		}
	}
	return true
}

var reSHA256 = regexp.MustCompile(`^[0-9a-f]{64}$`)

// readRequest reads and parses a request file.
func readRequest(path string) (*request, error) {
	f, err := os.Open(path) // #nosec G304 -- the request in the working folder.
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxRequest+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxRequest {
		return nil, fmt.Errorf("%s is larger than %d MB", filepath.Base(path), maxRequest>>20)
	}
	return parseRequest(filepath.Base(path), b)
}

// parseRequest parses the data lines of the request file name.
func parseRequest(name string, b []byte) (*request, error) {
	i := bytes.Index(b, []byte(requestData))
	if i < 0 {
		return nil, fmt.Errorf("%s has no data section", name)
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
		case len(f) == 2 && f[0] == "stage":
			q.Stage = f[1]
		case len(f) == 2 && f[0] == "made":
			q.Made, _ = time.Parse(time.RFC3339, f[1])
		case len(f) == 2 && f[0] == "status":
			q.Status = f[1]
		case len(f) == 2 && f[0] == "sheet":
			q.Sheet = f[1]
		case len(f) == 2 && f[0] == "others":
			q.Others, _ = strconv.Atoi(f[1])
		case len(f) == 9 && strings.Contains(" add replace keep differs dup ", " "+f[0]+" "):
			n, _ := strconv.ParseInt(f[4], 10, 64)
			c := change{kind: f[0], rel: f[1], sha: f[2], old: f[3], size: n, top: f[5], dup: f[6], origin: f[7], reason: f[8]}
			if !safeRel(c.rel) || (c.dup != "" && !safeRel(c.dup)) {
				return nil, fmt.Errorf("unsafe path %q in the request", c.rel)
			}
			if c.top != "" && (!safeRel(c.top) || !strings.HasPrefix(c.rel, c.top+"/")) {
				return nil, fmt.Errorf("unsafe folder %q in the request", c.top)
			}
			// What is copied must carry its checksum, or the review
			// would compare nothing with nothing.
			if (c.kind == kAdd || c.kind == kReplace) && !reSHA256.MatchString(c.sha) || c.kind == kReplace && !reSHA256.MatchString(c.old) {
				return nil, fmt.Errorf("%s: %s has no SHA-256 checksum in the request", name, c.rel)
			}
			q.Changes = append(q.Changes, c)
		case strings.TrimSpace(sc.Text()) == "":
		default:
			return nil, fmt.Errorf("%s: unreadable line %q", name, sc.Text())
		}
	}
	if q.From == "" || q.To == "" {
		return nil, fmt.Errorf("%s does not say where to copy from and to", name)
	}
	return q, sc.Err()
}
