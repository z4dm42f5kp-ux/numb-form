# NUMB-DATE(1)

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../numb-date.md). Maßgeblich ist die englische Fassung.

## NAME

**numb-date** — füllen Sie die *Report created* Spalte der Übersichtstabelle aus den Firmenformularen

## SYNOPSIS

```
numb-date [-f conf] [-y] [-root dir] [-timeout duration]
```

## DESCRIPTION

numb-date is started like numb-form, inside the working folder (or one level above it), usually right after numb-form. It is the only program that writes to the overview spreadsheet, and it writes nothing but report dates. The spreadsheet must be `.xlsx` or `.xlsm`, and closed in Excel and LibreOffice.

1. Die Unternehmensspalte wird von `header`, dann von gemeinsamen Namen (`Customer Name`, `Customer`, `Company Name`, `Kunde`, `Cliente` und andere, ignorieren Fall) gefunden; die Datumsspalte in der gleichen Kopfzeile von `Report created AB`, `Report created`, `Report date`, `Created` und ähnliche Namen, oder jede Überschrift, die einen Bericht und seine Erstellung oder Datum erwähnt.
2. Zeilen, deren Datumszelle eine Ziffer enthält, werden übersprungen: Sie haben bereits ein Datum.
3. Für die anderen Zeilen wird der Ordner des Unternehmens als numb-form gefunden (beinahe Übereinstimmungen werden gefragt oder mit `-y` aufgenommen). Die Formulare `.docx`, `.odt`, `.fodt` oder `.rtf` im Ordner (und eine Ebene unten) werden gelesen, das Risikobeurteilungsformular zuerst: das *Date of the evaluation* von Abschnitt A oder, in anders angelegten Formularen, das Datum, das nach einem Datumsetikett geschrieben wurde. Daten wie `16/10/2026`, `16.10.2026`, `2026-10-16` und `16 October 2026` werden verstanden. numb-date fragt nie nach einem datum. Ein Unternehmen ohne Ordner, ohne Formular oder ohne Datum in seiner Form wird mit dem Grund, z. B. `skipped: no form in the folder Foo GmbH`, übersprungen und vor den Daten erneut aufgelistet, so dass Sie es von Hand ausfüllen können.
4. Die Daten sind aufgelistet, nummeriert, mit dem Formular, aus dem jeder stammt. Sie akzeptieren sie, bearbeiten eine oder schreiben nichts. Sie werden als `DD-MM-YYYY` geschrieben, von der Festplatte zurückgelesen und angezeigt; Sie akzeptieren sie oder machen das Schreiben rückgängig.

Formulare werden nur gelesen.

## WHAT CHANGES IN THE SPREADSHEET

Nur die aufgeführten Zellen ändern sich. Sie behalten ihren eigenen Zellstil bei und erhalten keine neue Formatierung; jeder andere Teil der Arbeitsmappe wird Byte für Byte kopiert, und Zellen mit Formeln werden in Ruhe gelassen. Vor dem Schreiben wird jede Zelle jedes Blatts der neuen Arbeitsmappe mit dem Original verglichen: Alles andere, was sich unterscheidet, stoppt das Betäubungsdatum. SHA-256 schützt jeden Schritt: Die Datei muss immer noch die Prüfsumme haben, die sie beim Lesen hatte, die temporäre Kopie daneben muss übereinstimmen, bevor sie die Datei ersetzt, und die Datei muss danach übereinstimmen. Das Original wird im Protokollordner aufbewahrt, bis Sie es akzeptieren; *undo* legt es genau zurück.

| numb-date fragt | Was du tust |
| --- | --- |
| `Is the folder "…" this company's folder? (...) [y/N]` | `y`, wenn es das gleiche Unternehmen ist, andernfalls Enter |
| `Write these … date(s)? Nothing else in the spreadsheet changes. [a]ccept, [e]dit a date, [r]eject:` | `a` schreibt sie; `e` ändert ein Datum; `r` schreibt nichts |
| `Number to change (1-…, Enter to go back):` | Nummer des Datums in der Liste |
| `New date for … (DD-MM-YYYY, Enter = keep …, '-' = leave it out):` | Geben Sie das Datum ein, geben Sie ein, um es zu behalten, oder `-`, um dieses Unternehmen zu verlassen |
| `Is this correct? [a]ccept, [u]ndo:` | `a` behält sie, `u` stellt die Datei genau wieder her |

## OPTIONS

`-y`
: Nehmen Sie einen Ordner mit einem ähnlichen Namen wie der Ordner des Unternehmens, ohne zu fragen.

`-root dir`
: Verwenden Sie *dir* als Arbeitsordner, anstatt es zu erkennen.

`-timeout duration`
: Stoppen Sie sicher nach *Dauer *, zum Beispiel `30m`.

`-f file`
: Lesen Sie *Datei * nach dem System und den Konfigurationsdateien pro Benutzer.

`-allow-non-tty`
: Lesen Sie Antworten aus einer Pipe. Nur für automatisierte Tests.

`-version`
: Drucken Sie die Version.

## FILES

`<log folder>/<date>_<time>_<id>_date/numb-date.log`
: Log des Laufs, mit dem SHA-256 der Tabelle vor und nach.

`<log folder>/<date>_<time>_<id>_date/before-<spreadsheet>`
: Die ursprüngliche Tabellenkalkulation, gehalten, bis Sie akzeptieren oder rückgängig machen.

## EXIT STATUS

0 zum Erfolg oder wenn es nichts zu schreiben gibt, 1 zu einem Fehler, 2 zu einem Nutzungsfehler, 3 wenn mit Ctrl+C oder durch den Benutzer gestoppt.

## EXAMPLES

**Windows** (10, 11), in PowerShell, innerhalb des Arbeitsordners:

```
numb-form; if ($?) { numb-date }
```

**Unix-like** (macOS, Linux, BSD), im Terminal:

```
numb-form && numb-date
```

## WHEN SOMETHING GOES WRONG

**`… is open in Excel or LibreOffice`**
: Schließen Sie die Datei und führen Sie numb-date erneut aus.

**`numb-date writes only .xlsx and .xlsm spreadsheets`**
: Speichern Sie die Übersicht einmal als `.xlsx`.

**`no "Report created" column next to the company names`**
: Die Nachricht listet die gefundenen Überschriften auf. Benennen Sie die Überschrift der Datumsspalte in `Report created` oder ähnlich.

## SEE ALSO

[numb-form](usage.md), [numb-commit](numb-commit.md), [numb-form.conf](configuration.md), [../SECURITY.md](SECURITY.md)
