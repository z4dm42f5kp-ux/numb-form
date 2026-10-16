# Installing numb-form

numb-form is two programs, `numb-form` and `numb-commit`.
They are static binaries with no dependencies.

There are two ways to install them:

* Download a signed release binary.
* Build from source with Go 1.21 or later.

Every command below is a single line.
Copy it whole.
It stops at the first step that fails.

## Windows 10 and 11

Run these in **Command Prompt** (`cmd.exe`).
`curl.exe` and `tar.exe` are included in Windows 10 1803 and later.

### Release binary

Use `amd64` for most PCs and `arm64` for ARM devices (for example, Surface Pro X).
This downloads the release, checks its SHA-256 checksum against
`SHA256SUMS` and unpacks it into `%USERPROFILE%\numb-form`:

```
(if not exist "%USERPROFILE%\numb-form" mkdir "%USERPROFILE%\numb-form") && cd /d "%USERPROFILE%\numb-form" && curl.exe -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_windows_amd64.zip && curl.exe -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && powershell -NoProfile -Command "$h=(Get-FileHash numb-form_windows_amd64.zip -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch ($h+'  numb-form_windows_amd64.zip') -Quiet)) { Write-Error 'CHECKSUM MISMATCH'; exit 1 }" && tar -xf numb-form_windows_amd64.zip && del numb-form_windows_amd64.zip && echo OK
```

To also verify who built the release, install the GitHub CLI and check the
build provenance attestation:

```
winget install --id GitHub.cli -e && cd /d "%USERPROFILE%\numb-form" && gh attestation verify numb-form.exe --repo z4dm42f5kp-ux/numb-form && gh attestation verify numb-commit.exe --repo z4dm42f5kp-ux/numb-form
```

Add the folder to your user `PATH` (takes effect in new windows):

```
powershell -NoProfile -Command "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path','User') + ';' + $env:USERPROFILE + '\numb-form', 'User')" && echo Open a new Command Prompt
```

### From source

Install Go with winget, then open a **new** Command Prompt so `go` is on the
`PATH`:

```
winget install --id GoLang.Go -e
```

Build in a temporary Go cache that is deleted afterwards, so no build cache
is left on disk:

```
cd /d "%USERPROFILE%" && curl.exe -fL -o numb-form-src.zip https://github.com/z4dm42f5kp-ux/numb-form/archive/refs/heads/main.zip && tar -xf numb-form-src.zip && del numb-form-src.zip && cd numb-form-main && set "GOCACHE=%TEMP%\numb-form-gocache" && set "GOFLAGS=-trimpath -mod=readonly" && set "CGO_ENABLED=0" && go test ./... && go build -o "%USERPROFILE%\numb-form\numb-form.exe" ./cmd/numb-form && go build -o "%USERPROFILE%\numb-form\numb-commit.exe" ./cmd/numb-commit && go clean -cache && cd /d "%USERPROFILE%" && rmdir /s /q numb-form-main && echo OK
```

The project has no third-party Go modules, so nothing is downloaded during
the build. See [SECURITY.md](../SECURITY.md).

## Release binaries on other systems

Archives are named `numb-form_<os>_<arch>.tar.gz`, where *os* is
`linux`, `darwin` (macOS), `freebsd`, `openbsd` or `netbsd`, and *arch* is
`amd64` or `arm64` (Linux also has `386` and `armv7`).
Each archive holds `numb-form`, `numb-commit`, `README.md` and `LICENSE`.
The examples use `amd64`; replace it to suit the machine.

Every command downloads into a temporary directory, verifies the SHA-256
checksum, installs into `~/.local/bin` and removes the temporary directory.
Make sure `~/.local/bin` is in `PATH`.

### Linux (Debian, Ubuntu, Arch, Fedora, Gentoo, Void, Alpine)

Needs `curl`, `tar` and `sha256sum`, which every listed distribution
provides (Alpine through BusyBox).

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_linux_amd64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' numb-form_linux_amd64.tar.gz$' SHA256SUMS | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### macOS

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_darwin_arm64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' numb-form_darwin_arm64.tar.gz$' SHA256SUMS | shasum -a 256 -c - && tar -xzf numb-form_darwin_arm64.tar.gz && { xattr -d com.apple.quarantine numb-form numb-commit 2>/dev/null || true; } && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

Use `darwin_amd64` on Intel Macs.
The binaries are not notarised by Apple;
the `xattr` step removes the download quarantine flag after the checksum
has been verified.

### FreeBSD

```
cd "$(mktemp -d)" && fetch https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_freebsd_amd64.tar.gz && fetch https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && sha256 -c "$(awk '$2=="numb-form_freebsd_amd64.tar.gz"{print $1}' SHA256SUMS)" numb-form_freebsd_amd64.tar.gz && tar -xzf numb-form_freebsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### OpenBSD

The `SHA256` file uses the BSD format that sha256(1) checks directly.

```
cd "$(mktemp -d)" && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_openbsd_amd64.tar.gz && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256 && sha256 -C SHA256 numb-form_openbsd_amd64.tar.gz && tar -xzf numb-form_openbsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### NetBSD

```
cd "$(mktemp -d)" && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_netbsd_amd64.tar.gz && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256 && test "$(cksum -a SHA256 numb-form_netbsd_amd64.tar.gz)" = "$(grep '(numb-form_netbsd_amd64.tar.gz)' SHA256)" && tar -xzf numb-form_netbsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### Verifying who built a release

Checksums prove the download is intact.
To prove it was built by this repository's release workflow, verify the
GitHub build provenance attestation with the GitHub CLI:

```
gh attestation verify numb-form_linux_amd64.tar.gz --repo z4dm42f5kp-ux/numb-form
```

or the Sigstore signature of the checksum file with cosign:

```
cosign verify-blob --bundle SHA256SUMS.sigstore.json --certificate-identity-regexp '^https://github.com/z4dm42f5kp-ux/numb-form/\.github/workflows/release\.yml@refs/tags/v' --certificate-oidc-issuer https://token.actions.githubusercontent.com SHA256SUMS
```

## Building from source on other systems

Install Go 1.21 or later and git with the system package manager, then
run the build command below.

| System | Install Go and git |
| --- | --- |
| Debian 13, Ubuntu 24.04 or later | `sudo apt-get update && sudo apt-get install -y golang-go git` |
| Debian 12 | `echo 'deb http://deb.debian.org/debian bookworm-backports main' \| sudo tee /etc/apt/sources.list.d/backports.list && sudo apt-get update && sudo apt-get install -y -t bookworm-backports golang-go && sudo apt-get install -y git` |
| Arch Linux (official repositories only) | `sudo pacman -S --needed go git` |
| Fedora, RHEL, Rocky, Alma | `sudo dnf install -y golang git` |
| Gentoo | `sudo emerge --ask --noreplace dev-lang/go dev-vcs/git` |
| Void Linux | `sudo xbps-install -Sy go git` |
| Alpine Linux | `doas apk add go git` |
| macOS | `brew install go git` |
| FreeBSD | `su -m root -c 'pkg install -y go git'` |
| OpenBSD | `doas pkg_add go git` |
| NetBSD | `su -m root -c 'pkgin -y install go git' && PATH=$(ls -d /usr/pkg/go1*/bin \| tail -n 1):$PATH && export PATH` |

Then build, test and install, using a temporary Go build cache that is
deleted afterwards:

```
cd "$(mktemp -d)" && git clone --depth 1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$(mktemp -d)" GOFLAGS='-trimpath -mod=readonly' CGO_ENABLED=0 && go test ./... && go build -o numb-form ./cmd/numb-form && go build -o numb-commit ./cmd/numb-commit && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && go clean -cache && rm -rf "$GOCACHE" && cd / && rm -rf "$OLDPWD/.." && echo OK
```

On NetBSD, pkgsrc installs Go in a versioned directory such as
`/usr/pkg/go126/bin`; the table entry above adds the newest one to `PATH`
for the current shell.

If the installed Go is older than 1.21, the build stops with an error;
install a newer Go from the backports or ports collection rather than
letting it download a toolchain.

The project has no third-party modules: `go.mod` has no `require` lines
and the build downloads nothing.

## Containers

See [containers.md](containers.md) for Docker, Podman and Kubernetes.

### Optional free API key

The UK Companies House register needs a free key.
All other sources need none.
Run `numb-form keys` for instructions.
