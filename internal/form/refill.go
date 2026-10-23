package form

import (
	"archive/zip"
	"bytes"
)

// AsFilled reports whether doc is exactly the template tpl filled with
// the values doc holds in section A: a form that f filled and that nobody
// edited since. It writes no marker into documents; the comparison is
// the mark. Archives (.docx, .odt) are compared member by member,
// uncompressed, so a different compression of the same content counts as
// the same.
func AsFilled(f Filler, tpl, doc []byte) (bool, error) {
	v, err := f.ReadBack(doc)
	if err != nil {
		return false, err
	}
	out, _, err := f.Fill(tpl, v)
	if err != nil {
		return false, err
	}
	return sameContent(out, doc), nil
}

func sameContent(a, b []byte) bool {
	za, errA := zip.NewReader(bytes.NewReader(a), int64(len(a)))
	zb, errB := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if errA != nil || errB != nil {
		return errA != nil && errB != nil && bytes.Equal(a, b)
	}
	if len(za.File) != len(zb.File) {
		return false
	}
	for i, f := range za.File {
		g := zb.File[i]
		if f.Name != g.Name || f.CRC32 != g.CRC32 || f.UncompressedSize64 != g.UncompressedSize64 {
			return false
		}
		x, err1 := ZipEntry(za, f.Name)
		y, err2 := ZipEntry(zb, g.Name)
		if err1 != nil || err2 != nil || !bytes.Equal(x, y) {
			return false
		}
	}
	return true
}
