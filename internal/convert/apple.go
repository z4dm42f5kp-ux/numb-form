package convert

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// HasApp reports whether an Apple app (Pages or Numbers) is installed.
func HasApp(name string) bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	for _, d := range []string{"/Applications", "/System/Applications"} {
		if _, err := os.Stat(filepath.Join(d, name+".app")); err == nil {
			return true
		}
	}
	return false
}

// The scripts receive their paths as arguments, never inside the script
// text, so file names cannot inject AppleScript.
var scripts = map[string]string{
	"pages-to-docx": `on run argv
tell application "Pages"
set d to open (POSIX file (item 1 of argv))
export d to (POSIX file (item 2 of argv)) as Microsoft Word
close d saving no
end tell
end run`,
	"numbers-to-xlsx": `on run argv
tell application "Numbers"
set d to open (POSIX file (item 1 of argv))
export d to (POSIX file (item 2 of argv)) as Microsoft Excel
close d saving no
end tell
end run`,
	"docx-to-pages": `on run argv
tell application "Pages"
set d to open (POSIX file (item 1 of argv))
save d in (POSIX file (item 2 of argv))
close d saving no
end tell
end run`,
}

// Apple converts with Pages or Numbers through osascript. job is one of
// "pages-to-docx", "numbers-to-xlsx" and the experimental "docx-to-pages".
func Apple(ctx context.Context, job string, data []byte, inExt, outExt string) ([]byte, error) {
	script, ok := scripts[job]
	if !ok {
		return nil, fmt.Errorf("unknown conversion %q", job)
	}
	if runtime.GOOS != "darwin" {
		return nil, errors.New("Apple Pages and Numbers are only available on macOS")
	}
	if len(data) > MaxBytes {
		return nil, fmt.Errorf("file is larger than %d MB", MaxBytes>>20)
	}
	tmp, err := os.MkdirTemp("", "numb-form-convert-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	in := filepath.Join(tmp, "in"+inExt)
	out := filepath.Join(tmp, "out"+outExt)
	if err := os.WriteFile(in, data, 0o600); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	args := []string{}
	for _, l := range strings.Split(script, "\n") {
		args = append(args, "-e", l)
	}
	args = append(args, in, out)
	// #nosec G204 -- fixed script; paths are passed as arguments.
	msg, err := exec.CommandContext(ctx, "/usr/bin/osascript", args...).CombinedOutput()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("conversion took longer than %s", Timeout)
	}
	if err != nil {
		return nil, fmt.Errorf("%s failed: %v: %s", job, err, strings.TrimSpace(string(msg)))
	}
	return readCapped(out)
}
