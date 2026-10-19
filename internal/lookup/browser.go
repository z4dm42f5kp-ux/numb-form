package lookup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Browser is a locally installed Google Chrome (or, when Chrome is not
// installed, Chromium or Microsoft Edge) that numb-form starts headless:
// no window is shown. Every start uses a fresh, empty profile in the
// temporary folder that is deleted right after, so the user's own
// browser profile, history and sign-ins are never used or changed.
type Browser struct {
	Path string
	Name string
}

// browserMu allows one browser at a time, to keep memory use low.
var browserMu sync.Mutex

// maxDOM caps the page read from the browser.
const maxDOM = 4 << 20

func browserCandidates() []Browser {
	var out []Browser
	add := func(name string, paths ...string) {
		for _, p := range paths {
			if p != "" {
				out = append(out, Browser{Path: p, Name: name})
			}
		}
	}
	env := os.Getenv
	switch runtime.GOOS {
	case "windows":
		for _, base := range []string{env("ProgramFiles"), env("ProgramFiles(x86)"), env("LOCALAPPDATA")} {
			if base != "" {
				add("Google Chrome", filepath.Join(base, "Google", "Chrome", "Application", "chrome.exe"))
			}
		}
		for _, base := range []string{env("ProgramFiles(x86)"), env("ProgramFiles")} {
			if base != "" {
				add("Microsoft Edge", filepath.Join(base, "Microsoft", "Edge", "Application", "msedge.exe"))
			}
		}
	case "darwin":
		home, _ := os.UserHomeDir()
		for _, root := range []string{"/Applications", filepath.Join(home, "Applications")} {
			add("Google Chrome", filepath.Join(root, "Google Chrome.app", "Contents", "MacOS", "Google Chrome"))
		}
		add("Chromium", "/Applications/Chromium.app/Contents/MacOS/Chromium")
		add("Microsoft Edge", "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge")
	default:
		for _, n := range []string{"google-chrome", "google-chrome-stable"} {
			if p, err := exec.LookPath(n); err == nil {
				add("Google Chrome", p)
			}
		}
		for _, n := range []string{"chromium", "chromium-browser"} {
			if p, err := exec.LookPath(n); err == nil {
				add("Chromium", p)
			}
		}
		for _, n := range []string{"microsoft-edge", "microsoft-edge-stable"} {
			if p, err := exec.LookPath(n); err == nil {
				add("Microsoft Edge", p)
			}
		}
	}
	return out
}

// FindBrowser returns the browser to use: "auto" picks Google Chrome,
// else Chromium or Microsoft Edge; "none" or "" turns it off; anything
// else is the path of a Chromium-based browser.
func FindBrowser(setting string) *Browser {
	switch setting {
	case "", "none":
		return nil
	case "auto":
		for _, b := range browserCandidates() {
			if st, err := os.Stat(b.Path); err == nil && !st.IsDir() {
				b := b
				return &b
			}
		}
		return nil
	}
	if st, err := os.Stat(setting); err == nil && !st.IsDir() {
		return &Browser{Path: setting, Name: filepath.Base(setting)}
	}
	return nil
}

type capped struct {
	buf bytes.Buffer
	n   int
}

func (c *capped) Write(p []byte) (int, error) {
	if room := maxDOM - c.buf.Len(); room > 0 {
		if len(p) > room {
			c.buf.Write(p[:room])
		} else {
			c.buf.Write(p)
		}
	}
	c.n += len(p)
	return len(p), nil
}

// removeAllRetry deletes a folder; on Windows the browser may hold files
// for a moment after it exits.
func removeAllRetry(dir string) {
	for i := 0; i < 10; i++ {
		if err := os.RemoveAll(dir); err == nil {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
}

// Fetch loads an HTTPS page in the headless browser, lets its scripts run
// for a few seconds, and returns the resulting HTML.
func (b *Browser) Fetch(ctx context.Context, page string) ([]byte, error) {
	u, err := url.Parse(page)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return nil, fmt.Errorf("browser: refusing %q", page)
	}
	browserMu.Lock()
	defer browserMu.Unlock()
	profile, err := os.MkdirTemp("", "numb-form-browser-")
	if err != nil {
		return nil, err
	}
	defer removeAllRetry(profile)
	ctx, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()
	args := []string{
		"--headless=new", "--disable-gpu", "--no-first-run", "--no-default-browser-check", "--disable-extensions",
		"--disable-sync", "--disable-background-networking", "--disable-component-update", "--disable-default-apps",
		"--disable-features=Translate,MediaRouter,OptimizationHints", "--mute-audio", "--hide-scrollbars",
		"--disk-cache-size=1", "--user-data-dir=" + profile, "--virtual-time-budget=8000", "--dump-dom", u.String(),
	}
	if runtime.GOOS != "windows" && os.Geteuid() == 0 {
		args = append([]string{"--no-sandbox"}, args...) // the sandbox cannot run as root
	}
	cmd := exec.CommandContext(ctx, b.Path, args...) // #nosec G204 -- the user's own browser, fixed flags, checked HTTPS URL.
	var out capped
	cmd.Stdout = &out
	cmd.Stderr = nil
	cmd.WaitDelay = 5 * time.Second
	err = cmd.Run()
	if out.buf.Len() > 0 {
		return out.buf.Bytes(), nil
	}
	if err == nil {
		err = errors.New("empty page")
	}
	return nil, fmt.Errorf("%s: %w", b.Name, err)
}
