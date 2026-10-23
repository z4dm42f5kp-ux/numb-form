# The AI prompt

> Traducción automática (Argos Translate) del [original en inglés](../ai-prompt.md). La versión en inglés es la que vale.

numb-form no hace ninguna búsqueda en línea. Le pide la forma legal, la dirección y los representantes. Para encontrarlos, dale a un asistente de inteligencia artificial con la búsqueda web este aviso, luego el nombre de la compañía. Comprueba la respuesta contra tus fuentes internas antes de escribirla en numb-form.

El aviso es el archivo [ai-prompt.txt](../ai-prompt.txt). Es texto claro y no es parte de los programas.

## Usarlo

1. Abre el aviso y copia todo.
2. Pruébalo en el asistente de IA.
3. A continuación, añadir el nombre de la empresa exactamente como la hoja de cálculo lo escribe, por ejemplo su *Nombre del cliente*.
4. Envíenlo. La respuesta tiene tres líneas: *Formo legal*, *Headquarters/Address* y *Authorized Representatives*.
5. Dale esos valores a numb-form cuando les pida.

The standard texts in the prompt are the ones numb-form writes when you press Enter on an empty field. A line that starts with `Caution:` needs a check against internal sources.

## Descargarlo

Cada lanzamiento tiene el impulso como una descarga separada, `ai-prompt.txt`, con su suma de comprobación en `SHA256SUMS`.

**Same en todos los sistemas**: abrir <https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/ai-prompt.txt> en un navegador y guardar la página.

**Unix-like** (macOS, Linux, BSD): descargarlo en Downloads y comprobar su suma de comprobación.

```
cd "$HOME/Downloads" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/ai-prompt.txt && curl -fL -o ai-prompt.SHA256SUMS https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' ai-prompt.txt$' ai-prompt.SHA256SUMS | { sha256sum -c - 2>/dev/null || shasum -a 256 -c -; } && rm ai-prompt.SHA256SUMS && echo OK
```

En FreeBSD y OpenBSD, utilice `fetch` o `ftp` en lugar de `curl`, y `sha256 -C ai-prompt.SHA256SUMS ai-prompt.txt` para el cheque.

**Windows** (10, 11; PowerShell 5.1): descargarlo en Downloads, comprobar su suma de comprobación y abrirlo en Notepad.

```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $d = Join-Path $env:USERPROFILE 'Downloads'; $f = Join-Path $d 'ai-prompt.txt'; $s = Join-Path $d 'ai-prompt.SHA256SUMS'; try { Invoke-WebRequest -UseBasicParsing "$u/ai-prompt.txt" -OutFile $f; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile $s; $h = (Get-FileHash $f -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path $s -SimpleMatch "$h  ai-prompt.txt" -Quiet)) { Remove-Item $f; throw 'CHECKSUM MISMATCH' } } finally { Remove-Item -ErrorAction SilentlyContinue $s }; notepad $f }
```

## SEE ALSO

[the guide](guide.md), [usage.md](usage.md), [install.md](install.md)
