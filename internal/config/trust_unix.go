//go:build unix

package config

import (
	"os"
	"syscall"
)

// untrusted explains why a system file must be ignored: it must belong
// to root and must not be writable by group or others.
func untrusted(p string) string {
	st, err := os.Stat(p)
	if err != nil {
		return err.Error()
	}
	if st.Mode().Perm()&0o022 != 0 {
		return "writable by group or others"
	}
	if s, ok := st.Sys().(*syscall.Stat_t); ok && s.Uid != 0 {
		return "not owned by root"
	}
	return ""
}
