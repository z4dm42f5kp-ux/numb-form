// Package update replaces the installed numb-form and numb-commit
// programs with those of the newest GitHub release (stable) or of the
// edge development build. It does what the install commands in
// docs/install.md do: download the archive and SHA256SUMS over HTTPS,
// check the checksum, and only then replace the programs.
package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	Channel string // "stable" or "rolling"
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
	if strings.HasPrefix(version, "rolling-") || strings.HasPrefix(version, "edge-") {
		return "rolling"
	}
	return "stable"
}

// reTag matches a release number such as v0.1.1.
var reTag = regexp.MustCompile(`^v\d+\.\d+(?:\.\d+)?(?:-[0-9A-Za-z.]+)?$`)

// Channel normalises a channel name: "edge" is the older name of
// "rolling"; a release number such as v0.1.1 names that release.
func Channel(name string) (string, bool) {
	n := strings.ToLower(strings.TrimSpace(name))
	switch n {
	case "stable":
		return "stable", true
	case "rolling", "edge":
		return "rolling", true
	}
	if !strings.HasPrefix(n, "v") {
		n = "v" + n
	}
	if reTag.MatchString(n) {
		return n, true
	}
	return "", false
}

// Base returns the download address of a channel: stable is the newest
// published release, rolling the build of the newest change on main
// (kept under the release tag "edge").
func Base(channel string) (string, error) {
	switch c, _ := Channel(channel); c {
	case "stable":
		return Releases + "/latest/download", nil
	case "rolling":
		return Releases + "/download/edge", nil
	case "":
	default:
		return Releases + "/download/" + url.PathEscape(c), nil
	}
	return "", fmt.Errorf("unknown channel %q (use stable, rolling or a release number such as v0.1.1)", channel)
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
		return "", errors.New("CHECKSUM MISMATCH: the download does not match SHA256SUMS; nothing was changed (for rolling, a new build may be uploading: try again in five minutes)")
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
		// The version being replaced is kept as name.previous for
		// "numb-form rollback". A running .exe cannot be overwritten but
		// can be renamed, so this also works on Windows.
		prev := PreviousPath(o.Dir, name, ext)
		_ = os.Remove(prev)
		if err := os.Rename(dst, prev); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("replace %s: %w (close every numb-form window and try again)", filepath.Base(dst), err)
		}
		if err := os.Rename(staged[i], dst); err != nil {
			_ = os.Rename(prev, dst) // put the old one back
			return "", fmt.Errorf("replace %s: %w", filepath.Base(dst), err)
		}
	}
	staged = nil
	return v, nil
}

// PreviousPath is where an update keeps the version it replaced.
func PreviousPath(dir, name, ext string) string { return filepath.Join(dir, name+".previous"+ext) }

// Cleanup removes the .old programs that versions before rollback
// support left on Windows.
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
