// Safe file names and random names.

package fsutil

import (
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
	"strings"
)

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// RandHex returns n random bytes as hex.
func RandHex(n int) string { return randHex(n) }

var reserved = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com0": true, "com1": true, "com2": true, "com3": true, "com4": true, "com5": true, "com6": true, "com7": true, "com8": true, "com9": true,
	"lpt0": true, "lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true, "lpt5": true, "lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
	"com\u00b9": true, "com\u00b2": true, "com\u00b3": true, "lpt\u00b9": true, "lpt\u00b2": true, "lpt\u00b3": true,
}

// SafeName turns a literal company name into a name that is valid on
// Windows, macOS, Linux and the BSDs, changing as little as possible.
// Windows rules are applied everywhere because the tree may live on a
// Windows file share. It reports whether anything changed.
func SafeName(name string, file bool) (string, bool) {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r < 0x20, r == 0x7f:
			b.WriteRune('_')
		case strings.ContainsRune(`<>:"/\|?*`, r):
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	s := strings.TrimSpace(b.String())
	if !file {
		// Windows silently drops trailing dots and spaces on folders.
		s = strings.TrimRight(s, ". ")
	}
	// Windows also ignores spaces before the extension: "nul .txt" is "nul".
	base := strings.TrimRight(strings.ToLower(strings.SplitN(s, ".", 2)[0]), " ")
	if reserved[base] {
		s = "_" + s
	}
	if s == "" {
		s = "_"
	}
	return s, s != name
}

// IsUnder reports whether path p is parent itself or inside it.
func IsUnder(p, parent string) bool {
	ap, err1 := filepath.Abs(p)
	aq, err2 := filepath.Abs(parent)
	if err1 != nil || err2 != nil {
		return false
	}
	r, err := filepath.Rel(aq, ap)
	return err == nil && r != ".." && !strings.HasPrefix(r, ".."+string(filepath.Separator)) && !filepath.IsAbs(r)
}
