package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func zipOf(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.WriteString(w, body)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func tgzOf(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		_, _ = io.WriteString(tw, body)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sumLine(b []byte, name string) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:]) + "  " + name + "\n"
}

type fake struct {
	files map[string][]byte
	urls  []string
}

func (f *fake) get(_ context.Context, u string, _ int64) ([]byte, error) {
	f.urls = append(f.urls, u)
	b, ok := f.files[u[strings.LastIndex(u, "/")+1:]]
	if !ok {
		return nil, fmt.Errorf("404 %s", u)
	}
	return b, nil
}

func opts(dir, goos string, f *fake, version string) Options {
	return Options{
		Channel: "edge", Dir: dir, GOOS: goos, GOARCH: "amd64", Current: "v0.1.1", Out: io.Discard,
		Get: f.get,
		Probe: func(_ context.Context, p string) (string, error) {
			if _, err := os.Stat(p); err != nil {
				return "", err
			}
			return version, nil
		},
	}
}

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestWindowsUpdate(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "numb-form.exe"), []byte("old form"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "numb-commit.exe"), []byte("old commit"), 0o755)
	arc := zipOf(t, map[string]string{"numb-form.exe": "new form", "numb-commit.exe": "new commit", "README.md": "r", "../evil.exe": "x"})
	name := "numb-form_windows_amd64.zip"
	f := &fake{files: map[string][]byte{name: arc, "SHA256SUMS": []byte("00  other.zip\n" + sumLine(arc, name))}}
	v, err := Run(context.Background(), opts(dir, "windows", f, "edge-abc1234"))
	if err != nil || v != "edge-abc1234" {
		t.Fatalf("v %q err %v", v, err)
	}
	if read(t, filepath.Join(dir, "numb-form.exe")) != "new form" || read(t, filepath.Join(dir, "numb-commit.exe")) != "new commit" {
		t.Error("programs not replaced")
	}
	if read(t, filepath.Join(dir, "numb-form.old.exe")) != "old form" {
		t.Error("old program not kept for later removal")
	}
	if f.urls[0] != Releases+"/download/edge/SHA256SUMS" {
		t.Errorf("urls %v", f.urls)
	}
	Cleanup(dir)
	ents, _ := os.ReadDir(dir)
	var names []string
	for _, e := range ents {
		names = append(names, e.Name())
	}
	if strings.Join(names, " ") != "numb-commit.exe numb-form.exe" {
		t.Errorf("left behind: %v", names)
	}
}

func TestUnixUpdateAndUpToDate(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "numb-form"), []byte("old"), 0o755)
	arc := tgzOf(t, map[string]string{"numb-form": "new form", "numb-commit": "new commit", "LICENSE": "l"})
	name := "numb-form_linux_amd64.tar.gz"
	f := &fake{files: map[string][]byte{name: arc, "SHA256SUMS": []byte(sumLine(arc, name))}}
	o := opts(dir, "linux", f, "v0.1.1")
	o.Channel = "stable"
	v, err := Run(context.Background(), o)
	if err != nil || v != "v0.1.1" {
		t.Fatalf("v %q err %v", v, err)
	}
	if read(t, filepath.Join(dir, "numb-form")) != "old" {
		t.Error("same version must not replace anything")
	}
	if f.urls[0] != Releases+"/latest/download/SHA256SUMS" {
		t.Errorf("urls %v", f.urls)
	}
	o.Probe = func(context.Context, string) (string, error) { return "v0.2.0", nil }
	if v, err = Run(context.Background(), o); err != nil || v != "v0.2.0" {
		t.Fatalf("v %q err %v", v, err)
	}
	if read(t, filepath.Join(dir, "numb-form")) != "new form" {
		t.Error("not replaced")
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 2 {
		t.Errorf("%d files left, want 2", len(ents))
	}
}

func TestRefusals(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "numb-form.exe"), []byte("old"), 0o755)
	arc := zipOf(t, map[string]string{"numb-form.exe": "new", "numb-commit.exe": "new"})
	name := "numb-form_windows_amd64.zip"
	bad := sha256.Sum256([]byte("something else"))
	cases := map[string]*fake{
		"CHECKSUM MISMATCH": {files: map[string][]byte{name: arc, "SHA256SUMS": []byte(hex.EncodeToString(bad[:]) + "  " + name + "\n")}},
		"no entry":          {files: map[string][]byte{name: arc, "SHA256SUMS": []byte("")}},
		"has no numb-commit": func() *fake {
			a := zipOf(t, map[string]string{"numb-form.exe": "new", "sub/numb-commit.exe": "x"})
			return &fake{files: map[string][]byte{name: a, "SHA256SUMS": []byte(sumLine(a, name))}}
		}(),
	}
	for want, f := range cases {
		_, err := Run(context.Background(), opts(dir, "windows", f, "v9"))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("want %q, got %v", want, err)
		}
	}
	o := opts(dir, "windows", &fake{files: map[string][]byte{name: arc, "SHA256SUMS": []byte(sumLine(arc, name))}}, "")
	o.Probe = func(context.Context, string) (string, error) { return "", errors.New("exec format error") }
	if _, err := Run(context.Background(), o); err == nil || !strings.Contains(err.Error(), "does not start") {
		t.Errorf("probe failure: %v", err)
	}
	o.Current = "dev"
	if _, err := Run(context.Background(), o); err == nil {
		t.Error("dev build updated")
	}
	if read(t, filepath.Join(dir, "numb-form.exe")) != "old" {
		t.Error("a failed update changed the program")
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 1 {
		t.Errorf("%d files left after failures, want 1", len(ents))
	}
}

func TestAssetAndHosts(t *testing.T) {
	for _, c := range [][4]string{
		{"windows", "amd64", "", "numb-form_windows_amd64.zip"},
		{"linux", "arm", "7", "numb-form_linux_armv7.tar.gz"},
		{"darwin", "arm64", "", "numb-form_darwin_arm64.tar.gz"},
		{"netbsd", "amd64", "", "numb-form_netbsd_amd64.tar.gz"},
	} {
		if a, err := Asset(c[0], c[1], c[2]); err != nil || a != c[3] {
			t.Errorf("%v: %q %v", c, a, err)
		}
	}
	if _, err := Asset("plan9", "amd64", ""); err == nil {
		t.Error("plan9 accepted")
	}
	for h, ok := range map[string]bool{
		"github.com": true, "objects.githubusercontent.com": true, "release-assets.githubusercontent.com": true,
		"evil.com": false, "githubusercontent.com.evil.com": false, "github.com.evil.com": false,
	} {
		if allowedHost(h) != ok {
			t.Errorf("allowedHost(%q) != %v", h, ok)
		}
	}
	if _, err := httpsGet(context.Background(), "http://github.com/x", 1); err == nil {
		t.Error("plain HTTP accepted")
	}
	if DefaultChannel("edge-abc1234") != "edge" || DefaultChannel("v0.1.1") != "stable" {
		t.Error("default channel")
	}
}
