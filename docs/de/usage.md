# NUMB-FORM(1)

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../usage.md). Maßgeblich ist die englische Fassung.

## NAME

**numb-form**, **numb-commit** — Formulare für die Risikobewertung ausfüllen

## SYNOPSIS

```
numb-form [-f conf] [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form [-f conf] check-template template [sample]
numb-form [-f conf] config [init]
numb-form keys
numb-commit [-f conf] [-run dir] [-downloads dir]
```

## DESCRIPTION

**numb-form** wird im Ordner *über* dem Arbeitsordner gestartet. Der Arbeitsordner ist der einzelne Ordner, der sich dort befindet. Wenn es mehrere gibt, wird derjenige verwendet, der einen Übersichtsordner enthält; andernfalls wird der Benutzer gefragt.

Der Arbeitsordner muss Folgendes enthalten:

* einen Übersichtsordner, dessen Name `overview_dir` (Standard `Übersicht`, übereinstimmend mit `3. Übersichten`) enthält, der genau eine Tabelle enthält;
* die Vorlage mit der Bezeichnung `template_name` (Standard `Appendix A - Risk Assesment`) auf ihrer obersten Ebene.

Unterstützte Formate sind in [formats.md](formats.md) aufgeführt. Die Einstellungen sind in [configuration.md](configuration.md) beschrieben.

Ein Lauf läuft wie folgt ab:

1. Das Template wird überprüft. Der Lauf stoppt, wenn Abschnitt A oder eines seiner Labels nicht gefunden werden kann.
2. Der Name der Person, die das Formular verarbeitet, wird der Konfiguration entnommen oder einmal eingegeben.
3. Mit `mode = review` wird der gesamte Arbeitsordner in `Downloads/numb-form-runs/<date>_<time>_<id>/` kopiert und alles nach diesem Schritt geschieht in der Kopie. Mit `mode = direct` (dem Standard) wird der Arbeitsordner selbst verwendet und das Protokoll geht in den Statusordner pro Benutzer.
4. Firmennamen werden aus der Spalte mit der Überschrift `Company Name` von oben nach unten gelesen. Alle anderen Spalten werden ignoriert, einschließlich `Report created`. Die Tabelle wird nie geschrieben; seine SHA-256 Prüfsumme wird vor und nach dem Lauf überprüft.
5. Für jedes Unternehmen werden die Ordner im Arbeitsordner und alle Unterordner durchsucht. Namen werden ohne Fall, Akzente und Umlaute (ü = u = ue), Interpunktion und Leerzeichen verglichen. Ein Match bedeutet, dass das Formular bereits existiert und das Unternehmen übersprungen wird. Ein Beinahe-Match wird dem Benutzer zur Bestätigung angezeigt (oder mit `confirm = none` als *needs review* gemeldet und übersprungen). Ordner mit einer anderen Rechtsform stimmen nie überein: `Foo GmbH` und `Foo S.A.` sind unterschiedliche Unternehmen.
6. Für ein neues Unternehmen wird ein nach dem Unternehmen benannter Ordner erstellt. Die Vorlage wird als `<template name> - <Company Name>.<ext>` kopiert und Abschnitt A wird ausgefüllt.
7. Das Dokument wird von der Festplatte zurückgelesen und angezeigt. Der Benutzer akzeptiert es, bearbeitet ein Feld oder lehnt es ab; ein abgelehnter Ordner wird gelöscht. Bei `confirm = none` muss das Rücklesen den Werten genau entsprechen.

Nur die Zeilen von Abschnitt A, *Information About the Contracting Party*, sind gefüllt. Abschnitt B und alles außerhalb von Abschnitt A werden Byte für Byte unverändert gelassen und es wird keine Formatierung hinzugefügt.

| Feld | Wert |
| --- | --- |
| Name / Company Name of the Contracting Party | die wörtliche Tabellenkalkulationszelle |
| Legal form | das Suffix des Namens, wie geschrieben (`GmbH`, `S.A.`, `C.A. de C.V.`) |
| Headquarters / Address | ausgewählt aus den gefundenen Kandidaten |
| Commercial Registry Number | immer leer |
| Authorized Representatives | `Full Name (Role 1, Role 2), Full Name (Role)`, Rollen in Englisch |
| Date of the evaluation | heute `date_format` (Standard `DD/MM/YYYY`) |
| Person Processing the Form | aus der Konfiguration oder zu Beginn eingegeben |

Die Kandidaten werden gesammelt aus, in der Reihenfolge der Autorität:

1. UK Companies House (Regierungsregister; nur mit einem freien Schlüssel);
2. das GLEIF LEI-Register (Regulierungsregister; kein Schlüssel);
3. den eigenen rechtlichen Hinweis des Unternehmens (Impressum, Aviso legal, Rechtlicher Hinweis);
4. Wikidata (Community-editiert; hauptsächlich verwendet, um die Website zu finden).

Welche Quellen welches Feld speisen, die am wenigsten vertrauenswürdige Quelle, die für Vertreter akzeptiert wird, und die Fallbacks werden in der Konfiguration festgelegt. Personen und Adressen werden niemals von einem Elternteil, einer Tochtergesellschaft oder einer ähnlich benannten Einheit übernommen. Nicht verifizierte Werte erhalten genau diesen Text, wenn `caution` dies sagt:

> Vorsicht: Dieses Feld kann von internen Quellen abweichen oder aufgrund unzureichender Online-Informationen unpräzise sein

Wenn kein Vertreter gefunden wird, lautet das Feld:

> Es wurden keine Vertreter online gefunden, konsultieren Sie bitte andere interne Quellen für genauere Informationen

**numb-commit** gilt nur für Überprüfungsläufe. Es kopiert die akzeptierten Ordner eines Runs zurück in den echten Arbeitsordner. Es fügt nur neue Ordner hinzu, überschreibt oder löscht niemals etwas und lehnt Ordner ab, die in der Zwischenzeit im echten Baum erschienen sind. Jeder Ordner wird in einen temporären Namen kopiert, mit SHA-256 verifiziert und dann umbenannt. Dokumente, die während der Überprüfung bearbeitet werden, werden aufgelistet, bevor der Benutzer bestätigt. Danach bietet es an, die Überprüfungskopie zu löschen; die Protokolle werden beibehalten.

## COMMANDS

`check-template template [sample]`
: Überprüfen Sie, ob eine Vorlage gefüllt werden kann, und schreiben Sie optional ein ausgefülltes Sample in *sample*. Die Vorlage wird nur gelesen.

`config`
: Zeigen Sie die Konfigurationsdateien und effektiven Einstellungen an.

`config init`
: Schreiben Sie eine kommentierte Konfigurationsdatei pro Benutzer mit jedem Standard.

`keys`
: Erklären Sie die optionalen kostenlosen API-Schlüssel.

## OPTIONS

`-f file`
: Lesen Sie *Datei * nach dem System und den Konfigurationsdateien pro Benutzer.

`-offline`
: Kontaktieren Sie keine Online-Quelle.

`-root dir`
: Verwenden Sie *dir* als Arbeitsordner, anstatt es zu erkennen.

`-downloads dir`
: Review-Modus: Setzen Sie die Kopie in * dir *.

`-timeout duration`
: Stoppen Sie sicher nach *Dauer *; höchstens `25m`.

`-run dir`
: (numb-commit) Legen Sie diesen Run-Ordner anstelle des neuesten, nicht gebundenen Ordners fest.

`-allow-non-tty`
: Lesen Sie Antworten aus einer Pipe. Nur für automatisierte Tests.

`-version`
: Drucken Sie die Version.

## ENVIRONMENT

`NUMBFORM_DOWNLOADS`
: Downloads Ordner zu verwenden, zum Beispiel ein montiertes Volumen in einem Container.

`NUMBFORM_COMPANIES_HOUSE_KEY`
: Optional kostenlos UK Companies House API-Schlüssel. Nie protokolliert oder gespeichert.

`HTTPS_PROXY`, `NO_PROXY`
: Proxy-Einstellungen für Online-Lookups.

`TZ`
: Zeitzone des Bewertungsdatums.

## FILES

`<log folder>/<run>/numb-form.log`
: Log von allem, was der lauf gemacht hat.

`<log folder>/<run>/manifest.json`
: Maschinenlesbare Aufzeichnung des Laufs.

`Downloads/numb-form-runs/<run>/<working folder>/`
: Review-Modus: die Kopie.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Review-Modus: Protokoll des Commits.

Kein Cache oder temporäre Dateien bleiben zurück. Netzwerkantworten werden nur im Speicher gespeichert, Dokumente werden durch eine temporäre Datei geschrieben, die umbenannt oder entfernt wird, und Konverter arbeiten in einem privaten temporären Ordner, der gelöscht wird.

## EXIT STATUS

0 zum Erfolg, 1 zum Fehler, 2 zu einem Nutzungs- oder Konfigurationsfehler und 3 beim Stoppen durch das Zeitlimit, durch Strg + C oder durch den Benutzer.

## EXAMPLES

Überprüfen Sie die echte Vorlage und schreiben Sie ein Beispiel, ohne die Vorlage zu berühren:

```
numb-form check-template "Appendix A - Risk Assesment.docx" sample.docx
```

Füllen Sie die Formulare direkt im Arbeitsordner aus:

```
cd /path/to/folder-above && numb-form
```

Überprüfen Sie zuerst, dann Commit (numb-commit wartet auf Bestätigung):

```
cd /path/to/folder-above && numb-form -f review.conf && numb-commit
```

wobei `review.conf` `mode = review` enthält.

Laufen ohne Netzwerkzugang:

```
cd /path/to/folder-above && numb-form -offline
```

## SEE ALSO

[configuration.md](configuration.md), [formats.md](formats.md), [install.md](install.md), [containers.md](containers.md), [../SECURITY.md](SECURITY.md), [../AGENTS.md](../../AGENTS.md)
