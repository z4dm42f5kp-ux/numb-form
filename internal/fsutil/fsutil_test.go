package fsutil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeName(t *testing.T) {
	cases := []struct {
		in   string
		file bool
		want string
	}{
		{"Foo S.A.", false, "Foo S.A"},
		{"Foo S.A.", true, "Foo S.A."},
		{"A/B: C*", false, "A_B_ C_"},
		{"CON", false, "_CON"},
		{"Foo GmbH", false, "Foo GmbH"},
	}
	for _, c := range cases {
		if got, _ := SafeName(c.in, c.file); got != c.want {
			t.Errorf("SafeName(%q, %v) = %q, want %q", c.in, c.file, got, c.want)
		}
	}
}

func TestCopyTree(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src")
	dst := filepath.Join(t.TempDir(), "dst")
	if err := os.MkdirAll(filepath.Join(src, "a", "b"), 0o750); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"x.txt", "a/y.txt", "a/b/~$lock.docx"} {
		if err := os.WriteFile(filepath.Join(src, f), []byte(f), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	st, err := CopyTree(context.Background(), src, dst, func(string, ...any) {})
	if err != nil {
		t.Fatal(err)
	}
	if st.Files != 2 || len(st.Skipped) != 1 {
		t.Fatalf("stats %+v", st)
	}
	if b, _ := os.ReadFile(filepath.Join(dst, "a", "y.txt")); string(b) != "a/y.txt" {
		t.Fatal("content mismatch")
	}
	if _, err := CopyTree(context.Background(), src, dst, func(string, ...any) {}); err == nil {
		t.Fatal("copy over existing destination succeeded")
	}
	if _, err := CopyTree(context.Background(), src, filepath.Join(src, "inner"), func(string, ...any) {}); err == nil {
		t.Fatal("copy into itself succeeded")
	}
}

func TestWriteNewRefusesOverwrite(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f")
	if err := WriteNew(p, []byte("1")); err != nil {
		t.Fatal(err)
	}
	if err := WriteNew(p, []byte("2")); err == nil {
		t.Fatal("overwrote")
	}
	if b, _ := os.ReadFile(p); string(b) != "1" {
		t.Fatal("content changed")
	}
	ents, _ := os.ReadDir(filepath.Dir(p))
	if len(ents) != 1 {
		t.Fatal("temporary file left behind")
	}
}

func TestReadFileMax(t *testing.T) {
	p := filepath.Join(t.TempDir(), "form.docx")
	if err := os.WriteFile(p, make([]byte, 2<<20), 0o600); err != nil {
		t.Fatal(err)
	}
	if b, err := ReadFileMax(p, 2<<20); err != nil || len(b) != 2<<20 {
		t.Errorf("at the limit: %d bytes, %v", len(b), err)
	}
	if _, err := ReadFileMax(p, 2<<20-1); err == nil || !strings.Contains(err.Error(), "form.docx is larger than") {
		t.Errorf("over the limit: %v", err)
	}
}
