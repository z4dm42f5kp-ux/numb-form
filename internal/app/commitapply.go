package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
)

// copyChecked copies one file to dst, which must not exist, through a
// temporary name, and checks its SHA-256.
func copyChecked(src, dst, sha string) error {
	tmp := filepath.Join(filepath.Dir(dst), ".numb-commit-"+fsutil.RandHex(4))
	if err := fsutil.CopyFile(src, tmp); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if h := hashOf(tmp); h != sha {
		_ = os.Remove(tmp)
		return fmt.Errorf("%s does not match the request (sha256 %s)", src, h)
	}
	if _, err := os.Lstat(dst); err == nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("%s appeared in the real folder", dst)
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
