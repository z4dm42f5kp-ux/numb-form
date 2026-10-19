// Package update replaces the installed numb-form and numb-commit
// programs with those of the newest GitHub release (stable) or of the
// edge development build. It does what the install commands in
// docs/install.md do: download the archive and SHA256SUMS over HTTPS,
// check the checksum, and only then replace the programs.
package update

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Releases is the base address of the project's releases.
const Releases = "https://github.com/z4dm42f5kp-ux/numb-form/releases"

const (
	maxArchive = 64 << 20 // largest archive accepted
	maxSums    = 1 << 20  // largest checksum file accepted
	maxBinary  = 64 << 20 // largest program accepted from the archive
)

// Programs are the files replaced by an update, without extension.
var Programs = []string{"numb-form", "numb-commit"}

// Options describes one update.
type Options struct {
	Channel string // "stable" or "edge"
	Dir     string // folder holding the installed programs
	GOOS    string
	GOARCH  string
	GOARM   string // "7" for 32-bit ARM builds
	Current string // version of the running program
	Out     io.Writer

	// Get downloads a URL; nil means HTTPS from GitHub only.
	Get func(ctx context.Context, u string, limit int64) ([]byte, error)
	// Probe returns the version printed by "program -version"; nil runs it.
	Probe func(ctx context.Context, program string) (string, error)
}

// DefaultChannel is the channel the running build came from.
func DefaultChannel(version string) string {
	if strings.HasPrefix(version, "edge-") {
		return "edge"
	}
	return "stable"
}

// Base returns the download address of a channel.
func Base(channel string) (string, error) {
	switch channel {
	case "stable":
		return Releases + "/latest/download", nil
	case "edge":
		return Releases + "/download/edge", nil
	}
	return "", fmt.Errorf("unknown channel %q (use stable or edge)", channel)
}

// Asset returns the archive name of a platform, as scripts/build-release.sh
// names it.
func Asset(goos, goarch, goarm string) (string, error) {
	arch := goarch
	if goarch == "arm" {
		if goos != "linux" || (goarm != "" && goarm != "7") {
			return "", fmt.Errorf("no release archive for %s/arm", goos)
		}
		arch = "armv7"
	}
	ok := map[string][]string{
		"linux":   {"amd64", "arm64", "386", "armv7"},
		"darwin":  {"amd64", "arm64"},
		"windows": {"amd64", "arm64", "386"},
		"freebsd": {"amd64", "arm64"},
		"openbsd": {"amd64", "arm64"},
		"netbsd":  {"amd64", "arm64"},
	}
	for _, a := range ok[goos] {
		if a == arch {
			ext := ".tar.gz"
			if goos == "windows" {
				ext = ".zip"
			}
			return "numb-form_" + goos + "_" + arch + ext, nil
		}
	}
	return "", fmt.Errorf("no release archive for %s/%s", goos, goarch)
}

// Run performs the update and returns the installed version.
func Run(ctx context.Context, o Options) (string, error) {
	if o.Current == "dev" {
		return "", errors.New("this numb-form was built from source; update the source and build it again")
	}
	base, err := Base(o.Channel)
	if err != nil {
		return "", err
	}
	asset, err := Asset(o.GOOS, o.GOARCH, o.GOARM)
	if err != nil {
		return "", err
	}
	if o.Get == nil {
		o.Get = httpsGet
	}
	if o.Probe == nil {
		o.Probe = probe
	}
	if err := writable(o.Dir); err != nil {
		return "", err
	}
	fmt.Fprintf(o.Out, "Downloading %s (%s) ...\n", asset, o.Channel)
	sums, err := o.Get(ctx, base+"/SHA256SUMS", maxSums)
	if err != nil {
		return "", fmt.Errorf("download SHA256SUMS: %w", err)
	}
	want, err := lookupSum(sums, asset)
	if err != nil {
		return "", err
	}
	arc, err := o.Get(ctx, base+"/"+asset, maxArchive)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", asset, err)
	}
	got := sha256.Sum256(arc)
	if hex.EncodeToString(got[:]) != want {
		return "", errors.New("CHECKSUM MISMATCH: the download does not match SHA256SUMS; nothing was changed (if this is the edge build, a new one may be uploading: try again in five minutes)")
	}
	ext := ""
	if o.GOOS == "windows" {
		ext = ".exe"
	}
	files, err := extract(arc, asset, ext)
	if err != nil {
		return "", err
	}
	return install(ctx, o, files, ext)
}

// install writes the new programs next to the old ones, checks that the
// new numb-form starts, and then swaps them in.
func install(ctx context.Context, o Options, files map[string][]byte, ext string) (string, error) {
	var staged []string
	defer func() {
		for _, p := range staged {
			_ = os.Remove(p)
		}
	}()
	for _, name := range Programs {
		p := filepath.Join(o.Dir, name+".new"+ext)
		if err := os.WriteFile(p, files[name], 0o755); err != nil { // #nosec G306 -- programs must be executable.
			return "", err
		}
		staged = append(staged, p)
	}
	v, err := o.Probe(ctx, staged[0])
	if err != nil {
		return "", fmt.Errorf("the downloaded numb-form does not start (%w); nothing was changed", err)
	}
	if v == o.Current {
		return v, nil
	}
	for i, name := range Programs {
		dst := filepath.Join(o.Dir, name+ext)
		if o.GOOS == "windows" {
			// A running .exe cannot be overwritten but can be renamed;
			// the .old file is removed at the next start.
			old := filepath.Join(o.Dir, name+".old"+ext)
			_ = os.Remove(old)
			if err := os.Rename(dst, old); err != nil && !errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("replace %s: %w (close every numb-form window and try again)", filepath.Base(dst), err)
			}
		}
		if err := os.Rename(staged[i], dst); err != nil {
			return "", fmt.Errorf("replace %s: %w", filepath.Base(dst), err)
		}
	}
	staged = nil
	return v, nil
}

// Cleanup removes the .old programs an earlier update left on Windows.
func Cleanup(dir string) {
	for _, name := range Programs {
		_ = os.Remove(filepath.Join(dir, name+".old.exe"))
	}
}

func writable(dir string) error {
	f, err := os.CreateTemp(dir, ".numb-form-update-*")
	if err != nil {
		return fmt.Errorf("cannot write to %s, where numb-form is installed (%w); ask whoever installed it to update it", dir, err)
	}
	name := f.Name()
	_ = f.Close()
	return os.Remove(name)
}

func lookupSum(sums []byte, asset string) (string, error) {
	sc := bufio.NewScanner(bytes.NewReader(sums))
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == asset {
			s := strings.ToLower(f[0])
			if b, err := hex.DecodeString(s); err == nil && len(b) == sha256.Size {
				return s, nil
			}
		}
	}
	return "", fmt.Errorf("SHA256SUMS has no entry for %s", asset)
}

// extract returns the programs from a release archive. Only the two
// expected files at the top level are read; everything else is ignored.
func extract(arc []byte, asset, ext string) (map[string][]byte, error) {
	want := map[string]string{}
	for _, p := range Programs {
		want[p+ext] = p
	}
	out := map[string][]byte{}
	take := func(name string, r io.Reader) error {
		p, ok := want[name]
		if !ok || out[p] != nil {
			return nil
		}
		b, err := io.ReadAll(io.LimitReader(r, maxBinary+1))
		if err != nil {
			return err
		}
		if len(b) > maxBinary {
			return fmt.Errorf("%s in the archive is too large", name)
		}
		out[p] = b
		return nil
	}
	if strings.HasSuffix(asset, ".zip") {
		zr, err := zip.NewReader(bytes.NewReader(arc), int64(len(arc)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if f.FileInfo().IsDir() || path.Clean(f.Name) != f.Name || strings.Contains(f.Name, "/") {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			err = take(f.Name, rc)
			_ = rc.Close()
			if err != nil {
				return nil, err
			}
		}
	} else {
		gz, err := gzip.NewReader(bytes.NewReader(arc))
		if err != nil {
			return nil, err
		}
		tr := tar.NewReader(gz)
		for {
			h, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			if h.Typeflag != tar.TypeReg || strings.Contains(strings.TrimPrefix(h.Name, "./"), "/") {
				continue
			}
			if err := take(strings.TrimPrefix(h.Name, "./"), tr); err != nil {
				return nil, err
			}
		}
	}
	for _, p := range Programs {
		if len(out[p]) == 0 {
			return nil, fmt.Errorf("the archive has no %s%s", p, ext)
		}
	}
	return out, nil
}

// allowedHost reports whether downloads may come from host: GitHub and
// the hosts it serves release files from.
func allowedHost(host string) bool {
	host = strings.ToLower(host)
	return host == "github.com" || strings.HasSuffix(host, ".githubusercontent.com")
}

func httpsGet(ctx context.Context, u string, limit int64) ([]byte, error) {
	pu, err := url.Parse(u)
	if err != nil || pu.Scheme != "https" || !allowedHost(pu.Hostname()) {
		return nil, fmt.Errorf("refusing to download from %s", u)
	}
	hc := &http.Client{
		Timeout: 5 * time.Minute,
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
			TLSHandshakeTimeout: 15 * time.Second,
			ForceAttemptHTTP2:   true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "https" || !allowedHost(req.URL.Hostname()) {
				return fmt.Errorf("refusing redirect to %s", req.URL.Host)
			}
			return nil
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "numb-form-update")
	resp, err := hc.Do(req) // #nosec G107 -- scheme and host are checked above.
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d", path.Base(pu.Path), resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("%s is larger than %d MB", path.Base(pu.Path), limit>>20)
	}
	return b, nil
}

func probe(ctx context.Context, program string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, program, "-version").Output() // #nosec G204 -- the checksummed program just downloaded.
	if err != nil {
		return "", err
	}
	f := strings.Fields(string(out))
	if len(f) != 2 || f[0] != "numb-form" {
		return "", fmt.Errorf("unexpected -version output %q", strings.TrimSpace(string(out)))
	}
	return f[1], nil
}
