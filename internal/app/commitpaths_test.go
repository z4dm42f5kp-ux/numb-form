package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

func TestCleanPath(t *testing.T) {
	for in, want := range map[string]string{
		`"C:\Users\Foo\14. GWG"`: `C:\Users\Foo\14. GWG`,
		`  'F:\14. GWG'  `:       `F:\14. GWG`,
		`" /srv/a b "`:           `/srv/a b`,
		`"`:                      `"`,
		`/srv/a"b`:               `/srv/a"b`,
	} {
		if got := cleanPath(in); got != want {
			t.Errorf("cleanPath(%q) = %q, want %q", in, got, want)
		}
	}
	if firstOf("", "b", "c") != "b" || firstOf("", "") != "" {
		t.Error("firstOf")
	}
}

// The prompts ask again until both answers are different working folders
// that are not inside each other.
func TestPickFolders(t *testing.T) {
	_, real, dl := setup(t)
	cp := filepath.Join(dl, filepath.Base(real))
	if _, err := fsutil.CopyTree(context.Background(), real, cp, func(string, ...any) {}); err != nil {
		t.Fatal(err)
	}
	cfg := config.Defaults()
	answers := strings.Join([]string{
		"relative/folder",
		`"` + filepath.Join(dl, "missing") + `"`,
		dl,
		cp,
		cp,
		filepath.Join(cp, "Foo GmbH"),
		"'" + real + "'",
	}, "\n") + "\n"
	var out bytes.Buffer
	u := ui.New(context.Background(), strings.NewReader(answers), &out)
	from, to, err := pickFolders(u, "", "", cfg)
	if err != nil || from != cp || to != real {
		t.Fatalf("got %q %q %v\n%s", from, to, err, out.String())
	}
	for _, want := range []string{askFrom, askTo, "Give the whole path", "is not reachable", "is not a working folder",
		"the target is the folder to commit", "the two folders must not be inside each other"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}

	// Folders given on the command line or in the configuration are not
	// asked again: a wrong one is an error.
	u = ui.New(context.Background(), strings.NewReader(""), &out)
	if _, _, err := pickFolders(u, cp, cp, cfg); err == nil || !strings.Contains(err.Error(), "the target is the folder to commit") {
		t.Errorf("same folder given twice: %v", err)
	}
	if _, _, err := pickFolders(u, dl, real, cfg); err == nil || !strings.Contains(err.Error(), "is not a working folder") {
		t.Errorf("given folder is not a working folder: %v", err)
	}
	if from, to, err := pickFolders(u, `"`+cp+`"`, real, cfg); err != nil || from != cp || to != real {
		t.Errorf("given folders: %q %q %v", from, to, err)
	}
	// Without answers, asking ends with an error instead of looping.
	if _, _, err := pickFolders(u, "", "", cfg); err == nil {
		t.Error("no error at the end of the input")
	}
}
