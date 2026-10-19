# Installation von numb-form

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../install.md). Maßgeblich ist die englische Fassung.

Kein Programmierer? Die [step-by-step-Anleitung](guide.md) hat die Installationszeile für Windows und erklärt, was nach der Installation zu tun ist.

numb-form besteht aus zwei statischen Programmen, `numb-form` und `numb-commit`, ohne Abhängigkeiten. Jedes System erhält die gleichen Programme und das gleiche Verhalten.

Es gibt zwei Möglichkeiten, sie zu installieren:

* Laden Sie ein Release Binary herunter. Dies erfordert niemals Administratorrechte.
* Erstellen Sie aus der Quelle mit Go 1.21 oder höher, mit oder ohne Administratorrechte.

Jeder Befehl unten ist eine einzelne Zeile. Kopieren Sie es ganz. Es stoppt beim ersten Schritt, der fehlschlägt.

## Binärdateien freigeben (keine Administratorrechte)

Archive heißen `numb-form_<os>_<arch>`, wobei *os* `linux`, `darwin` (macOS), `windows`, `freebsd`, `openbsd` oder `netbsd` und *arch* `amd64` oder `arm64` ist (Linux und Windows haben auch `386`; Linux hat `armv7`). Die Befehle verwenden `amd64`; ersetzen Sie es, um die Maschine zu passen. Jeder Befehl wird in einen temporären Ordner heruntergeladen, überprüft die SHA-256 Prüfsumme, installiert sich in einem Ordner, den Sie besitzen, und bereinigt.

### Linux (Debian, Ubuntu, Arch, Fedora, Gentoo, Void, Alpine)

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_linux_amd64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' numb-form_linux_amd64.tar.gz$' SHA256SUMS | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### macOS

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_darwin_arm64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' numb-form_darwin_arm64.tar.gz$' SHA256SUMS | shasum -a 256 -c - && tar -xzf numb-form_darwin_arm64.tar.gz && { xattr -d com.apple.quarantine numb-form numb-commit 2>/dev/null || true; } && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

Verwenden Sie `darwin_amd64` auf Intel Macs. Die Binärdateien werden nicht von Apple beglaubigt; der `xattr`-Schritt entfernt das Download-Quarantäne-Flag, nachdem die Prüfsumme verifiziert wurde.

### Windows 10 und 11

Befehle für beide Shells folgen. Verwenden Sie das Fenster, das Sie geöffnet haben: Windows PowerShell (blau, prompt `PS C:\>`) oder Command Prompt (schwarz, prompt `C:\>`). Sie installieren in `%LOCALAPPDATA%\Programs\numb-form` und fügen es zu Ihrem eigenen `PATH` hinzu; Es sind keine Administratorrechte erforderlich. Verwenden Sie `arm64` anstelle von `amd64` auf ARM-Geräten.

Windows PowerShell 5.1 hat kein `&&`. Die PowerShell-Befehle stoppen daher beim ersten Fehler mit `$ErrorActionPreference = 'Stop'`, was den gleichen Effekt hat.

Windows PowerShell:

```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Set-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; 'OK, open a new PowerShell window' }
```

Command Prompt:

```
(if not exist "%LOCALAPPDATA%\Programs\numb-form" mkdir "%LOCALAPPDATA%\Programs\numb-form") && cd /d "%LOCALAPPDATA%\Programs\numb-form" && curl.exe -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_windows_amd64.zip && curl.exe -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && powershell -NoProfile -Command "$h=(Get-FileHash numb-form_windows_amd64.zip -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch ($h+'  numb-form_windows_amd64.zip') -Quiet)) { Write-Error 'CHECKSUM MISMATCH'; exit 1 }; $d=$env:LOCALAPPDATA+'\Programs\numb-form'; $p=[Environment]::GetEnvironmentVariable('Path','User'); if (-not $p) { $p='' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';')+';'+$d).TrimStart(';'), 'User') }" && tar -xf numb-form_windows_amd64.zip && del numb-form_windows_amd64.zip SHA256SUMS && echo OK, open a new Command Prompt
```

Wenn ein Download mit `404` fehlschlägt, wurde noch kein Release veröffentlicht; Build from source wie unten beschrieben.

### FreeBSD

```
cd "$(mktemp -d)" && fetch https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_freebsd_amd64.tar.gz && fetch https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && sha256 -c "$(awk '$2=="numb-form_freebsd_amd64.tar.gz"{print $1}' SHA256SUMS)" numb-form_freebsd_amd64.tar.gz && tar -xzf numb-form_freebsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### OpenBSD

Die `SHA256`-Datei verwendet das BSD-Format, das sha256(1) direkt überprüft.

```
cd "$(mktemp -d)" && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_openbsd_amd64.tar.gz && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256 && sha256 -C SHA256 numb-form_openbsd_amd64.tar.gz && tar -xzf numb-form_openbsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### NetBSD

```
cd "$(mktemp -d)" && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_netbsd_amd64.tar.gz && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256 && test "$(cksum -a SHA256 numb-form_netbsd_amd64.tar.gz)" = "$(grep '(numb-form_netbsd_amd64.tar.gz)' SHA256)" && tar -xzf numb-form_netbsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

Stellen Sie unter Linux, macOS und den BSDs sicher, dass `~/.local/bin` in `PATH` enthalten ist:

```
echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.profile" && . "$HOME/.profile"
```

### Überprüfen, wer ein Release erstellt hat

Checksums beweisen, dass der Download intakt ist. Um zu beweisen, dass es vom Release-Workflow dieses Repositorys erstellt wurde, überprüfen Sie die Build-Provenienzbescheinigung mit dem GitHub CLI:

```
gh attestation verify numb-form_linux_amd64.tar.gz --repo z4dm42f5kp-ux/numb-form
```

oder die Sigstore-Signatur der Prüfsummendatei mit cosign:

```
cosign verify-blob --bundle SHA256SUMS.sigstore.json --certificate-identity-regexp '^https://github.com/z4dm42f5kp-ux/numb-form/\.github/workflows/release\.yml@refs/tags/v' --certificate-oidc-issuer https://token.actions.githubusercontent.com SHA256SUMS
```

## Bauen aus der Quelle

Das Projekt hat keine Module von Drittanbietern: `go.mod` hat keine `require`-Linien und der Build lädt nichts herunter. Die folgenden Builds verwenden eine temporäre Go Build Cache, der danach gelöscht wird, so dass kein Cache auf der Festplatte verbleibt.

### Gehen Sie ohne Administratorrechte

Die offiziellen Archive von Go werden in Ihrem Home-Ordner ausgepackt. Ersetzen Sie `go1.27.2` durch die neueste Version, die bei https://go.dev/dl/. aufgeführt ist Die Prüfsumme stammt aus der Releaseliste von go.dev.

Linux (verwenden Sie `sha256sum`) und macOS (verwenden Sie `shasum -a 256`; Datei `go1.27.2.darwin-arm64.tar.gz`):

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

Verwenden Sie auf den BSDs die Binärversion, wenn Sie keine Administratorrechte haben; Gos Archive für sie brauchen sowieso die eigenen Pakete des Systems.

### Gehen Sie mit Administratorrechten

| System | Installieren Sie Go und Git |
| --- | --- |
| Alpines Linux | `doas apk add go git` |
| Arch Linux (nur offizielle Repositories) | `sudo pacman -S --needed go git` |
| Debian 13, Ubuntu 24.04 oder höher | `sudo apt-get update && sudo apt-get install -y golang-go git` |
| Debian 12 | `echo 'deb http://deb.debian.org/debian bookworm-backports main' \| sudo tee /etc/apt/sources.list.d/backports.list && sudo apt-get update && sudo apt-get install -y -t bookworm-backports golang-go && sudo apt-get install -y git` |
| Fedora, RHEL, Rocky, Alma | `sudo dnf install -y golang git` |
| FreeBSD | `su -m root -c 'pkg install -y go git'` |
| Gentoo | `sudo emerge --ask --noreplace dev-lang/go dev-vcs/git` |
| macOS | `brew install go git` |
| NetBSD | `su -m root -c 'pkgin -y install go git' && PATH=$(ls -d /usr/pkg/go1*/bin \| tail -n 1):$PATH && export PATH` |
| OpenBSD | `doas pkg_add go git` |
| Void Linux | `sudo xbps-install -Sy go git` |
| Windows (Command Prompt oder PowerShell 7) | `winget install --id GoLang.Go -e && winget install --id Git.Git -e` |
| Windows PowerShell 5.1 | `winget install --id GoLang.Go -e; if ($?) { winget install --id Git.Git -e }` |

Auf NetBSD installiert pkgsrc Go in einem versionierten Ordner wie `/usr/pkg/go126/bin`; der Tabelleneintrag fügt den neuesten zu `PATH` hinzu. Wenn das installierte Go älter als 1,21 ist, stoppt der Build mit einem Fehler; Installieren Sie ein neueres Go, anstatt es eine Toolchain herunterladen zu lassen.

### Build und Installation

Linux, macOS und die BSDs (installiert in `~/.local/bin`):

```
cd "$(mktemp -d)" && git clone --depth 1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$(mktemp -d)" GOFLAGS='-trimpath -mod=readonly' CGO_ENABLED=0 && go build -o numb-form ./cmd/numb-form && go build -o numb-commit ./cmd/numb-commit && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && go clean -cache && rm -rf "$GOCACHE" && cd / && rm -rf "$OLDPWD/.." && echo OK
```

Windows (kein Git erforderlich; installiert in `%LOCALAPPDATA%\Programs\numb-form` und fügt es Ihrem `PATH` hinzu). Windows PowerShell:

```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $t = Join-Path $env:TEMP ('numb-form-' + [guid]::NewGuid()); $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $t, $d | Out-Null; try { Set-Location $t; Invoke-WebRequest -UseBasicParsing 'https://github.com/z4dm42f5kp-ux/numb-form/archive/refs/heads/main.zip' -OutFile src.zip; Expand-Archive src.zip -DestinationPath $t; Set-Location (Join-Path $t 'numb-form-main'); $env:GOCACHE = Join-Path $t 'cache'; $env:GOFLAGS = '-trimpath -mod=readonly'; $env:CGO_ENABLED = '0'; foreach ($c in 'numb-form', 'numb-commit') { go build -o (Join-Path $d "$c.exe") "./cmd/$c"; if ($LASTEXITCODE -ne 0) { throw "building $c failed" } } } finally { Set-Location $env:TEMP; Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $t; Remove-Item Env:GOCACHE, Env:GOFLAGS, Env:CGO_ENABLED -ErrorAction SilentlyContinue }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; 'OK, open a new PowerShell window' }
```

Command Prompt:

```
cd /d "%TEMP%" && curl.exe -fL -o numb-form-src.zip https://github.com/z4dm42f5kp-ux/numb-form/archive/refs/heads/main.zip && tar -xf numb-form-src.zip && del numb-form-src.zip && cd numb-form-main && set "GOCACHE=%TEMP%\numb-form-gocache" && set "GOFLAGS=-trimpath -mod=readonly" && set "CGO_ENABLED=0" && (if not exist "%LOCALAPPDATA%\Programs\numb-form" mkdir "%LOCALAPPDATA%\Programs\numb-form") && go build -o "%LOCALAPPDATA%\Programs\numb-form\numb-form.exe" ./cmd/numb-form && go build -o "%LOCALAPPDATA%\Programs\numb-form\numb-commit.exe" ./cmd/numb-commit && go clean -cache && cd /d "%TEMP%" && rmdir /s /q numb-form-main && powershell -NoProfile -Command "$d=$env:LOCALAPPDATA+'\Programs\numb-form'; $p=[Environment]::GetEnvironmentVariable('Path','User'); if (-not $p) { $p='' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';')+';'+$d).TrimStart(';'), 'User') }" && echo OK, open a new Command Prompt
```

## Bürosoftware

numb-form bearbeitet .docx, .odt, .fodt und .rtf Vorlagen selbst. .doc, .pages und .numbers Dateien benötigen LibreOffice, Apache OpenOffice oder Apple Pages/Numbers auf demselben Computer; siehe [formats.md](formats.md).

## Optional kostenloser API Key

Das UK Companies House Register benötigt einen freien Schlüssel. Alle anderen Quellen brauchen keine. Führen Sie `numb-form keys` für Anweisungen aus.

## Container

Siehe [containers.md](containers.md) für Docker, Podman (rootless, keine Administratorrechte) und Kubernetes.

## Nach der Installation

Speichern Sie Ihren Namen einmal und starten Sie dann numb-form im Arbeitsordner:

```
numb-form config set person "Your Name" && cd /path/to/GWG && numb-form
```

Die [Schritt-für-Schritt-Anleitung](guide.md) erklärt jede Frage, die sie stellt.
