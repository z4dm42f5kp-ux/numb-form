package numbform

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
)

// The documentation follows the code (.github/DOCUMENTATION.md): these
// tests fail when a flag or a configuration key has no manual entry, or
// when a link in the English documentation points nowhere.

// english returns the English documentation: README, SECURITY, CHANGES
// and docs/*.md, by path.
func english(t *testing.T) map[string]string {
	t.Helper()
	files, err := filepath.Glob("docs/*.md")
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, f := range append(files, "README.md", "SECURITY.md", "CHANGES.md") {
		b, err := os.ReadFile(f) // #nosec G304 -- the repository's own files.
		if err != nil {
			t.Fatal(err)
		}
		out[f] = string(b)
	}
	return out
}

// flags returns the flags each program defines with fs.String, fs.Bool
// and the like, by program.
func flags(t *testing.T) map[string][]string {
	t.Helper()
	mains, err := filepath.Glob("cmd/*/main.go")
	if err != nil || len(mains) == 0 {
		t.Fatalf("no programs found: %v", err)
	}
	out := map[string][]string{}
	for _, m := range mains {
		f, err := parser.ParseFile(token.NewFileSet(), m, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if id, ok := sel.X.(*ast.Ident); !ok || id.Name != "fs" {
				return true
			}
			switch sel.Sel.Name {
			case "String", "Bool", "Duration", "Int", "Int64", "Uint", "Float64":
			default:
				return true
			}
			if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				if name, err := strconv.Unquote(lit.Value); err == nil {
					prog := filepath.Base(filepath.Dir(m))
					out[prog] = append(out[prog], name)
				}
			}
			return true
		})
	}
	return out
}

func TestFlagsDocumented(t *testing.T) {
	docs := english(t)
	all := strings.Join(func() []string {
		var s []string
		for _, d := range docs {
			s = append(s, d)
		}
		return s
	}(), "\n")
	n := 0
	for prog, names := range flags(t) {
		for _, name := range names {
			n++
			if !strings.Contains(all, "`-"+name) {
				t.Errorf("%s -%s is not documented: describe it in its manual page", prog, name)
			}
		}
	}
	if n == 0 {
		t.Fatal("found no flags; the test no longer reads cmd/*/main.go correctly")
	}
}

func TestKeysDocumented(t *testing.T) {
	b, err := os.ReadFile("docs/configuration.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range config.Keys() {
		if !strings.Contains(string(b), "\n`"+k+" =") {
			t.Errorf("configuration key %s has no entry (`%s = ...`) in docs/configuration.md", k, k)
		}
	}
}

var reLink = regexp.MustCompile(`\]\(([^)\s]+)\)`)

func TestLinksResolve(t *testing.T) {
	for path, text := range english(t) {
		for _, m := range reLink.FindAllStringSubmatch(text, -1) {
			target := m[1]
			if i := strings.IndexByte(target, '#'); i >= 0 {
				target = target[:i]
			}
			if target == "" || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			p := filepath.Join(filepath.Dir(path), filepath.FromSlash(target))
			if _, err := os.Stat(p); err != nil {
				t.Errorf("%s links to %s, which does not exist", path, m[1])
			}
		}
	}
}
