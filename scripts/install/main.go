// Command install keeps the install commands of the documentation pinned
// to one stable release (.github/workflows/install-docs.yml).
//
//	go run ./scripts/install render
//	go run ./scripts/install pin vX.Y.Z SHA256SUMS program...
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/installdoc"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "install:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	switch {
	case len(args) == 1 && args[0] == "render":
		changed, err := installdoc.Stale(".", true)
		for _, d := range changed {
			fmt.Println("rewrote", d)
		}
		return err
	case len(args) >= 4 && args[0] == "pin":
		sums, err := os.ReadFile(args[2])
		if err != nil {
			return err
		}
		p, err := installdoc.ParsePins(append([]byte(fmt.Sprintf("version %s\nprograms %s\n", args[1], strings.Join(args[3:], " "))), sums...))
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(installdoc.Dir, "pins.txt"), p.Format(), 0o644) // #nosec G306 -- documentation.
	}
	return fmt.Errorf("usage: install render, or install pin vX.Y.Z SHA256SUMS program [program]")
}
