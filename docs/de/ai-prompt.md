# Der AI prompt

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../ai-prompt.md). Maßgeblich ist die englische Fassung.

numb-form macht keine Online-Lookups. Es fragt Sie nach der Rechtsform, der Adresse und den Vertretern. Um sie zu finden, geben Sie einem KI-Assistenten mit Web-Suche diese Aufforderung, dann den Namen des Unternehmens. Überprüfen Sie die Antwort mit Ihren internen Quellen, bevor Sie sie in numb-form eingeben.

Die Eingabeaufforderung ist die Datei [ai-prompt.txt](../ai-prompt.txt). Es ist Klartext und ist nicht Teil der Programme.

## Verwenden Sie

1. Öffnen Sie die Eingabeaufforderung und kopieren Sie alles.
2. Fügen Sie es in den AI-Assistenten ein.
3. Fügen Sie darunter den Namen des Unternehmens genau so hinzu, wie die Tabelle es schreibt, zum Beispiel seinen * Kundennamen *.
4. Senden Sie es. Die Antwort hat drei Zeilen: *Rechtsform *, *Hauptquartier / Adresse * und * Authorized Representatives *.
5. Geben Sie diese Werte an numb-form, wenn es danach fragt.

Die Standardtexte in der Eingabeaufforderung sind die, die numb-form schreibt, wenn Sie Enter in einem leeren Feld drücken. Eine Zeile, die mit `Caution:` beginnt, benötigt eine Überprüfung mit internen Quellen.

## Herunterladen

Jede Version hat die Eingabeaufforderung als separaten Download, `ai-prompt.txt`, mit seiner Prüfsumme in `SHA256SUMS`.

**Same auf allen Systemen**: Öffnen Sie <https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/ai-prompt.txt> in einem Browser und speichern Sie die Seite.

**Unix-like** (macOS, Linux, BSD): Laden Sie es in Downloads herunter und überprüfen Sie seine Prüfsumme.

```
cd "$HOME/Downloads" && curl -fLO https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/ai-prompt.txt && curl -fL -o ai-prompt.SHA256SUMS https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && grep ' ai-prompt.txt$' ai-prompt.SHA256SUMS | { sha256sum -c - 2>/dev/null || shasum -a 256 -c -; } && rm ai-prompt.SHA256SUMS && echo OK
```

Verwenden Sie auf FreeBSD und OpenBSD `fetch` oder `ftp` anstelle von `curl` und `sha256 -C ai-prompt.SHA256SUMS ai-prompt.txt` für die Prüfung.

**Windows** (10, 11; PowerShell 5.1): Laden Sie es in Downloads herunter, überprüfen Sie seine Prüfsumme und öffnen Sie es in Notepad.

```
& { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $d = Join-Path $env:USERPROFILE 'Downloads'; $f = Join-Path $d 'ai-prompt.txt'; $s = Join-Path $d 'ai-prompt.SHA256SUMS'; try { Invoke-WebRequest -UseBasicParsing "$u/ai-prompt.txt" -OutFile $f; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile $s; $h = (Get-FileHash $f -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path $s -SimpleMatch "$h  ai-prompt.txt" -Quiet)) { Remove-Item $f; throw 'CHECKSUM MISMATCH' } } finally { Remove-Item -ErrorAction SilentlyContinue $s }; notepad $f }
```

## SEE ALSO

[der guide](guide.md), [usage.md](usage.md), [install.md](install.md)
