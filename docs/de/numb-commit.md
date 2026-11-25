# NUMB-COMMIT(1)

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../numb-commit.md). Maßgeblich ist die englische Fassung.

## NAME

**numb-commit** — copy the new forms into the real working folder, after showing and checking exactly what changes

## SYNOPSIS

```
numb-commit [-f conf] [-from dir] [-to dir]
numb-commit [-f conf] [-from dir] [-to dir] review
numb-commit [-f conf] [-run dir] [-downloads dir]
```

## DESCRIPTION

numb-form und numb-date arbeiten auf einem Ordner auf Ihrem eigenen Computer oder auf dem echten Arbeitsordner selbst. numb-commit bringt das Ergebnis in den echten Arbeitsordner, normalerweise auf einem Netzwerklaufwerk, und sonst nichts.

Es gibt zwei Möglichkeiten zu arbeiten; numb-commit behandelt beide.

**A Arbeitskopie** (empfohlen für ein Netzwerklaufwerk). Sie kopieren den echten Ordner, zum Beispiel `F:\14. GWG`, auf Ihren Computer, zum Beispiel `C:\Users\you\Downloads\14. GWG`, und führen numb-form und numb-date dort aus. Dann starten Sie numb-commit; es fragt nach den beiden Ordnern:

```
Which directory should be committed (insert absolute path)?: C:\Users\you\Downloads\14. GWG
Which directory is the target?: F:\14. GWG
```

Pfade, die mit *Copy as path* (mit Anführungszeichen) kopiert wurden, werden akzeptiert. Um ein für alle Mal zu antworten, legen Sie sie in der Konfiguration fest:

```
numb-form config set commit_from "C:\Users\you\Downloads\14. GWG"
numb-form config set commit_to "F:\14. GWG"
```

`-from` und `-to` beantworten sie für einen Lauf.

**A review run** (`mode = review`). numb-form erstellt die Kopie selbst in `Downloads\numb-form-runs` und numb-commit kopiert die akzeptierten Ordner des neuesten Runs zurück in den echten Arbeitsordner. Es fügt nur neue Ordner hinzu, überschreibt oder löscht niemals etwas und lehnt Ordner ab, die zwischenzeitlich im echten Ordner erschienen sind. Jeder Ordner wird in einen temporären Namen kopiert, mit SHA-256 verifiziert und dann umbenannt. Dokumente, die während der Überprüfung bearbeitet wurden, werden aufgelistet, bevor Sie bestätigen. Danach bietet es an, die Überprüfungskopie zu löschen; die Protokolle werden beibehalten. numb-commit tut dies, wenn `mode = review` gesetzt ist oder Überprüfungsläufe warten, und weder `-from`, `-to`, `commit_from`, `commit_to` noch `review` nennen einen Ordner; `-run` tut es immer.

## WHAT NUMB-COMMIT DOES WITH A WORKING COPY

1. Es vergleicht die Arbeitskopie mit dem Ziel, Datei für Datei. Dateien mit der gleichen Größe und Zeit sind die gleichen; für die anderen entscheidet die SHA-256 Prüfsumme.
2. Es kopiert genau die Dateien, die es in einen **staged copy** in Ihrem Downloads-Ordner (`numb-commit staged copy <date> <time> <id>`) übertragen wird, der wie im Ziel angelegt ist, überprüft jede Prüfsumme und macht sie schreibgeschützt. Nichts berührt das Ziel noch.
3. Es schreibt **`numb-commit-request.txt`** in die Arbeitskopie und zeigt es an: die Liste der Änderungen und den Ordnerbaum des Ziels davor und danach (siehe DIE ANTRAGDATEI).
4. Es überprüft die anfrage (siehe review) und weigert sich, fortzufahren, wenn eine Überprüfung fehlschlägt. Schauen Sie sich die inszenierte Kopie und die Anfrage selbst an.
5. Es fragt `Copy this into F:\14. GWG now? [y/N]`. Nur `y` Kopien.
6. Es kopiert die inszenierten Dateien in das Ziel, überprüft jede Prüfsumme erneut und markiert die Anforderung als angewendet.
7. Es überprüft das Ziel noch einmal: Jede Datei muss mit ihrer Prüfsumme da sein.
8. Es entfernt die inszenierte Kopie, auch wenn Sie mit Nein antworten, wenn ein Check fehlschlägt, wenn Sie Ctrl+C drücken oder wenn Sie das Fenster schließen. Eine inszenierte Kopie, die von einem getöteten numb-commit hinterlassen wurde (z. B. durch einen Stromausfall), wird von dem nächsten numb-commit nach einem Tag entfernt.

Was kopiert wird:

* **Neue Ordner** (normalerweise eine pro neue Firma): wurde unter einem temporären Namen im Ziel erstellt und nach Fertigstellung umbenannt, so dass ein Ordner ganz oder gar nicht erscheint.
* **Neue Dateien** in bereits vorhandenen Ordnern.
* **Die Übersichtstabelle**, wenn numb-date Berichtsdaten hinzugefügt. Seine alte Version wird zuerst im Protokollordner aufbewahrt.

Was nie kopiert wird:

* **A-Datei, die mit anderen Inhalten im Ziel vorhanden ist**: numb-commit überschreibt niemals eine Datei außer der Tabellenkalkulation. Es listet diese als unterschiedlich auf, so dass Sie sehen, was sich in der Arbeitskopie geändert hat.
* die Vorlage (`Appendix A - Risk Assesment.docx`), auch wenn sie sich geändert hat;
* Logs, `manifest.json`, `numb-form-runs` Ordner und die Anforderungsdatei;
* Office und LibreOffice Sperrdateien (`~$…`, `.~lock.…#`), versteckte und temporäre Dateien, `Thumbs.db`, `desktop.ini`.

Nichts im Ziel wird jemals gelöscht.

## FORMS MADE BY NUMB-FORM

numb-commit sagt, welche Formen numb-form gemacht hat und seitdem niemand geändert hat. numb-form schreibt keine Markierung in die Dokumente. Es zeichnet die SHA-256-Prüfsumme jedes Formulars auf, das es in den `manifest.json` seines Laufs schreibt: in `Downloads/numb-form-runs` für einen Überprüfungslauf, im Protokollordner für einen direkten Lauf. numb-commit vergleicht die Prüfsumme jedes Formulars mit diesen Datensätzen. Ein Formblatt wird gekennzeichnet:

* `made by numb-form, unchanged`, wenn ein Lauf seine Prüfsumme aufgezeichnet hat;
* `a form edited after filling, made by hand or filled on another computer` ansonsten;
* Nichts, für andere Dateien.

Dies funktioniert für jedes Template-Format.

## THE REQUEST FILE

`numb-commit-request.txt` liegt in der Arbeitskopie, niemals im Ziel. Es sagt, in einfachen Worten, genau das, was sich ändert:

```
numb-commit request: what numb-commit copies into the real folder

From:   C:\Users\you\Downloads\14. GWG
To:     F:\14. GWG
Staged: C:\Users\you\Downloads\numb-commit staged copy 2026-10-23 14.05.12 3fa1
Made:   23/10/2026 14:05
Status: waiting

New folder  Bar Ltda
  + Bar Ltda\Appendix A - Risk Assesment - Bar Ltda.docx  (18 KB, made by numb-form, unchanged, sha256 4be1…)

Replaced  3. Übersichten\Kundenliste.xlsx
  sha256 9f02… in the real folder now
  sha256 c771… afterwards
  the replaced version is kept in the log folder

Left alone
  - Appendix A - Risk Assesment.docx  (not copied: the template)

Nothing in the real folder is deleted, and no file there is overwritten except the ones under Replaced.

Before: F:\14. GWG
├── 3. Übersichten
│   └── Kundenliste.xlsx
└── … 1204 more, unchanged

After the commit:
├── 3. Übersichten
│   └── Kundenliste.xlsx                                ~ replaced
├── Bar Ltda                                            + new folder
│   └── Appendix A - Risk Assesment - Bar Ltda.docx     + new, made by numb-form, unchanged
└── … 1204 more, unchanged
```

Die Bäume zeigen nur, was die Anforderung berührt; die letzte Zeile zählt die anderen Ordner und Dateien oben auf dem Ziel. Im Baum nach dem Commit:

| Markierung | Bedeutung |
| --- | --- |
| `+` | neu: kopiert ins Ziel |
| `~` | ersetzt: die Tabellenkalkulation mit neuen Berichtsdaten |
| `≠` | unterscheidet sich in der Arbeitskopie; die Version des Ziels bleibt |

*Status* ist `waiting`, bis die Kopie fertig ist, dann `applied`. Unterhalb des lesbaren Teils wiederholt ein Datenabschnitt die Liste für numb-commit; Bearbeiten Sie sie nicht. Eine Kopie der angewendeten Anforderung wird im Protokollordner aufbewahrt.

## REVIEW

Bei der Überprüfung wird geprüft, ob

* jede zu kopierende Datei hat die Prüfsumme in der Anforderung, in der gestaffelten Kopie und in der Arbeitskopie (seitdem wurde nichts mehr bearbeitet);
* noch keine davon im Ziel vorhanden ist;
* Die Tabelle des Ziels enthält immer noch die Prüfsumme in der Anforderung (in der Zwischenzeit hat sie niemand geändert);
* Die Tabelle erhält nur Berichtsdaten: nur leere oder ziffernfreie Zellen der Spalte * Report created* ändern sich, und jeder neue Wert sieht aus wie `DD-MM-YYYY`;
* jeder neue Ordner gehört zu einem Unternehmen in der Excel-Liste und verfügt über ein Formular zur Risikobewertung, dessen Abschnitt A dieses Unternehmen benennt und keine leere Zeile hat und das Ziel keinen ähnlichen Ordner dafür hat;
* Nichts, was zurückbleiben muss (Vorlage, Protokolle, Sperrdateien) wird aufgeführt.

Nachdem die Anforderung angewendet wurde, wird überprüft, ob sich jede aufgelistete Datei mit ihrer Prüfsumme im Ziel befindet.

Jede Zeile ist mit `ok`, `check` (siehe es vor dem Begehen) oder `FAIL`, gefolgt von dem Urteil:

| Urteil | Bedeutung |
| --- | --- |
| `OK to commit.` | Jeder Scheck bestanden. |
| `commit after checking the lines marked check.` | Nichts ist falsch mit den Dateien, aber etwas verdient einen Blick, zum Beispiel eine leere Adresse. |
| `do not commit.` | Etwas hat sich geändert oder gehört nicht dazu; numb-commit weigert sich zu kopieren. |

`numb-commit review` führt die Überprüfung der Anfrage in der Arbeitskopie jederzeit durch, ohne zu kopieren, beispielsweise auf eine Anfrage, die ein Kollege vorbereitet hat.

## OPTIONS

`-from dir`
: Die Arbeitskopie. Standard: `commit_from` in der Konfiguration; wenn diese leer ist, fragt numb-commit.

`-to dir`
: Das Ziel, der echte Arbeitsordner. Standard: `commit_to` in der Konfiguration; wenn diese leer ist, fragt numb-commit.

`review`
: Überprüfen Sie nur die Anforderung in der Arbeitskopie; kopieren Sie nichts.

`-run dir`
: Review Runs: Begehen Sie diesen Run-Ordner anstelle des neuesten nicht verpflichteten Ordners.

`-downloads dir`
: Wobei `numb-form-runs` und die inszenierte Kopie sind. Standard: `downloads` in der Konfiguration, ansonsten Ihr Downloads-Ordner.

`-f file`
: Lesen Sie *Datei * nach dem System und den Konfigurationsdateien pro Benutzer.

`-allow-non-tty`
: Lesen Sie Antworten aus einer Pipe. Nur für automatisierte Tests.

`-version`
: Drucken Sie die Version.

## FILES

`<working copy>/numb-commit-request.txt`
: Die Anfrage: Was ist (oder wurde) kopiert. UTF-8 mit einer Byte-Ordermarkierung, also zeigen Notepad und Windows PowerShell 5.1 (`Get-Content`) den Baum korrekt an.

`<Downloads>/numb-commit staged copy <date> <time> <id>/`
: Die Dateien zu kopieren, während numb-commit läuft. Am Ende entfernt.

`<log folder>/<date>_<time>_<id>_commit/numb-commit.log`
: Protokoll der Kopie, mit jeder Prüfsumme.

`<log folder>/<date>_<time>_<id>_commit/replaced/`
: Die vorherige Version des Ziels jeder ersetzten Datei.

`<log folder>/<date>_<time>_<id>_commit/numb-commit-request.txt`
: Der beantragte Antrag.

Der Protokollordner ist `log_dir`, standardmäßig `%LOCALAPPDATA%\numb-form\runs` (Windows), `~/Library/Logs/numb-form` (macOS) oder `~/.local/state/numb-form/runs` (Linux, BSD).

## EXIT STATUS

0 wenn die Kopie fertig ist, nichts zu tun ist oder Sie mit Nein geantwortet haben; 1 bei einem Fehler oder einer fehlgeschlagenen Überprüfung; 2 bei einem Verwendungsfehler; 3 wenn Sie mit Ctrl+C gestoppt werden oder das Fenster schließen.

## EXAMPLES

Einmal: Sagen Sie, wo sich beide Ordner befinden (alle Systeme):

```
numb-form config set commit_from "C:\Users\you\Downloads\14. GWG"; numb-form config set commit_to "F:\14. GWG"
```

Jedes Mal.

**Windows** (10, 11), in PowerShell:

```
cd "$env:USERPROFILE\Downloads\14. GWG"; numb-form; if ($?) { numb-date }; if ($?) { numb-commit }
```

**Unix-like** (macOS, Linux, BSD), im Terminal:

```
cd ~/Downloads/14.\ GWG && numb-form && numb-date && numb-commit
```

Lassen Sie numb-commit eine Anfrage überprüfen, z. B. eine, die ein Kollege erstellt hat, ohne zu kopieren (alle Systeme):

```
numb-commit review
```

## WHEN SOMETHING GOES WRONG

**`F:\14. GWG is not reachable (is the drive connected?)`**
: Verbinden Sie das Netzwerklaufwerk (öffnen Sie es einmal in File Explorer) und geben Sie den Pfad erneut ein.

**`… changed after the request was made`**
: Eine Datei wurde bearbeitet, nachdem sie inszeniert wurde. Führen Sie numb-commit erneut aus.

**`… in the real folder changed after the request was made`**
: Jemand anderes hat die Tabellenkalkulation auf dem Netzwerklaufwerk geändert. numb-commit überschreibt ihre Arbeit nicht. Kopieren Sie die reale Tabellenkalkulation erneut in die Arbeitskopie, führen Sie dort numb-date aus, dann numb-commit.

**`… already exists in the real folder and is never overwritten`**
: Ein Kollege hat unterdessen die gleiche Datei erstellt. Vergleichen Sie die beiden von Hand.

**`spreadsheet cell … changed; only report dates may change`**
: Die Tabelle der Arbeitskopie wurde von Hand bearbeitet. numb-commit trägt nur Berichtsdaten; andere Änderungen in der realen Tabelle vornehmen.

**`… is not a working folder: it has no template and "Übersicht" folder`**
: Geben Sie den Ordner, der die Vorlage und den Übersichtsordner enthält, nicht einen Ordner darin.

**`no request to review`**
: Führen Sie `numb-commit` zuerst aus; `numb-commit review` überprüft die Anforderung, die es schreibt.

## SEE ALSO

[numb-form](usage.md), [numb-date](numb-date.md), [configuration.md](configuration.md), [der guide](guide.md)
