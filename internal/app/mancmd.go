package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	numbform "github.com/z4dm42f5kp-ux/numb-form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/manual"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

func envInt(k string, def, lo, hi int) int {
	if n, err := strconv.Atoi(os.Getenv(k)); err == nil && n >= lo {
		return min(n, hi)
	}
	return def
}

// Man shows the built-in manual: "numb-form man [topic] [en|es|de]",
// "numb-form man topics", "numb-form man -install", "numb-form man -roff".
func Man(out io.Writer, args []string) int {
	lang, topic := manual.Lang(), ""
	for _, a := range args {
		switch strings.ToLower(a) {
		case "-install", "--install", "install-man":
			roff, err := manual.Roff(numbform.Docs, Version, time.Now())
			if err == nil {
				var paths []string
				if paths, err = manual.Install(roff); err == nil {
					fmt.Fprint(out, manual.Hint(paths))
					return ExitOK
				}
			}
			fmt.Fprintln(out, "man:", err)
			return ExitError
		case "-roff", "--roff":
			roff, err := manual.Roff(numbform.Docs, Version, time.Now())
			if err != nil {
				fmt.Fprintln(out, "man:", err)
				return ExitError
			}
			fmt.Fprint(out, roff)
			return ExitOK
		case "topics", "-h", "-help", "--help":
			fmt.Fprint(out, manual.TopicList())
			return ExitOK
		case "en", "es", "de":
			lang = strings.ToLower(a)
		case "numb-form", "numb-commit", "numb-date":
			// "numb-form man numb-form" (from man.cmd) means the whole manual.
		default:
			if _, ok := manual.Find(a); !ok {
				fmt.Fprintf(out, "man: no topic %q\n\n%s", a, manual.TopicList())
				return ExitUsage
			}
			topic = a
		}
	}
	render := func(width int) (string, error) {
		width = max(40, min(width, 100))
		if topic == "" {
			return manual.Full(numbform.Docs, lang, width)
		}
		t, _ := manual.Find(topic)
		md, err := manual.Source(numbform.Docs, t, lang)
		return manual.Text(md, width), err
	}
	interactive := out == os.Stdout && ui.IsTerminal(os.Stdout) && ui.IsTerminal(os.Stdin)
	if interactive && runtime.GOOS != "windows" {
		if less, err := exec.LookPath("less"); err == nil {
			text, err := render(envInt("COLUMNS", 80, 40, 200) - 1)
			if err != nil {
				fmt.Fprintln(out, "man:", err)
				return ExitError
			}
			cmd := exec.Command(less, "-F", "-X") // #nosec G204 -- fixed program and arguments.
			cmd.Stdin, cmd.Stdout, cmd.Stderr = strings.NewReader(text), os.Stdout, os.Stderr
			if cmd.Run() == nil {
				return ExitOK
			}
		}
	}
	if interactive {
		// A less(1)-like view: keys act at once, without Enter.
		if restore, rows, cols, err := manual.RawTerminal(); err == nil {
			text, err := render(cols - 1)
			if err == nil {
				err = manual.Less(os.Stdout, os.Stdin, text, rows, cols)
			}
			restore()
			if err != nil {
				fmt.Fprintln(out, "man:", err)
				return ExitError
			}
			return ExitOK
		}
	}
	text, err := render(envInt("COLUMNS", 80, 40, 200) - 1)
	if err != nil {
		fmt.Fprintln(out, "man:", err)
		return ExitError
	}
	if !interactive {
		_, _ = io.WriteString(out, text)
		return ExitOK
	}
	if err := manual.Page(os.Stdout, os.Stdin, text, envInt("LINES", 24, 5, 200)); err != nil && !errors.Is(err, manual.ErrQuit) {
		return ExitError
	}
	return ExitOK
}
