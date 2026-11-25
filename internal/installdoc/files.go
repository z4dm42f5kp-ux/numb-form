package installdoc

import (
	"os"
	"path/filepath"
	"strings"
)

// Docs are the documents with install commands, relative to the
// repository root.
var Docs = []string{"docs/install.md", "docs/guide.md", "docs/containers.md"}

// Dir holds pins.txt and templates/, relative to the repository root.
const Dir = "scripts/install"

// Load reads the pins and the templates under root.
func Load(root string) (Pins, map[string]string, error) {
	b, err := os.ReadFile(filepath.Join(root, Dir, "pins.txt")) // #nosec G304 -- the repository's own file.
	if err != nil {
		return Pins{}, nil, err
	}
	p, err := ParsePins(b)
	if err != nil {
		return p, nil, err
	}
	files, err := filepath.Glob(filepath.Join(root, Dir, "templates", "*"))
	if err != nil {
		return p, nil, err
	}
	t := map[string]string{}
	for _, f := range files {
		b, err := os.ReadFile(f) // #nosec G304 -- the repository's own templates.
		if err != nil {
			return p, nil, err
		}
		t[filepath.Base(f)] = strings.ReplaceAll(string(b), "\r\n", "\n")
	}
	return p, t, nil
}

// Stale returns the documents under root whose install commands differ
// from what the pins and templates give; with write, it rewrites them.
func Stale(root string, write bool) ([]string, error) {
	p, t, err := Load(root)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, d := range Docs {
		path := filepath.Join(root, filepath.FromSlash(d))
		b, err := os.ReadFile(path) // #nosec G304 -- the repository's own documents.
		if err != nil {
			return nil, err
		}
		got, err := Apply(string(b), t, p)
		if err != nil {
			return nil, err
		}
		if got == string(b) {
			continue
		}
		out = append(out, d)
		if write {
			if err := os.WriteFile(path, []byte(got), 0o644); err != nil { // #nosec G306 -- documentation.
				return nil, err
			}
		}
	}
	return out, nil
}
