# Office-Suiten und Dateiformate

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../formats.md). Maßgeblich ist die englische Fassung.

numb-form funktioniert mit jeder Office-Suite gleich. Es bearbeitet die gängigen offenen Formate selbst und verwendet Office-Software bereits auf dem Computer nur für Formate, die es nicht sicher bearbeiten kann.

## Meldebögen

| Format | Geschrieben von | Wie numb-form es füllt | Gefüllte Datei |
| --- | --- | --- | --- |
| `.docx` | Microsoft Word, LibreOffice, Apache OpenOffice, Proton Docs, Apple Pages (Export) | direkt | `.docx` |
| `.odt` | LibreOffice, Apache OpenOffice, Collabora, Microsoft Word | direkt | `.odt` |
| `.fodt` | LibreOffice (flach XML) | direkt | `.fodt` |
| `.rtf` | Jeder Textverarbeiter | direkt | `.rtf` |
| `.pages` | Apple Pages | umgewandelt, dann gefüllt | `.docx` (siehe unten) |
| `.doc` | Wort 97-2003 | Legacy: umgewandelt, dann gefüllt | `.docx` oder `.doc` mit `legacy_output = same` |

"Direkt" bedeutet, dass die Datei nie wieder von einem anderen Programm gespeichert wird: Werte werden in den eigenen Text des Labels eingefügt, und alles andere bleibt Byte für Byte, wie es war. Das Ergebnis wird nach jeder Füllung zurückgelesen und überprüft.

Wenn der Arbeitsordner die Vorlage in mehreren Formaten hält, gewinnt das erste in der Tabelle.

## Tabellenkalkulation

| Format | Geschrieben von | Wie numb-form es liest |
| --- | --- | --- |
| `.xlsx`, `.xlsm` | Microsoft Excel, LibreOffice Calc, Proton Sheets, Apple Numbers (Export) | direkt; Makros werden nie ausgeführt |
| `.ods`, `.fods` | LibreOffice Calc, Apache OpenOffice | direkt |
| `.csv`, `.tsv` | Jedes Tabellenkalkulationsprogramm | direkt; UTF-8 oder Windows-1252, `;` `,` oder tab |
| `.numbers` | Apple Numbers | konvertiert im Speicher |
| `.xls` | Excel 97-2003 | Legacy: direkt, mit dem eigenen Leser von numb-form |

Die Tabelle wird nur gelesen, nie geschrieben.

## Konvertierte Formate

`.doc`, `.pages` und `.numbers` werden bereits mit Software auf dem Computer konvertiert:

* LibreOffice oder Apache OpenOffice, auf jedem System. numb-form sucht die üblichen Installationsorte; Setzen Sie `soffice = <path>`, wenn es sich an anderer Stelle befindet.
* Unter macOS, Apple Pages und Numbers, wenn installiert (`pages_converter` und `numbers_converter`, Standard `auto`).

Die Konvertierung läuft in einem privaten temporären Ordner mit einem Wegwerf-Office-Profil und alles wird danach gelöscht. Nichts wird heruntergeladen.

### Apple Pages

Pages Dateien verwenden ein undokumentiertes Format. numb-form liest daher eine `.pages`-Vorlage, indem sie sie in das Word-Format konvertiert, füllt diese aus und speichert die ausgefüllten Formulare als `.docx`, die Pages öffnet. Überprüfen Sie das Layout des ersten ausgefüllten Formulars.

Auf macOS mit installiertem Pages speichert `pages_write = yes` die ausgefüllten Formulare als `.pages` durch Pages selbst. Dies ist **experimental**: Es wird getestet, aber standardmäßig deaktiviert, und jede Datei wird nach dem Speichern zurückkonvertiert und überprüft. Wenn Sie keine `.pages`-Dateien benötigen, lassen Sie sie aus.

### Legacy-Formate (.doc, .xls)

Legacy-Binärformate sind standardmäßig deaktiviert. Ein Administrator aktiviert sie in der Systemkonfigurationsdatei (siehe [configuration.md](configuration.md)):

```
legacy_formats = yes
```

Runs, die ein Legacy-Format verwenden, sind immer vollständig offline. Der `.xls`-Reader ist numb-form-eigen und wird kontinuierlich fuzz-getestet.

## Bürosuiten

* **Microsoft 365 / Office**: Speichern Sie die Vorlage als `.docx` und die Übersicht als `.xlsx` (die Standardwerte).
* **LibreOffice** und **Apache OpenOffice**: `.odt` und `.ods` arbeiten direkt; so auch `.docx` und `.xlsx`.
* **Proton Drive (Docs und Sheets)**: Laden Sie die Vorlage als `.docx` und die Übersicht als `.xlsx` in den Arbeitsordner herunter. Dies funktioniert auch auf Systemen ohne Desktop-Office-Suite, einschließlich der BSDs, aus einem Webbrowser.
* **Apple Pages und Numbers**: Verwenden Sie `.pages` und `.numbers` wie oben beschrieben, oder exportieren Sie nach `.docx` und `.xlsx` (File → Export To).
* **Die BSDs**: LibreOffice ist als Paket verfügbar (`pkg install libreoffice`, `pkg_add libreoffice`, `pkgin install libreoffice`); Proton Docs in einem Browser benötigt keine Installation.
