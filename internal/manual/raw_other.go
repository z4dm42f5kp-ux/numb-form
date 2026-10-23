//go:build !windows

package manual

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func stty(args ...string) (string, error) {
	cmd := exec.Command("stty", args...) // #nosec G204 -- fixed program, fixed or saved arguments.
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// RawTerminal switches the terminal to single-key input with stty(1) and
// returns a function that restores it, and the window size.
func RawTerminal() (restore func(), rows, cols int, err error) {
	saved, err := stty("-g")
	if err != nil {
		return nil, 0, 0, err
	}
	rows, cols = 24, 80
	if size, err := stty("size"); err == nil {
		_, _ = fmt.Sscan(size, &rows, &cols)
	}
	if _, err := stty("raw", "-echo"); err != nil {
		return nil, 0, 0, err
	}
	return func() { _, _ = stty(saved) }, rows, cols, nil
}
