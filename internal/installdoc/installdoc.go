// Package installdoc writes the install commands of the documentation
// from templates and the pinned release in scripts/install, so that every
// stable install command names one release and its SHA-256, and a test can
// tell when the documentation and the pins disagree.
package installdoc

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Pins is the release the documentation installs: its tag, the programs
// in its archives and the SHA-256 of every file, from its SHA256SUMS.
type Pins struct {
	Version  string
	Programs []string
	Sums     map[string]string
}

var (
	reVersion = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	reSum     = regexp.MustCompile(`^[0-9a-f]{64}$`)
	reAsset   = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	reProgram = regexp.MustCompile(`^numb-[a-z]+$`)
)

// ParsePins reads pins.txt: "version vX.Y.Z", "programs a b c" and the
// lines of SHA256SUMS ("<sha256>  <file>").
func ParsePins(b []byte) (Pins, error) {
	p := Pins{Sums: map[string]string{}}
	sc := bufio.NewScanner(bytes.NewReader(b))
	for n := 1; sc.Scan(); n++ {
		f := strings.Fields(sc.Text())
		switch {
		case len(f) == 0 || strings.HasPrefix(f[0], "#"):
		case f[0] == "version" && len(f) == 2 && reVersion.MatchString(f[1]):
			p.Version = f[1]
		case f[0] == "programs" && len(f) > 1:
			for _, prog := range f[1:] {
				if !reProgram.MatchString(prog) {
					return p, fmt.Errorf("line %d: bad program %q", n, prog)
				}
			}
			p.Programs = f[1:]
		case len(f) == 2 && reSum.MatchString(f[0]) && reAsset.MatchString(f[1]):
			p.Sums[f[1]] = f[0]
		default:
			return p, fmt.Errorf("line %d: unreadable: %q", n, sc.Text())
		}
	}
	if p.Version == "" || len(p.Programs) == 0 || len(p.Sums) == 0 {
		return p, fmt.Errorf("pins need a version, programs and checksums")
	}
	return p, sc.Err()
}

// Format writes p in the form ParsePins reads.
func (p Pins) Format() []byte {
	var b bytes.Buffer
	b.WriteString("# The release that the install commands of the documentation install.\n")
	b.WriteString("# Written by .github/workflows/install-docs.yml; do not edit by hand.\n")
	fmt.Fprintf(&b, "version %s\nprograms %s\n", p.Version, strings.Join(p.Programs, " "))
	for _, a := range p.assets() {
		fmt.Fprintf(&b, "%s  %s\n", p.Sums[a], a)
	}
	return b.Bytes()
}

func (p Pins) assets() []string {
	var out []string
	for a := range p.Sums {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}

var rePlaceholder = regexp.MustCompile(`@[A-Z_]+(:[A-Za-z0-9._-]+)?@`)

// Render fills a template's placeholders: @VERSION@, @VERSION_NUM@
// (without the v), @PROGRAMS@ (space
// separated), @PROGRAMS_COMMA@ (comma separated), @EXES_COMMA@ (with
// .exe), @SHA:<file>@ and @SUMS_TABLE@ (a Markdown table body). An unknown
// placeholder or a file without a checksum is an error.
func Render(tmpl string, p Pins) (string, error) {
	var err error
	out := rePlaceholder.ReplaceAllStringFunc(tmpl, func(m string) string {
		name := strings.Trim(m, "@")
		switch {
		case name == "VERSION":
			return p.Version
		case name == "VERSION_NUM":
			return strings.TrimPrefix(p.Version, "v")
		case name == "PROGRAMS":
			return strings.Join(p.Programs, " ")
		case name == "PROGRAMS_COMMA":
			return strings.Join(p.Programs, ", ")
		case name == "EXES_COMMA":
			return strings.Join(p.Programs, ".exe, ") + ".exe"
		case name == "SUMS_TABLE":
			var rows []string
			for _, a := range p.assets() {
				rows = append(rows, "| `"+a+"` | `"+p.Sums[a]+"` |")
			}
			return strings.Join(rows, "\n")
		case strings.HasPrefix(name, "SHA:"):
			if s, ok := p.Sums[name[4:]]; ok {
				return s
			}
			err = fmt.Errorf("%s: no checksum for %s in the pins", m, name[4:])
		default:
			err = fmt.Errorf("unknown placeholder %s", m)
		}
		return m
	})
	return out, err
}

var reInclude = regexp.MustCompile(`@INCLUDE:[a-z0-9-]+@`)

var reMarker = regexp.MustCompile(`(?m)^([ \t]*)<!-- install:([a-z0-9-]+) -->\n`)

// Apply replaces every block between "<!-- install:NAME -->" and
// "<!-- /install:NAME -->" in doc with the rendered template NAME, keeping
// the marker's indentation. Templates ending in ".md" are inserted as
// they are; the others become a code block. "@INCLUDE:NAME@" in a
// template is replaced by the template NAME.cmd.
func Apply(doc string, templates map[string]string, p Pins) (string, error) {
	var b strings.Builder
	for {
		m := reMarker.FindStringSubmatchIndex(doc)
		if m == nil {
			b.WriteString(doc)
			return b.String(), nil
		}
		indent, name := doc[m[2]:m[3]], doc[m[4]:m[5]]
		end := indent + "<!-- /install:" + name + " -->"
		rest := doc[m[1]:]
		e := strings.Index(rest, end)
		if e < 0 {
			return "", fmt.Errorf("install:%s has no end marker", name)
		}
		body, md := templates[name+".md"], true
		if body == "" {
			body, md = templates[name+".cmd"], false
		}
		if body == "" {
			return "", fmt.Errorf("no template for install:%s", name)
		}
		body = reInclude.ReplaceAllStringFunc(body, func(m string) string {
			return strings.TrimRight(templates[strings.Trim(m, "@")[len("INCLUDE:"):]+".cmd"], "\n")
		})
		r, err := Render(strings.TrimRight(body, "\n"), p)
		if err != nil {
			return "", fmt.Errorf("install:%s: %w", name, err)
		}
		b.WriteString(doc[:m[1]])
		if !md {
			r = "```\n" + r + "\n```"
		}
		for _, l := range strings.Split(r, "\n") {
			if l == "" {
				b.WriteString("\n")
				continue
			}
			b.WriteString(indent + l + "\n")
		}
		doc = rest[e:]
	}
}
