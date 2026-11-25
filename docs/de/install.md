# Installation von numb-form

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../install.md). Maßgeblich ist die englische Fassung.

Kein Programmierer? Die [step-by-step-Anleitung](guide.md) hat die Installationszeile für Windows und erklärt, was nach der Installation zu tun ist.

numb-form besteht aus vier statischen Programmen, `numb-form`, `numb-commit`, `numb-date` und `numb-typo`, ohne Abhängigkeiten. Jedes System erhält die gleichen Programme und das gleiche Verhalten.

Es gibt zwei Möglichkeiten, sie zu installieren:

* Laden Sie ein Release Binary herunter. Dies erfordert niemals Administratorrechte.
* Erstellen Sie aus der Quelle mit Go 1.21 oder höher, mit oder ohne Administratorrechte.

Jeder Befehl unten ist eine einzelne Zeile. Kopieren Sie es ganz. Es stoppt beim ersten Schritt, der fehlschlägt.

## Binärdateien freigeben (keine Administratorrechte)

Jeder Befehl installiert ein stabiles Release, das im Befehl benannt ist, und überprüft seine SHA-256 Prüfsumme mit dem im Befehl selbst geschriebenen Wert. Ein geänderter oder beschädigter Download stoppt mit einem Fehler, bevor etwas installiert wird. Die Prüfsumme kommt von dieser Seite, nicht vom Server, von dem die Dateien stammen, so dass eine auf dem Server ersetzte Datei nicht übergeben werden kann. Die Befehle werden nur über HTTPS heruntergeladen, in einen Ordner installiert, den Sie besitzen, und nach sich selbst bereinigt.

Die Befehle auf dieser Seite werden von einem Programm aus den signierten Prüfsummen des Releases geschrieben (`scripts/install`). Nach jedem stabilen Release überprüft ein Workflow die Sigstore-Signatur des Releases und schlägt die neuen Befehle vor; ein täglicher Check fängt ein verpasstes Release. `numb-form update` installiert neuere Releases auf die gleiche Weise.

Archive heißen `numb-form_<os>_<arch>`, wobei *os* `linux`, `darwin` (macOS), `windows`, `freebsd`, `openbsd` oder `netbsd` und *arch* `amd64` oder `arm64` ist (Linux und Windows haben auch `386`; Linux hat `armv7`). Ersetzen Sie für ein anderes Archiv seinen Namen an beiden Stellen des Befehls und nehmen Sie seine Prüfsumme von [die Tabelle](#checksums).

### Linux

Ein Befehl funktioniert auf jeder Distribution; er benötigt `curl`, `tar`, `gzip` und `sha256sum`.

<!-- install:linux -->
```
cd "$(mktemp -d)" && curl --proto '=https' --tlsv1.2 -fsSLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_linux_amd64.tar.gz && echo 'e336d789a9e758ea1d919f2adc5521d14d159572b644ddc325174d01e0dcc6b8  numb-form_linux_amd64.tar.gz' | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:linux -->

Auf ARM-Computern (arm64):

<!-- install:linux-arm64 -->
```
cd "$(mktemp -d)" && curl --proto '=https' --tlsv1.2 -fsSLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_linux_arm64.tar.gz && echo 'c44b5089a40a6883e9a1531f0bcde6e9afaf9698f05e857f11f81475042405b0  numb-form_linux_arm64.tar.gz' | sha256sum -c - && tar -xzf numb-form_linux_arm64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:linux-arm64 -->

Wenn der Befehl meldet, dass `curl` oder ein anderes Programm fehlt, installieren Sie es zuerst:

<!-- install:linux-tools -->
| Verteilung | Installieren Sie, was der Befehl benötigt (einmal) |
| --- | --- |
| Debian 13 und darauf basierende Systeme | `sudo apt-get update && sudo apt-get install -y curl ca-certificates` |
| Fedora 44 (Curl ist installiert) | `sudo dnf install -y tar gzip` |
| Arch Linux | `sudo pacman -S --needed curl tar gzip` |
| Gentoo | `sudo emerge --ask --noreplace net-misc/curl` |
| Void Linux | `sudo xbps-install -Sy curl` |
| openSUSE Tumbleweed und Leap 16 | `sudo zypper install -y curl tar gzip` |
| Alpines Linux | `doas apk add curl` |
| NixOS | `nix-shell -p curl` (eine Shell mit Curl; führen Sie den Befehl darin aus) |
| Guix-System | `guix shell curl tar gzip coreutils` (eine Shell mit ihnen; führen Sie den Befehl darin aus) |
<!-- /install:linux-tools -->

Diese Distributionen werden vor jedem Minor Release (vX.Y.0) getestet: Debian 13, Fedora 44, Arch Linux, Gentoo, Void Linux, openSUSE Tumbleweed und Leap 16, Alpine Linux, NixOS und Guix.

### macOS

Apple Silizium (M1 und höher):

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

Die Binärdateien werden nicht von Apple beglaubigt; der Schritt `xattr` entfernt das Download-Quarantäne-Flag, nachdem die Prüfsumme verifiziert wurde.

### Windows 10 und 11

Befehle für beide Shells folgen. Verwenden Sie das Fenster, das Sie geöffnet haben: Windows PowerShell (blau, prompt `PS C:\>`) oder Command Prompt (schwarz, prompt `C:\>`). Sie installieren in `%LOCALAPPDATA%\Programs\numb-form` und fügen es zu Ihrem eigenen `PATH` hinzu; Es sind keine Administratorrechte erforderlich.

Windows PowerShell 5.1 hat kein `&&`. Die PowerShell-Befehle stoppen daher beim ersten Fehler mit `$ErrorActionPreference = 'Stop'`, was den gleichen Effekt hat.

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

### Neuinstallation oder Aktualisierung unter Windows

`numb-form update` macht dies von selbst (siehe [usage](usage.md)); es funktioniert auf jedem System gleich. Wenn es keine Verbindung herstellen kann, schließen Sie jedes numb-form-Fenster und führen Sie dies in Windows PowerShell aus. Es ersetzt die Programme, behält Ihre Einstellungen und druckt die installierte Version.

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

Stellen Sie unter Linux, macOS und den BSDs sicher, dass `~/.local/bin` in `PATH` enthalten ist:

```
echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.profile" && . "$HOME/.profile"
```

### Prüfsummen

<!-- install:checksums -->
| Datei von v1.2.1 | SHA-256 |
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

### Walzkörper

Nach jeder Änderung an `main` führt GitHub Actions die Tests durch, baut alle Archive neu auf und ersetzt die Dateien des rollenden Pre-Release (Release-Tag `edge`). Seine Archive sind checksummed, signiert und wie ein stabiles Release attestiert. `releases/latest` bedeutet immer das neueste stabile Release; der rollende Build wird nie als neuestes markiert. `numb-form update rolling` wechselt dazu, `numb-form update stable` zurück.

Der rollende Build ändert sich nach jeder Änderung in `main`, so dass seine Prüfsumme nicht auf diese Seite geschrieben werden kann. Seine Befehle überprüfen die heruntergeladene Datei mit der eigenen `SHA256SUMS` des rollenden Builds, die vor beschädigten Downloads schützt, aber nicht vor einer auf dem Server ersetzten Datei. Verwenden Sie stabile Releases für echte Arbeit. Unter Linux:

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge/numb-form_linux_amd64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge/SHA256SUMS && grep ' numb-form_linux_amd64.tar.gz$' SHA256SUMS | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date numb-typo "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && numb-form -version
```

Das erneute Ausführen eines Installationsbefehls ist auch die Neuinstallation: Es überschreibt die Programme und lässt die Konfiguration in Ruhe. `numb-form -version` zeigt `rolling-<commit>` für einen rollenden Build (`edge-<commit>` für ältere).

### Überprüfen, wer ein Release erstellt hat

Checksums beweisen, dass der Download intakt ist. Um zu beweisen, dass es vom Release-Workflow dieses Repositorys erstellt wurde, überprüfen Sie die Build-Provenienzbescheinigung mit dem GitHub CLI:

```
gh attestation verify numb-form_linux_amd64.tar.gz --repo z4dm42f5kp-ux/numb-form
```

oder die Sigstore-Signatur der Prüfsummendatei mit cosign:

```
cosign verify-blob --bundle SHA256SUMS.sigstore.json --certificate-identity-regexp '^https://github.com/z4dm42f5kp-ux/numb-form/\.github/workflows/release\.yml@refs/tags/v' --certificate-oidc-issuer https://token.actions.githubusercontent.com SHA256SUMS
```

Eine Version kann auch `SHA256SUMS.asc`, eine PGP-Signatur der Prüfsummendatei, für Systeme mit GnuPG, aber nicht cosign haben. Es wird nur gemacht, wenn der Besitzer einen Release-Schlüssel eingerichtet hat; sein öffentlicher Schlüssel ist `.github/release-key.asc` im Repository. Überprüfen Sie es mit GnuPG, im Ordner mit den Downloads:

```
curl -fLO https://raw.githubusercontent.com/z4dm42f5kp-ux/numb-form/main/.github/release-key.asc && gpg --import release-key.asc && gpg --verify SHA256SUMS.asc SHA256SUMS
```

Eine gute Signatur und eine passende Prüfsumme zusammen beweisen, dass das Archiv das ist, was der Release-Workflow erstellt hat.

## Bauen aus der Quelle

Das Projekt hat keine Module von Drittanbietern: `go.mod` hat keine `require`-Linien und der Build lädt nichts als Go selbst herunter. Die Befehle unten bauen die gleiche stabile Freigabe wie die Binärdateien oben, von seinem Tag, mit Go 1.21 oder höher. Sie verwenden einen temporären Build-Cache, der danach gelöscht wird, legen `GOTOOLCHAIN=local` so fest, dass Go niemals eine andere Toolchain herunterlädt, und installieren in `~/.local/bin` (Windows: `%LOCALAPPDATA%\Programs\numb-form`). Ein Programm, das aus der Quelle erstellt wurde, druckt `dev` als Version, und `numb-form update` ersetzt es nicht.

1. Installieren Sie Go und git, mit oder ohne Administratorrechte (unten).
2. Führen Sie den Build-Befehl Ihres Systems aus.
3. Überprüfen Sie das Ergebnis: `numb-form -version` druckt `dev`.

### Gehen Sie mit Administratorrechten

<!-- install:go-packages -->
| System | Installieren Sie Go und Git |
| --- | --- |
| Debian 13 und darauf basierende Systeme | `sudo apt-get update && sudo apt-get install -y golang-go git ca-certificates` |
| Fedora 44 | `sudo dnf install -y golang git` |
| Arch Linux (offizielle Repositories, keine AUR) | `sudo pacman -S --needed go git` |
| Gentoo | `sudo emerge --ask --noreplace dev-lang/go dev-vcs/git` |
| Void Linux | `sudo xbps-install -Sy go git` |
| openSUSE Tumbleweed und Leap 16 | `sudo zypper install -y go git` |
| Alpines Linux | `doas apk add go git` |
| NixOS | none: Der Befehl NixOS unten bringt Go und git |
| Guix-System | none: Der Guix-Befehl unten bringt Go und git |
| macOS | `brew install go git` |
| FreeBSD | `su -m root -c 'pkg install -y go git'` |
| OpenBSD | `doas pkg_add go git` |
| NetBSD | `su -m root -c 'pkgin -y install go git' && PATH=$(ls -d /usr/pkg/go1*/bin \| tail -n 1):$PATH && export PATH` |
| Windows (Command Prompt oder PowerShell 7) | `winget install --id GoLang.Go -e && winget install --id Git.Git -e` |
| Windows PowerShell 5.1 | `winget install --id GoLang.Go -e; if ($?) { winget install --id Git.Git -e }` |
<!-- /install:go-packages -->

Auf NetBSD installiert pkgsrc Gehen Sie in einen versionierten Ordner wie `/usr/pkg/go126/bin`; Der Tabelleneintrag fügt den neuesten zu `PATH` hinzu. Wenn das installierte Go älter als 1,21 ist, stoppt der Build mit einem Fehler; installieren Sie ein neueres Go.

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

### Build und Installation

Linux, macOS und die BSDs:

<!-- install:source-unix -->
```
t="$(mktemp -d)" && cd "$t" && git clone --depth 1 --branch v1.2.1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$t/cache" GOFLAGS="-trimpath -mod=readonly" CGO_ENABLED=0 GOTOOLCHAIN=local && go build -o bin/ ./cmd/... && mkdir -p "$HOME/.local/bin" && install -m 0755 bin/* "$HOME/.local/bin/" && cd / && rm -rf "$t" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:source-unix -->

NixOS (bringt Go und git nur für diesen Befehl):

<!-- install:source-nixos -->
```
nix-shell -p go git cacert --run 't="$(mktemp -d)" && cd "$t" && git clone --depth 1 --branch v1.2.1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$t/cache" GOFLAGS="-trimpath -mod=readonly" CGO_ENABLED=0 GOTOOLCHAIN=local && go build -o bin/ ./cmd/... && mkdir -p "$HOME/.local/bin" && install -m 0755 bin/* "$HOME/.local/bin/" && cd / && rm -rf "$t" && "$HOME/.local/bin/numb-form" -version'
```
<!-- /install:source-nixos -->

Guix System (bringt Go und git nur für diesen Befehl):

<!-- install:source-guix -->
```
guix shell go git nss-certs coreutils -- sh -c 'export SSL_CERT_DIR="$GUIX_ENVIRONMENT/etc/ssl/certs" GIT_SSL_CAINFO="$GUIX_ENVIRONMENT/etc/ssl/certs/ca-certificates.crt" && t="$(mktemp -d)" && cd "$t" && git clone --depth 1 --branch v1.2.1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$t/cache" GOFLAGS="-trimpath -mod=readonly" CGO_ENABLED=0 GOTOOLCHAIN=local && go build -o bin/ ./cmd/... && mkdir -p "$HOME/.local/bin" && install -m 0755 bin/* "$HOME/.local/bin/" && cd / && rm -rf "$t" && "$HOME/.local/bin/numb-form" -version'
```
<!-- /install:source-guix -->

Windows PowerShell (kein Git erforderlich):

<!-- install:source-windows -->
```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $t = Join-Path $env:TEMP ('numb-form-' + [guid]::NewGuid()); $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $t, $d | Out-Null; try { Set-Location $t; Invoke-WebRequest -UseBasicParsing 'https://github.com/z4dm42f5kp-ux/numb-form/archive/refs/tags/v1.2.1.zip' -OutFile src.zip; Expand-Archive src.zip -DestinationPath $t; Set-Location (Join-Path $t 'numb-form-1.2.1'); $env:GOCACHE = Join-Path $t 'cache'; $env:GOFLAGS = '-trimpath -mod=readonly'; $env:CGO_ENABLED = '0'; $env:GOTOOLCHAIN = 'local'; go build -o $d ./cmd/...; if ($LASTEXITCODE -ne 0) { throw 'the build failed' } } finally { Set-Location $env:TEMP; Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $t; Remove-Item Env:GOCACHE, Env:GOFLAGS, Env:CGO_ENABLED, Env:GOTOOLCHAIN -ErrorAction SilentlyContinue }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, open a new PowerShell window' }
```
<!-- /install:source-windows -->

## Bürosoftware

numb-form bearbeitet .docx, .odt, .fodt und .rtf Vorlagen selbst. .doc, .pages und .numbers Dateien benötigen LibreOffice, Apache OpenOffice oder Apple Pages/Numbers auf demselben Computer; siehe [formats.md](formats.md).

## Container

Siehe [containers.md](containers.md) für Docker, Podman (rootless, keine Administratorrechte) und Kubernetes.

## Nach der Installation

Speichern Sie Ihren Namen einmal und starten Sie dann numb-form im Arbeitsordner.

**Unix-like** (macOS, Linux, BSD), im Terminal:

```
numb-form config set person "Your Name" && cd /path/to/GWG && numb-form
```

**Windows** (10, 11), in PowerShell:

```
numb-form config set person "Your Name"; cd "H:\GWG"; numb-form
```

Die [Schritt-für-Schritt-Anleitung](guide.md) erklärt jede Frage, die sie stellt.
