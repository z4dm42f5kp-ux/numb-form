# NUMB-FORM(1)

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../usage.md). Maßgeblich ist die englische Fassung.

## NAME

**numb-form**, **numb-commit**, **numb-date** — Formulare für die Risikobewertung ausfüllen

## SYNOPSIS

```
numb-form [-f conf] [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form [-f conf] check-template template [sample]
numb-form [-f conf] config [init | set key value]
numb-form [-f conf] update [stable | rolling | v1.2.3]
numb-form rollback
numb-commit [-f conf] [-run dir] [-downloads dir]
numb-date [-f conf] [-root dir] [-timeout duration]
```

## DESCRIPTION

**numb-form** wird innerhalb des Arbeitsordners, des Ordners, der die Vorlage enthält, und des Übersichtsordners (z. B. `GWG`) gestartet. Es kann auch eine Ebene darüber gestartet werden: Der Arbeitsordner ist dann der einzelne Ordner, der sich dort befindet, oder derjenige, der einen Übersichtsordner enthält, oder der Benutzer wird gefragt. `-root` benennt den Arbeitsordner explizit.

Büropersonal: siehe [step-by-step guide](guide.md).

Der Arbeitsordner muss Folgendes enthalten:

* einen Übersichtsordner, dessen Name `overview_dir` (Standard `Übersicht`, übereinstimmend mit `3. Übersichten`) enthält, der genau eine Tabelle enthält;
* die Vorlage mit der Bezeichnung `template_name` (Standard `Appendix A - Risk Assesment`) auf ihrer obersten Ebene.

Unterstützte Formate sind in [formats.md](formats.md) aufgeführt. Die Einstellungen sind in [configuration.md](configuration.md) beschrieben.

Ein Lauf läuft wie folgt ab:

1. Die Vorlage wird mit einer Probefüllung jedes Feldes überprüft. Der Lauf stoppt, wenn Abschnitt A oder ein anderes Label als *Commercial Registry Number* nicht gefunden werden kann oder nicht leer ist.
2. Der Name der Person, die das Formular verarbeitet, wird der Konfiguration entnommen oder einmal eingegeben.
3. Mit `mode = review` wird der gesamte Arbeitsordner in `Downloads/numb-form-runs/<date>_<time>_<id>/` kopiert und alles nach diesem Schritt geschieht in der Kopie. Mit `mode = direct` (dem Standard) wird der Arbeitsordner selbst verwendet und das Protokoll geht in den Statusordner pro Benutzer.
4. Firmennamen werden aus der Spalte mit der Überschrift `Customer Name` von oben nach unten gelesen. Alle anderen Spalten werden ignoriert, einschließlich `Report created`. numb-form schreibt niemals die Tabelle; seine SHA-256 Prüfsumme wird vor und nach dem Lauf überprüft.
5. Für jedes Unternehmen werden die Ordner im Arbeitsordner und alle Unterordner durchsucht. Namen werden ohne Fall, Akzente und Umlaute (ü = u = ue), Interpunktion und Leerzeichen verglichen. Ein Match bedeutet, dass das Formular bereits existiert und das Unternehmen übersprungen wird. Ein Near Match wird dem Benutzer zur Bestätigung angezeigt. Ordner mit einer anderen Rechtsform stimmen nie überein: `Foo GmbH` und `Foo S.A.` sind unterschiedliche Unternehmen.
6. Bei einem neuen Unternehmen gibt der Benutzer die Adresse des Hauptsitzes und die autorisierten Vertreter ein. Alles andere wird ohne Fragen gefüllt.
7. Ein nach dem Unternehmen benannter Ordner wird erstellt. Die Vorlage wird als `<template name> - <Company Name>.<ext>` kopiert und Abschnitt A wird ausgefüllt.
8. Das Dokument wird von der Festplatte zurückgelesen und angezeigt. Der Benutzer akzeptiert es, bearbeitet ein Feld oder lehnt es ab; ein abgelehnter Ordner wird gelöscht.

Nur die Zeilen von Abschnitt A, *Information About the Contracting Party*, sind gefüllt. Abschnitt B und alles außerhalb von Abschnitt A werden Byte für Byte unverändert gelassen und es wird keine Formatierung hinzugefügt.

| Feld | Wert |
| --- | --- |
| Name / Company Name of the Contracting Party | die wörtliche Tabellenkalkulationszelle |
| Legal form | gefunden im Namen: am Ende, vor einer Zweig- oder Ländernote (`Foo GmbH (Wien)`), vor einer Stadt oder einem Land (`Foo SAS, Bogotá, Colombia`, `Foo S.A.C.I. Paraguay`) oder zuerst (`Ets Foo`); Akzente, Punkte und Leerzeichen können sich unterscheiden (`SOCIEDAD ANONIMA`, `S A`); geschrieben wie im Namen mit herkömmlichen Großbuchstaben (`GMBH` gibt `GmbH`); nur gefragt, wenn der Name keine hat |
| Headquarters / Address | vom Benutzer eingegeben; Enter lässt es leer |
| Commercial Registry Number | immer `No commercial registry number was found online consult internal sources for further information`; es kann durch die Nummer mit *edit* ersetzt werden |
| Authorized Representatives | vom Benutzer als `Full Name (Role 1, Role 2), Full Name (Role)` eingegeben, Rollen in Englisch; Enter schreibt den Text ohne Vertreter unten, `-` lässt ihn leer |
| Date of the evaluation | heute `date_format` (Standard `DD/MM/YYYY`) |
| Person Processing the Form | aus der Konfiguration oder zu Beginn eingegeben |

numb-form macht keine Online-Lookups für Unternehmen. Nur `numb-form update` kontaktiert GitHub.

Wenn die Vertreter leer bleiben, lautet das Feld:

> Es wurden keine Vertreter online gefunden, konsultieren Sie bitte andere interne Quellen für genauere Informationen

**numb-commit** applies to review runs only. It copies the accepted folders of a run back into the real working folder. It only adds new folders, never overwrites or deletes anything, and refuses folders that appeared in the real tree in the meantime. Each folder is copied to a temporary name, verified with SHA-256 and then renamed. Documents edited during review are listed before the user confirms. Afterwards it offers to delete the review copy; the logs are kept.

**numb-date** füllt die Berichtsdatumsspalte der Tabelle. Es wird wie numb-form in oder über dem Arbeitsordner gestartet und findet die gleiche Übersichtstabelle, die `.xlsx` oder `.xlsm` sein muss.

1. Die Unternehmensspalte wird von `header`, dann von gemeinsamen Namen (`Customer Name`, `Customer`, `Company Name`, `Kunde`, `Cliente` und andere, ignorieren Fall) gefunden; die Datumsspalte in der gleichen Kopfzeile von `Report created AB`, `Report created`, `Report date`, `Created` und ähnliche Namen, oder jede Überschrift, die einen Bericht und seine Erstellung oder Datum erwähnt.
2. Zeilen, deren Datumszelle eine Ziffer enthält, werden übersprungen: Sie haben bereits ein Datum.
3. Für die anderen Zeilen wird der Ordner des Unternehmens als numb-form gefunden (beinahe Übereinstimmungen werden gefragt). Die Formulare `.docx`, `.odt`, `.fodt` oder `.rtf` im Ordner (und eine Ebene unten) werden gelesen, das Risikobeurteilungsformular zuerst: das *Date of the evaluation* von Abschnitt A oder, in anders angelegten Formularen, das Datum, das nach einem Datumsetikett geschrieben wurde. Daten wie `16/10/2026`, `16.10.2026`, `2026-10-16` und `16 October 2026` werden verstanden. Wenn kein Datum gefunden wird, gibt der Benutzer es ein oder überspringt das Unternehmen.
4. Die Daten werden aufgelistet, als `DD-MM-YYYY` geschrieben, nachdem der Benutzer zugestimmt hat, von der Festplatte zurückgelesen und angezeigt; Der Benutzer akzeptiert sie oder macht das Schreiben rückgängig.

Nur die aufgeführten Zellen ändern sich; siehe [../SECURITY.md](SECURITY.md) für die Prüfungen. Formulare werden nur gelesen. Die Datei darf nicht in Excel oder LibreOffice geöffnet sein. Das Protokoll ist `numb-date.log` im Statusordner pro Benutzer (oder `log_dir`).

## COMMANDS

`check-template template [sample]`
: Überprüfen Sie, ob eine Vorlage gefüllt werden kann, und schreiben Sie optional ein ausgefülltes Sample in *sample*. Die Vorlage wird nur gelesen.

`config`
: Zeigen Sie die Konfigurationsdateien und effektiven Einstellungen an.

`config init`
: Schreiben Sie eine kommentierte Konfigurationsdatei pro Benutzer mit jedem Standard.

`config set key value`
: Speichern Sie eine Einstellung in der Konfigurationsdatei pro Benutzer, z. B. `numb-form config set person "Erika Mustermann"`. Der Wert wird überprüft, bevor er gespeichert wird.

`rollback`
: Wechseln Sie zurück zu der Version, die das letzte `update` ersetzt hat. Die zurückgerollte Version wird an ihrem Platz gehalten, so dass ein zweiter `rollback` zu ihr zurückkehrt. Ohne eine gespeicherte Version installieren Sie stattdessen ein Release nach Nummer: `numb-form update v0.1.1`.

`man` [*topic*] [`en` | `es` | `de`]
: Zeigen Sie dieses Handbuch im Terminal, jeweils einen Bildschirm (Enter: next screen; q und Enter: quit). Ohne ein Thema wird das gesamte Handbuch angezeigt, beginnend mit dem Leitfaden. Themen: `guide`, `usage`, `configuration`, `install`, `formats`, `containers`, `security`; `numb-form man topics` listet sie auf. Die Sprache folgt `LANG` oder wird angegeben. `numb-form man -install` lässt `man numb-form` funktionieren: Unter Linux, macOS und den BSDs schreibt es die Handbuchseite zu `~/.local/share/man/man1` (erneuert von `numb-form update`); Unter Windows schreibt es `man.cmd` neben `numb-form.exe` für Command Prompt. Windows PowerShell behält `man` für seine eigene Hilfe: dort, Typ `numb-form man`.

`update` [`stable` | `rolling` | *v1.2.3*]
: Ersetzen Sie das installierte `numb-form` und `numb-commit`: **stable** durch das neueste stabile Release, **rolling** durch den aktuellen Rolling Build von `main` (nach jeder Änderung neu erstellt; `edge` wird als älterer Name akzeptiert), oder *v1.2.3* mit genau diesem Release. Die ersetzte Version wird neben den Programmen als `numb-form.previous` und `numb-commit.previous` für `rollback` beibehalten. Es funktioniert wie die Installationsbefehle: Das Archiv und `SHA256SUMS` werden nur von GitHub über HTTPS heruntergeladen, die Prüfsumme muss übereinstimmen und das neue `numb-form` muss starten, bevor etwas ersetzt wird. Ohne Argument wird der Kanal des laufenden Builds verwendet. Einstellungen werden beibehalten. Verweigert mit `-offline`, mit `online = no` und für Builds von der Quelle. Unter Windows wird das alte Programm in `numb-form.old.exe` umbenannt und beim nächsten Start entfernt.

## OPTIONS

`-f file`
: Lesen Sie *Datei * nach dem System und den Konfigurationsdateien pro Benutzer.

`-offline`
: Mit `update`: weigern Sie sich, GitHub zu kontaktieren.

`-root dir`
: Verwenden Sie *dir* als Arbeitsordner, anstatt es zu erkennen.

`-downloads dir`
: Review-Modus: Setzen Sie die Kopie in * dir *.

`-timeout duration`
: Stoppen Sie sicher nach *Dauer*, z. B. `30m`; überschreibt ein längeres `timeout` in der Konfiguration. Ohne sie gibt es keine zeitliche Begrenzung.

`-run dir`
: (numb-commit) Legen Sie diesen Run-Ordner anstelle des neuesten, nicht gebundenen Ordners fest.

`-allow-non-tty`
: Lesen Sie Antworten aus einer Pipe. Nur für automatisierte Tests.

`-version`
: Drucken Sie die Version.

## ENVIRONMENT

`NUMBFORM_DOWNLOADS`
: Downloads Ordner zu verwenden, zum Beispiel ein montiertes Volumen in einem Container.

`HTTPS_PROXY`, `NO_PROXY`
: Proxy-Einstellungen für `numb-form update`.

`TZ`
: Zeitzone des Bewertungsdatums.

## FILES

`<log folder>/<run>/numb-form.log`
: Log von allem, was der lauf gemacht hat.

`<log folder>/<run>/manifest.json`
: Maschinenlesbare Aufzeichnung des Laufs.

`<log folder>/<run>_date/numb-date.log`
: Protokoll eines numb-date-laufs, mit dem sha-256 der tabelle davor und danach.

`Downloads/numb-form-runs/<run>/<working folder>/`
: Review-Modus: die Kopie.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Review-Modus: Protokoll des Commits.

Kein Cache oder temporäre Dateien bleiben zurück. Dokumente werden durch eine temporäre Datei geschrieben, die umbenannt oder entfernt wird, und Konverter arbeiten in einem privaten temporären Ordner, der gelöscht wird.

## EXIT STATUS

0 auf Erfolg, 1 auf Fehler, 2 auf einen Nutzungs- oder Konfigurationsfehler und 3 wenn durch das Zeitlimit, durch Ctrl+C oder durch den Benutzer gestoppt.

## EXAMPLES

Überprüfen Sie die echte Vorlage und schreiben Sie ein Beispiel, ohne die Vorlage zu berühren:

```
numb-form check-template "Appendix A - Risk Assesment.docx" sample.docx
```

Füllen Sie die Formulare direkt im Arbeitsordner aus:

```
cd /path/to/GWG && numb-form
```

Überprüfen Sie zuerst, dann Commit (numb-commit wartet auf Bestätigung):

```
cd /path/to/GWG && numb-form -f review.conf && numb-commit
```

wobei `review.conf` `mode = review` enthält.

## SEE ALSO

[configuration.md](configuration.md), [formats.md](formats.md), [install.md](install.md), [containers.md](containers.md), [../SECURITY.md](SECURITY.md), [../AGENTS.md](../../AGENTS.md)
