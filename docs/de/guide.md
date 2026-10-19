# Schrittweise Anleitung

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../guide.md). Maßgeblich ist die englische Fassung.

Dieser Leitfaden richtet sich an alle, die die Formulare zur Risikobewertung ausfüllen. Sie benötigen keine technischen Kenntnisse. Sie kopieren, fügen nur ein und drücken Enter.

## Was numb-form tut

Ihr Team führt einen Arbeitsordner, zum Beispiel **GWG**. Es sieht so aus:

```
GWG
├── 3. Übersichten
│   └── the Excel list of companies
├── Appendix A - Risk Assesment.docx      (the empty form)
├── Foo GmbH                              (one folder per company
└── Bar S.A.                               that already has a form)
```

numb-form geht durch die Excel-Liste von oben nach unten. Für jedes Unternehmen, das noch keinen Ordner hat, ist es:

1. erstellt einen Ordner mit dem Firmennamen,
2. eine Kopie des Formulars mit dem Namen `Appendix A - Risk Assesment - <company>.docx` einreicht,
3. füllt Teil A des Vordrucks aus (Name, Rechtsform, Anschrift, Vertreter, Datum und Name),
4. zeigt Ihnen, was es geschrieben hat und wartet auf Ihr OK.

Unternehmen, die bereits einen Ordner haben, werden übersprungen. Die Excel-Liste wird nie geändert. Teil B der Form wird nie berührt.

## 1. Installieren Sie es (einmal)

1. Öffnen Sie das Menü **Start**, geben Sie `PowerShell` ein und öffnen Sie **Windows PowerShell**. Ein blaues Fenster erscheint.
2. Kopieren Sie diese ganze Zeile, fügen Sie sie in das blaue Fenster ein (Rechtsklick-Pasten) und drücken Sie **Enter**:

   ```
   & { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Set-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; 'OK, open a new PowerShell window' }
   ```

3. Warten Sie, bis es `OK, open a new PowerShell window` sagt. Schließen Sie das Fenster.

Sie benötigen keine Administratorrechte. Wenn dort `CHECKSUM MISMATCH` steht, wurde der Download beschädigt oder geändert: Fahren Sie nicht fort und sagen Sie, wer sich um das Tool kümmert.

Um numb-form später zu aktualisieren, siehe [Reinstall oder update](#reinstall-or-update).

## 2. Speichern Sie Ihren Namen (einmal)

Dein Name geht in das Feld *Person Processing the Form*. Öffnen Sie Windows PowerShell erneut und geben Sie mit Ihrem eigenen Namen ein:

```
numb-form config set person "Erika Mustermann"
```

Es beantwortet `Saved person = Erika Mustermann`. Wenn Sie diesen Schritt überspringen, fragt numb-form jedes Mal nach Ihrem Namen.

## 3. Formulare ausfüllen (jedes Mal)

1. Öffnen Sie den Ordner **GWG** in File Explorer, dem Ordner, der `3. Übersichten` enthält.
2. Klicken Sie einmal in die **addressleiste** oben im Fenster (wo der Ordnerpfad angezeigt wird), geben Sie `powershell` ein und drücken Sie **Enter**. Es öffnet sich ein blaues Fenster, das bereits im GWG-Ordner "in" ist.
3. Geben Sie `numb-form` ein und drücken Sie **Enter**.

numb-form zeigt zuerst, was es gefunden hat:

```
Working folder: H:\GWG
Spreadsheet:    3. Übersichten\Kundenliste.xlsx (read only)
Template:       Appendix A - Risk Assesment.docx
```

Wenn das richtig ist, geht es durch die Unternehmen eins nach dem anderen.

## 4. Die Fragen und was zu beantworten ist

Viele Fragen zeigen einen Vorschlag in eckigen Klammern, wie `[GmbH]` oder `[y/N]`. **Pressing Enter akzeptiert den Vorschlag** (für `[y/N]` ist das *nein*, für `[Y/n]` ist es *ja*). Um eine andere Antwort zu geben, geben Sie sie ein und drücken Sie Enter.

Jedes Unternehmen beginnt mit einer Linie wie `[3/11] Foo GmbH  (Sheet1!C5)`: das dritte von elf Unternehmen, und die Zelle, von der es genommen wurde.

Dann sieht numb-form das Unternehmen selbst, was einige Sekunden dauert: offizielle Register, die eigene Website des Unternehmens (es findet es allein), offizielle Nummern, die dort gedruckt werden, Regierungsblätter und eine Websuche. Es zeigt die von ihm verwendete Website, jeden dort gedruckten Registereintrag (z. B. `HRB 12345`) und wie viele Adressen und Vertreter es gefunden hat. Die Rechtsform wird dem Firmennamen entnommen, ohne zu fragen. *Commercial Registry Number* erhält immer den Satz "Keine kommerzielle Registernummer wurde online gefunden, konsultieren Sie interne Quellen für weitere Informationen"; Um die Nummer stattdessen einzugeben, wählen Sie `e` (Bearbeitung) und Feld 4 am Ende.

| numb-form fragt | Was es bedeutet | Was Sie normalerweise tun |
| --- | --- | --- |
| `Full name of the person processing the form:` | Nur gefragt, ob Sie Schritt 2 übersprungen haben | Geben Sie Ihren Namen ein; Enter bei der nächsten Frage verwendet es für jedes Formular in diesem Lauf |
| `Is the existing folder "Foo Gmbh" this company's folder? (...) [y/N]` | Ein Ordner mit einem sehr ähnlichen Namen existiert bereits | `y`, wenn es sich um dieselbe Firma handelt (es wird nichts erstellt), andernfalls Enter |
| `Legal form (Enter = keep, '-' = blank) [GmbH]` | Gefragt nur, wenn der Firmenname keine Rechtsform wie `GmbH` oder `S.A.` hat; der Vorschlag stammt aus dem Register oder der Website | Eingang |
| `Headquarters / Address candidates` und eine nummerierte Liste, dann `Choose a number, 'm' to type it, 'b' to leave blank [1]` | Adressen gefunden, das zuverlässigste (offizielle Register) zuerst; jedes zeigt seine Quelle | Enter nimmt Nummer 1; geben Sie eine andere Nummer ein; `m`, um die Adresse selbst einzugeben; `b`, um sie leer zu lassen |
| `Address (Enter = keep, or correct it) [...]` | Die Adresse, die Sie ausgewählt haben, damit Sie einen Tippfehler korrigieren können | Geben Sie die korrigierte Adresse ein |
| `Add the caution text to the address (not fully verified)? [y/N]` | Fügt den Satz "Vorsicht" hinzu: Dieses Feld kann sich von internen Quellen unterscheiden ... Die gleiche Frage kann für die Vertreter kommen | Press Enter: Der Vorschlag hängt bereits davon ab, wie zuverlässig die Quelle ist |
| `Authorized representative candidates` und eine nummerierte Liste, dann `Numbers to include (e.g. 1,3), 'a' all, 'm' type them, 'n' none found [1,2]` | Personen gefunden, jeweils markiert `[verified]` (offizielles Register) oder `[unverified]` | Enter nimmt die verifizierten; Typnummern wie `1,3`; `a` für alle; `m`, um sie selbst einzugeben; `n` schreibt "Keine Vertreter wurden online gefunden ..." |
| `Is this correct? [a]ccept, [e]dit a field, [r]eject:` | Das Formular wurde geschrieben; numb-form zeigt Teil A als das Formular jetzt liest | Geben Sie `a` ein, wenn es richtig ist; `e`, um eine Zeile zu ändern; `r`, um sie wegzuwerfen. Es gibt hier keinen Standard: Sie müssen einen Brief eingeben |

Am Ende zeigt numb-form eine Zusammenfassung, zum Beispiel `Done: 3 accepted, 8 already had a folder, 0 rejected, 0 need review, 0 skipped, 0 errors.` und wo sich das Protokoll befindet.

Sie können jederzeit mit **Ctrl+C** aufhören. Formulare, die Sie bereits akzeptiert haben, bleiben; das, an dem Sie gearbeitet haben, wird entfernt.

## 5. Was sich am Laufwerk geändert hat

* Ein neuer Ordner pro akzeptiertem Unternehmen, mit einem ausgefüllten Formular darin.
* Nichts anderes. Keine vorhandene Datei oder Ordner wird geändert oder gelöscht, und die Excel-Liste wird nie geändert.
* Eine abgelehnte Form lässt nichts zurück.

Um ein akzeptiertes Formular rückgängig zu machen, löschen Sie den Firmenordner.

Eine Aufzeichnung von allem, was numb-form getan hat (das Protokoll) wird in Ihrem eigenen Benutzerordner gespeichert, nicht in GWG: `%LOCALAPPDATA%\numb-form\runs`. Die letzte Zeile jedes Laufs zeigt die genaue Datei an.

## 6. Überprüfen Sie Ihre Formularvorlage (einmal, empfohlen)

Vor dem ersten echten Durchlauf lassen Sie numb-form die Vorlage testen. Geben Sie in einem PowerShell-Fenster, das in GWG geöffnet ist (siehe Schritt 3), ein:

```
numb-form check-template "Appendix A - Risk Assesment.docx" "$env:USERPROFILE\Desktop\sample.docx"
```

Es liest die Vorlage, ohne sie zu ändern, und speichert ein ausgefülltes Beispiel auf Ihrem Desktop. Öffnen Sie `sample.docx` und überprüfen Sie, ob sich die Werte an den richtigen Stellen befinden. Danach löschen.

## Möchten Sie lieber alles überprüfen, bevor es in GWG landet?

Wechseln Sie einmal in den Überprüfungsmodus:

```
numb-form config set mode review
```

Von da an kopiert numb-form GWG in Ihren **Downloads** Ordner und funktioniert nur noch auf der Kopie. Schauen Sie sich die neuen Ordner dort an. Wenn alles in Ordnung ist, geben Sie `numb-commit` ein: Es kopiert die neuen Ordner in den echten GWG-Ordner, nachdem Sie noch einmal gefragt wurden. Wechseln Sie zurück mit `numb-form config set mode direct`.

## Neuinstallation oder Update

Der einfache Weg: Windows PowerShell öffnen und Typ

```
numb-form update
```

Es lädt die neueste Version herunter, überprüft sie, ersetzt das Programm und druckt die neue Version, z. B. `Updated to numb-form v0.1.2 (stable)`. Ihre gespeicherten Einstellungen (Ihr Name, die Überschrift, der Modus) bleiben erhalten. Es bleibt auf der Version, die Sie haben: Die Entwicklungsversion aktualisiert ständig auf den neuesten Entwicklungs-Build, die stabile Version auf das neueste stabile Release. Um zu wechseln, Typ `numb-form update edge` (Entwicklung) oder `numb-form update stable`.

Versionen 0.1.1 und älter haben `numb-form update` nicht. Für diejenigen, oder wenn `numb-form update` keine Verbindung herstellen kann, verwenden Sie eine der beiden folgenden Zeilen. Geben Sie sie nicht ein: Öffnen Sie diese Seite in Ihrem Browser, klicken Sie auf das Kopiersymbol am rechten Rand des grauen Feldes und klicken Sie mit der rechten Maustaste auf das PowerShell-Fenster, um es einzufügen. Schließen Sie zuerst jedes numb-form-Fenster.

**Stable version** (die neueste veröffentlichte Version; verwenden Sie diese für echte Arbeit):

```
& { $ErrorActionPreference = 'Stop'; if (Get-Process numb-form, numb-commit -ErrorAction SilentlyContinue) { throw 'Close every numb-form window first, then run this again' }; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Push-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Remove-Item -Force -ErrorAction SilentlyContinue numb-form.exe, numb-commit.exe; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS; Pop-Location }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, reinstalled. Your settings are kept. Open a new PowerShell window' }
```

**Entwicklungsversion** (die neuesten Änderungen, automatisch nach jedem Update auf GitHub neu erstellt; verwenden Sie dies, um sofort eine Korrektur zu versuchen):

```
& { $ErrorActionPreference = 'Stop'; if (Get-Process numb-form, numb-commit -ErrorAction SilentlyContinue) { throw 'Close every numb-form window first, then run this again' }; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Push-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Remove-Item -Force -ErrorAction SilentlyContinue numb-form.exe, numb-commit.exe; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS; Pop-Location }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, reinstalled. Your settings are kept. Open a new PowerShell window' }
```

Beide drucken die von ihnen installierte Version, z. B. `v0.1.1` (stabil) oder `edge-3f976cc` (Entwicklung), und enden mit `OK, reinstalled`. Öffnen Sie danach ein neues PowerShell-Fenster.

Wenn PowerShell `Unexpected token` oder `Missing ...` rot beantwortet, wurde die Zeile beim Kopieren geändert (normalerweise ging ein Zeichen verloren): Kopieren Sie es erneut mit dem Kopiersymbol.

Wenn die Entwicklungslinie `CHECKSUM MISMATCH` sagt, wurde in diesem Moment ein neuer Build hochgeladen: Warten Sie fünf Minuten und führen Sie ihn noch einmal aus. Wenn es wieder passiert, oder mit der stabilen Linie, stoppen und sagen, wer sich um das Werkzeug kümmert.

## Lesen Sie dieses Handbuch im Fenster

Geben Sie `numb-form man` ein, um alles im blauen Fenster zu lesen, jeweils einen Bildschirm (Enter: next screen; q und Enter: quit). `numb-form man guide` zeigt nur diese Anleitung; fügen Sie `de` oder `es` für Deutsch oder Spanisch hinzu: `numb-form man guide de`. In Command Prompt (dem schwarzen Fenster) funktioniert `man numb-form` auch, nachdem `numb-form man -install` einmal ausgeführt wurde.

## Wenn etwas schief geht

**`numb-form` wird nicht erkannt**
: Schließen Sie das Fenster und öffnen Sie ein neues. Wenn es immer noch passiert, installieren Sie es erneut (Schritt 1).

**`... is not a working folder`** oder **`several folders could be the working folder`**
: Das blaue Fenster wurde in GWG nicht geöffnet. Wiederholen Sie Schritt 3 aus dem Ordner GWG. Sie können den Ordner auch direkt benennen, zum Beispiel `numb-form -root "H:\GWG"`.

**`template "Appendix A - Risk Assesment" ... not found`**
: Das leere Formular muss direkt in GWG sein, nicht in einem Unterordner, und sein Name muss mit `Appendix A - Risk Assesment` beginnen.

**`no folder whose name contains "Übersicht" found`**
: GWG muss den Ordner `3. Übersichten` enthalten (jeder Name, der `Übersicht` enthält, funktioniert).

**`2 spreadsheets in ...; expected exactly one`**
: `3. Übersichten` muss genau eine Excel-Datei enthalten. Verlegen Sie die anderen woanders.

**`no sheet has a "Customer Name" header cell`**
: numb-form sucht nach der Überschrift `Customer Name` und nimmt die Namen darunter. Die Nachricht listet die Überschriften auf, die sie gefunden hat. Wenn Ihre Liste eine andere Überschrift verwendet, z. B. `Firmenname`, speichern Sie sie einmal mit `numb-form config set header "Firmenname"`. Versionen 0.1.1 und älter suchen stattdessen nach `Company Name`; dort führen Sie `numb-form config set header "Customer Name"` einmal aus.

**`Time limit of 25m0s reached`**
: Ein Lauf stoppt nach 25 Minuten aus Sicherheitsgründen. Starten Sie es erneut; Unternehmen, die bereits fertig sind, werden übersprungen.

**Version 0.1.0 startet nicht innerhalb von GWG**
: Aktualisieren (siehe [Reinstall oder update](#reinstall-or-update)) oder `numb-form -root "H:\GWG"` verwenden.

**Running Skripte oder Programme werden von Ihrem Unternehmen** blockiert
: Bitten Sie Ihre IT-Abteilung, `%LOCALAPPDATA%\Programs\numb-form\numb-form.exe` zuzulassen.

Für alles andere, siehe das [reference-Handbuch](usage.md) und die [settings](configuration.md).
