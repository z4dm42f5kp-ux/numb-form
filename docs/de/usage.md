# NUMB-FORM(1)

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../usage.md). Maßgeblich ist die englische Fassung.

## NAME

**numb-form**, **numb-commit** — füllen Sie die Risikobeurteilungsformulare aus und verpflichten Sie sie

## SYNOPSIS

```
numb-form [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form check-template template.docx [sample.docx]
numb-form keys
numb-commit [-run dir] [-downloads dir]
```

## DESCRIPTION

**numb-form** wird im Ordner *über* dem Arbeitsordner gestartet.
Der Arbeitsordner ist der einzelne Ordner, der sich dort befindet.
Wenn mehrere vorhanden sind, wird der Ordner mit einem Übersichtsordner verwendet;
Ansonsten wird der Benutzer gefragt.

Der Arbeitsordner muss Folgendes enthalten:

* einen Übersichtsordner, dessen Name `3. Übersichten` ähnelt,
  Halten genau einer `.xlsx`-Datei, der Tabelle;
* die Vorlage `Appendix A - Risk Assesment.docx` auf ihrer obersten Ebene.

Ein Lauf läuft wie folgt ab:

1. Das Template wird überprüft.
   Der Lauf stoppt, wenn Abschnitt A oder eines seiner Labels nicht gefunden werden kann.
2. Der Benutzer gibt einmal den Namen der Person ein, die das Formular verarbeitet.
3. Der gesamte Arbeitsordner wird kopiert
   `Downloads/numb-form-runs/<date>_<time>_<id>/`.
   Alles nach diesem Schritt geschieht in der Kopie.
4. Firmennamen werden aus der Tabellenkalkulation gelesen Spalte Überschrift
   `Company Name`, von oben nach unten.
   Alle anderen Spalten werden ignoriert, einschließlich `Report created`.
   Die Tabellenkalkulation wird nie geschrieben.
   Seine SHA-256 Prüfsumme wird vor und nach dem Lauf überprüft.
5. Für jedes Unternehmen, die Ordner im Arbeitsordner und alle seine
   Unterordner durchsucht werden.
   Namen werden ohne Fall, Akzente und Umlaute verglichen (ü = u = ue),
   Interpunktion und Leerzeichen.
   Ein Match bedeutet, dass das Formular bereits existiert und das Unternehmen übersprungen wird.
   Ein Near Match wird dem Benutzer zur Bestätigung angezeigt.
   Ordner mit einer anderen Rechtsform stimmen nie überein:
   `Foo GmbH` und `Foo S.A.` sind verschiedene Unternehmen.
6. Für ein neues Unternehmen wird ein nach dem Unternehmen benannter Ordner erstellt.
   Die Vorlage wird in sie kopiert als
   `Appendix A - Risk Assesment - <Company Name>.docx`, und Abschnitt A ist gefüllt.
7. Das Dokument wird von der Festplatte zurückgelesen und angezeigt.
   Der Benutzer akzeptiert es, bearbeitet ein Feld oder lehnt es ab.
   Ein abgelehnter Ordner wird aus der Kopie gelöscht.

Nur die Zeilen von Abschnitt A, *Informationen über die Vertragspartei*,
gefüllt sind.
Abschnitt B und alles außerhalb von Abschnitt A bleibt Byte für Byte
unverändert.
Werte werden an den eigenen Textlauf des Labels angehängt, so dass sie die
Formatierung des Labels; es wird keine neue Formatierung hinzugefügt.

| Feld | Wert |
| --- | --- |
| Name/Firmenname der Vertragspartei | die wörtliche Tabellenkalkulationszelle |
| Rechtsform | das Suffix des Namens, wie geschrieben (`GmbH`, `S.A.`, `C.A. de C.V.`) |
| Sitz/Anschrift | vom Benutzer aus den gefundenen Kandidaten ausgewählt |
| Handelsregisternummer | immer leer |
| Bevollmächtigte Vertreter | `Full Name (Role 1, Role 2), Full Name (Role)`, Rollen in Englisch |
| Datum der Bewertung | Heute `DD/MM/YYYY` |
| Person, die das Formular bearbeitet | der zu Beginn eingegebene Name |

Die Kandidaten werden gesammelt aus, in der Reihenfolge der Autorität:

1. UK Companies House (Regierungsregister; nur mit einem freien Schlüssel);
2. das GLEIF-LEI-Register (Regulierungsregister; kein Schlüssel);
3. den eigenen rechtlichen Hinweis des Unternehmens (Impressum, Aviso legal, Legal Notice);
4. Wikidata (Community-editiert; hauptsächlich verwendet, um die Website zu finden).

Personen und Adressen werden niemals von einem Elternteil, einer Tochtergesellschaft oder
ähnlich benannte Entität.
Personen, die nur auf der Website oder nur auf Wikidata gefunden werden, sind markiert
*nicht verifiziert*, und standardmäßig ist dieser genaue Text angehängt:

> Vorsicht: Dieses Feld kann von internen Quellen abweichen oder aufgrund unzureichender Online-Informationen unpräzise sein

Wenn kein Vertreter gefunden werden kann, lautet das Feld:

> Es wurden keine Vertreter online gefunden, konsultieren Sie bitte andere interne Quellen für genauere Informationen

**numb-commit** kopiert die akzeptierten Ordner eines Run Back in die reale
Arbeitsordner.
Es fügt nur neue Ordner hinzu.
Es überschreibt oder löscht niemals etwas.
Es lehnt Ordner ab, die in der Zwischenzeit im echten Baum erschienen sind.
Jeder Ordner wird auf einen temporären Namen kopiert, mit SHA-256 verifiziert und dann
umbenannt.
Dokumente, die während der Überprüfung bearbeitet werden, werden aufgelistet, bevor der Benutzer bestätigt.
Danach bietet es an, die Überprüfungskopie zu löschen; die Protokolle werden beibehalten.

## OPTIONS

`-offline`
: Kontaktieren Sie keine Online-Quelle. Jeder Wert wird vom Benutzer eingegeben.

`-root dir`
: Verwenden Sie *dir* als Arbeitsordner, anstatt es zu erkennen.

`-downloads dir`
: Setzen Sie die Bewertungskopie in * dir * anstelle des Downloads-Ordners.

`-timeout duration`
: Halten Sie nach *Dauer* sicher an (z. B. `15m`). Das Maximum und
  Standard ist `25m`.

`-run dir`
: (numb-commit) Begehen Sie diesen Run-Ordner anstelle des neuesten
  Eine unverbindliche.

`-allow-non-tty`
: Lesen Sie Antworten aus einer Pipe. Nur für automatisierte Tests.

`-version`
: Drucken Sie die Version.

## ENVIRONMENT

`NUMBFORM_DOWNLOADS`
: Downloads Ordner zu verwenden, zum Beispiel ein Mounted Volume in einem Container.

`NUMBFORM_COMPANIES_HOUSE_KEY`
: Optional kostenlos UK Companies House API-Schlüssel. Nie protokolliert oder gespeichert.

`HTTPS_PROXY`, `NO_PROXY`
: Proxy-Einstellungen für Online-Lookups.

## FILES

`Downloads/numb-form-runs/<run>/<working folder>/`
: Die Überprüfungskopie.

`Downloads/numb-form-runs/<run>/numb-form.log`
: Log von allem, was der lauf gemacht hat.

`Downloads/numb-form-runs/<run>/manifest.json`
: Maschinenlesbare Aufzeichnung, die von numb-commit verwendet wird.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Log des Commits.

Kein Cache oder temporäre Dateien bleiben zurück.
Netzwerkantworten werden nur im Speicher gespeichert und Dokumente werden geschrieben
durch eine temporäre Datei, die umbenannt oder entfernt wird.
Ein Durchlauf schreibt höchstens 100 MB Dokumente neben der Kopie.

## EXIT STATUS

0 auf Erfolg, 1 auf Fehler, 2 auf einen Nutzungsfehler und 3 wenn durch die
Zeitlimit, durch Ctrl + C oder durch den Benutzer.

## EXAMPLES

Überprüfen Sie die echte Vorlage und schreiben Sie eine Probe an, ohne zu berühren
Muster:

```
numb-form check-template "Appendix A - Risk Assesment.docx" "%USERPROFILE%\Downloads\sample.docx"
```

Run, Review, dann Commit (numb-commit wartet auf Bestätigung):

```
cd /d "S:\Kunden" && numb-form && numb-commit
```

Laufen ohne Netzwerkzugang:

```
cd /d "S:\Kunden" && numb-form -offline && numb-commit
```

## SEE ALSO

[install.md](install.md), [containers.md](containers.md),
[../SECURITY.md](SECURITY.md), [../AGENTS.md](../../AGENTS.md)
