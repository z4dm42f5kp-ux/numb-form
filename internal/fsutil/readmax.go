package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ReadFileMax reads the file at path, which must not be larger than max
// bytes, so that a huge file cannot fill the memory.
func ReadFileMax(path string, max int64) ([]byte, error) {
	f, err := os.Open(path) // #nosec G304 -- callers name files of the working folders.
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, fmt.Errorf("%s is larger than %d MB", filepath.Base(path), max>>20)
	}
	return b, nil
}
