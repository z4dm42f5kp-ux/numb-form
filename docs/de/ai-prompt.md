# Der AI prompt

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../ai-prompt.md). Maßgeblich ist die englische Fassung.

numb-form macht keine Online-Lookups. Es fragt Sie nach der Rechtsform, der Adresse und den Vertretern. Um sie zu finden, geben Sie einem KI-Assistenten mit Web-Suche diese Aufforderung und den Namen des Unternehmens. Überprüfen Sie die Antwort mit Ihren internen Quellen, bevor Sie sie in numb-form eingeben.

Die Eingabeaufforderung ist die Datei [ai-prompt.txt](../ai-prompt.txt). Es ist Klartext und ist nicht Teil der Programme.

## Verwenden Sie

1. Öffnen Sie die Eingabeaufforderung und kopieren Sie alles.
2. Fügen Sie es in einen KI-Assistenten ein, der das Web durchsuchen kann.
3. Am Ende, nach `Company name:`, fügen Sie den Namen des Unternehmens genau so hinzu, wie die Tabelle es schreibt. Fügen Sie nach `Country` das Land hinzu, wenn Sie es kennen.
4. Senden Sie es. Die Antwort ist drei Zeilen:

   ```
   Legal Form: S.A. de C.V.
   Headquarters/Address: Calle Foo 12, 06600, Ciudad de México, CDMX, Mexico
   Authorized Representatives: Ana María Pérez Gómez (Legal Representative)
   ```

5. Überprüfen Sie jede Zeile mit Ihren internen Quellen.
6. Fügen Sie jeden Wert in numb-form ein, wenn er danach gefragt wird.

Die Aufforderung sagt dem Assistenten:

* welches Unternehmen zählt (genaue Rechtsbezeichnung und Rechtsform, niemals Mutter-, Tochter- oder ähnliche Bezeichnung);
* welche Quellen in der Reihenfolge zu vertrauen sind: offizielle Register, die eigenen rechtlichen Hinweise des Unternehmens, der LEI-Index, dann Geschäftsinformationsseiten und Nachrichten, mit den Registern, die für jedes Land funktionieren,
* wie man jedes Feld schreibt: die Rechtsform, wie der Name es schreibt (`GmbH`, `S.A.`, `Ltda.`), die Adresse auf einer Zeile, die Vertreter als `Full Name (Role 1, Role 2)` mit jeder Rolle in Englisch (`General Director`, niemals `Director General`) und keine Erklärung, warum eine Person es hält,
* ein Feld mit `Caution: This field might differ from internal sources or be inaccurate due to insufficient information online` zu beenden, wenn keine offizielle Quelle es bestätigt,
* den Standardtext von numb-form zu schreiben, wenn nichts gefunden wird,
* Klartext ohne Em Dashes, En Dashes oder Semikolons zu schreiben.

Eine Zeile mit `Caution:` benötigt eine Überprüfung mit internen Quellen, bevor Sie sie verwenden.

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
