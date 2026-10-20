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
	"os/exec"
	"path"
	"strings"
	"time"
)

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
