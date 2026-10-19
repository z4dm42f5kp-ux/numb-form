package form

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
)

// MaxPart caps how many decompressed bytes are read from one archive
// member (zip bomb guard).
const MaxPart = 64 << 20

// ZipEntry returns one member of a zip archive.
func ZipEntry(zr *zip.Reader, name string) ([]byte, error) {
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		b, err := io.ReadAll(io.LimitReader(rc, MaxPart+1))
		if err != nil {
			return nil, err
		}
		if len(b) > MaxPart {
			return nil, fmt.Errorf("%s is too large", name)
		}
		return b, nil
	}
	return nil, fmt.Errorf("%s not found", name)
}

// ReplaceZipEntry rebuilds an archive with one member replaced. All other
// members are copied byte for byte, in the same order (so an ODF
// "mimetype" member stays first and uncompressed), and the result is
// re-read and compared before it is returned.
func ReplaceZipEntry(src []byte, name string, data []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(src), int64(len(src)))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range zr.File {
		if f.Name == name {
			w, err := zw.CreateHeader(&zip.FileHeader{Name: f.Name, Method: zip.Deflate, Modified: f.Modified})
			if err != nil {
				return nil, err
			}
			if _, err := w.Write(data); err != nil {
				return nil, err
			}
			continue
		}
		if err := zw.Copy(f); err != nil {
			return nil, fmt.Errorf("copy %s: %w", f.Name, err)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	out := buf.Bytes()
	zo, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	if err != nil || len(zo.File) != len(zr.File) {
		return nil, errors.New("self-check: archive did not round-trip")
	}
	for i, f := range zr.File {
		g := zo.File[i]
		if g.Name != f.Name {
			return nil, errors.New("self-check: member order changed")
		}
		if f.Name != name && (g.CRC32 != f.CRC32 || g.UncompressedSize64 != f.UncompressedSize64 || g.Method != f.Method) {
			return nil, fmt.Errorf("self-check: %s changed", f.Name)
		}
	}
	if got, err := ZipEntry(zo, name); err != nil || !bytes.Equal(got, data) {
		return nil, errors.New("self-check: replaced member did not round-trip")
	}
	return out, nil
}
