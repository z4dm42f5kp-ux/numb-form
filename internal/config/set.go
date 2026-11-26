package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SetUserValue writes key = value into the per-user file, replacing an
// existing setting of the same key. The value is checked first, so an
// invalid setting is never saved. It returns the file's path.
func SetUserValue(key, value string) (string, error) {
	key = strings.TrimSpace(key)
	value = strings.Join(strings.Fields(value), " ")
	if retired[key] {
		return "", fmt.Errorf("%s is no longer used", key)
	}
	if systemOnly[key] {
		return "", fmt.Errorf("%s can only be set by an administrator in the system configuration", key)
	}
	line := key + " = " + value
	if strings.Contains(value, "#") {
		line = key + ` = "` + value + `"`
	}
	check := Defaults()
	if _, err := check.Apply([]byte(line+"\n"), "value", false); err != nil {
		return "", errors.New(strings.TrimPrefix(err.Error(), "value:1: "))
	}
	p, err := UserPath()
	if err != nil {
		return "", err
	}
	old, err := os.ReadFile(p) // #nosec G304 -- the per-user configuration file.
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	var out []string
	replaced := false
	for _, l := range strings.Split(strings.TrimRight(string(old), "\n"), "\n") {
		k, _, ok := strings.Cut(strings.TrimSpace(l), "=")
		if ok && !strings.HasPrefix(strings.TrimSpace(l), "#") && strings.TrimSpace(k) == key {
			if !replaced {
				out = append(out, line)
				replaced = true
			}
			continue
		}
		if l != "" || len(out) > 0 {
			out = append(out, l)
		}
	}
	if !replaced {
		out = append(out, line)
	}
	data := strings.Join(out, "\n") + "\n"
	all := Defaults()
	if _, err := all.Apply([]byte(data), p, false); err != nil {
		return "", fmt.Errorf("the existing file %s has an error; fix or delete it first: %w", p, err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return "", err
	}
	// A new file with a random name: a planted file or link named like
	// the temporary file cannot redirect the write.
	f, err := os.CreateTemp(filepath.Dir(p), ".numb-form.conf.*.tmp")
	if err != nil {
		return "", err
	}
	tmp := f.Name()
	_, err = f.WriteString(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp, p)
	}
	if err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return p, nil
}
