# Installing numb-form

Not a programmer?
The [step-by-step guide](guide.md) has the install line for Windows and
explains what to do after installing.

numb-form is four static programs, `numb-form`, `numb-commit`,
`numb-date` and `numb-typo`, with no dependencies.
Every system gets the same programs and the same behaviour.

There are two ways to install them:

* Download a release binary.
  This never needs administrator rights.
* Build from source with Go 1.21 or later, with or without administrator
  rights.

Every command below is a single line.
Copy it whole.
It stops at the first step that fails.

## Release binaries (no administrator rights)

Each command installs one stable release, named in the command, and
checks its SHA-256 checksum against the value written in the command
itself.
A changed or damaged download stops with an error before anything is
installed.
The checksum comes from this page, not from the server the files come
from, so a file replaced on the server cannot pass.
The commands download over HTTPS only, install into a folder you own
and clean up after themselves.

The commands on this page are written by a program from the release's
signed checksums (`scripts/install`).
After every stable release, a workflow checks the release's Sigstore
signature and proposes the new commands; a daily check catches a
missed release.
`numb-form update` installs newer releases the same way.

Archives are named `numb-form_<os>_<arch>`, where *os* is `linux`,
`darwin` (macOS), `windows`, `freebsd`, `openbsd` or `netbsd`, and *arch*
is `amd64` or `arm64` (Linux and Windows also have `386`; Linux has
`armv7`).
For another archive, replace its name in both places of the command and
take its checksum from [the table](#checksums).

### Linux

One command works on every distribution; it needs `curl`, `tar`,
`gzip` and `sha256sum`.

<!-- install:linux -->
```
cd "$(mktemp -d)" && curl --proto '=https' --tlsv1.2 -fsSLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_linux_amd64.tar.gz && echo 'e336d789a9e758ea1d919f2adc5521d14d159572b644ddc325174d01e0dcc6b8  numb-form_linux_amd64.tar.gz' | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:linux -->

On ARM computers (arm64):

<!-- install:linux-arm64 -->
```
cd "$(mktemp -d)" && curl --proto '=https' --tlsv1.2 -fsSLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_linux_arm64.tar.gz && echo 'c44b5089a40a6883e9a1531f0bcde6e9afaf9698f05e857f11f81475042405b0  numb-form_linux_arm64.tar.gz' | sha256sum -c - && tar -xzf numb-form_linux_arm64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:linux-arm64 -->

If the command reports that `curl` or another program is missing,
install it first:

<!-- install:linux-tools -->
| Distribution | Install what the command needs (once) |
| --- | --- |
| Debian 13 and systems based on it | `sudo apt-get update && sudo apt-get install -y curl ca-certificates` |
| Fedora 44 (curl is installed) | `sudo dnf install -y tar gzip` |
| Arch Linux | `sudo pacman -S --needed curl tar gzip` |
| Gentoo | `sudo emerge --ask --noreplace net-misc/curl` |
| Void Linux | `sudo xbps-install -Sy curl` |
| openSUSE Tumbleweed and Leap 16 | `sudo zypper install -y curl tar gzip` |
| Alpine Linux | `doas apk add curl` |
| NixOS | `nix-shell -p curl` (a shell with curl; run the command inside it) |
| Guix System | `guix shell curl tar gzip coreutils` (a shell with them; run the command inside it) |
<!-- /install:linux-tools -->

These distributions are tested before every minor release (vX.Y.0):
Debian 13, Fedora 44, Arch Linux, Gentoo, Void Linux, openSUSE
Tumbleweed and Leap 16, Alpine Linux, NixOS and Guix.

### macOS

Apple silicon (M1 and later):

<!-- install:macos -->
```
cd "$(mktemp -d)" && curl --proto '=https' --tlsv1.2 -fsSLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_darwin_arm64.tar.gz && echo 'bb3284a2412fe68075be90e5f4df81284c536af9b304a10ffd5112cef21cb396  numb-form_darwin_arm64.tar.gz' | shasum -a 256 -c - && tar -xzf numb-form_darwin_arm64.tar.gz && { xattr -d com.apple.quarantine numb-form numb-commit numb-date 2>/dev/null || true; } && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:macos -->

Intel Macs:

<!-- install:macos-intel -->
```
cd "$(mktemp -d)" && curl --proto '=https' --tlsv1.2 -fsSLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_darwin_amd64.tar.gz && echo '54ea02b6c4e055aceed0fefce86677afc043a10cd0ee0cae2fe3532ce0260d0b  numb-form_darwin_amd64.tar.gz' | shasum -a 256 -c - && tar -xzf numb-form_darwin_amd64.tar.gz && { xattr -d com.apple.quarantine numb-form numb-commit numb-date 2>/dev/null || true; } && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:macos-intel -->

The binaries are not notarised by Apple; the `xattr` step removes the
download quarantine flag after the checksum has been verified.

### Windows 10 and 11

Commands for both shells follow.
Use the one that matches the window you have open: Windows PowerShell
(blue, prompt `PS C:\>`) or Command Prompt (black, prompt `C:\>`).
They install into `%LOCALAPPDATA%\Programs\numb-form` and add it to your
own `PATH`; no administrator rights are needed.

Windows PowerShell 5.1 has no `&&`.
The PowerShell commands therefore stop at the first error with
`$ErrorActionPreference = 'Stop'` instead, which has the same effect.

Windows PowerShell:

<!-- install:windows -->
```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Set-Location $d; try { Invoke-WebRequest -UseBasicParsing "https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/$z" -OutFile $z; if ((Get-FileHash $z -Algorithm SHA256).Hash.ToLower() -ne '26f893d6306babb42a2b2a801519e9136f1c91850233cf02206706c7701d8a91') { throw 'CHECKSUM MISMATCH' }; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, open a new PowerShell window' }
```
<!-- /install:windows -->

Command Prompt:

<!-- install:windows-cmd -->
```
(if not exist "%LOCALAPPDATA%\Programs\numb-form" mkdir "%LOCALAPPDATA%\Programs\numb-form") && cd /d "%LOCALAPPDATA%\Programs\numb-form" && curl.exe --proto =https --tlsv1.2 -fsSLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_windows_amd64.zip && powershell -NoProfile -Command "if ((Get-FileHash numb-form_windows_amd64.zip -Algorithm SHA256).Hash.ToLower() -ne '26f893d6306babb42a2b2a801519e9136f1c91850233cf02206706c7701d8a91') { Write-Error 'CHECKSUM MISMATCH'; exit 1 }; $d=$env:LOCALAPPDATA+'\Programs\numb-form'; $p=[Environment]::GetEnvironmentVariable('Path','User'); if (-not $p) { $p='' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';')+';'+$d).TrimStart(';'), 'User') }" && tar -xf numb-form_windows_amd64.zip && del numb-form_windows_amd64.zip && numb-form.exe -version && echo OK, open a new Command Prompt
```
<!-- /install:windows-cmd -->

### Reinstalling or updating on Windows

`numb-form update` does this by itself (see [usage](usage.md)); it works
the same on every system.
If it cannot connect, close every numb-form window, then run this in
Windows PowerShell.
It replaces the programs, keeps your settings and prints the installed
version.

<!-- install:windows-reinstall -->
```
& { $ErrorActionPreference = 'Stop'; if (Get-Process numb-form, numb-commit, numb-date, numb-typo -ErrorAction SilentlyContinue) { throw 'Close every numb-form window first, then run this again' }; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Push-Location $d; try { Invoke-WebRequest -UseBasicParsing "https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/$z" -OutFile $z; if ((Get-FileHash $z -Algorithm SHA256).Hash.ToLower() -ne '26f893d6306babb42a2b2a801519e9136f1c91850233cf02206706c7701d8a91') { throw 'CHECKSUM MISMATCH' }; Remove-Item -Force -ErrorAction SilentlyContinue numb-form.exe, numb-commit.exe, numb-date.exe; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z; Pop-Location }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, reinstalled. Your settings are kept. Open a new PowerShell window' }
```
<!-- /install:windows-reinstall -->

### FreeBSD

<!-- install:freebsd -->
```
cd "$(mktemp -d)" && fetch https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_freebsd_amd64.tar.gz && sha256 -c a535bdcbf5f93320ad0bfae865dd9100dc5f9c95b58591aaa565c5c34458c932 numb-form_freebsd_amd64.tar.gz && tar -xzf numb-form_freebsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:freebsd -->

### OpenBSD

<!-- install:openbsd -->
```
cd "$(mktemp -d)" && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_openbsd_amd64.tar.gz && test "$(sha256 -q numb-form_openbsd_amd64.tar.gz)" = dd0c3b9b04c614c0b1a634ee84fded10011b7b4abdef2539cb88237fa1ec42cb && tar -xzf numb-form_openbsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:openbsd -->

### NetBSD

<!-- install:netbsd -->
```
cd "$(mktemp -d)" && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_netbsd_amd64.tar.gz && test "$(cksum -a SHA256 numb-form_netbsd_amd64.tar.gz)" = 'SHA256 (numb-form_netbsd_amd64.tar.gz) = 949174f1434f2e3e529a4f55f5dcff4aaa314b507a1c7a4bf1ab1887134e1c8b' && tar -xzf numb-form_netbsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:netbsd -->

On Linux, macOS and the BSDs, make sure `~/.local/bin` is in `PATH`:

```
echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.profile" && . "$HOME/.profile"
```

### Checksums

<!-- install:checksums -->
| File of v1.2.1 | SHA-256 |
| --- | --- |
| `numb-form_darwin_amd64.tar.gz` | `54ea02b6c4e055aceed0fefce86677afc043a10cd0ee0cae2fe3532ce0260d0b` |
| `numb-form_darwin_arm64.tar.gz` | `bb3284a2412fe68075be90e5f4df81284c536af9b304a10ffd5112cef21cb396` |
| `numb-form_freebsd_amd64.tar.gz` | `a535bdcbf5f93320ad0bfae865dd9100dc5f9c95b58591aaa565c5c34458c932` |
| `numb-form_freebsd_arm64.tar.gz` | `3d6f362be408cb282d862f4a442ef6663495fc5fdc15de6cd7ff87a3ba0ebf01` |
| `numb-form_linux_386.tar.gz` | `c1f4eeb5192169bdbb938abf6a55c4a9584b1ba3b78a4624719fee3874c151a8` |
| `numb-form_linux_amd64.tar.gz` | `e336d789a9e758ea1d919f2adc5521d14d159572b644ddc325174d01e0dcc6b8` |
| `numb-form_linux_arm64.tar.gz` | `c44b5089a40a6883e9a1531f0bcde6e9afaf9698f05e857f11f81475042405b0` |
| `numb-form_linux_armv7.tar.gz` | `633be78a65423d29c5b33c2c909801f080d86c78643d9539b041c0dfe2c32952` |
| `numb-form_netbsd_amd64.tar.gz` | `949174f1434f2e3e529a4f55f5dcff4aaa314b507a1c7a4bf1ab1887134e1c8b` |
| `numb-form_netbsd_arm64.tar.gz` | `c17809e0afbff7bafc78d53927d3dbc6a575557c43c0c4687e8aa50dae8f672f` |
| `numb-form_openbsd_amd64.tar.gz` | `dd0c3b9b04c614c0b1a634ee84fded10011b7b4abdef2539cb88237fa1ec42cb` |
| `numb-form_openbsd_arm64.tar.gz` | `8aff77e9100dad9708f7b5131162d2828d01ba7eceb78f2c24ab762b2b890e4e` |
| `numb-form_windows_386.zip` | `3a02f2809d5162e91ecd9ffc3282685e895217cecd97e0b2787a3ffa71bfb003` |
| `numb-form_windows_amd64.zip` | `26f893d6306babb42a2b2a801519e9136f1c91850233cf02206706c7701d8a91` |
| `numb-form_windows_arm64.zip` | `d7655c63e379bbd6d1cc1556a8be305cc42464730b5bc4fe0eb9c2fafbe8f40e` |
<!-- /install:checksums -->

### Rolling builds

After every change to `main`, GitHub Actions runs the tests, rebuilds all
archives and replaces the files of the rolling pre-release (release tag
`edge`).
Its archives are checksummed, signed and attested like a stable release.
`releases/latest` always means the newest stable release; the rolling
build is never marked latest.
`numb-form update rolling` switches to it, `numb-form update stable`
back.

The rolling build changes after every change to `main`, so its
checksum cannot be written on this page.
Its commands check the downloaded file against the rolling build's own
`SHA256SUMS` instead, which protects against damaged downloads but not
against a file replaced on the server.
Use stable releases for real work.
On Linux:

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge/numb-form_linux_amd64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge/SHA256SUMS && grep ' numb-form_linux_amd64.tar.gz$' SHA256SUMS | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date numb-typo "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && numb-form -version
```

Running an install command again is also how to reinstall: it
overwrites the programs and leaves the configuration alone.
`numb-form -version` shows `rolling-<commit>` for a rolling build
(`edge-<commit>` for older ones).

### Verifying who built a release

Checksums prove the download is intact.
To prove it was built by this repository's release workflow, check the
build provenance attestation with the GitHub CLI:

```
gh attestation verify numb-form_linux_amd64.tar.gz --repo z4dm42f5kp-ux/numb-form
```

or the Sigstore signature of the checksum file with cosign:

```
cosign verify-blob --bundle SHA256SUMS.sigstore.json --certificate-identity-regexp '^https://github.com/z4dm42f5kp-ux/numb-form/\.github/workflows/release\.yml@refs/tags/v' --certificate-oidc-issuer https://token.actions.githubusercontent.com SHA256SUMS
```

## Building from source

The project has no third-party modules: `go.mod` has no `require` lines
and the build downloads nothing but Go itself.
The commands below build the same stable release as the binaries above,
from its tag, with Go 1.21 or later.
They use a temporary build cache that is deleted afterwards, set
`GOTOOLCHAIN=local` so that Go never downloads another toolchain, and
install into `~/.local/bin` (Windows: `%LOCALAPPDATA%\Programs\numb-form`).
A program built from source prints `dev` as its version, and
`numb-form update` does not replace it.

1. Install Go and git, with or without administrator rights (below).
2. Run the build command of your system.
3. Check the result: `numb-form -version` prints `dev`.

### Go with administrator rights

<!-- install:go-packages -->
| System | Install Go and git |
| --- | --- |
| Debian 13 and systems based on it | `sudo apt-get update && sudo apt-get install -y golang-go git ca-certificates` |
| Fedora 44 | `sudo dnf install -y golang git` |
| Arch Linux (official repositories, no AUR) | `sudo pacman -S --needed go git` |
| Gentoo | `sudo emerge --ask --noreplace dev-lang/go dev-vcs/git` |
| Void Linux | `sudo xbps-install -Sy go git` |
| openSUSE Tumbleweed and Leap 16 | `sudo zypper install -y go git` |
| Alpine Linux | `doas apk add go git` |
| NixOS | none: the NixOS command below brings Go and git |
| Guix System | none: the Guix command below brings Go and git |
| macOS | `brew install go git` |
| FreeBSD | `su -m root -c 'pkg install -y go git'` |
| OpenBSD | `doas pkg_add go git` |
| NetBSD | `su -m root -c 'pkgin -y install go git' && PATH=$(ls -d /usr/pkg/go1*/bin \| tail -n 1):$PATH && export PATH` |
| Windows (Command Prompt or PowerShell 7) | `winget install --id GoLang.Go -e && winget install --id Git.Git -e` |
| Windows PowerShell 5.1 | `winget install --id GoLang.Go -e; if ($?) { winget install --id Git.Git -e }` |
<!-- /install:go-packages -->

On NetBSD, pkgsrc installs Go in a versioned folder such as
`/usr/pkg/go126/bin`; the table entry adds the newest one to `PATH`.
If the installed Go is older than 1.21, the build stops with an error;
install a newer Go.

### Go without administrator rights

Go's official archives unpack into your home folder.
Replace `go1.27.2` with the newest release listed at https://go.dev/dl/.
The checksum is taken from go.dev's release list.

Linux (use `sha256sum`) and macOS (use `shasum -a 256`; file
`go1.27.2.darwin-arm64.tar.gz`):

```
cd "$(mktemp -d)" && f=go1.27.2.linux-amd64.tar.gz && curl -fLO "https://go.dev/dl/$f" && sum=$(curl -fsSL 'https://go.dev/dl/?mode=json&include=all' | grep -A8 "\"$f\"" | sed -n 's/.*"sha256": *"\([0-9a-f]\{64\}\)".*/\1/p' | head -n 1) && test -n "$sum" && echo "$sum  $f" | sha256sum -c - && rm -rf "$HOME/.local/go" && mkdir -p "$HOME/.local" && tar -C "$HOME/.local" -xzf "$f" && cd / && rm -rf "$OLDPWD" && echo 'export PATH="$HOME/.local/go/bin:$HOME/.local/bin:$PATH"' >> "$HOME/.profile" && . "$HOME/.profile" && go version
```

Windows PowerShell:

```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $f = 'go1.27.2.windows-amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs'; New-Item -ItemType Directory -Force $d | Out-Null; Set-Location $d; try { Invoke-WebRequest -UseBasicParsing "https://go.dev/dl/$f" -OutFile $f; $j = Invoke-RestMethod -UseBasicParsing 'https://go.dev/dl/?mode=json&include=all'; $s = ($j | ForEach-Object { $_.files } | Where-Object { $_.filename -eq $f } | Select-Object -First 1).sha256; if (-not $s -or (Get-FileHash $f -Algorithm SHA256).Hash.ToLower() -ne $s) { throw 'CHECKSUM MISMATCH' }; Expand-Archive -Force $f -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $f }; $b = Join-Path $d 'go\bin'; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $b) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $b).TrimStart(';'), 'User') }; 'OK, open a new PowerShell window' }
```

Command Prompt:

```
(if not exist "%LOCALAPPDATA%\Programs" mkdir "%LOCALAPPDATA%\Programs") && cd /d "%LOCALAPPDATA%\Programs" && curl.exe -fLO https://go.dev/dl/go1.27.2.windows-amd64.zip && powershell -NoProfile -Command "$f='go1.27.2.windows-amd64.zip'; $j=Invoke-RestMethod -UseBasicParsing 'https://go.dev/dl/?mode=json&include=all'; $s=($j | ForEach-Object { $_.files } | Where-Object { $_.filename -eq $f } | Select-Object -First 1).sha256; if (-not $s -or (Get-FileHash $f -Algorithm SHA256).Hash.ToLower() -ne $s) { Write-Error 'CHECKSUM MISMATCH'; exit 1 }; $b=$env:LOCALAPPDATA+'\Programs\go\bin'; $p=[Environment]::GetEnvironmentVariable('Path','User'); if (-not $p) { $p='' }; if (($p -split ';') -notcontains $b) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';')+';'+$b).TrimStart(';'), 'User') }" && tar -xf go1.27.2.windows-amd64.zip && del go1.27.2.windows-amd64.zip && echo OK, open a new Command Prompt
```

On the BSDs, use the release binary when you have no administrator
rights; Go's archives for them need the system's own packages anyway.

### Build and install

Linux, macOS and the BSDs:

<!-- install:source-unix -->
```
cd "$(mktemp -d)" && git clone --depth 1 --branch v1.2.1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$(mktemp -d)" GOFLAGS="-trimpath -mod=readonly" CGO_ENABLED=0 GOTOOLCHAIN=local && go build -o bin/ ./cmd/... && mkdir -p "$HOME/.local/bin" && install -m 0755 bin/* "$HOME/.local/bin/" && go clean -cache && rm -rf "$GOCACHE" && cd / && rm -rf "$OLDPWD/.." && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:source-unix -->

NixOS (brings Go and git for this command only):

<!-- install:source-nixos -->
```
nix-shell -p go git cacert --run 'cd "$(mktemp -d)" && git clone --depth 1 --branch v1.2.1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$(mktemp -d)" GOFLAGS="-trimpath -mod=readonly" CGO_ENABLED=0 GOTOOLCHAIN=local && go build -o bin/ ./cmd/... && mkdir -p "$HOME/.local/bin" && install -m 0755 bin/* "$HOME/.local/bin/" && go clean -cache && rm -rf "$GOCACHE" && cd / && rm -rf "$OLDPWD/.." && "$HOME/.local/bin/numb-form" -version'
```
<!-- /install:source-nixos -->

Guix System (brings Go and git for this command only):

<!-- install:source-guix -->
```
guix shell go git nss-certs coreutils -- sh -c 'export SSL_CERT_DIR="$GUIX_ENVIRONMENT/etc/ssl/certs" GIT_SSL_CAINFO="$GUIX_ENVIRONMENT/etc/ssl/certs/ca-certificates.crt" && cd "$(mktemp -d)" && git clone --depth 1 --branch v1.2.1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$(mktemp -d)" GOFLAGS="-trimpath -mod=readonly" CGO_ENABLED=0 GOTOOLCHAIN=local && go build -o bin/ ./cmd/... && mkdir -p "$HOME/.local/bin" && install -m 0755 bin/* "$HOME/.local/bin/" && go clean -cache && rm -rf "$GOCACHE" && cd / && rm -rf "$OLDPWD/.." && "$HOME/.local/bin/numb-form" -version'
```
<!-- /install:source-guix -->

Windows PowerShell (no git needed):

<!-- install:source-windows -->
```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $t = Join-Path $env:TEMP ('numb-form-' + [guid]::NewGuid()); $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $t, $d | Out-Null; try { Set-Location $t; Invoke-WebRequest -UseBasicParsing 'https://github.com/z4dm42f5kp-ux/numb-form/archive/refs/tags/v1.2.1.zip' -OutFile src.zip; Expand-Archive src.zip -DestinationPath $t; Set-Location (Join-Path $t 'numb-form-1.2.1'); $env:GOCACHE = Join-Path $t 'cache'; $env:GOFLAGS = '-trimpath -mod=readonly'; $env:CGO_ENABLED = '0'; $env:GOTOOLCHAIN = 'local'; go build -o $d ./cmd/...; if ($LASTEXITCODE -ne 0) { throw 'the build failed' } } finally { Set-Location $env:TEMP; Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $t; Remove-Item Env:GOCACHE, Env:GOFLAGS, Env:CGO_ENABLED, Env:GOTOOLCHAIN -ErrorAction SilentlyContinue }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, open a new PowerShell window' }
```
<!-- /install:source-windows -->

## Office software

numb-form edits .docx, .odt, .fodt and .rtf templates itself.
.doc, .pages and .numbers files need LibreOffice, Apache OpenOffice or
Apple Pages/Numbers on the same computer; see [formats.md](formats.md).

## Containers

See [containers.md](containers.md) for Docker, Podman (rootless, no
administrator rights) and Kubernetes.

## After installing

Save your name once, then start numb-form inside the working folder.

**Unix-like** (macOS, Linux, BSD), in Terminal:

```
numb-form config set person "Your Name" && cd /path/to/GWG && numb-form
```

**Windows** (10, 11), in PowerShell:

```
numb-form config set person "Your Name"; cd "H:\GWG"; numb-form
```

The [step-by-step guide](guide.md) explains every question it asks.
