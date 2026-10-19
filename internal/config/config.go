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
	Mode    string // direct | review
	Confirm string // each | none
	Online  bool

	OverviewDir  string // folder name pattern of the overview folder
	Header       string // spreadsheet column with company names
	IDHeader     string // spreadsheet column with customer ids
	TemplateName string // template file name without extension
	Person       string // person processing the form; empty = ask
	DateFormat   string // DD, MM, YYYY, YY tokens
	Timeout      time.Duration
	MaxWriteMB   int
	Fuzzy        float64

	Downloads string // review mode: where the copy goes; empty = Downloads
	LogDir    string // direct mode: where logs go; empty = per-user state folder

	AddressSources  []string
	AddressFallback string // blank | ask
	RepsSources     []string
	RepsMinTrust    string // register | website | any
	RepsFallback    string // no-reps-text | blank
	Caution         string // auto | always | never
	Website         string // auto | ask | never
	Search          string // duckduckgo | none
	GuessDomains    bool   // also try addresses such as companyname.de

	Soffice          string // path to LibreOffice/OpenOffice soffice; empty = search
	PagesConverter   string // auto | pages-app | libreoffice
	PagesWrite       bool   // experimental: write .pages through Apple Pages
	NumbersConverter string // auto | numbers-app | libreoffice
	LegacyFormats    bool   // .doc and .xls; system file only
	LegacyOutput     string // docx | same

	// Origin records which file set each key, for "numb-form config".
	Origin map[string]string
}

// Sources understood in *_sources keys, most authoritative first.
var Sources = []string{"companies-house", "vies", "brasil-cnpj", "official", "gleif", "website", "search", "wikidata"}

// MaxTimeout is the hard upper bound for a run.
const MaxTimeout = 25 * time.Minute

// Defaults returns the built-in configuration, equal to the documented
// behaviour.
func Defaults() Config {
	return Config{
		Mode:             "direct",
		Confirm:          "each",
		Online:           true,
		OverviewDir:      "Übersicht",
		Header:           "Customer Name",
		IDHeader:         "Customer ID",
		TemplateName:     "Appendix A - Risk Assesment",
		DateFormat:       "DD/MM/YYYY",
		Timeout:          MaxTimeout,
		MaxWriteMB:       100,
		Fuzzy:            0.85,
		AddressSources:   append([]string(nil), Sources...),
		AddressFallback:  "ask",
		RepsSources:      append([]string(nil), Sources...),
		RepsMinTrust:     "any",
		RepsFallback:     "no-reps-text",
		Caution:          "auto",
		Website:          "auto",
		Search:           "duckduckgo",
		GuessDomains:     true,
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

// Allowed reports whether source is listed.
func Allowed(list []string, source string) bool {
	for _, s := range list {
		if s == source {
			return true
		}
	}
	return false
}

// Validate checks combinations that individual keys cannot.
func (c *Config) Validate() error {
	if c.Confirm == "none" && c.Person == "" {
		return fmt.Errorf("confirm = none needs person = <full name> in the configuration")
	}
	if c.Confirm == "none" && c.AddressFallback == "ask" {
		c.AddressFallback = "blank"
	}
	if c.Website == "wikidata" {
		c.Website = "auto" // older name of the same setting
	}
	if c.Confirm == "none" && c.Website == "ask" {
		c.Website = "auto"
	}
	if c.Timeout <= 0 || c.Timeout > MaxTimeout {
		return fmt.Errorf("timeout must be between 1s and 25m")
	}
	if c.MaxWriteMB < 1 || c.MaxWriteMB > 100 {
		return fmt.Errorf("max_write_mb must be between 1 and 100")
	}
	_, err := GoDateLayout(c.DateFormat)
	return err
}
