package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestDefaultsValid(t *testing.T) {
	c := Defaults()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.Mode != "direct" || !c.Online || c.LegacyFormats || c.PagesWrite {
		t.Fatalf("unexpected defaults %+v", c)
	}
}

func TestApply(t *testing.T) {
	c := Defaults()
	notes, err := c.Apply([]byte("\ufeff# comment\nmode = review # inline\nperson = \"Max Muster\"\ntimeout = 10m\nlegacy_formats = yes\n"), "user.conf", false)
	if err != nil {
		t.Fatal(err)
	}
	if c.Mode != "review" || c.Person != "Max Muster" || c.Timeout != 10*time.Minute {
		t.Fatalf("got %+v", c)
	}
	if c.LegacyFormats || len(notes) != 1 {
		t.Fatalf("legacy_formats must need the system file: %v %v", c.LegacyFormats, notes)
	}
	if _, err := c.Apply([]byte("legacy_formats = yes\n"), "sys", true); err != nil || !c.LegacyFormats {
		t.Fatal("system file could not enable legacy formats")
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}

// Keys of the removed online lookup are skipped with a note, so older
// files keep working.
func TestRetiredKeys(t *testing.T) {
	c := Defaults()
	notes, err := c.Apply([]byte("confirm = none\nreps_sources = gleif\nsearch = duckduckgo\nbrowser = none\ncaution = always\n"), "old.conf", false)
	if err != nil || len(notes) != 5 {
		t.Fatalf("%v %q", err, notes)
	}
}

func TestErrors(t *testing.T) {
	for _, in := range []string{"nokey\n", "colour = red\n", "mode = maybe\n", "timeout = soon\n", "online = maybe\n"} {
		c := Defaults()
		if _, err := c.Apply([]byte(in), "x", false); err == nil {
			t.Errorf("%q accepted", in)
		}
	}
	c := Defaults()
	c.Timeout = time.Hour
	if c.Validate() == nil {
		t.Error("timeout over 25m accepted")
	}
}

func TestTemplateParses(t *testing.T) {
	// Uncommenting every line of the template must give the defaults.
	var b strings.Builder
	for _, l := range strings.Split(Template(), "\n") {
		if strings.HasPrefix(l, "#") && strings.Contains(l, " = ") && !strings.HasPrefix(l, "# ") {
			b.WriteString(strings.TrimPrefix(l, "#") + "\n")
		}
	}
	c := Defaults()
	if _, err := c.Apply([]byte(b.String()), "template", true); err != nil {
		t.Fatal(err)
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(c.Origin) != len(Keys()) {
		t.Fatalf("template covers %d of %d keys", len(c.Origin), len(Keys()))
	}
}

func TestDateLayout(t *testing.T) {
	if l, err := GoDateLayout("DD/MM/YYYY"); err != nil || l != "02/01/2006" {
		t.Fatal(l, err)
	}
	if _, err := GoDateLayout("MM-YYYY"); err == nil {
		t.Fatal("missing day accepted")
	}
}

func FuzzApply(f *testing.F) {
	f.Add([]byte(Template()))
	f.Add([]byte("mode = review\nperson = P\n"))
	f.Fuzz(func(t *testing.T, b []byte) {
		c := Defaults()
		_, _ = c.Apply(b, "fuzz", false)
		_ = c.Validate()
	})
}

func TestSetUserValue(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", d)
	t.Setenv("HOME", d)
	t.Setenv("APPDATA", d)
	if _, err := SetUserValue("person", "Erika Mustermann"); err != nil {
		t.Fatal(err)
	}
	if _, err := SetUserValue("mode", "review"); err != nil {
		t.Fatal(err)
	}
	p, err := SetUserValue("person", "Max Muster")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if strings.Count(string(b), "person =") != 1 || !strings.Contains(string(b), "person = Max Muster") || !strings.Contains(string(b), "mode = review") {
		t.Fatalf("file:\n%s", b)
	}
	c, _, err := Load("")
	if err != nil || c.Person != "Max Muster" || c.Mode != "review" {
		t.Fatalf("load %+v %v", c, err)
	}
	for _, bad := range [][2]string{{"mode", "maybe"}, {"colour", "red"}, {"legacy_formats", "yes"}, {"confirm", "none"}} {
		if _, err := SetUserValue(bad[0], bad[1]); err == nil {
			t.Errorf("%v accepted", bad)
		}
	}
}
