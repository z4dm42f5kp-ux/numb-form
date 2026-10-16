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
`curl.exe` and `tar.exe` ship with Windows 10 1803 and later.

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

### Optional free API key

The UK Companies House register needs a free key.
All other sources need none.
Run `numb-form keys` for instructions.
