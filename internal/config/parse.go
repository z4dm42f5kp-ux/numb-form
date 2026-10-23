package config

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type setter func(c *Config, v string) error

func oneOf(field *string, allowed ...string) setter {
	return func(c *Config, v string) error {
		for _, a := range allowed {
			if v == a {
				*field = v
				return nil
			}
		}
		return fmt.Errorf("must be one of: %s", strings.Join(allowed, ", "))
	}
}

func yesNo(v string) (bool, error) {
	switch strings.ToLower(v) {
	case "yes", "true", "on", "1":
		return true, nil
	case "no", "false", "off", "0":
		return false, nil
	}
	return false, fmt.Errorf("must be yes or no")
}

// keys maps every configuration key to its parser. Pointers into a
// throwaway Config are rebound per call by keyTable.
func keyTable(c *Config) map[string]setter {
	str := func(p *string) setter { return func(_ *Config, v string) error { *p = v; return nil } }
	boolean := func(p *bool) setter {
		return func(_ *Config, v string) error { b, err := yesNo(v); *p = b; return err }
	}
	return map[string]setter{
		"mode":              oneOf(&c.Mode, "direct", "review"),
		"online":            boolean(&c.Online),
		"overview_dir":      str(&c.OverviewDir),
		"header":            str(&c.Header),
		"id_header":         str(&c.IDHeader),
		"template_name":     str(&c.TemplateName),
		"person":            str(&c.Person),
		"date_format":       str(&c.DateFormat),
		"downloads":         str(&c.Downloads),
		"log_dir":           str(&c.LogDir),
		"soffice":           str(&c.Soffice),
		"pages_converter":   oneOf(&c.PagesConverter, "auto", "pages-app", "libreoffice"),
		"pages_write":       boolean(&c.PagesWrite),
		"numbers_converter": oneOf(&c.NumbersConverter, "auto", "numbers-app", "libreoffice"),
		"legacy_formats":    boolean(&c.LegacyFormats),
		"legacy_output":     oneOf(&c.LegacyOutput, "docx", "same"),
		"timeout": func(_ *Config, v string) error {
			switch strings.ToLower(v) {
			case "none", "off", "no", "0":
				c.Timeout = 0
				return nil
			}
			d, err := time.ParseDuration(v)
			if err == nil && d <= 0 {
				err = fmt.Errorf("must be positive, or none")
			}
			c.Timeout = d
			return err
		},
		"max_write_mb": func(_ *Config, v string) error {
			n, err := strconv.Atoi(v)
			c.MaxWriteMB = n
			return err
		},
		"fuzzy_threshold": func(_ *Config, v string) error {
			f, err := strconv.ParseFloat(v, 64)
			if err == nil && (f < 0.5 || f > 1) {
				err = fmt.Errorf("must be between 0.5 and 1")
			}
			c.Fuzzy = f
			return err
		},
	}
}

// retired keys belonged to the online lookup and to running without
// questions, both removed; files that still set them keep working.
var retired = map[string]bool{
	"confirm": true, "address_sources": true, "address_fallback": true, "reps_sources": true, "reps_min_trust": true,
	"reps_fallback": true, "caution": true, "website": true, "search": true, "browser": true, "guess_domains": true,
}

// systemOnly keys are ignored unless they come from the system file.
var systemOnly = map[string]bool{"legacy_formats": true}

// Keys returns every known key, for documentation and tests.
func Keys() []string {
	var c Config
	var out []string
	for k := range keyTable(&c) {
		out = append(out, k)
	}
	return out
}

// Apply parses one configuration file into c. origin names the file in
// messages; system tells whether system-only keys are honoured.
func (c *Config) Apply(data []byte, origin string, system bool) ([]string, error) {
	var notes []string
	table := keyTable(c)
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), 64*1024)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if n == 1 {
			line = strings.TrimPrefix(line, "\ufeff")
		}
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return notes, fmt.Errorf("%s:%d: expected key = value", origin, n)
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if i := strings.Index(v, " #"); i >= 0 && !strings.HasPrefix(v, `"`) {
			v = strings.TrimSpace(v[:i])
		}
		if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
			v = v[1 : len(v)-1]
		}
		if retired[k] {
			notes = append(notes, fmt.Sprintf("%s:%d: %s is no longer used and is ignored", origin, n, k))
			continue
		}
		set, known := table[k]
		if !known {
			return notes, fmt.Errorf("%s:%d: unknown key %q", origin, n, k)
		}
		if systemOnly[k] && !system {
			notes = append(notes, fmt.Sprintf("%s:%d: %s is ignored here; an administrator must set it in the system configuration", origin, n, k))
			continue
		}
		if err := set(c, v); err != nil {
			return notes, fmt.Errorf("%s:%d: %s: %w", origin, n, k, err)
		}
		c.Origin[k] = origin
	}
	return notes, sc.Err()
}
