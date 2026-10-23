# The AI prompt

numb-form makes no online lookups.
It asks you for the legal form, the address and the representatives.
To find them, give an AI assistant with web search this prompt, then the
company's name.
Check the answer against your internal sources before you type it into
numb-form.

The prompt is the file [ai-prompt.txt](ai-prompt.txt).
It is plain text and is not part of the programs.

## Using it

1. Open the prompt and copy all of it.
2. Paste it into the AI assistant.
3. Below it, add the company's name exactly as the spreadsheet writes
   it, for example its *Customer Name*.
4. Send it. The answer has three lines: *Legal Form*,
   *Headquarters/Address* and *Authorized Representatives*.
5. Give those values to numb-form when it asks for them.

The standard texts in the prompt are the ones numb-form writes when you
press Enter on an empty field.
A line that starts with `Caution:` needs a check against internal
sources.

## Downloading it

Every release has the prompt as a separate download, `ai-prompt.txt`,
with its checksum in `SHA256SUMS`.

**Same on all systems**: open
<https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/ai-prompt.txt>
in a browser and save the page.

**Unix-like** (macOS, Linux, BSD): download it into Downloads and check
its checksum.

```
cd "$HOME/Downloads" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/ai-prompt.txt && curl -fL -o ai-prompt.SHA256SUMS https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' ai-prompt.txt$' ai-prompt.SHA256SUMS | { sha256sum -c - 2>/dev/null || shasum -a 256 -c -; } && rm ai-prompt.SHA256SUMS && echo OK
```

On FreeBSD and OpenBSD, use `fetch` or `ftp` instead of `curl`, and
`sha256 -C ai-prompt.SHA256SUMS ai-prompt.txt` for the check.

**Windows** (10, 11; PowerShell 5.1): download it into Downloads, check
its checksum and open it in Notepad.

```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $d = Join-Path $env:USERPROFILE 'Downloads'; $f = Join-Path $d 'ai-prompt.txt'; $s = Join-Path $d 'ai-prompt.SHA256SUMS'; try { Invoke-WebRequest -UseBasicParsing "$u/ai-prompt.txt" -OutFile $f; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile $s; $h = (Get-FileHash $f -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path $s -SimpleMatch "$h  ai-prompt.txt" -Quiet)) { Remove-Item $f; throw 'CHECKSUM MISMATCH' } } finally { Remove-Item -ErrorAction SilentlyContinue $s }; notepad $f }
```

## SEE ALSO

[the guide](guide.md), [usage.md](usage.md), [install.md](install.md)
