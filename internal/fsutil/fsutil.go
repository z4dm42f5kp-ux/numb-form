// Package fsutil holds file-system helpers: locating the Downloads
// folder, copying trees, hashing and safe file names.
package fsutil

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DownloadsDir returns the user's Downloads folder. The NUMBFORM_DOWNLOADS
// environment variable overrides it (useful in containers).
func DownloadsDir() (string, error) {
	if d := strings.TrimSpace(os.Getenv("NUMBFORM_DOWNLOADS")); d != "" {
		return filepath.Abs(d)
	}
	return downloadsDir()
}

// SHA256File returns the hex SHA-256 of a file.
func SHA256File(p string) (string, int64, error) {
	f, err := os.Open(p) // #nosec G304 -- paths come from the user's own tree.
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

// SHA256Bytes returns the hex SHA-256 of b.
func SHA256Bytes(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// Stats summarises a tree.
type Stats struct {
	Files, Dirs int
	Bytes       int64
	Skipped     []string
}

// Measure walks root and counts what CopyTree would copy.
func Measure(ctx context.Context, root string) (Stats, error) {
	var st Stats
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		switch {
		case d.IsDir():
			st.Dirs++
		case d.Type().IsRegular():
			if skipFile(d.Name()) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			st.Files++
			st.Bytes += info.Size()
		}
		return nil
	})
	return st, err
}

func skipFile(name string) bool {
	// Office owner/lock files: transient and often unreadable.
	return strings.HasPrefix(name, "~$")
}

// CopyTree copies the regular files and directories of src into dst,
// which must not exist yet. Symbolic links and special files are
// skipped and reported. Modification times are preserved.
func CopyTree(ctx context.Context, src, dst string, logf func(string, ...any)) (Stats, error) {
	var st Stats
	src, err := filepath.Abs(src)
	if err != nil {
		return st, err
	}
	dst, err = filepath.Abs(dst)
	if err != nil {
		return st, err
	}
	if rel, err := filepath.Rel(src, dst); err == nil && !strings.HasPrefix(rel, "..") {
		return st, errors.New("destination is inside the source tree")
	}
	if _, err := os.Lstat(dst); err == nil {
		return st, fmt.Errorf("%s already exists", dst)
	}
	buf := make([]byte, 1<<20)
	err = filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			st.Dirs++
			// Keep the source permissions so shared folders stay shared.
			return os.MkdirAll(target, info.Mode().Perm()|0o700)
		case d.Type().IsRegular():
			if skipFile(d.Name()) {
				st.Skipped = append(st.Skipped, rel)
				logf("skip lock file %s", rel)
				return nil
			}
			n, err := copyFile(p, target, info, buf)
			if err != nil {
				return fmt.Errorf("copy %s: %w", rel, err)
			}
			st.Files++
			st.Bytes += n
			return nil
		default:
			st.Skipped = append(st.Skipped, rel)
			logf("skip non-regular file %s (%s)", rel, d.Type())
			return nil
		}
	})
	if err != nil {
		return st, err
	}
	// Directory times last, after their contents were written.
	_ = filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(src, p)
		if info, err := d.Info(); err == nil {
			_ = os.Chtimes(filepath.Join(dst, rel), info.ModTime(), info.ModTime())
		}
		return nil
	})
	return st, nil
}

func copyFile(src, dst string, info fs.FileInfo, buf []byte) (int64, error) {
	in, err := os.Open(src) // #nosec G304 -- walking the user's own tree.
	if err != nil {
		return 0, err
	}
	defer in.Close()
	// Keep the source permissions (the copy must match the original).
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm()|0o600) // #nosec G302 G304 -- mirrors the source file.
	if err != nil {
		return 0, err
	}
	n, err := io.CopyBuffer(out, in, buf)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return n, err
	}
	if n != info.Size() {
		return n, fmt.Errorf("size mismatch: copied %d of %d bytes", n, info.Size())
	}
	return n, os.Chtimes(dst, info.ModTime(), info.ModTime())
}

// WriteNew writes data to path, which must not exist, via a temporary
// file in the same directory and a rename.
func WriteNew(path string, data []byte) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}
	dir := filepath.Dir(path)
	tmp := filepath.Join(dir, ".numb-form-tmp-"+randHex(6))
	// Business documents on a team share must stay readable by the team;
	// the process umask still applies.
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644) // #nosec G302 G304
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		_ = os.Remove(tmp)
	}
	return err
}

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
	"com1": true, "com2": true, "com3": true, "com4": true, "com5": true, "com6": true, "com7": true, "com8": true, "com9": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true, "lpt5": true, "lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
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
	base := strings.ToLower(strings.SplitN(s, ".", 2)[0])
	if reserved[base] {
		s = "_" + s
	}
	if s == "" {
		s = "_"
	}
	return s, s != name
}
