// Package convert turns formats numb-form cannot edit directly (.doc,
// .pages, .numbers) into ones it can, using software already on the
// computer: LibreOffice or Apache OpenOffice anywhere, and Apple Pages or
// Numbers on macOS. Nothing is downloaded. Work happens in a private
// temporary folder, including a throwaway office profile, which is
// deleted afterwards, so no cache is left behind.
package convert

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// MaxBytes caps the size of files going through a converter.
const MaxBytes = 40 << 20

// Timeout bounds one conversion.
const Timeout = 2 * time.Minute

// sofficeCandidates are the usual install locations, after PATH.
func sofficeCandidates() []string {
	switch runtime.GOOS {
	case "windows":
		var out []string
		for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "LOCALAPPDATA"} {
			if d := os.Getenv(env); d != "" {
				out = append(out,
					filepath.Join(d, "LibreOffice", "program", "soffice.exe"),
					filepath.Join(d, "Programs", "LibreOffice", "program", "soffice.exe"),
					filepath.Join(d, "OpenOffice 4", "program", "soffice.exe"))
			}
		}
		return out
	case "darwin":
		home, _ := os.UserHomeDir()
		return []string{
			"/Applications/LibreOffice.app/Contents/MacOS/soffice",
			filepath.Join(home, "Applications/LibreOffice.app/Contents/MacOS/soffice"),
			"/Applications/OpenOffice.app/Contents/MacOS/soffice",
		}
	}
	return []string{"/usr/bin/soffice", "/usr/local/bin/soffice", "/usr/lib/libreoffice/program/soffice", "/opt/libreoffice/program/soffice"}
}

// FindSoffice returns the LibreOffice/OpenOffice program, or an error
// that says how to install one.
func FindSoffice(configured string) (string, error) {
	if configured != "" {
		if st, err := os.Stat(configured); err == nil && !st.IsDir() {
			return configured, nil
		}
		return "", fmt.Errorf("soffice = %s: not found", configured)
	}
	for _, n := range []string{"soffice", "libreoffice"} {
		if p, err := exec.LookPath(n); err == nil {
			return p, nil
		}
	}
	for _, p := range sofficeCandidates() {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", errors.New("LibreOffice or Apache OpenOffice is needed for this format; install it, or set soffice = <path> in numb-form.conf")
}

func fileURL(p string) string {
	s := filepath.ToSlash(p)
	if !strings.HasPrefix(s, "/") {
		s = "/" + s // C:/Users/... on Windows
	}
	return (&url.URL{Scheme: "file", Path: s}).String()
}

// Soffice converts data (a file with extension inExt) to outExt, such
// as "docx", "xlsx" or "doc", and returns the result.
func Soffice(ctx context.Context, soffice string, data []byte, inExt, outExt string) ([]byte, error) {
	if len(data) > MaxBytes {
		return nil, fmt.Errorf("file is larger than %d MB", MaxBytes>>20)
	}
	tmp, err := os.MkdirTemp("", "numb-form-convert-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	in := filepath.Join(tmp, "in"+inExt)
	if err := os.WriteFile(in, data, 0o600); err != nil {
		return nil, err
	}
	outDir := filepath.Join(tmp, "out")
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	// #nosec G204 -- soffice is the configured or discovered office program;
	// arguments are fixed except paths inside our private temp folder.
	cmd := exec.CommandContext(ctx, soffice,
		"-env:UserInstallation="+fileURL(filepath.Join(tmp, "profile")),
		"--headless", "--norestore", "--nolockcheck", "--nodefault", "--nologo",
		"--convert-to", outExt, "--outdir", outDir, in)
	cmd.Env = append(os.Environ(), "HOME="+tmp, "TMPDIR="+tmp, "TMP="+tmp, "TEMP="+tmp)
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("conversion took longer than %s", Timeout)
	}
	if err != nil {
		return nil, fmt.Errorf("conversion failed: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return readCapped(filepath.Join(outDir, "in."+outExt))
}

func readCapped(p string) ([]byte, error) {
	f, err := os.Open(p) // #nosec G304 -- file inside our temp folder.
	if err != nil {
		return nil, fmt.Errorf("conversion produced no output: %w", err)
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, MaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(b) > MaxBytes {
		return nil, errors.New("converted file is too large")
	}
	return b, nil
}
