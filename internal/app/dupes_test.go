package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompanyFolder(t *testing.T) {
	folders := []realFolder{
		{rel: "D10629 Foo S.A.", name: "D10629 Foo S.A."},
		{rel: "Archive/X20001 - Bar GmbH", name: "X20001 - Bar GmbH"},
		{rel: "Baz Ltda.", name: "Baz Ltda."},
	}
	for _, c := range []struct{ company, id, want string }{
		{"Foo S.A.", "D10629", "D10629 Foo S.A."},
		{"Foo S.A.", "", "D10629 Foo S.A."},
		{"Foo Sociedad", "d10629", "D10629 Foo S.A."}, // the id decides
		{"Bar GmbH", "X99999", "Archive/X20001 - Bar GmbH"},
		{"BAZ LTDA", "", "Baz Ltda."},
		{"Foo GmbH", "", ""},
		{"Qux S.A.", "D1062", ""},
	} {
		f, why, ok := companyFolder(folders, c.company, c.id)
		if f.rel != c.want || ok != (c.want != "") {
			t.Errorf("%s (%s): got %q (%s), want %q", c.company, c.id, f.rel, why, c.want)
		}
	}
}

// listFolders sees what numb-form sees: no overview folder, no hidden
// folders. Folder names end without a dot, as SafeName makes them:
// Windows drops a trailing dot.
func TestListFolders(t *testing.T) {
	_, real, _ := setup(t)
	for _, d := range []string{".hidden/Foo Corp", "D10629 Foo S.A/2025"} {
		if err := os.MkdirAll(filepath.Join(real, filepath.FromSlash(d)), 0o750); err != nil {
			t.Fatal(err)
		}
	}
	got := map[string]bool{}
	for _, f := range listFolders(real, filepath.Join(real, "3. Übersichten")) {
		got[f.rel] = true
		if f.name != filepath.Base(filepath.FromSlash(f.rel)) {
			t.Errorf("%s named %q", f.rel, f.name)
		}
	}
	for _, want := range []string{"Foo GmbH", "Archive", "Archive/foo-corp", "D10629 Foo S.A", "D10629 Foo S.A/2025"} {
		if !got[want] {
			t.Errorf("%s not listed: %v", want, got)
		}
	}
	for _, not := range []string{"3. Übersichten", ".hidden", ".hidden/Foo Corp"} {
		if got[not] {
			t.Errorf("%s listed", not)
		}
	}
}

// A new folder of a company the real folder already has under its
// customer id is not copied, and the review and the tree say so.
func TestSkipKnownCompanies(t *testing.T) {
	real, cp, rc, q := scanned(t)
	if err := os.Mkdir(filepath.Join(real, "D00000 Foo Ltd"), 0o750); err != nil {
		t.Fatal(err)
	}
	q.skipKnownCompanies(rc)
	foo := "Foo Ltd/Appendix A - Risk Assesment - Foo Ltd.docx"
	for _, c := range q.Changes {
		switch {
		case c.rel == foo && (c.kind != kDup || c.dup != "D00000 Foo Ltd"):
			t.Errorf("%s: %s %q", c.rel, c.kind, c.dup)
		case c.rel == foo && c.reason != `not copied: the real folder already has "D00000 Foo Ltd" for Foo Ltd (it starts with the customer id D00000)`:
			t.Errorf("reason %q", c.reason)
		case c.rel != foo && c.kind == kDup:
			t.Errorf("%s is a duplicate", c.rel)
		}
	}
	v := reviewRequest(q, rc)
	if ok, _ := has(v, "ok", `"Foo Ltd" is not copied: the real folder already has "D00000 Foo Ltd" for that company`); !ok || v.count("fail") != 0 {
		t.Errorf("review:\n%s", dump(v))
	}
	if !strings.Contains(q.tree(), `= unchanged; "Foo Ltd" of the working copy is the same company, not copied`) {
		t.Errorf("tree:\n%s", q.tree())
	}

	// More than numb-form's form in that folder asks for a look.
	if err := os.WriteFile(filepath.Join(cp, "Foo Ltd", "notes.txt"), []byte("notes"), 0o600); err != nil {
		t.Fatal(err)
	}
	q, err := scanRequest(context.Background(), cp, real, rc)
	if err != nil {
		t.Fatal(err)
	}
	q.skipKnownCompanies(rc)
	v = reviewRequest(q, rc)
	if ok, _ := has(v, "warn", "holds more than numb-form's form (notes.txt)"); !ok {
		t.Errorf("review:\n%s", dump(v))
	}
}
