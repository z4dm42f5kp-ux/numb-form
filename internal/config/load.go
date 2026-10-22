package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// FileName is the configuration file name.
const FileName = "numb-form.conf"

// SystemPaths lists where the administrator's file may be, per OS. On
// Windows it lives under Program Files, which ordinary users cannot write.
func SystemPaths() []string {
	switch runtime.GOOS {
	case "windows":
		pf := os.Getenv("ProgramFiles")
		if pf == "" {
			pf = `C:\Program Files`
		}
		return []string{filepath.Join(pf, "numb-form", FileName)}
	case "darwin":
		return []string{"/Library/Application Support/numb-form/" + FileName}
	case "freebsd", "openbsd", "netbsd", "dragonfly":
		return []string{"/usr/local/etc/" + FileName, "/etc/" + FileName}
	}
	return []string{"/etc/" + FileName}
}

// UserPath is the per-user file: %APPDATA%\numb-form on Windows,
// ~/Library/Application Support/numb-form on macOS and
// $XDG_CONFIG_HOME/numb-form (~/.config/numb-form) elsewhere.
func UserPath() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "numb-form", FileName), nil
}

// StateDir is where direct-mode logs go by default.
func StateDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if d := os.Getenv("LOCALAPPDATA"); d != "" {
			return filepath.Join(d, "numb-form", "runs"), nil
		}
	case "darwin":
		h, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(h, "Library", "Logs", "numb-form"), nil
	}
	if d := os.Getenv("XDG_STATE_HOME"); d != "" && filepath.IsAbs(d) {
		return filepath.Join(d, "numb-form", "runs"), nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".local", "state", "numb-form", "runs"), nil
}

func readIfExists(p string) ([]byte, bool, error) {
	b, err := os.ReadFile(p) // #nosec G304 -- fixed configuration paths or -f.
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if len(b) > 1<<20 {
		return nil, false, fmt.Errorf("%s is larger than 1 MB", p)
	}
	return b, true, nil
}

// Load returns the effective configuration. explicit is the -f file, or
// empty.
func Load(explicit string) (Config, []string, error) {
	c := Defaults()
	var notes []string
	for _, p := range SystemPaths() {
		b, ok, err := readIfExists(p)
		if err != nil {
			return c, notes, err
		}
		if !ok {
			continue
		}
		if why := untrusted(p); why != "" {
			notes = append(notes, fmt.Sprintf("%s ignored: %s", p, why))
			continue
		}
		n, err := c.Apply(b, p, true)
		notes = append(notes, n...)
		if err != nil {
			return c, notes, err
		}
		break
	}
	files := []string{}
	if up, err := UserPath(); err == nil {
		files = append(files, up)
	}
	if explicit != "" {
		files = append(files, explicit)
	}
	for i, p := range files {
		b, ok, err := readIfExists(p)
		if err != nil {
			return c, notes, err
		}
		if !ok {
			if p == explicit && i == len(files)-1 {
				return c, notes, fmt.Errorf("configuration file %s not found", p)
			}
			continue
		}
		n, err := c.Apply(b, p, false)
		notes = append(notes, n...)
		if err != nil {
			return c, notes, err
		}
	}
	return c, notes, c.Validate()
}

// Template returns a commented configuration file with every default.
func Template() string {
	d := Defaults()
	var b strings.Builder
	b.WriteString("# numb-form.conf: every line shows the built-in default.\n")
	b.WriteString("# Remove the leading '#' to change a value. See docs/configuration.md.\n\n")
	line := func(comment, k, v string) { fmt.Fprintf(&b, "# %s\n#%s = %s\n\n", comment, k, v) }
	line("direct: write into the working folder; review: work on a copy in Downloads, then numb-commit", "mode", d.Mode)
	line("each: confirm every document; none: run without questions (needs person)", "confirm", d.Confirm)
	line("look up registers and websites online", "online", "yes")
	line("name pattern of the folder that holds the spreadsheet", "overview_dir", d.OverviewDir)
	line("spreadsheet column with the company names", "header", d.Header)
	line("spreadsheet column with the customer ids", "id_header", d.IDHeader)
	line("template file name without extension", "template_name", d.TemplateName)
	line("person processing the form; empty means ask once per run", "person", "")
	line("evaluation date format (DD, MM, YYYY, YY)", "date_format", d.DateFormat)
	line("stop safely after this long (at most 25m)", "timeout", "25m")
	line("most megabytes of documents one run may write (at most 100)", "max_write_mb", "100")
	line("how close a folder name must be to be offered as a match (0.5-1)", "fuzzy_threshold", "0.85")
	line("review mode: folder for the copy; empty means your Downloads folder", "downloads", "")
	line("direct mode: folder for logs; empty means the per-user state folder", "log_dir", "")
	line("sources for the address, in order: "+strings.Join(Sources, " "), "address_sources", strings.Join(d.AddressSources, " "))
	line("when no address is found: ask or blank", "address_fallback", d.AddressFallback)
	line("sources for representatives", "reps_sources", strings.Join(d.RepsSources, " "))
	line("least trusted source accepted for representatives: register, website or any", "reps_min_trust", d.RepsMinTrust)
	line("when no representative is found: no-reps-text or blank", "reps_fallback", d.RepsFallback)
	line("caution text: auto (unverified values), always or never", "caution", d.Caution)
	line("company website and its legal notice: auto (found by itself), ask (you type it) or never", "website", d.Website)
	line("web search engines, asked in this order when one refuses: serper (Google results; needs NUMB_FORM_SERPER_KEY) duckduckgo brave, or none", "search", d.Search)
	line("invisible browser for script-built websites and refused searches: auto (Google Chrome, else Chromium or Edge), none, or a path", "browser", d.Browser)
	line("also try addresses such as companyname.de or companyname.com.mx", "guess_domains", "yes")
	line("path to LibreOffice or OpenOffice soffice; empty means search", "soffice", "")
	line(".pages templates: auto, pages-app (macOS) or libreoffice", "pages_converter", d.PagesConverter)
	line("EXPERIMENTAL, macOS only: save filled forms as .pages through Apple Pages", "pages_write", "no")
	line(".numbers spreadsheets: auto, numbers-app (macOS) or libreoffice", "numbers_converter", d.NumbersConverter)
	line("SYSTEM FILE ONLY: allow legacy .doc and .xls (always offline)", "legacy_formats", "no")
	line("legacy .doc templates: write docx, or the same format as the template", "legacy_output", d.LegacyOutput)
	return b.String()
}
