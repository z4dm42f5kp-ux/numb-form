// Package ui implements the line-based terminal dialogue. Every read is
// bounded by the run's deadline.
//
// In plain words: it prints the questions, reads your answers, and stops
// waiting when the time limit is reached.
package ui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// ErrClosed is returned when standard input reaches end of file.
var ErrClosed = errors.New("input closed")

type line struct {
	s   string
	err error
}

// UI reads answers from the user.
type UI struct {
	ctx   context.Context
	out   io.Writer
	lines chan line
}

// New starts reading lines from in.
func New(ctx context.Context, in io.Reader, out io.Writer) *UI {
	u := &UI{ctx: ctx, out: out, lines: make(chan line)}
	go func() {
		r := bufio.NewReader(in)
		for {
			s, err := r.ReadString('\n')
			if err != nil && s == "" {
				u.lines <- line{err: ErrClosed}
				return
			}
			u.lines <- line{s: strings.TrimRight(s, "\r\n")}
		}
	}()
	return u
}

// IsTerminal reports whether f is an interactive terminal.
func IsTerminal(f *os.File) bool {
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

// Printf writes to the terminal.
func (u *UI) Printf(format string, a ...any) { fmt.Fprintf(u.out, format, a...) }

// Println writes a line to the terminal.
func (u *UI) Println(a ...any) { fmt.Fprintln(u.out, a...) }

// Safe prints remote text with control characters removed.
func Safe(s string) string { return textutil.SafeTerminal(s) }

// Ask prints prompt and returns the trimmed answer.
func (u *UI) Ask(prompt string) (string, error) {
	fmt.Fprint(u.out, prompt)
	select {
	case <-u.ctx.Done():
		fmt.Fprintln(u.out)
		return "", u.ctx.Err()
	case l := <-u.lines:
		if l.err != nil {
			return "", l.err
		}
		return strings.TrimSpace(l.s), nil
	}
}

// AskDefault is Ask with a default used for an empty answer.
func (u *UI) AskDefault(prompt, def string) (string, error) {
	p := prompt
	if def != "" {
		p = fmt.Sprintf("%s [%s]", prompt, Safe(def))
	}
	a, err := u.Ask(p + ": ")
	if err != nil {
		return "", err
	}
	if a == "" {
		return def, nil
	}
	return a, nil
}

// Choose asks until the answer is one of the allowed single-letter keys.
func (u *UI) Choose(prompt string, keys string, def byte) (byte, error) {
	for {
		a, err := u.Ask(prompt)
		if err != nil {
			return 0, err
		}
		if a == "" && def != 0 {
			return def, nil
		}
		if len(a) >= 1 {
			c := strings.ToLower(a)[0]
			if strings.IndexByte(keys, c) >= 0 {
				return c, nil
			}
		}
		u.Printf("Please answer one of: %s\n", strings.Join(strings.Split(keys, ""), ", "))
	}
}

// YesNo asks a yes/no question.
func (u *UI) YesNo(prompt string, def bool) (bool, error) {
	d := byte('n')
	hint := "[y/N]"
	if def {
		d, hint = 'y', "[Y/n]"
	}
	c, err := u.Choose(prompt+" "+hint+" ", "yn", d)
	return c == 'y', err
}
