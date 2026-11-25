# NUMB-TYPO(1)

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../numb-typo.md). Maßgeblich ist die englische Fassung.

## NAME

**numb-typo** — korrekte falsch geschriebene oder abgeschnittene Städte und Länder in den Firmennamen der Übersichtstabelle

## SYNOPSIS

```
numb-typo [-f conf] [-root dir] [-timeout duration]
```

## DESCRIPTION

Firmennamen in der Excel-Liste enden oft in einer Stadt oder einem Land, das falsch geschrieben oder abgeschnitten ist, z. B. `Foo S.A. de C.V., Ciudad de Me` oder `Foo SAS, Bogta, Colomiba`. numb-typo findet sie und schreibt, nachdem Sie zugestimmt haben, die korrigierten Namen: `Foo S.A. de C.V., Ciudad de México`, `Foo SAS, Bogotá, Colombia`.

Führen Sie es wie numb-form vor numb-form im Arbeitsordner aus, damit die neuen Ordner und Formulare die korrigierten Namen erhalten. Die Tabelle muss `.xlsx` oder `.xlsm` sein und in Excel und LibreOffice geschlossen sein.

## WHAT IS CORRECTED

Nur die Teile nach einem Komma (und nach ` - `), die wie eine Stadt oder ein Land aussehen:

* **Cut off**: der letzte Teil einer Zelle, wenn es der Anfang eines bekannten Ortes ist: `Ciudad de Me` wird `Ciudad de México`, `Nairob` wird `Nairobi`.
* **Typo**: ein Teil, der ein oder zwei Buchstaben von genau einem bekannten Ort entfernt ist: `Colomiba` wird zu `Colombia`, `Medelin` wird zu `Medellín`, `Duabi` wird zu `Dubái`. Teile von weniger als fünf Buchstaben werden nie korrigiert; ein zusätzlicher Buchstabe wird nur von sieben Buchstaben korrigiert, zwei Fehler nur von neun, und der erste Buchstabe muss übereinstimmen.
* Wenn mehrere Orte passen (`Santa` könnte Santa Cruz, Santa Marta, ... sein), fragt numb-typo, welchen, oder Sie lassen es.

Die Großbuchstaben werden wie geschrieben aufbewahrt: `GUAYAQIL` wird zu `GUAYAQUIL`.

Was sich nie geändert hat:

* Name des Unternehmens selbst vor dem ersten Komma;
* Rechtsformen (`S.A.`, `Ltda`, `LLC`), Zahlen und Wörter, die keine Orte sind;
* eine Stelle, die bereits richtig ist; fehlende Akzente (`Bogota`) bleiben wie geschrieben;
* die Namen von Unternehmen, die bereits einen Ordner haben: Umbenennung würde sie von ihrem Ordner und Formular trennen, so dass sie nur im Protokoll aufgeführt sind.

Die Liste der Orte umfasst die Länder in Spanisch, Englisch, Portugiesisch, Französisch und Deutsch sowie die wichtigsten Städte Lateinamerikas, Afrikas, des Golfs, Europas und Nordamerikas.

## WHAT HAPPENS

1. numb-typo geht durch die Firmennamen und zeigt jede Korrektur:

   ```
   Foo S.A. de C.V., Ciudad de Me  (Customer ID D1001, row 315)
     → Foo S.A. de C.V., Ciudad de México
   ```

2. Es listet sie alle auf und fragt `Write these 3 correction(s)? [a]ll, [o]ne by one, [n]one:`. Mit `o` fragt es nach jedem.
3. Es schreibt sie, liest die Tabellenkalkulation von der Festplatte zurück, zeigt die Zellen und fragt `Is this correct? [a]ccept, [u]ndo:`.

Nur die Zellen der Spalte Firmenname ändern sich, mit den gleichen Prüfungen wie numb-date: jede andere Zelle jedes Blattes wird vor dem Schreiben verglichen, jede Zelle behält ihren Stil, SHA-256 schützt jeden Schritt und *undo* stellt das genaue Original wieder her, das im Protokollordner aufbewahrt wird, bis Sie akzeptieren.

## OPTIONS

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

`<log folder>/<date>_<time>_<id>_typo/numb-typo.log`
: Protokoll des Laufs: jede Korrektur, jeder Name allein gelassen und warum, und die SHA-256 der Tabelle vorher und nachher.

## EXIT STATUS

0 bei Erfolg oder wenn nichts zu korrigieren ist, 1 bei einem Fehler, 2 bei einem Nutzungsfehler, 3 bei Beendigung mit Ctrl+C, durch Schließen des Fensters oder durch den Benutzer.

## EXAMPLES

Korrigieren Sie die Namen und füllen Sie dann die Formulare und die Daten aus.

**Windows** (10, 11), in PowerShell, innerhalb des Arbeitsordners:

```
numb-typo; if ($?) { numb-form }; if ($?) { numb-date }
```

**Unix-like** (macOS, Linux, BSD), im Terminal:

```
numb-typo && numb-form && numb-date
```

## SEE ALSO

[numb-form](usage.md), [numb-date](numb-date.md), [numb-commit](numb-commit.md)
