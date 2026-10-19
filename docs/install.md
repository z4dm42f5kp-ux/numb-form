# Installing numb-form

Not a programmer?
The [step-by-step guide](guide.md) has the install line for Windows and
explains what to do after installing.

numb-form is two static programs, `numb-form` and `numb-commit`, with no
dependencies.
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

Archives are named `numb-form_<os>_<arch>`, where *os* is `linux`,
`darwin` (macOS), `windows`, `freebsd`, `openbsd` or `netbsd`, and *arch*
is `amd64` or `arm64` (Linux and Windows also have `386`; Linux has
`armv7`).
The commands use `amd64`; replace it to suit the machine.
Each command downloads into a temporary folder, verifies the SHA-256
checksum, installs into a folder you own and cleans up.

### Linux (Debian, Ubuntu, Arch, Fedora, Gentoo, Void, Alpine)

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_linux_amd64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' numb-form_linux_amd64.tar.gz$' SHA256SUMS | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### macOS

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_darwin_arm64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' numb-form_darwin_arm64.tar.gz$' SHA256SUMS | shasum -a 256 -c - && tar -xzf numb-form_darwin_arm64.tar.gz && { xattr -d com.apple.quarantine numb-form numb-commit 2>/dev/null || true; } && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

Use `darwin_amd64` on Intel Macs.
The binaries are not notarised by Apple; the `xattr` step removes the
download quarantine flag after the checksum has been verified.

### Windows 10 and 11

Commands for both shells follow.
Use the one that matches the window you have open: Windows PowerShell
(blue, prompt `PS C:\>`) or Command Prompt (black, prompt `C:\>`).
They install into `%LOCALAPPDATA%\Programs\numb-form` and add it to your
own `PATH`; no administrator rights are needed.
Use `arm64` instead of `amd64` on ARM devices.

Windows PowerShell 5.1 has no `&&`.
The PowerShell commands therefore stop at the first error with
`$ErrorActionPreference = 'Stop'` instead, which has the same effect.

Windows PowerShell:

```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Set-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; 'OK, open a new PowerShell window' }
```

Command Prompt:

```
(if not exist "%LOCALAPPDATA%\Programs\numb-form" mkdir "%LOCALAPPDATA%\Programs\numb-form") && cd /d "%LOCALAPPDATA%\Programs\numb-form" && curl.exe -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_windows_amd64.zip && curl.exe -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && powershell -NoProfile -Command "$h=(Get-FileHash numb-form_windows_amd64.zip -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch ($h+'  numb-form_windows_amd64.zip') -Quiet)) { Write-Error 'CHECKSUM MISMATCH'; exit 1 }; $d=$env:LOCALAPPDATA+'\Programs\numb-form'; $p=[Environment]::GetEnvironmentVariable('Path','User'); if (-not $p) { $p='' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';')+';'+$d).TrimStart(';'), 'User') }" && tar -xf numb-form_windows_amd64.zip && del numb-form_windows_amd64.zip SHA256SUMS && echo OK, open a new Command Prompt
```

If a download fails with `404`, no release has been published yet;
build from source as described below.

### Reinstalling or updating on Windows

Close every numb-form window, then run one of these in Windows
PowerShell.
They replace the programs, keep your settings and print the installed
version.

Stable release:

```
& { $ErrorActionPreference = 'Stop'; if (Get-Process numb-form, numb-commit -ErrorAction SilentlyContinue) { throw 'Close every numb-form window first, then run this again' }; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Push-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Remove-Item -Force -ErrorAction SilentlyContinue numb-form.exe, numb-commit.exe; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS; Pop-Location }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, reinstalled. Your settings are kept. Open a new PowerShell window' }
```

Development build (the `edge` pre-release, rebuilt from `main` after every
change; see below):

```
& { $ErrorActionPreference = 'Stop'; if (Get-Process numb-form, numb-commit -ErrorAction SilentlyContinue) { throw 'Close every numb-form window first, then run this again' }; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Push-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Remove-Item -Force -ErrorAction SilentlyContinue numb-form.exe, numb-commit.exe; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS; Pop-Location }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, reinstalled. Your settings are kept. Open a new PowerShell window' }
```

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

On Linux, macOS and the BSDs, make sure `~/.local/bin` is in `PATH`:

```
echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.profile" && . "$HOME/.profile"
```

### Development builds

After every change to `main`, GitHub Actions runs the tests, rebuilds all
archives and replaces the files of the `edge` pre-release.
Its archives are checksummed, signed and attested like a stable release.
`releases/latest` always means the newest stable release; edge is never
marked latest.

To install or reinstall the development build on any system, run the
install command of that system with `releases/latest/download` replaced
by `releases/download/edge`.
For example, on Linux:

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge/numb-form_linux_amd64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge/SHA256SUMS && grep ' numb-form_linux_amd64.tar.gz$' SHA256SUMS | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && numb-form -version
```

Running an install command again is also how to reinstall: it
overwrites the programs and leaves the configuration alone.
`numb-form -version` shows `edge-<commit>` for a development build.

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
and the build downloads nothing.
The builds below use a temporary Go build cache that is deleted
afterwards, so no cache is left on disk.

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

### Go with administrator rights

| System | Install Go and git |
| --- | --- |
| Alpine Linux | `doas apk add go git` |
| Arch Linux (official repositories only) | `sudo pacman -S --needed go git` |
| Debian 13, Ubuntu 24.04 or later | `sudo apt-get update && sudo apt-get install -y golang-go git` |
| Debian 12 | `echo 'deb http://deb.debian.org/debian bookworm-backports main' \| sudo tee /etc/apt/sources.list.d/backports.list && sudo apt-get update && sudo apt-get install -y -t bookworm-backports golang-go && sudo apt-get install -y git` |
| Fedora, RHEL, Rocky, Alma | `sudo dnf install -y golang git` |
| FreeBSD | `su -m root -c 'pkg install -y go git'` |
| Gentoo | `sudo emerge --ask --noreplace dev-lang/go dev-vcs/git` |
| macOS | `brew install go git` |
| NetBSD | `su -m root -c 'pkgin -y install go git' && PATH=$(ls -d /usr/pkg/go1*/bin \| tail -n 1):$PATH && export PATH` |
| OpenBSD | `doas pkg_add go git` |
| Void Linux | `sudo xbps-install -Sy go git` |
| Windows (Command Prompt or PowerShell 7) | `winget install --id GoLang.Go -e && winget install --id Git.Git -e` |
| Windows PowerShell 5.1 | `winget install --id GoLang.Go -e; if ($?) { winget install --id Git.Git -e }` |

On NetBSD, pkgsrc installs Go in a versioned folder such as
`/usr/pkg/go126/bin`; the table entry adds the newest one to `PATH`.
If the installed Go is older than 1.21, the build stops with an error;
install a newer Go rather than letting it download a toolchain.

### Build and install

Linux, macOS and the BSDs (installs into `~/.local/bin`):

```
cd "$(mktemp -d)" && git clone --depth 1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$(mktemp -d)" GOFLAGS='-trimpath -mod=readonly' CGO_ENABLED=0 && go build -o numb-form ./cmd/numb-form && go build -o numb-commit ./cmd/numb-commit && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && go clean -cache && rm -rf "$GOCACHE" && cd / && rm -rf "$OLDPWD/.." && echo OK
```

Windows (no git needed; installs into `%LOCALAPPDATA%\Programs\numb-form`
and adds it to your `PATH`).
Windows PowerShell:

```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $t = Join-Path $env:TEMP ('numb-form-' + [guid]::NewGuid()); $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $t, $d | Out-Null; try { Set-Location $t; Invoke-WebRequest -UseBasicParsing 'https://github.com/z4dm42f5kp-ux/numb-form/archive/refs/heads/main.zip' -OutFile src.zip; Expand-Archive src.zip -DestinationPath $t; Set-Location (Join-Path $t 'numb-form-main'); $env:GOCACHE = Join-Path $t 'cache'; $env:GOFLAGS = '-trimpath -mod=readonly'; $env:CGO_ENABLED = '0'; foreach ($c in 'numb-form', 'numb-commit') { go build -o (Join-Path $d "$c.exe") "./cmd/$c"; if ($LASTEXITCODE -ne 0) { throw "building $c failed" } } } finally { Set-Location $env:TEMP; Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $t; Remove-Item Env:GOCACHE, Env:GOFLAGS, Env:CGO_ENABLED -ErrorAction SilentlyContinue }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; 'OK, open a new PowerShell window' }
```

Command Prompt:

```
cd /d "%TEMP%" && curl.exe -fL -o numb-form-src.zip https://github.com/z4dm42f5kp-ux/numb-form/archive/refs/heads/main.zip && tar -xf numb-form-src.zip && del numb-form-src.zip && cd numb-form-main && set "GOCACHE=%TEMP%\numb-form-gocache" && set "GOFLAGS=-trimpath -mod=readonly" && set "CGO_ENABLED=0" && (if not exist "%LOCALAPPDATA%\Programs\numb-form" mkdir "%LOCALAPPDATA%\Programs\numb-form") && go build -o "%LOCALAPPDATA%\Programs\numb-form\numb-form.exe" ./cmd/numb-form && go build -o "%LOCALAPPDATA%\Programs\numb-form\numb-commit.exe" ./cmd/numb-commit && go clean -cache && cd /d "%TEMP%" && rmdir /s /q numb-form-main && powershell -NoProfile -Command "$d=$env:LOCALAPPDATA+'\Programs\numb-form'; $p=[Environment]::GetEnvironmentVariable('Path','User'); if (-not $p) { $p='' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';')+';'+$d).TrimStart(';'), 'User') }" && echo OK, open a new Command Prompt
```

## Office software

numb-form edits .docx, .odt, .fodt and .rtf templates itself.
.doc, .pages and .numbers files need LibreOffice, Apache OpenOffice or
Apple Pages/Numbers on the same computer; see [formats.md](formats.md).

## Optional free API key

The UK Companies House register needs a free key.
All other sources need none.
Run `numb-form keys` for instructions.

## Containers

See [containers.md](containers.md) for Docker, Podman (rootless, no
administrator rights) and Kubernetes.

## After installing

Save your name once, then start numb-form inside the working folder:

```
numb-form config set person "Your Name" && cd /path/to/GWG && numb-form
```

The [step-by-step guide](guide.md) explains every question it asks.
