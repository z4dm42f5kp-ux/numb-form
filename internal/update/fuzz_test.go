package update

import "testing"

// FuzzExtract feeds arbitrary release archives and SHA256SUMS files to
// numb-form update.
func FuzzExtract(f *testing.F) {
	f.Add([]byte("PK\x05\x06\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"), "x.zip", ".exe")
	f.Add([]byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff"), "x.tar.gz", "")
	f.Fuzz(func(t *testing.T, b []byte, asset, ext string) {
		_, _ = extract(b, asset, ext)
		_, _ = lookupSum(b, asset)
	})
}
