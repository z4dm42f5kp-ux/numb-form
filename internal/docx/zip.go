package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
)

const maxPart = 64 << 20

func mainXML(docx []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		return nil, err
	}
	main, err := mainPart(zr)
	if err != nil {
		return nil, err
	}
	return readEntry(zr, main)
}

func readEntry(zr *zip.Reader, name string) ([]byte, error) {
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		b, err := io.ReadAll(io.LimitReader(rc, maxPart+1))
		if err != nil {
			return nil, err
		}
		if len(b) > maxPart {
			return nil, fmt.Errorf("%s is too large", name)
		}
		return b, nil
	}
	return nil, fmt.Errorf("%s not found", name)
}

func mainPart(zr *zip.Reader) (string, error) {
	rels, err := readEntry(zr, "_rels/.rels")
	if err != nil {
		return "", errors.New("not a .docx file (_rels/.rels missing)")
	}
	var r struct {
		Rel []struct {
			Type   string `xml:"Type,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if err := xml.Unmarshal(rels, &r); err != nil {
		return "", fmt.Errorf("_rels/.rels: %w", err)
	}
	for _, x := range r.Rel {
		if strings.HasSuffix(x.Type, "/officeDocument") {
			return strings.TrimPrefix(path.Clean("/"+x.Target), "/"), nil
		}
	}
	return "", errors.New("main document part not found")
}

// rewrite copies every zip entry byte-for-byte except the main document
// part, which is replaced by fn's output. It then re-reads the result and
// checks that all other entries are identical.
func rewrite(src []byte, fn func([]byte) ([]byte, []string, error)) ([]byte, []string, error) {
	zr, err := zip.NewReader(bytes.NewReader(src), int64(len(src)))
	if err != nil {
		return nil, nil, fmt.Errorf("template is not a valid .docx (zip): %w", err)
	}
	main, err := mainPart(zr)
	if err != nil {
		return nil, nil, err
	}
	ct, err := readEntry(zr, "[Content_Types].xml")
	if err != nil {
		return nil, nil, err
	}
	if bytes.Contains(ct, []byte("macroEnabled")) {
		return nil, nil, errors.New("template contains macros; refusing to use it")
	}
	doc, err := readEntry(zr, main)
	if err != nil {
		return nil, nil, err
	}
	newDoc, warnings, err := fn(doc)
	if err != nil {
		return nil, nil, err
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range zr.File {
		if f.Name == main {
			w, err := zw.CreateHeader(&zip.FileHeader{Name: f.Name, Method: zip.Deflate, Modified: f.Modified})
			if err != nil {
				return nil, nil, err
			}
			if _, err := w.Write(newDoc); err != nil {
				return nil, nil, err
			}
			continue
		}
		if err := zw.Copy(f); err != nil {
			return nil, nil, fmt.Errorf("copy %s: %w", f.Name, err)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, nil, err
	}
	out := buf.Bytes()
	zo, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	if err != nil {
		return nil, nil, fmt.Errorf("self-check: %w", err)
	}
	if len(zo.File) != len(zr.File) {
		return nil, nil, errors.New("self-check: entry count changed")
	}
	for i, f := range zr.File {
		g := zo.File[i]
		if g.Name != f.Name {
			return nil, nil, errors.New("self-check: entry order changed")
		}
		if f.Name == main {
			continue
		}
		if g.CRC32 != f.CRC32 || g.UncompressedSize64 != f.UncompressedSize64 {
			return nil, nil, fmt.Errorf("self-check: %s changed", f.Name)
		}
	}
	got, err := readEntry(zo, main)
	if err != nil || !bytes.Equal(got, newDoc) {
		return nil, nil, errors.New("self-check: document part did not round-trip")
	}
	return out, warnings, nil
}
