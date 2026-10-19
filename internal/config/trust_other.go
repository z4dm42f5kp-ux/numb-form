//go:build !unix

package config

// untrusted trusts the system file on Windows because it lives under
// Program Files, which only administrators can write by default.
func untrusted(string) string { return "" }
