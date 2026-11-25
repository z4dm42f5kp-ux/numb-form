# Instalación de numb-form

> Traducción automática (Argos Translate) del [original en inglés](../install.md). La versión en inglés es la que vale.

¿No es un programador? La guía [step-by-step](guide.md) tiene la línea de instalación para Windows y explica qué hacer después de la instalación.

numb-form es cuatro programas estáticos, `numb-form`, `numb-commit`, `numb-date` y `numb-typo`, sin dependencias. Cada sistema recibe los mismos programas y el mismo comportamiento.

Hay dos maneras de instalarlos:

* Descargar un binario de liberación. Esto nunca necesita derechos de administrador.
* Construir de fuente con Go 1.21 o posterior, con o sin derechos de administrador.

Cada comando de abajo es una sola línea. Entendido. Se detiene en el primer paso que falla.

## Liberar binarios (sin derechos de administrador)

Cada comando instala una versión estable, llamada en el comando, y comprueba su suma de comprobación SHA-256 contra el valor escrito en el comando mismo. Una descarga cambiada o dañada se detiene con un error antes de instalar cualquier cosa. La suma de comprobación viene de esta página, no desde el servidor los archivos vienen de, por lo que un archivo reemplazado en el servidor no puede pasar. Los comandos descargan sobre HTTPS solamente, instalan en una carpeta que posee y limpian después de sí mismos.

Los comandos de esta página están escritos por un programa de los cheques firmados de la versión (`scripts/install`). Después de cada lanzamiento estable, un flujo de trabajo revisa la firma Sigstore de la versión y propone los nuevos comandos; un cheque diario captura una liberación perdida. `numb-form update` instala nuevas versiones de la misma manera.

Archives are named `numb-form_<os>_<arch>`, where *os* is `linux`, `darwin` (macOS), `windows`, `freebsd`, `openbsd` or `netbsd`, and *arch* is `amd64` or `arm64` (Linux and Windows also have `386`; Linux has `armv7`). For another archive, replace its name in both places of the command and take its checksum from [the table](#checksums).

### Linux

Un comando funciona en cada distribución; necesita `curl`, `tar`, `gzip` y `sha256sum`.

<!-- install:linux -->
```
cd "$(mktemp -d)" && curl --proto '=https' --tlsv1.2 -fsSLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_linux_amd64.tar.gz && echo 'e336d789a9e758ea1d919f2adc5521d14d159572b644ddc325174d01e0dcc6b8  numb-form_linux_amd64.tar.gz' | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:linux -->

En las computadoras ARM (arm64):

<!-- install:linux-arm64 -->
```
cd "$(mktemp -d)" && curl --proto '=https' --tlsv1.2 -fsSLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/v1.2.1/numb-form_linux_arm64.tar.gz && echo 'c44b5089a40a6883e9a1531f0bcde6e9afaf9698f05e857f11f81475042405b0  numb-form_linux_arm64.tar.gz' | sha256sum -c - && tar -xzf numb-form_linux_arm64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:linux-arm64 -->

Si el comando informa que `curl` u otro programa falta, instálelo primero:

<!-- install:linux-tools -->
| Distribución | Instalar lo que el comando necesita (una vez) |
| --- | --- |
| Debian 13 y sistemas basados en él | `sudo apt-get update && sudo apt-get install -y curl ca-certificates` |
| Fedora 44 (curl está instalado) | `sudo dnf install -y tar gzip` |
| Arch Linux | `sudo pacman -S --needed curl tar gzip` |
| Gentoo | `sudo emerge --ask --noreplace net-misc/curl` |
| Void Linux | `sudo xbps-install -Sy curl` |
| openSUSE Tumbleweed and Leap 16 | `sudo zypper install -y curl tar gzip` |
| Alpine Linux | `doas apk add curl` |
| NixOS | `nix-shell -p curl` (una concha con rizo; ejecutar el comando dentro de él) |
| Sistema Guix | `guix shell curl tar gzip coreutils` (una concha con ellos; ejecutar el comando dentro de él) |
<!-- /install:linux-tools -->

Estas distribuciones se prueban antes de cada lanzamiento menor (vX.Y.0): Debian 13, Fedora 44, Arch Linux, Gentoo, Void Linux, openSUSE Tumbleweed y Leap 16, Alpine Linux, NixOS y Guix.

### macOS

Silicio de manzana (M1 y posterior):

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

Los binarios no son notarizados por Apple; el paso `xattr` elimina la bandera de cuarentena de descarga después de que se haya verificado la suma de comprobación.

### Windows 10 y 11

Los comandos de ambos proyectiles siguen. Use el que coincida con la ventana que tiene abierto: Windows PowerShell (azul, rápido `PS C:\>`) o Command Prompt (negro, rápido `C:\>`). Se instalan en `%LOCALAPPDATA%\Programs\numb-form` y se la añaden a su propio `PATH`; no se necesitan derechos de administrador.

Windows PowerShell 5.1 no tiene `&&`. Por lo tanto, los comandos PowerShell se detienen en el primer error con `$ErrorActionPreference = 'Stop'`, que tiene el mismo efecto.

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

### Reinstalación o actualización en Windows

`numb-form update` hace esto por sí mismo (ver [usage](usage.md)); funciona igual en cada sistema. Si no puede conectarse, cierre cada ventana numb-form, ejecute esto en Windows PowerShell. Reemplaza los programas, mantiene su configuración e imprime la versión instalada.

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

En Linux, macOS y BSD, asegúrate de que `~/.local/bin` esté en `PATH`:

```
echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.profile" && . "$HOME/.profile"
```

### Sumas de comprobación

<!-- install:checksums -->
| Archivo de v1.2.1 | SHA-256 |
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

Después de cada cambio a `main`, GitHub Actions ejecuta las pruebas, reconstruye todos los archivos y reemplaza los archivos de la versión preliminar (marca de liberación `edge`). Sus archivos son verificados, firmados y certificados como una versión estable. `releases/latest` siempre significa la nueva versión estable; la construcción rodante nunca se marca más reciente. `numb-form update rolling` cambia a él, `numb-form update stable` de vuelta.

La construcción de rodaje cambia después de cada cambio a `main`, por lo que su suma de comprobación no se puede escribir en esta página. Sus comandos comprueban el archivo descargado contra el propio `SHA256SUMS` de la construcción enrollada, que protege contra las descargas dañadas pero no contra un archivo reemplazado en el servidor. Utilice versiones estables para el trabajo real. En Linux:

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge/numb-form_linux_amd64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge/SHA256SUMS && grep ' numb-form_linux_amd64.tar.gz$' SHA256SUMS | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit numb-date numb-typo "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && numb-form -version
```

Ejecutar un comando de instalación de nuevo es también cómo reinstalar: sobrescribe los programas y deja la configuración sola. `numb-form -version` muestra `rolling-<commit>` para una construcción rodante (`edge-<commit>` para los más antiguos).

### Verificar quién construyó una liberación

Checksums prueba que la descarga está intacta. Para probar que fue construido por el flujo de trabajo de liberación de este repositorio, compruebe el certificado de procedencia de construcción con el GitHub CLI:

```
gh attestation verify numb-form_linux_amd64.tar.gz --repo z4dm42f5kp-ux/numb-form
```

o la firma Sigstore del archivo checksum con cosign:

```
cosign verify-blob --bundle SHA256SUMS.sigstore.json --certificate-identity-regexp '^https://github.com/z4dm42f5kp-ux/numb-form/\.github/workflows/release\.yml@refs/tags/v' --certificate-oidc-issuer https://token.actions.githubusercontent.com SHA256SUMS
```

Una versión también puede tener `SHA256SUMS.asc`, una firma PGP del archivo de suma de comprobación, para sistemas que tienen GnuPG pero no cosign. Se hace sólo una vez que el propietario ha establecido una llave de liberación; su clave pública es `.github/release-key.asc` en el repositorio. Compruébalo con GnuPG, en la carpeta con las descargas:

```
curl -fLO https://raw.githubusercontent.com/z4dm42f5kp-ux/numb-form/main/.github/release-key.asc && gpg --import release-key.asc && gpg --verify SHA256SUMS.asc SHA256SUMS
```

Una buena firma y una suma de verificación coincidente juntos prueban que el archivo es el que el flujo de trabajo de liberación construido.

## Aprovechamiento de la fuente

El proyecto no tiene módulos de terceros: `go.mod` no tiene líneas `require` y la construcción no descarga nada más que Go. Los comandos a continuación construyen la misma liberación estable que los binarios arriba, de su etiqueta, con Go 1.21 o posterior. Utilizan un caché de construcción temporal que se elimina después, establece `GOTOOLCHAIN=local` para que Go nunca descargue otra cadena de herramientas, e instale en `~/.local/bin` (Windows: `%LOCALAPPDATA%\Programs\numb-form`). Un programa construido a partir de las impresiones fuente `dev` como su versión, y `numb-form update` no lo reemplaza.

1. Instala Go y git, con o sin derechos de administrador (bajo).
2. Ejecute el comando de construcción de su sistema.
3. Compruebe el resultado: `numb-form -version` prints `dev`.

### Ir con derechos de administrador

<!-- install:go-packages -->
| Sistema | Instalar Go y git |
| --- | --- |
| Debian 13 y sistemas basados en él | `sudo apt-get update && sudo apt-get install -y golang-go git ca-certificates` |
| Fedora 44 | `sudo dnf install -y golang git` |
| Arch Linux (repositorios oficiales, no AUR) | `sudo pacman -S --needed go git` |
| Gentoo | `sudo emerge --ask --noreplace dev-lang/go dev-vcs/git` |
| Void Linux | `sudo xbps-install -Sy go git` |
| openSUSE Tumbleweed and Leap 16 | `sudo zypper install -y go git` |
| Alpine Linux | `doas apk add go git` |
| NixOS | ninguno: el comando NixOS a continuación trae Go y git |
| Sistema Guix | ninguno: el comando Guix abajo trae Go y git |
| macOS | `brew install go git` |
| FreeBSD | `su -m root -c 'pkg install -y go git'` |
| OpenBSD | `doas pkg_add go git` |
| NetBSD | `su -m root -c 'pkgin -y install go git' && PATH=$(ls -d /usr/pkg/go1*/bin \| tail -n 1):$PATH && export PATH` |
| Windows (Command Prompt o PowerShell 7) | `winget install --id GoLang.Go -e && winget install --id Git.Git -e` |
| Windows PowerShell 5.1 | `winget install --id GoLang.Go -e; if ($?) { winget install --id Git.Git -e }` |
<!-- /install:go-packages -->

En NetBSD, pkgsrc instala Vaya en una carpeta versionada como `/usr/pkg/go126/bin`; la entrada de la tabla añade la más reciente a `PATH`. Si el Go instalado es mayor de 1.21, la construcción se detiene con un error; instalar un nuevo Go.

### Ir sin derechos de administrador

Vaya a los archivos oficiales desempacar en su carpeta de inicio. Reemplazar `go1.27.2` con la versión más reciente listada en https://go.dev/dl/. La suma de comprobación se toma de la lista de liberación de Go.dev.

Linux (uso `sha256sum`) y macOS (uso `shasum -a 256`; archivo `go1.27.2.darwin-arm64.tar.gz`):

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

En los BSD, utilice el binario de liberación cuando no tenga derechos de administrador; Los archivos de Go para ellos necesitan los propios paquetes del sistema de todos modos.

### Construir e instalar

Linux, macOS y BSD:

<!-- install:source-unix -->
```
t="$(mktemp -d)" && cd "$t" && git clone --depth 1 --branch v1.2.1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$t/cache" GOFLAGS="-trimpath -mod=readonly" CGO_ENABLED=0 GOTOOLCHAIN=local && go build -o bin/ ./cmd/... && mkdir -p "$HOME/.local/bin" && install -m 0755 bin/* "$HOME/.local/bin/" && cd / && rm -rf "$t" && "$HOME/.local/bin/numb-form" -version
```
<!-- /install:source-unix -->

NixOS (Brings Go and git for this command only):

<!-- install:source-nixos -->
```
nix-shell -p go git cacert --run 't="$(mktemp -d)" && cd "$t" && git clone --depth 1 --branch v1.2.1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$t/cache" GOFLAGS="-trimpath -mod=readonly" CGO_ENABLED=0 GOTOOLCHAIN=local && go build -o bin/ ./cmd/... && mkdir -p "$HOME/.local/bin" && install -m 0755 bin/* "$HOME/.local/bin/" && cd / && rm -rf "$t" && "$HOME/.local/bin/numb-form" -version'
```
<!-- /install:source-nixos -->

Sistema Guix (brings Go and git for this command only):

<!-- install:source-guix -->
```
guix shell go git nss-certs coreutils -- sh -c 'export SSL_CERT_DIR="$GUIX_ENVIRONMENT/etc/ssl/certs" GIT_SSL_CAINFO="$GUIX_ENVIRONMENT/etc/ssl/certs/ca-certificates.crt" && t="$(mktemp -d)" && cd "$t" && git clone --depth 1 --branch v1.2.1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$t/cache" GOFLAGS="-trimpath -mod=readonly" CGO_ENABLED=0 GOTOOLCHAIN=local && go build -o bin/ ./cmd/... && mkdir -p "$HOME/.local/bin" && install -m 0755 bin/* "$HOME/.local/bin/" && cd / && rm -rf "$t" && "$HOME/.local/bin/numb-form" -version'
```
<!-- /install:source-guix -->

Windows PowerShell (sin git needed):

<!-- install:source-windows -->
```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $t = Join-Path $env:TEMP ('numb-form-' + [guid]::NewGuid()); $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $t, $d | Out-Null; try { Set-Location $t; Invoke-WebRequest -UseBasicParsing 'https://github.com/z4dm42f5kp-ux/numb-form/archive/refs/tags/v1.2.1.zip' -OutFile src.zip; Expand-Archive src.zip -DestinationPath $t; Set-Location (Join-Path $t 'numb-form-1.2.1'); $env:GOCACHE = Join-Path $t 'cache'; $env:GOFLAGS = '-trimpath -mod=readonly'; $env:CGO_ENABLED = '0'; $env:GOTOOLCHAIN = 'local'; go build -o $d ./cmd/...; if ($LASTEXITCODE -ne 0) { throw 'the build failed' } } finally { Set-Location $env:TEMP; Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $t; Remove-Item Env:GOCACHE, Env:GOFLAGS, Env:CGO_ENABLED, Env:GOTOOLCHAIN -ErrorAction SilentlyContinue }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, open a new PowerShell window' }
```
<!-- /install:source-windows -->

## Programas informáticos de oficina

numb-form edita .docx, .odt, .fodt y .rtf plantillas en sí. .doc, .pages y .numbers archivos necesitan LibreOffice, Apache OpenOffice o Apple Pages/Numbers en el mismo equipo; vea [formats.md](formats.md).

## Containers

Ver [containers.md](containers.md) para Docker, Podman (sin raíces, sin derechos de administrador) y Kubernetes.

## Después de instalar

Guarda tu nombre una vez, y luego comienza numb-form dentro de la carpeta de trabajo.

**Unix-like** (macOS, Linux, BSD), en Terminal:

```
numb-form config set person "Your Name" && cd /path/to/GWG && numb-form
```

**Windows** (10, 11), en PowerShell:

```
numb-form config set person "Your Name"; cd "H:\GWG"; numb-form
```

La guía [step-by-step](guide.md) explica cada pregunta que hace.
