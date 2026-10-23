package manual

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Keys of the pager, as in less(1).
const lessHelp = `Moving
  Space  f  PageDown        next page        b  PageUp          previous page
  Enter  j  e  Down         next line        k  y  Up           previous line
  d                         half page down   u                  half page up
  g  <  Home                first line       G  >  End          last line
Searching
  /text Enter               search forward   n  N               next, previous match
Leaving
  q  Q  Ctrl+C              quit             h                  this help (any key returns)`

// pager is the state of a less-like view of lines in a window of rows
// lines (the last one is the status line).
type pager struct {
	lines   []string
	rows    int
	top     int
	pattern string // last search, lower case
	status  string // a message instead of the position, once
	help    bool
	width   int // characters per row, one less than the window
}

func (p *pager) page() int { return max(1, p.rows-1) }

func (p *pager) scroll(n int) {
	p.top = max(0, min(p.top+n, len(p.lines)-p.page()))
}

// search moves to the next (dir 1) or previous (dir -1) line containing
// the pattern, starting after the current top line.
func (p *pager) search(dir int) {
	if p.pattern == "" {
		p.status = "no previous search"
		return
	}
	for i := p.top + dir; i >= 0 && i < len(p.lines); i += dir {
		if strings.Contains(strings.ToLower(p.lines[i]), p.pattern) {
			p.top = i
			p.scroll(0)
			return
		}
	}
	p.status = "pattern not found: " + p.pattern
}

// key applies one key and reports whether the pager should quit.
func (p *pager) key(k string) bool {
	if p.help {
		p.help = false
		return false
	}
	switch k {
	case "q", "Q", "\x03":
		return true
	case " ", "f", "F", "\x06", "\x1b[6~":
		p.scroll(p.page())
	case "b", "B", "\x02", "\x1b[5~":
		p.scroll(-p.page())
	case "\r", "\n", "j", "J", "e", "E", "\x0e", "\x1b[B", "\x1bOB":
		p.scroll(1)
	case "k", "K", "y", "Y", "\x10", "\x1b[A", "\x1bOA":
		p.scroll(-1)
	case "d", "D", "\x04":
		p.scroll(p.page() / 2)
	case "u", "U", "\x15":
		p.scroll(-p.page() / 2)
	case "g", "<", "\x1b[H", "\x1b[1~", "\x1bOH":
		p.top = 0
	case "G", ">", "\x1b[F", "\x1b[4~", "\x1bOF":
		p.scroll(len(p.lines))
	case "n":
		p.search(1)
	case "N":
		p.search(-1)
	case "h", "H":
		p.help = true
	}
	return false
}

// fit cuts a line to the window width, in characters.
func fit(s string, width int) string {
	if r := []rune(s); len(r) > width {
		return string(r[:width])
	}
	return s
}

// softWrap breaks lines longer than width into several, so that every
// line takes exactly one row of the window; long commands in code blocks
// would otherwise wrap by themselves and push the window out of place.
func softWrap(lines []string, width int) []string {
	var out []string
	for _, l := range lines {
		// A tab would take up to eight columns.
		r := []rune(strings.ReplaceAll(l, "\t", "    "))
		for len(r) > width {
			out = append(out, string(r[:width]))
			r = r[width:]
		}
		out = append(out, string(r))
	}
	return out
}

// draw writes the window: the visible lines and the status line. It
// overwrites the rows in place instead of clearing the screen, so the
// window does not flicker, and hides the cursor while it does.
func (p *pager) draw(w io.Writer) {
	var b strings.Builder
	b.WriteString("\x1b[?25l\x1b[H")
	body := p.lines[p.top:min(len(p.lines), p.top+p.page())]
	if p.help {
		body = strings.Split(lessHelp, "\n")
	}
	for _, l := range body {
		b.WriteString(fit(l, p.width) + "\x1b[K\r\n")
	}
	for i := len(body); i < p.page(); i++ {
		b.WriteString("~\x1b[K\r\n")
	}
	status := p.status
	if status == "" {
		end := min(len(p.lines), p.top+p.page())
		pos := "END"
		if end < len(p.lines) {
			pos = fmt.Sprintf("%d%%", end*100/len(p.lines))
		}
		status = fmt.Sprintf("lines %d-%d of %d  %s  (h: help, q: quit)", p.top+1, end, len(p.lines), pos)
	}
	p.status = ""
	b.WriteString("\x1b[7m" + fit(status, p.width) + "\x1b[0m\x1b[K\x1b[?25h")
	_, _ = io.WriteString(w, b.String())
}

// readKey reads one key: a byte, or an escape sequence such as the
// arrow and page keys ("\x1b[A", "\x1b[6~").
func readKey(r *bufio.Reader) (string, error) {
	c, err := r.ReadByte()
	if err != nil || c != 0x1b {
		// One byte, not the character with that number: bytes of an
		// accented letter arrive one by one.
		return string([]byte{c}), err
	}
	n, err := r.ReadByte()
	if err != nil {
		return "\x1b", err
	}
	if n != '[' && n != 'O' {
		return string(n), nil // Esc and then a key: the key
	}
	seq := []byte{c, n}
	for len(seq) < 8 {
		x, err := r.ReadByte()
		if err != nil {
			return string(seq), err
		}
		seq = append(seq, x)
		if (x >= 'A' && x <= 'Z') || x == '~' {
			break
		}
	}
	return string(seq), nil
}

// readLine reads search text after "/", echoing it, until Enter.
func readLine(r *bufio.Reader, w io.Writer) (string, error) {
	var s []rune
	for {
		k, err := readKey(r)
		if err != nil {
			return "", err
		}
		switch {
		case k == "\r" || k == "\n":
			return string(s), nil
		case k == "\x03" || k == "\x1b":
			return "", nil
		case (k == "\x7f" || k == "\b") && len(s) > 0:
			// Remove a whole character, including all bytes of an accented one.
			for len(s) > 1 && s[len(s)-1] >= 0x80 && s[len(s)-1] < 0xc0 {
				s = s[:len(s)-1]
			}
			s = s[:len(s)-1]
			fmt.Fprint(w, "\b \b")
		case len(k) == 1 && k[0] >= 0x20 && k[0] < 0x7f, len(k) == 1 && k[0] >= 0x80:
			// Bytes of a UTF-8 character arrive one by one; they are
			// collected and decoded when the line ends.
			s = append(s, rune(k[0]))
			fmt.Fprint(w, k)
		}
	}
}

// Less shows text like less(1) in a terminal of rows by cols that is
// already in raw mode: keys move without Enter, and the window is redrawn
// after each one.
func Less(w io.Writer, in io.Reader, text string, rows, cols int) error {
	width := max(cols-1, 20)
	p := &pager{lines: softWrap(strings.Split(strings.TrimRight(text, "\n"), "\n"), width), rows: max(rows, 3), width: width}
	r := bufio.NewReader(in)
	fmt.Fprint(w, "\x1b[2J")
	defer fmt.Fprint(w, "\x1b[0m\x1b[?25h\x1b[H\x1b[2J")
	for {
		p.draw(w)
		k, err := readKey(r)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if k == "/" && !p.help {
			fmt.Fprint(w, "\r\x1b[K/")
			s, err := readLine(r, w)
			if err != nil {
				return err
			}
			if s = strings.ToLower(strings.TrimSpace(utf8Bytes(s))); s != "" {
				p.pattern = s
				p.search(1)
			}
			continue
		}
		if p.key(k) {
			return nil
		}
	}
}

// utf8Bytes turns runes that each hold one byte back into the UTF-8 text
// they spell.
func utf8Bytes(s string) string {
	b := make([]byte, 0, len(s))
	for _, r := range s {
		b = append(b, byte(r))
	}
	return string(b)
}
