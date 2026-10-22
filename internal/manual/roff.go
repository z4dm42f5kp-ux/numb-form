package manual

import (
	"fmt"
	"io/fs"
	"strings"
	"time"
)

// roffEscape protects text for roff: backslashes, and lines that would
// start with a control character.
func roffEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\e`)
	if strings.HasPrefix(s, ".") || strings.HasPrefix(s, "'") {
		s = `\&` + s
	}
	return s
}

// roffText renders one document's Markdown as man(7) requests.
func roffText(md string, b *strings.Builder) {
	var para []string
	flush := func() {
		if len(para) > 0 {
			b.WriteString(".PP\n" + roffEscape(inline(strings.Join(para, " "))) + "\n")
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
			b.WriteString(".PP\n.RS 4\n.nf\n")
			for i++; i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```"); i++ {
				b.WriteString(roffEscape(strings.TrimSpace(lines[i])) + "\n")
			}
			b.WriteString(".fi\n.RE\n")
		case strings.HasPrefix(t, "# "):
			flush()
			b.WriteString(".SH " + roffEscape(strings.ToUpper(inline(t[2:]))) + "\n")
		case strings.HasPrefix(t, "## "):
			flush()
			b.WriteString(".SS " + roffEscape(inline(t[3:])) + "\n")
		case strings.HasPrefix(t, "### ") || strings.HasPrefix(t, "#### "):
			flush()
			b.WriteString(".PP\n.B " + roffEscape(inline(strings.TrimLeft(t, "# "))) + "\n")
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
			b.WriteString(".IP \\(bu 2\n" + roffEscape(strings.Join(cells, " | ")) + "\n")
		case strings.HasPrefix(t, ": "):
			text := t[2:]
			for i+1 < len(lines) && strings.HasPrefix(lines[i+1], "  ") && strings.TrimSpace(lines[i+1]) != "" {
				i++
				text += " " + strings.TrimSpace(lines[i])
			}
			b.WriteString(roffEscape(inline(text)) + "\n")
		case i+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i+1]), ": ") && para == nil:
			b.WriteString(".TP\n.B " + roffEscape(inline(t)) + "\n")
		case reList.MatchString(l):
			flush()
			m := reList.FindStringSubmatch(l)
			text := m[3]
			for i+1 < len(lines) && strings.HasPrefix(lines[i+1], strings.Repeat(" ", len(m[1])+2)) && !reList.MatchString(lines[i+1]) && strings.TrimSpace(lines[i+1]) != "" && !strings.HasPrefix(strings.TrimSpace(lines[i+1]), "```") {
				i++
				text += " " + strings.TrimSpace(lines[i])
			}
			mark := `\(bu`
			if m[2] != "*" && m[2] != "+" && m[2] != "-" {
				mark = m[2]
			}
			fmt.Fprintf(b, ".IP %s %d\n%s\n", mark, 4+len(m[1]), roffEscape(inline(text)))
		case reNote.MatchString(t):
			para = append(para, reNote.ReplaceAllString(t, ""))
		default:
			para = append(para, t)
		}
	}
	flush()
}

// Roff returns the whole English manual as a man(7) page for section 1.
func Roff(docs fs.FS, version string, now time.Time) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, ".TH NUMB-FORM 1 %q %q \"User Commands\"\n", now.Format("2006-01-02"), "numb-form "+version)
	b.WriteString(".SH NAME\nnumb-form, numb-commit \\- fill the risk assessment form for every new company\n")
	b.WriteString(".SH SYNOPSIS\n.nf\nnumb-form [-f conf] [-offline] [-root dir] [-downloads dir] [-timeout d]\n" +
		"numb-form [-f conf] check-template template [sample]\nnumb-form [-f conf] config [init | set key value]\n" +
		"numb-form [-f conf] update [stable | rolling | v1.2.3]\nnumb-form rollback\nnumb-form man [topic] [en | es | de]\n" +
		"numb-commit [-f conf] [-run dir] [-downloads dir]\n.fi\n")
	for _, t := range Topics {
		md, err := Source(docs, t, "en")
		if err != nil {
			return "", err
		}
		roffText(md, &b)
	}
	return b.String(), nil
}
