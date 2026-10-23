package app

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/formats"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// finding is one result of reviewing a request.
type finding struct {
	level string // ok, warn, fail
	msg   string
}

type reqReview struct {
	findings []finding
}

func (v *reqReview) add(level, format string, a ...any) {
	v.findings = append(v.findings, finding{level, fmt.Sprintf(format, a...)})
}

func (v *reqReview) count(level string) int {
	n := 0
	for _, f := range v.findings {
		if f.level == level {
			n++
		}
	}
	return n
}

func hashOf(p string) string {
	h, _, err := fsutil.SHA256File(p)
	if err != nil {
		return ""
	}
	return h
}

var reReportDate = regexp.MustCompile(`^\d{2}-\d{2}-\d{4}$`)

// reviewRequest checks a request against both folders. Before it is
// applied: the working copy still holds exactly the listed files, the
// real folder does not have them yet, the replaced spreadsheet only gains
// report dates, every new company folder belongs to a company in the
// list and holds a filled form, and nothing that must stay behind is
// listed. After it is applied: every listed file is in the real folder
// with its checksum.
func reviewRequest(q *request, cfg reviewConfig) *reqReview {
	v := &reqReview{}
	applied := q.Status == "applied"
	for _, c := range q.Changes {
		if c.kind == "keep" {
			continue
		}
		if why := residue(filepath.FromSlash(c.rel), false); why != "" || strings.EqualFold(c.rel, cfg.template) {
			v.add("fail", "%s is listed for copying but must stay behind", c.rel)
		}
	}
	for _, c := range q.of("add") {
		src := filepath.Join(q.From, filepath.FromSlash(c.rel))
		dst := filepath.Join(q.To, filepath.FromSlash(c.rel))
		switch {
		case applied && hashOf(dst) == c.sha:
		case applied:
			v.add("fail", "%s is missing in the real folder or differs from the request", c.rel)
		case hashOf(src) != c.sha:
			v.add("fail", "%s changed after the request was made; run numb-commit again", c.rel)
		default:
			if _, err := os.Lstat(dst); !errors.Is(err, fs.ErrNotExist) {
				v.add("fail", "%s already exists in the real folder", c.rel)
			}
		}
	}
	for _, c := range q.of("replace") {
		src := filepath.Join(q.From, filepath.FromSlash(c.rel))
		dst := filepath.Join(q.To, filepath.FromSlash(c.rel))
		now := hashOf(dst)
		switch {
		case applied && now == c.sha:
		case applied:
			v.add("fail", "%s in the real folder is not the committed version", c.rel)
		case hashOf(src) != c.sha:
			v.add("fail", "%s changed after the request was made; run numb-commit again", c.rel)
		case now != c.old:
			v.add("fail", "%s in the real folder changed after the request was made (someone else edited it?)", c.rel)
		case strings.EqualFold(c.rel, q.Sheet):
			reviewSheet(v, dst, src, cfg)
		}
	}
	if !applied {
		reviewFolders(v, q, cfg)
	}
	if v.count("fail") == 0 {
		if applied {
			v.add("ok", "every listed file is in the real folder with its SHA-256 checksum")
		} else {
			v.add("ok", "%d file(s) to add and %d to replace match their SHA-256 checksums", len(q.of("add")), len(q.of("replace")))
		}
	}
	return v
}

// reviewConfig is what a review needs to know about the working folder.
type reviewConfig struct {
	header, idHeader string
	overviewDir      string
	template         string            // rel path of the template
	templateBase     string            // its name without extension
	names            map[string]string // folder name -> company, from the spreadsheet
}

// reviewSheet allows only report dates written into empty or digit-free
// cells of the report-date column.
func reviewSheet(v *reqReview, oldPath, newPath string, cfg reviewConfig) {
	ob, err1 := os.ReadFile(oldPath) // #nosec G304 -- the overview spreadsheet.
	nb, err2 := os.ReadFile(newPath) // #nosec G304 -- the overview spreadsheet.
	if err1 != nil || err2 != nil {
		v.add("fail", "the spreadsheet cannot be read")
		return
	}
	ext := strings.ToLower(filepath.Ext(newPath))
	if ext != ".xlsx" && ext != ".xlsm" {
		v.add("warn", "the spreadsheet changed; only .xlsx changes can be checked cell by cell")
		return
	}
	t, err := sheet.ReadTable(nb, cfg.header, cfg.idHeader)
	oc, err1 := sheet.AllCells(ob)
	nc, err2 := sheet.AllCells(nb)
	if err != nil || err1 != nil || err2 != nil {
		v.add("fail", "the spreadsheet cannot be compared cell by cell")
		return
	}
	dateCol := map[string]bool{}
	for _, r := range t.Rows {
		dateCol[t.Sheet+"!"+r.DateRef] = true
	}
	vals := func(m map[string][]sheet.Cell) map[string]string {
		out := map[string]string{}
		for name, cells := range m {
			for _, c := range cells {
				if c.Val != "" {
					out[name+"!"+c.Ref] = c.Val
				}
			}
		}
		return out
	}
	before, after := vals(oc), vals(nc)
	dates := 0
	for ref, nv := range after {
		ov := before[ref]
		if ov == nv {
			continue
		}
		switch {
		case !dateCol[ref]:
			v.add("fail", "spreadsheet cell %s changed from %q to %q; only report dates may change", ref, ov, nv)
		case hasDigit(ov):
			v.add("fail", "spreadsheet cell %s already had a date (%q) and was changed to %q", ref, ov, nv)
		case !reReportDate.MatchString(nv):
			v.add("warn", "spreadsheet cell %s holds %q, not a DD-MM-YYYY date", ref, nv)
		default:
			dates++
		}
	}
	for ref, ov := range before {
		if _, ok := after[ref]; !ok {
			v.add("fail", "spreadsheet cell %s (%q) was emptied", ref, ov)
		}
	}
	v.add("ok", "the spreadsheet only gains %d report date(s) in column %q", dates, t.DateHeader)
}

// reviewFolders checks every new top-level company folder.
func reviewFolders(v *reqReview, q *request, cfg reviewConfig) {
	groups, order := q.newFolders()
	var real *match.Index
	if ov, err := findOverview(q.To, cfg.overviewDir); err == nil {
		real, _ = indexFolders(q.To, ov)
	}
	for _, top := range order {
		if top == "" || strings.Contains(top, "/") {
			continue
		}
		company, known := cfg.names[strings.ToLower(top)]
		if !known {
			v.add("warn", "new folder %q matches no company in the Excel list", top)
			company = top
		}
		if real != nil {
			exact, fuzzy := real.Find(company)
			switch {
			case len(exact) > 0:
				v.add("warn", "the real folder already has %q for %s", exact[0].Rel, company)
			case len(fuzzy) > 0:
				v.add("warn", "the real folder has a similar folder %q for %s (%s)", fuzzy[0].Folder.Rel, company, fuzzy[0].Reason)
			}
		}
		var formRel string
		for _, c := range groups[top] {
			b := path.Base(c.rel)
			if strings.HasPrefix(strings.ToLower(b), strings.ToLower(cfg.templateBase)) && formats.IsTemplate(path.Ext(b)) {
				formRel = c.rel
			}
		}
		if formRel == "" {
			v.add("warn", "new folder %q has no risk assessment form", top)
			continue
		}
		b, err := os.ReadFile(filepath.Join(q.From, filepath.FromSlash(formRel))) // #nosec G304 -- a form in the working copy.
		if err != nil {
			v.add("fail", "%s cannot be read", formRel)
			continue
		}
		vals, err := formats.ReadFilled(b, path.Ext(formRel))
		if err != nil {
			v.add("warn", "%s: section A cannot be read (%v)", formRel, err)
			continue
		}
		var blank []string
		for _, f := range form.AllFields {
			if strings.TrimSpace(vals[f]) == "" {
				blank = append(blank, f.Title())
			}
		}
		switch {
		case known && textutil.OneLine(vals[form.FieldName]) != company:
			v.add("warn", "%s names %q, not %q", formRel, vals[form.FieldName], company)
		case len(blank) > 0:
			v.add("warn", "%s: blank %s", formRel, strings.Join(blank, ", "))
		default:
			v.add("ok", "%s: section A is filled for %s", top, company)
		}
	}
}
