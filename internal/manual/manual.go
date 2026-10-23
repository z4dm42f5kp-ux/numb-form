// Package manual shows the built-in documentation: as text in a simple
// pager (numb-form man), or as a Unix manual page (man numb-form).
package manual

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Topic is one document of the manual.
type Topic struct {
	Name  string // what the user types: guide, usage, ...
	File  string // English file in the documentation tree
	Title string
}

// Topics in manual order: the step-by-step guide first.
var Topics = []Topic{
	{"guide", "docs/guide.md", "Step-by-step guide"},
	{"usage", "docs/usage.md", "Reference manual"},
	{"numb-date", "docs/numb-date.md", "numb-date: fill the report dates"},
	{"configuration", "docs/configuration.md", "Settings (numb-form.conf)"},
	{"install", "docs/install.md", "Installing and updating"},
	{"ai-prompt", "docs/ai-prompt.md", "The AI prompt for the address and representatives"},
	{"formats", "docs/formats.md", "File formats"},
	{"containers", "docs/containers.md", "Docker, Podman and Kubernetes"},
	{"security", "SECURITY.md", "Security"},
	{"changes", "CHANGES.md", "Changes in each release"},
}

var aliases = map[string]string{"date": "numb-date", "config": "configuration", "settings": "configuration", "conf": "configuration", "update": "install", "docker": "containers", "kubernetes": "containers", "podman": "containers", "help": "guide", "start": "guide", "prompt": "ai-prompt", "ai": "ai-prompt"}

// Find returns the topic called name.
func Find(name string) (Topic, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if a, ok := aliases[name]; ok {
		name = a
	}
	for _, t := range Topics {
		if t.Name == name {
			return t, true
		}
	}
	return Topic{}, false
}

// Lang picks "es" or "de" from the LANG/LC_ALL/LANGUAGE settings, or "en".
func Lang() string {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"} {
		v := strings.ToLower(os.Getenv(k))
		switch {
		case strings.HasPrefix(v, "es"):
			return "es"
		case strings.HasPrefix(v, "de"):
			return "de"
		case v != "" && v != "c" && v != "posix":
			return "en"
		}
	}
	return "en"
}

// Source returns the Markdown of a topic in a language, falling back to
// English when there is no translation.
func Source(docs fs.FS, t Topic, lang string) (string, error) {
	if lang == "es" || lang == "de" {
		p := "docs/" + lang + "/" + path.Base(t.File)
		if b, err := fs.ReadFile(docs, p); err == nil {
			return string(b), nil
		}
	}
	b, err := fs.ReadFile(docs, t.File)
	return string(b), err
}

var (
	reLink   = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)
	reCode   = regexp.MustCompile("`([^`]*)`")
	reStrong = regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	reEm     = regexp.MustCompile(`(^|[\s(])\*([^*\s][^*]*)\*`)
	reList   = regexp.MustCompile(`^(\s*)([*+-]|\d+\.)\s+(.*)$`)
	reNote   = regexp.MustCompile(`^>\s?`)
)

// inline removes Markdown inline markup. Links to other documents name
// the topic to open; web links keep their address.
func inline(s string) string {
	s = reLink.ReplaceAllStringFunc(s, func(m string) string {
		sm := reLink.FindStringSubmatch(m)
		text, target := sm[1], sm[2]
		if strings.HasPrefix(target, "http") {
			if strings.Contains(text, target) {
				return text
			}
			return text + " <" + target + ">"
		}
		if strings.HasPrefix(target, "#") {
			return text
		}
		base := strings.TrimSuffix(path.Base(strings.SplitN(target, "#", 2)[0]), ".md")
		if t, ok := Find(strings.ToLower(base)); ok {
			return text + " (numb-form man " + t.Name + ")"
		}
		return text
	})
	s = reStrong.ReplaceAllString(s, "$1$2")
	s = reEm.ReplaceAllString(s, "$1$2")
	s = reCode.ReplaceAllString(s, "$1")
	return s
}

func wrap(text string, width int, first, rest string) []string {
	var out []string
	line := first
	empty := true
	for _, w := range strings.Fields(text) {
		if !empty && utf8.RuneCountInString(line)+1+utf8.RuneCountInString(w) > width {
			out = append(out, line)
			line, empty = rest, true
		}
		if empty {
			line += w
			empty = false
		} else {
			line += " " + w
		}
	}
	if !empty {
		out = append(out, line)
	}
	return out
}

// Text renders Markdown as plain text for a terminal of the given width.
func Text(md string, width int) string {
	var out []string
	var para []string
	blank := func() {
		if len(out) > 0 && out[len(out)-1] != "" {
			out = append(out, "")
		}
	}
	flush := func() {
		if len(para) > 0 {
			out = append(out, wrap(inline(strings.Join(para, " ")), width, "    ", "    ")...)
			out = append(out, "")
			para = nil
		}
	}
	lines := strings.Split(strings.ReplaceAll(md, "\r\n", "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		t := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(t, "```"):
			flush()
			blank()
			for i++; i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```"); i++ {
				out = append(out, "        "+strings.TrimSpace(lines[i]))
			}
			out = append(out, "")
		case strings.HasPrefix(t, "# "):
			flush()
			out = append(out, strings.ToUpper(inline(t[2:])), "")
		case strings.HasPrefix(t, "## "):
			flush()
			out = append(out, strings.ToUpper(inline(t[3:])), "")
		case strings.HasPrefix(t, "### ") || strings.HasPrefix(t, "#### "):
			flush()
			out = append(out, "  "+inline(strings.TrimLeft(t, "# ")), "")
		case t == "":
			flush()
		case strings.HasPrefix(t, "|"):
			flush()
			if strings.Trim(t, "|-: ") == "" {
				continue
			}
			cells := strings.Split(strings.Trim(t, "|"), "|")
			for j := range cells {
				cells[j] = inline(strings.TrimSpace(cells[j]))
			}
			out = append(out, wrap(strings.Join(cells, "  |  "), width, "    ", "        ")...)
		case strings.HasPrefix(t, ": "):
			// Definition: "term" on the line before, ": text" here.
			flush()
			text := t[2:]
			for i+1 < len(lines) && strings.HasPrefix(lines[i+1], "  ") && strings.TrimSpace(lines[i+1]) != "" {
				i++
				text += " " + strings.TrimSpace(lines[i])
			}
			out = append(out, wrap(inline(text), width, "        ", "        ")...)
			out = append(out, "")
		case i+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i+1]), ": ") && para == nil:
			out = append(out, "    "+inline(t))
		case reList.MatchString(l):
			flush()
			m := reList.FindStringSubmatch(l)
			indent := strings.Repeat(" ", 4+len(m[1]))
			bullet := m[2]
			if bullet == "*" || bullet == "+" || bullet == "-" {
				bullet = "•"
			}
			text := m[3]
			for i+1 < len(lines) && strings.HasPrefix(lines[i+1], strings.Repeat(" ", len(m[1])+2)) && !reList.MatchString(lines[i+1]) && strings.TrimSpace(lines[i+1]) != "" && !strings.HasPrefix(strings.TrimSpace(lines[i+1]), "```") {
				i++
				text += " " + strings.TrimSpace(lines[i])
			}
			pad := strings.Repeat(" ", utf8.RuneCountInString(bullet)+1)
			out = append(out, wrap(inline(text), width, indent+bullet+" ", indent+pad)...)
		case reNote.MatchString(t):
			para = append(para, reNote.ReplaceAllString(t, ""))
		default:
			if para == nil {
				blank()
			}
			para = append(para, t)
		}
	}
	flush()
	return strings.Join(out, "\n") + "\n"
}

// Full returns the whole manual as text: every topic, guide first.
func Full(docs fs.FS, lang string, width int) (string, error) {
	var b strings.Builder
	b.WriteString("NUMB-FORM(1)\n\n")
	b.WriteString(wrapPlain("Topics: "+names()+". Show one with: numb-form man <topic>. Languages: add es (Español) or de (Deutsch).", width))
	b.WriteString("\n")
	for _, t := range Topics {
		md, err := Source(docs, t, lang)
		if err != nil {
			return "", err
		}
		b.WriteString(strings.Repeat("=", min(width, 72)) + "\n")
		b.WriteString(Text(md, width))
	}
	return b.String(), nil
}

func names() string {
	var n []string
	for _, t := range Topics {
		n = append(n, t.Name)
	}
	return strings.Join(n, ", ")
}

func wrapPlain(s string, width int) string {
	return strings.Join(wrap(s, width, "    ", "    "), "\n") + "\n"
}

// TopicList describes the topics.
func TopicList() string {
	var b strings.Builder
	b.WriteString("numb-form man [topic] [en | es | de]\n\nTopics:\n")
	for _, t := range Topics {
		fmt.Fprintf(&b, "  %-14s %s\n", t.Name, t.Title)
	}
	b.WriteString("\nWithout a topic, the whole manual is shown, starting with the guide.\n")
	return b.String()
}

// ErrQuit is returned by Page when the reader quits early.
var ErrQuit = errors.New("quit")

// Page writes text one screen at a time. After each screen it waits for
// Enter (next screen) or q and Enter (quit). It works the same in every
// terminal, including Windows consoles, and keeps accents intact.
func Page(w io.Writer, r io.Reader, text string, height int) error {
	if height < 5 {
		height = 24
	}
	in := bufio.NewReader(r)
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for i := 0; i < len(lines); i += height - 1 {
		end := min(len(lines), i+height-1)
		if _, err := io.WriteString(w, strings.Join(lines[i:end], "\n")+"\n"); err != nil {
			return err
		}
		if end == len(lines) {
			return nil
		}
		fmt.Fprintf(w, "-- %d%% -- Enter: more   q + Enter: quit ", end*100/len(lines))
		a, err := in.ReadString('\n')
		if err != nil || strings.HasPrefix(strings.TrimSpace(strings.ToLower(a)), "q") {
			return ErrQuit
		}
	}
	return nil
}
