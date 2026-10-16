# Instalación de forma de tuerca

> Traducción automática (Argos Translate) del [original en inglés](../install.md). La versión en inglés es la que vale.

numb-form es dos programas, `numb-form` y `numb-commit`.
Son binarios estáticos sin dependencia.

Hay dos maneras de instalarlos:

* Descargar un binario de lanzamiento firmado.
* Construir de la fuente con Go 1.21 o posterior.

Cada comando de abajo es una sola línea.
Entendido.
Se detiene en el primer paso que falla.

## Windows 10 y 11

Ejecutar estos en **Command Prompt** (`cmd.exe`).
`curl.exe` y `tar.exe` nave con Windows 10 1803 y posterior.

### Liberar binario

Utilice `amd64` para la mayoría de PCs y `arm64` para dispositivos ARM (por ejemplo, Surface Pro X).
Esto descarga el lanzamiento, comprueba su suma de comprobación SHA-256 contra
`SHA256SUMS` y lo desempaqueta en `%USERPROFILE%\numb-form`:

```
(if not exist "%USERPROFILE%\numb-form" mkdir "%USERPROFILE%\numb-form") && cd /d "%USERPROFILE%\numb-form" && curl.exe -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_windows_amd64.zip && curl.exe -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && powershell -NoProfile -Command "$h=(Get-FileHash numb-form_windows_amd64.zip -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch ($h+'  numb-form_windows_amd64.zip') -Quiet)) { Write-Error 'CHECKSUM MISMATCH'; exit 1 }" && tar -xf numb-form_windows_amd64.zip && del numb-form_windows_amd64.zip && echo OK
```

Para verificar también quién construyó el lanzamiento, instale el GitHub CLI y compruebe el
atestiguación de la procedencia:

```
winget install --id GitHub.cli -e && cd /d "%USERPROFILE%\numb-form" && gh attestation verify numb-form.exe --repo z4dm42f5kp-ux/numb-form && gh attestation verify numb-commit.exe --repo z4dm42f5kp-ux/numb-form
```

Añada la carpeta a su usuario `PATH` (hace efecto en nuevas ventanas):

```
powershell -NoProfile -Command "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path','User') + ';' + $env:USERPROFILE + '\numb-form', 'User')" && echo Open a new Command Prompt
```

### De la fuente

Instala Ve con el alaet, luego abre un **new** Command Prompt para que `go` esté en el
`PATH`:

```
winget install --id GoLang.Go -e
```

Construir en un temporal Vaya caché que se elimina después, por lo que no construir caché
se deja en el disco:

```
cd /d "%USERPROFILE%" && curl.exe -fL -o numb-form-src.zip https://github.com/z4dm42f5kp-ux/numb-form/archive/refs/heads/main.zip && tar -xf numb-form-src.zip && del numb-form-src.zip && cd numb-form-main && set "GOCACHE=%TEMP%\numb-form-gocache" && set "GOFLAGS=-trimpath -mod=readonly" && set "CGO_ENABLED=0" && go test ./... && go build -o "%USERPROFILE%\numb-form\numb-form.exe" ./cmd/numb-form && go build -o "%USERPROFILE%\numb-form\numb-commit.exe" ./cmd/numb-commit && go clean -cache && cd /d "%USERPROFILE%" && rmdir /s /q numb-form-main && echo OK
```

El proyecto no tiene módulos Go de terceros, así que nada se descarga durante
la construcción. Ver [SECURITY.md](SECURITY.md).

## Liberar binarios en otros sistemas

Los archivos se llaman `numb-form_<os>_<arch>.tar.gz`, donde *os* es
`linux`, `darwin` (macOS), `freebsd`, `openbsd` o `netbsd`, y *arch* es
`amd64` o `arm64` (Linux también tiene `386` y `armv7`).
Cada archivo tiene `numb-form`, `numb-commit`, `README.md` y `LICENSE`.
Los ejemplos utilizan `amd64`; reemplazarlo para adaptarse a la máquina.

Cada comando descarga en un directorio temporal, verifica el SHA-256
checksum, se instala en `~/.local/bin` y elimina el directorio temporal.
Asegúrese de que `~/.local/bin` esté en `PATH`.

### Linux (Debian, Ubuntu, Arch, Fedora, Gentoo, Void, Alpine)

Necesita `curl`, `tar` y `sha256sum`, que cada distribución lista
proporciona (Alpine a través de BusyBox).

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_linux_amd64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' numb-form_linux_amd64.tar.gz$' SHA256SUMS | sha256sum -c - && tar -xzf numb-form_linux_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### macOS

```
cd "$(mktemp -d)" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_darwin_arm64.tar.gz && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' numb-form_darwin_arm64.tar.gz$' SHA256SUMS | shasum -a 256 -c - && tar -xzf numb-form_darwin_arm64.tar.gz && { xattr -d com.apple.quarantine numb-form numb-commit 2>/dev/null || true; } && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

Use `darwin_amd64` en Intel Macs.
Los binarios no son notarizados por Apple;
el paso `xattr` elimina la bandera de cuarentena de descarga después del checksum
ha sido verificada.

### FreeBSD

```
cd "$(mktemp -d)" && fetch https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_freebsd_amd64.tar.gz && fetch https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && sha256 -c "$(awk '$2=="numb-form_freebsd_amd64.tar.gz"{print $1}' SHA256SUMS)" numb-form_freebsd_amd64.tar.gz && tar -xzf numb-form_freebsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### OpenBSD

El archivo `SHA256` utiliza el formato BSD que sha256(1) verifica directamente.

```
cd "$(mktemp -d)" && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_openbsd_amd64.tar.gz && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256 && sha256 -C SHA256 numb-form_openbsd_amd64.tar.gz && tar -xzf numb-form_openbsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### NetBSD

```
cd "$(mktemp -d)" && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_netbsd_amd64.tar.gz && ftp https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256 && test "$(cksum -a SHA256 numb-form_netbsd_amd64.tar.gz)" = "$(grep '(numb-form_netbsd_amd64.tar.gz)' SHA256)" && tar -xzf numb-form_netbsd_amd64.tar.gz && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && cd / && rm -rf "$OLDPWD" && echo OK
```

### Verificar quién construyó una liberación

Checksums prueba que la descarga está intacta.
Para probar que fue construido por el flujo de trabajo de liberación de este repositorio, verifique el
GitHub construye el certificado de procedencia con el GitHub CLI:

```
gh attestation verify numb-form_linux_amd64.tar.gz --repo z4dm42f5kp-ux/numb-form
```

o la firma Sigstore del archivo checksum con cosign:

```
cosign verify-blob --bundle SHA256SUMS.sigstore.json --certificate-identity-regexp '^https://github.com/z4dm42f5kp-ux/numb-form/\.github/workflows/release\.yml@refs/tags/v' --certificate-oidc-issuer https://token.actions.githubusercontent.com SHA256SUMS
```

## Aprovechamiento de la fuente en otros sistemas

Instalar Go 1.21 o posterior y git con el administrador del paquete del sistema, luego
ejecutar el comando de construcción abajo.

| Sistema | Instalar Go y git |
| --- | --- |
| Debian 13, Ubuntu 24.04 o posterior | `sudo apt-get update && sudo apt-get install -y golang-go git` |
| Debian 12 | `echo 'deb http://deb.debian.org/debian bookworm-backports main' \| sudo tee /etc/apt/sources.list.d/backports.list && sudo apt-get update && sudo apt-get install -y -t bookworm-backports golang-go && sudo apt-get install -y git` |
| Arch Linux (repositorios oficiales solamente) | `sudo pacman -S --needed go git` |
| Fedora, RHEL, Rocky, Alma | `sudo dnf install -y golang git` |
| Gentoo | `sudo emerge --ask --noreplace dev-lang/go dev-vcs/git` |
| Void Linux | `sudo xbps-install -Sy go git` |
| Alpine Linux | `doas apk add go git` |
| macOS | `brew install go git` |
| FreeBSD | `su -m root -c 'pkg install -y go git'` |
| OpenBSD | `doas pkg_add go git` |
| NetBSD | `su -m root -c 'pkgin -y install go git'` |

Luego construir, probar e instalar, utilizando un temporal Go build cache that is
suprimida después:

```
cd "$(mktemp -d)" && git clone --depth 1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && export GOCACHE="$(mktemp -d)" GOFLAGS='-trimpath -mod=readonly' CGO_ENABLED=0 && go test ./... && go build -o numb-form ./cmd/numb-form && go build -o numb-commit ./cmd/numb-commit && mkdir -p "$HOME/.local/bin" && install -m 0755 numb-form numb-commit "$HOME/.local/bin/" && go clean -cache && rm -rf "$GOCACHE" && cd / && rm -rf "$OLDPWD/.." && echo OK
```

Si el Go instalado es mayor de 1.21, la construcción se detiene con un error;
instalar un nuevo Go from the backports or ports collection rather than
dejándolo descargar una cadena de herramientas.

El proyecto no tiene módulos de terceros: `go.mod` no tiene líneas `require`
y la construcción no descarga nada.

## Containers

Ver [containers.md](containers.md) para Docker, Podman y Kubernetes.

### Opcional API gratuita clave

El registro UK Companies House necesita una llave gratuita.
Todas las demás fuentes no necesitan nada.
Ejecute `numb-form keys` para instrucciones.
