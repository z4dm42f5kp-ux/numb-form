// The pager of numb-form man.

package manual

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

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
