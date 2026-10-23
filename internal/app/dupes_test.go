package app

import (
	"os"
	"path/filepath"
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
// folders.
func TestListFolders(t *testing.T) {
	_, real, _ := setup(t)
	for _, d := range []string{".hidden/Foo Corp", "D10629 Foo S.A./2025"} {
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
	for _, want := range []string{"Foo GmbH", "Archive", "Archive/foo-corp", "D10629 Foo S.A.", "D10629 Foo S.A./2025"} {
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
