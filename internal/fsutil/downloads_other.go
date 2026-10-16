//go:build !windows

package fsutil

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// downloadsDir returns ~/Downloads, honouring XDG_DOWNLOAD_DIR from
// user-dirs.dirs on Linux and the BSDs.
func downloadsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS != "darwin" {
		cfg := os.Getenv("XDG_CONFIG_HOME")
		if cfg == "" {
			cfg = filepath.Join(home, ".config")
		}
		if d := xdgDownload(filepath.Join(cfg, "user-dirs.dirs"), home); d != "" {
			return d, nil
		}
	}
	return filepath.Join(home, "Downloads"), nil
}

func xdgDownload(file, home string) string {
	f, err := os.Open(file) // #nosec G304 -- fixed per-user config file.
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		v, ok := strings.CutPrefix(line, "XDG_DOWNLOAD_DIR=")
		if !ok {
			continue
		}
		v = strings.Trim(v, `"`)
		v = strings.ReplaceAll(v, "$HOME", home)
		if filepath.IsAbs(v) && v != home {
			return v
		}
	}
	return ""
}
