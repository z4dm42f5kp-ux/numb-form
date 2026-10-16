# Installieren von numb-form

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../install.md). Maßgeblich ist die englische Fassung.

numb-form ist zwei Programme, `numb-form` und `numb-commit`.
Sie sind statische Binärdateien ohne Abhängigkeiten.

Es gibt zwei Möglichkeiten, sie zu installieren:

* Laden Sie eine signierte Version Binary herunter.
* Bauen Sie aus der Quelle mit Go 1.21 oder höher.

Jeder Befehl unten ist eine einzelne Zeile.
Kopieren Sie es ganz.
Es stoppt beim ersten Schritt, der fehlschlägt.

## Windows 10 und 11

Führen Sie diese in **Command Prompt** (`cmd.exe`) aus.
`curl.exe` und `tar.exe` werden mit Windows 10 1803 und höher ausgeliefert.

### Release Binär

Verwenden Sie `amd64` für die meisten PCs und `arm64` für ARM-Geräte (z. B. Surface Pro X).
Dieser lädt das Release herunter, überprüft seine SHA-256 Prüfsumme gegen
`SHA256SUMS` und entpackt es in `%USERPROFILE%\numb-form`:

```
(if not exist "%USERPROFILE%\numb-form" mkdir "%USERPROFILE%\numb-form") && cd /d "%USERPROFILE%\numb-form" && curl.exe -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_windows_amd64.zip && curl.exe -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && powershell -NoProfile -Command "$h=(Get-FileHash numb-form_windows_amd64.zip -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch ($h+'  numb-form_windows_amd64.zip') -Quiet)) { Write-Error 'CHECKSUM MISMATCH'; exit 1 }" && tar -xf numb-form_windows_amd64.zip && del numb-form_windows_amd64.zip && echo OK
```

Um auch zu überprüfen, wer das Release erstellt hat, installieren Sie die GitHub CLI und überprüfen Sie die
Bauprovenienzbescheinigung:

```
winget install --id GitHub.cli -e && cd /d "%USERPROFILE%\numb-form" && gh attestation verify numb-form.exe --repo z4dm42f5kp-ux/numb-form && gh attestation verify numb-commit.exe --repo z4dm42f5kp-ux/numb-form
```

Fügen Sie den Ordner Ihrem Benutzer `PATH` hinzu (wird in neuen Fenstern wirksam):

```
powershell -NoProfile -Command "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path','User') + ';' + $env:USERPROFILE + '\numb-form', 'User')" && echo Open a new Command Prompt
```

### aus der Quelle

Installieren Gehen Sie mit winget, dann öffnen Sie eine **new** Befehlsaufforderung, so dass `go` auf der
`PATH`:

```
winget install --id GoLang.Go -e
```

Bauen Sie in einem temporären Cache gehen, der danach gelöscht wird, also kein Build-Cache
links auf der Festplatte:

```
cd /d "%USERPROFILE%" && curl.exe -fL -o numb-form-src.zip https://github.com/z4dm42f5kp-ux/numb-form/archive/refs/heads/main.zip && tar -xf numb-form-src.zip && del numb-form-src.zip && cd numb-form-main && set "GOCACHE=%TEMP%\numb-form-gocache" && set "GOFLAGS=-trimpath -mod=readonly" && set "CGO_ENABLED=0" && go test ./... && go build -o "%USERPROFILE%\numb-form\numb-form.exe" ./cmd/numb-form && go build -o "%USERPROFILE%\numb-form\numb-commit.exe" ./cmd/numb-commit && go clean -cache && cd /d "%USERPROFILE%" && rmdir /s /q numb-form-main && echo OK
```

Das Projekt hat keine Go-Module von Drittanbietern, so dass nichts heruntergeladen wird
den Build. Siehe [SECURITY.md](SECURITY.md).

## Binärdateien auf anderen Systemen freigeben

Archive heißen `numb-form_<os>_<arch>.tar.gz`, wobei *os*
`linux`, `darwin` (macOS), `freebsd`, `openbsd` oder `netbsd` und *arch* ist
`amd64` oder `arm64` (Linux hat auch `386` und `armv7`).
Jedes Archiv enthält `numb-form`, `numb-commit`, `README.md` und `LICENSE`.
Die Beispiele verwenden `amd64`; ersetzen Sie es, um die Maschine zu passen.

Jeder Befehl, der in ein temporäres Verzeichnis heruntergeladen wird, überprüft den SHA-256
Prüfsumme, installiert in `~/.local/bin` und entfernt das temporäre Verzeichnis.
Stellen Sie sicher, dass `~/.local/bin` in `PATH` ist.

### Linux (Debian, Ubuntu, Arch, Fedora, Gentoo, Void, Alpine)

Benötigt `curl`, `tar` und `sha256sum`, die jede gelistete Distribution
bietet (Alpine durch BusyBox).

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_linux_amd64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' numb-form_linux_amd64.tar.gz$' SHA256SUMS | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### macOS

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_darwin_arm64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' numb-form_darwin_arm64.tar.gz$' SHA256SUMS | shasum -a 256 -c - && tar -xzf numb-form_darwin_arm64.tar.gz && { xattr -d com.apple.quarantine numb-form numb-commit 2>/dev/null || true; } && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

Verwenden Sie `darwin_amd64` auf Intel Macs.
Die Binärdateien werden nicht von Apple beglaubigt;
Der `xattr`-Schritt entfernt das Download-Quarantäne-Flag nach der Prüfsumme
geprüft wurde.

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

### Überprüfen, wer ein Release erstellt hat

Checksums beweisen, dass der Download intakt ist.
Um zu beweisen, dass es vom Release-Workflow dieses Repositorys erstellt wurde, überprüfen Sie die
GitHub bauen Provenienzbescheinigung mit dem GitHub CLI:

```
gh attestation verify numb-form_linux_amd64.tar.gz --repo z4dm42f5kp-ux/numb-form
```

oder die Sigstore-Signatur der Prüfsummendatei mit Cosign:

```
cosign verify-blob --bundle SHA256SUMS.sigstore.json --certificate-identity-regexp '^https://github.com/z4dm42f5kp-ux/numb-form/\.github/workflows/release\.yml@refs/tags/v' --certificate-oidc-issuer https://token.actions.githubusercontent.com SHA256SUMS
```

## Bauen von der Quelle auf anderen Systemen

Installieren Sie Go 1.21 oder höher und git mit dem Systempaketmanager, dann
Führen Sie den Build-Befehl unten aus.

| System | Installieren Sie Go und Git |
| --- | --- |
| Debian 13, Ubuntu 24.04 oder höher | `sudo apt-get update && sudo apt-get install -y golang-go git` |
| Debian 12 | `echo 'deb http://deb.debian.org/debian bookworm-backports main' \| sudo tee /etc/apt/sources.list.d/backports.list && sudo apt-get update && sudo apt-get install -y -t bookworm-backports golang-go && sudo apt-get install -y git` |
| Arch Linux (nur offizielle Repositories) | `sudo pacman -S --needed go git` |
| Fedora, RHEL, Rocky, Alma | `sudo dnf install -y golang git` |
| Gentoo | `sudo emerge --ask --noreplace dev-lang/go dev-vcs/git` |
| Void Linux | `sudo xbps-install -Sy go git` |
| Alpines Linux | `doas apk add go git` |
| macOS | `brew install go git` |
| FreeBSD | `su -m root -c 'pkg install -y go git'` |
| OpenBSD | `doas pkg_add go git` |
| NetBSD | `su -m root -c 'pkgin -y install go git'` |

Dann bauen, testen und installieren, mit einem temporären Bauen Sie den Cache auf, der
später gestrichen:

```
cd "$(mktemp -d)" && git clone --depth 1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$(mktemp -d)" GOFLAGS='-trimpath -mod=readonly' CGO_ENABLED=0 && go test ./... && go build -o numb-form ./cmd/numb-form && go build -o numb-commit ./cmd/numb-commit && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && go clean -cache && rm -rf "$GOCACHE" && cd / && rm -rf "$OLDPWD/.." && echo OK
```

Wenn das installierte Go älter als 1,21 ist, stoppt der Build mit einem Fehler;
Installieren Sie ein neueres Go aus der Backports- oder Ports-Sammlung, anstatt
Lassen Sie es eine Toolchain herunterladen.

Das Projekt hat keine Module von Drittanbietern: `go.mod` hat keine `require`-Linien
Und der Build lädt nichts herunter.

## Container

Siehe [containers.md](containers.md) für Docker, Podman und Kubernetes.

### Optional kostenloser API Key

Das UK Companies House Register benötigt einen kostenlosen Schlüssel.
Alle anderen Quellen brauchen keine.
Führen Sie `numb-form keys` für Anweisungen aus.
