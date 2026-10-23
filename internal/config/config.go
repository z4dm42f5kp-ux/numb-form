// Package config reads numb-form.conf.
//
// Files are read in this order, later ones overriding earlier ones:
//
//  1. the system file (administrators only),
//  2. the per-user file,
//  3. a file given with -f.
//
// Some keys, such as legacy_formats, are honoured only in the system
// file, so enabling them needs administrator rights.
package config

import (
	"fmt"
	"strings"
	"time"
)

// Config is the effective configuration.
type Config struct {
	Mode   string // direct | review
	Online bool   // numb-form update may contact GitHub

	OverviewDir  string        // folder name pattern of the overview folder
	Header       string        // spreadsheet column with company names
	IDHeader     string        // spreadsheet column with customer ids
	TemplateName string        // template file name without extension
	Person       string        // person processing the form; empty = ask
	DateFormat   string        // DD, MM, YYYY, YY tokens
	Timeout      time.Duration // 0: no time limit
	MaxWriteMB   int
	Fuzzy        float64

	Downloads  string // review mode: where the copy goes; empty = Downloads
	LogDir     string // direct mode: where logs go; empty = per-user state folder
	CommitFrom string // numb-commit: the working copy to commit; empty = ask
	CommitTo   string // numb-commit: the real working folder it goes to; empty = ask

	Soffice          string // path to LibreOffice/OpenOffice soffice; empty = search
	PagesConverter   string // auto | pages-app | libreoffice
	PagesWrite       bool   // experimental: write .pages through Apple Pages
	NumbersConverter string // auto | numbers-app | libreoffice
	LegacyFormats    bool   // .doc and .xls; system file only
	LegacyOutput     string // docx | same

	// Origin records which file set each key, for "numb-form config".
	Origin map[string]string
}

// Defaults returns the built-in configuration, equal to the documented
// behaviour.
func Defaults() Config {
	return Config{
		Mode:             "direct",
		Online:           true,
		OverviewDir:      "Übersicht",
		Header:           "Customer Name",
		IDHeader:         "Customer ID",
		TemplateName:     "Appendix A - Risk Assesment",
		DateFormat:       "DD/MM/YYYY",
		MaxWriteMB:       100,
		Fuzzy:            0.85,
		PagesConverter:   "auto",
		NumbersConverter: "auto",
		LegacyOutput:     "docx",
		Origin:           map[string]string{},
	}
}

// GoDateLayout turns DD/MM/YYYY style formats into a Go time layout.
func GoDateLayout(f string) (string, error) {
	r := strings.NewReplacer("YYYY", "2006", "YY", "06", "MM", "01", "DD", "02")
	out := r.Replace(f)
	if strings.ContainsAny(out, "DMY") || !strings.Contains(out, "02") || !strings.Contains(out, "01") {
		return "", fmt.Errorf("date_format %q: use DD, MM and YYYY (for example DD/MM/YYYY)", f)
	}
	return out, nil
}

// Validate checks combinations that individual keys cannot.
func (c *Config) Validate() error {
	if c.Timeout < 0 {
		return fmt.Errorf("timeout must be a duration such as 30m, or none")
	}
	if c.MaxWriteMB < 1 || c.MaxWriteMB > 100 {
		return fmt.Errorf("max_write_mb must be between 1 and 100")
	}
	_, err := GoDateLayout(c.DateFormat)
	return err
}
