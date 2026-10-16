package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/docx"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/lookup"
	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
	"github.com/z4dm42f5kp-ux/numb-form/internal/xlsx"
)

// company handles one spreadsheet row. A returned error stops the run;
// problems specific to this company are recorded in its status instead.
func (r *run) company(row xlsx.Row) (runlog.Company, error) {
	c := runlog.Company{Name: row.Name, Cell: row.Sheet + "!" + row.Cell}
	exact, fuzzy := r.ix.Find(row.Name)
	if len(exact) > 0 {
		c.Status, c.Folder, c.Reason = "exists", exact[0].Rel, "folder name matches"
		r.ui.Printf("  already has a folder: %s — nothing to do\n", ui.Safe(exact[0].Rel))
		r.logf("exists %q -> %s", row.Name, exact[0].Rel)
		return c, nil
	}
	for _, cand := range fuzzy {
		ok, err := r.ui.YesNo(fmt.Sprintf("  Is the existing folder %q this company's folder? (%s)", ui.Safe(cand.Folder.Rel), cand.Reason), false)
		if err != nil {
			return c, err
		}
		r.logf("fuzzy %q vs %s (%.2f): user said %v", row.Name, cand.Folder.Rel, cand.Score, ok)
		if ok {
			c.Status, c.Folder, c.Reason = "exists", cand.Folder.Rel, "confirmed by user: "+cand.Reason
			return c, nil
		}
	}

	folderName, changed := fsutil.SafeName(row.Name, false)
	if changed {
		r.ui.Printf("  note: folder will be named %q (characters not allowed on Windows were adjusted)\n", folderName)
		r.logf("folder name %q adjusted to %q", row.Name, folderName)
	}
	folder := filepath.Join(r.copyRoot, folderName)
	if _, err := os.Lstat(folder); err == nil {
		c.Status, c.Reason = "skipped", "a file or folder named "+folderName+" already exists"
		r.ui.Printf("  skipped: %s\n", c.Reason)
		return c, nil
	}
	fileBase, _ := fsutil.SafeName(row.Name, true)
	fileName := strings.TrimSuffix(r.tplName, filepath.Ext(r.tplName)) + " - " + fileBase + ".docx"

	vals, sources, err := r.research(row.Name)
	if err != nil {
		return c, err
	}

	if err := os.Mkdir(folder, 0o750); err != nil {
		c.Status, c.Reason = "error", err.Error()
		return c, nil
	}
	r.logf("created folder %s", folderName)
	accepted := false
	defer func() {
		if !accepted {
			if err := removeCreated(r.copyRoot, folder); err != nil {
				r.logf("ERROR could not remove %s: %v", folder, err)
			} else {
				r.logf("removed folder %s", folderName)
			}
		}
	}()
	target := filepath.Join(folder, fileName)
	for {
		if err := r.write(target, vals); err != nil {
			c.Status, c.Reason = "error", err.Error()
			r.ui.Printf("  error: %v\n", err)
			r.logf("ERROR %q: %v", row.Name, err)
			return c, nil
		}
		shown, err := r.show(target)
		if err != nil {
			c.Status, c.Reason = "error", err.Error()
			return c, nil
		}
		ch, err := r.ui.Choose("  Is this correct? [a]ccept, [e]dit a field, [r]eject: ", "aer", 0)
		if err != nil {
			return c, err
		}
		switch ch {
		case 'a':
			h, n, err := fsutil.SHA256File(target)
			if err != nil {
				return c, err
			}
			accepted = true
			c.Status, c.Folder = "accepted", folderName
			c.Files = []runlog.File{{Rel: filepath.ToSlash(filepath.Join(folderName, fileName)), SHA256: h, Size: n}}
			c.Fields = map[string]string{}
			for _, f := range docx.AllFields {
				c.Fields[f.Title()] = shown[f]
			}
			c.Sources = sources
			r.ix.Add(folder, folderName, folderName)
			r.logf("accepted %q -> %s sha256=%s fields=%q", row.Name, c.Files[0].Rel, h, c.Fields)
			return c, nil
		case 'r':
			c.Status, c.Reason = "rejected", "rejected by user"
			r.logf("rejected %q", row.Name)
			return c, nil
		case 'e':
			if err := r.editField(vals); err != nil {
				return c, err
			}
			if err := os.Remove(target); err != nil {
				return c, err
			}
		}
	}
}

// removeCreated deletes a folder this run created, after checking that it
// lies inside the Downloads copy.
func removeCreated(copyRoot, folder string) error {
	rr, err := filepath.Rel(copyRoot, folder)
	if err != nil || rr == "." || strings.HasPrefix(rr, "..") || filepath.IsAbs(rr) {
		return errors.New("refusing to remove a folder outside the copy")
	}
	return os.RemoveAll(folder)
}

func (r *run) write(target string, vals docx.Values) error {
	out, warns, err := docx.Fill(r.template, vals)
	if err != nil {
		return err
	}
	for _, w := range warns {
		r.logf("fill note: %s", w)
	}
	r.scratch += int64(len(out))
	if r.scratch > MaxScratchByte {
		return fmt.Errorf("this run would write more than %d MB of documents; stopping", MaxScratchByte>>20)
	}
	return fsutil.WriteNew(target, out)
}

// show re-reads the written file from disk and prints what is in it.
func (r *run) show(target string) (docx.Values, error) {
	b, err := os.ReadFile(target) // #nosec G304 -- file written by this run.
	if err != nil {
		return nil, err
	}
	got, err := docx.ReadBack(b)
	if err != nil {
		return nil, err
	}
	r.ui.Printf("\n  Written to %s\n  Section A now reads:\n", filepath.Base(target))
	for i, f := range docx.AllFields {
		v := got[f]
		if v == "" {
			v = "(blank)"
		}
		r.ui.Printf("    %d. %s: %s\n", i+1, f.Title(), ui.Safe(v))
	}
	r.ui.Println("  Section B and everything else is unchanged.")
	return got, nil
}

func (r *run) editField(vals docx.Values) error {
	for {
		a, err := r.ui.Ask("  Field number to change (1-7, Enter to go back): ")
		if err != nil {
			return err
		}
		if a == "" {
			return nil
		}
		n, err := strconv.Atoi(a)
		if err != nil || n < 1 || n > len(docx.AllFields) {
			continue
		}
		f := docx.AllFields[n-1]
		if f == docx.FieldRegistry {
			r.ui.Println("  The commercial registry number is always left blank.")
			continue
		}
		v, err := r.ui.Ask(fmt.Sprintf("  New value for %s (empty = blank): ", f.Title()))
		if err != nil {
			return err
		}
		vals[f] = textutil.OneLine(v)
		r.logf("user edited %s", f.Title())
		return nil
	}
}

// research collects the values for one company and lets the user decide
// on each of them.
func (r *run) research(name string) (docx.Values, []string, error) {
	vals := docx.Values{
		docx.FieldName:   name,
		docx.FieldDate:   r.date,
		docx.FieldPerson: r.person,
	}
	q := lookup.NewQuery(name)
	var res lookup.Result
	if !r.opt.Offline {
		r.ui.Println("  Looking up official sources…")
		res = lookup.Registries(r.ctx, r.opt.Client, r.opt.Lookup, q)
		if err := r.ctx.Err(); err != nil {
			return nil, nil, err
		}
		for _, n := range res.Notes {
			r.ui.Printf("  · %s\n", ui.Safe(n))
		}
		site, err := r.askWebsite(res.Websites)
		if err != nil {
			return nil, nil, err
		}
		if site != "" {
			r.ui.Println("  Reading the company's legal notice…")
			w := lookup.Website(r.ctx, r.opt.Client, site)
			for _, n := range w.Notes {
				r.ui.Printf("  · %s\n", ui.Safe(n))
			}
			res.Merge(w)
			for _, e := range w.Excerpts {
				r.ui.Printf("\n  Excerpt from %s:\n", ui.Safe(e.Source.URL))
				for _, l := range strings.Split(e.Text, "\n") {
					r.ui.Printf("    | %s\n", ui.Safe(l))
				}
			}
		}
		for _, s := range res.Sources {
			r.logf("source for %q: %s %s", name, s.Name, s.URL)
		}
	}
	var sources []string
	for _, s := range res.Sources {
		sources = append(sources, s.String()+" "+s.URL)
	}

	// Legal form: taken from the name, conventional casing.
	form, _ := match.LegalForm(name)
	r.ui.Println()
	if len(res.FormHints) > 0 {
		r.ui.Printf("  Register lists the legal form as: %s\n", ui.Safe(strings.Join(res.FormHints, "; ")))
	}
	lf, err := r.ui.AskDefault("  Legal form (Enter = keep, '-' = blank)", form)
	if err != nil {
		return nil, nil, err
	}
	if lf == "-" {
		lf = ""
	}
	vals[docx.FieldLegalForm] = textutil.OneLine(lf)

	addr, err := r.chooseAddress(res.Addresses)
	if err != nil {
		return nil, nil, err
	}
	vals[docx.FieldAddress] = addr

	reps, err := r.chooseReps(res.People)
	if err != nil {
		return nil, nil, err
	}
	vals[docx.FieldReps] = reps
	return vals, sources, nil
}

func (r *run) askWebsite(suggest []string) (string, error) {
	def := ""
	if len(suggest) > 0 {
		def = suggest[0]
		r.ui.Printf("  Website suggested by Wikidata: %s\n", ui.Safe(def))
	}
	for {
		p := "  Official company website to read its legal notice (Enter = suggestion, '-' = skip, or type an address): "
		if def == "" {
			p = "  Official company website to read its legal notice (type an address, Enter = skip): "
		}
		a, err := r.ui.Ask(p)
		if err != nil {
			return "", err
		}
		switch {
		case a == "-":
			return "", nil
		case a == "":
			return def, nil
		}
		if _, ok := lookup.NormalizeWebsite(a); ok {
			return a, nil
		}
		r.ui.Println("  That is not a usable web address.")
	}
}

func (r *run) chooseAddress(cands []lookup.Address) (string, error) {
	r.ui.Println("\n  Headquarters / Address candidates (most authoritative first):")
	if len(cands) > 9 {
		cands = cands[:9]
	}
	for i, a := range cands {
		r.ui.Printf("    %d) %s\n       source: %s\n", i+1, ui.Safe(a.Text), ui.Safe(a.Source.String()))
	}
	if len(cands) == 0 {
		r.ui.Println("    (none found)")
	}
	def := "m"
	if len(cands) > 0 {
		def = "1"
	}
	for {
		a, err := r.ui.AskDefault("  Choose a number, 'm' to type it, 'b' to leave blank", def)
		if err != nil {
			return "", err
		}
		switch a {
		case "b":
			return "", nil
		case "m":
			v, err := r.ui.Ask("  Address on one line: ")
			if err != nil {
				return "", err
			}
			v = textutil.OneLine(v)
			if v == "" {
				continue
			}
			return r.withCaution(v, false, "address")
		}
		n, err := strconv.Atoi(a)
		if err != nil || n < 1 || n > len(cands) {
			continue
		}
		c := cands[n-1]
		v, err := r.ui.AskDefault("  Address (Enter = keep, or correct it)", c.Text)
		if err != nil {
			return "", err
		}
		return r.withCaution(textutil.OneLine(v), c.Source.Kind > lookup.KindRegulatory, "address")
	}
}

func (r *run) withCaution(v string, def bool, what string) (string, error) {
	ok, err := r.ui.YesNo(fmt.Sprintf("  Add the caution text to the %s (not fully verified)?", what), def)
	if err != nil {
		return "", err
	}
	if ok {
		return v + " " + CautionText, nil
	}
	return v, nil
}

// FormatPeople renders "Jane Doe (General Manager, Legal Representative), John Roe (Director)".
func FormatPeople(ps []lookup.Person) string {
	var parts []string
	for _, p := range ps {
		roles := append(append([]string(nil), p.Roles...), p.Untranslated...)
		if len(roles) == 0 {
			parts = append(parts, p.Name)
			continue
		}
		parts = append(parts, p.Name+" ("+strings.Join(roles, ", ")+")")
	}
	return strings.Join(parts, ", ")
}

func (r *run) chooseReps(people []lookup.Person) (string, error) {
	r.ui.Println("\n  Authorized representative candidates (most authoritative first):")
	if len(people) > 20 {
		people = people[:20]
	}
	var verified []string
	for i, p := range people {
		state := "unverified"
		if p.Verified() {
			state = "verified"
			verified = append(verified, strconv.Itoa(i+1))
		}
		var src []string
		for _, s := range p.Sources {
			src = append(src, s.String())
		}
		r.ui.Printf("    %d) %s  [%s]\n       source: %s\n", i+1, ui.Safe(FormatPeople([]lookup.Person{p})), state, ui.Safe(strings.Join(src, "; ")))
		if len(p.Untranslated) > 0 {
			r.ui.Printf("       note: role not translated automatically: %s\n", ui.Safe(strings.Join(p.Untranslated, ", ")))
		}
	}
	if len(people) == 0 {
		r.ui.Println("    (none found)")
	}
	def := "n"
	switch {
	case len(verified) > 0:
		def = strings.Join(verified, ",")
	case len(people) > 0:
		def = "m"
	}
	for {
		a, err := r.ui.AskDefault("  Numbers to include (e.g. 1,3), 'a' all, 'm' type them, 'n' none found", def)
		if err != nil {
			return "", err
		}
		var line string
		unverified := false
		switch a {
		case "n":
			if len(people) > 0 {
				ok, err := r.ui.YesNo("  Candidates were found. Use the 'No representatives were found online' text anyway?", false)
				if err != nil {
					return "", err
				}
				if !ok {
					continue
				}
			}
			return NoRepsText, nil
		case "m":
			r.ui.Println("  Format: Full Name (Role 1, Role 2), Full Name (Role) — roles in English.")
			v, err := r.ui.Ask("  Representatives: ")
			if err != nil {
				return "", err
			}
			line = textutil.OneLine(v)
			if line == "" {
				continue
			}
			unverified = true
		default:
			var sel []lookup.Person
			if a == "a" {
				sel = people
			} else {
				bad := false
				for _, s := range strings.Split(a, ",") {
					n, err := strconv.Atoi(strings.TrimSpace(s))
					if err != nil || n < 1 || n > len(people) {
						bad = true
						break
					}
					sel = append(sel, people[n-1])
				}
				if bad || len(sel) == 0 {
					continue
				}
			}
			for _, p := range sel {
				if !p.Verified() {
					unverified = true
				}
			}
			line = FormatPeople(sel)
			v, err := r.ui.AskDefault("  Line to write (Enter = keep, or correct it)", line)
			if err != nil {
				return "", err
			}
			line = textutil.OneLine(v)
		}
		return r.withCaution(line, unverified, "representatives")
	}
}
