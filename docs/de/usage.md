# NUMB-FORM(1)

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../usage.md). Maßgeblich ist die englische Fassung.

## NAME

**numb-form**, **numb-commit** — Formulare für die Risikobewertung ausfüllen und diese festlegen

## SYNOPSIS

```
numb-form [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form check-template template.docx [sample.docx]
numb-form keys
numb-commit [-run dir] [-downloads dir]
```

## DESCRIPTION

**numb-form** wird im Ordner *über* dem Arbeitsordner gestartet. Der Arbeitsordner ist der einzelne Ordner, der sich dort befindet. Wenn es mehrere gibt, wird derjenige verwendet, der einen Übersichtsordner enthält; andernfalls wird der Benutzer gefragt.

Der Arbeitsordner muss Folgendes enthalten:

* einen Übersichtsordner, dessen Name `3. Übersichten` ähnelt und der genau eine `.xlsx`-Datei enthält, die Tabelle;
* die Vorlage `Appendix A - Risk Assesment.docx` auf ihrer obersten Ebene.

Ein Lauf läuft wie folgt ab:

1. Das Template wird überprüft. Der Lauf stoppt, wenn Abschnitt A oder eines seiner Labels nicht gefunden werden kann.
2. Der Benutzer gibt einmal den Namen der Person ein, die das Formular verarbeitet.
3. Der gesamte Arbeitsordner wird in `Downloads/numb-form-runs/<date>_<time>_<id>/` kopiert. Alles nach diesem Schritt geschieht in der Kopie.
4. Firmennamen werden aus der Tabellenkalkulationsspalte mit der Überschrift `Company Name` von oben nach unten gelesen. Alle anderen Spalten werden ignoriert, einschließlich `Report created`. Die Tabellenkalkulation wird nie geschrieben. Seine SHA-256 Prüfsumme wird vor und nach dem Lauf überprüft.
5. Für jedes Unternehmen werden die Ordner im Arbeitsordner und alle Unterordner durchsucht. Namen werden ohne Fall, Akzente und Umlaute (ü = u = ue), Interpunktion und Leerzeichen verglichen. Ein Match bedeutet, dass das Formular bereits existiert und das Unternehmen übersprungen wird. Ein Near Match wird dem Benutzer zur Bestätigung angezeigt. Ordner mit einer anderen Rechtsform stimmen nie überein: `Foo GmbH` und `Foo S.A.` sind unterschiedliche Unternehmen.
6. Für ein neues Unternehmen wird ein nach dem Unternehmen benannter Ordner erstellt. Die Vorlage wird als `Appendix A - Risk Assesment - <Company Name>.docx` kopiert und Abschnitt A wird ausgefüllt.
7. Das Dokument wird von der Festplatte zurückgelesen und angezeigt. Der Benutzer akzeptiert es, bearbeitet ein Feld oder lehnt es ab. Ein abgelehnter Ordner wird aus der Kopie gelöscht.

Nur die Zeilen von Abschnitt A, *Information About the Contracting Party*, sind gefüllt. Abschnitt B und alles außerhalb von Abschnitt A werden Byte für Byte unverändert gelassen. Werte werden an den eigenen Textlauf des Labels angehängt, so dass sie die Formatierung des Labels annehmen; es wird keine neue Formatierung hinzugefügt.

| Feld | Wert |
| --- | --- |
| Name / Company Name of the Contracting Party | die wörtliche Tabellenkalkulationszelle |
| Legal form | das Suffix des Namens, wie geschrieben (`GmbH`, `S.A.`, `C.A. de C.V.`) |
| Headquarters / Address | vom Benutzer aus den gefundenen Kandidaten ausgewählt |
| Commercial Registry Number | immer leer |
| Authorized Representatives | `Full Name (Role 1, Role 2), Full Name (Role)`, Rollen in Englisch |
| Date of the evaluation | Heute `DD/MM/YYYY` |
| Person Processing the Form | der zu Beginn eingegebene Name |

Die Kandidaten werden gesammelt aus, in der Reihenfolge der Autorität:

1. UK Companies House (Regierungsregister; nur mit einem freien Schlüssel);
2. das GLEIF LEI-Register (Regulierungsregister; kein Schlüssel);
3. den eigenen rechtlichen Hinweis des Unternehmens (Impressum, Aviso legal, Rechtlicher Hinweis);
4. Wikidata (Community-editiert; hauptsächlich verwendet, um die Website zu finden).

Personen und Adressen werden niemals von einem Elternteil, einer Tochtergesellschaft oder einer ähnlich benannten Einheit übernommen. Personen, die nur auf der Website oder nur auf Wikidata gefunden werden, sind * nicht verifiziert * markiert, und standardmäßig ist dieser genaue Text angehängt:

> Vorsicht: Dieses Feld kann von internen Quellen abweichen oder aufgrund unzureichender Online-Informationen unpräzise sein

Wenn kein Vertreter gefunden werden kann, lautet das Feld:

> Es wurden keine Vertreter online gefunden, konsultieren Sie bitte andere interne Quellen für genauere Informationen

**numb-commit** kopiert die akzeptierten Ordner eines Runs zurück in den echten Arbeitsordner. Es fügt nur neue Ordner hinzu. Es überschreibt oder löscht niemals etwas. Es lehnt Ordner ab, die in der Zwischenzeit im echten Baum erschienen sind. Jeder Ordner wird in einen temporären Namen kopiert, mit SHA-256 verifiziert und dann umbenannt. Dokumente, die während der Überprüfung bearbeitet werden, werden aufgelistet, bevor der Benutzer bestätigt. Danach bietet es an, die Überprüfungskopie zu löschen; die Protokolle werden beibehalten.

## OPTIONS

`-offline`
: Kontaktieren Sie keine Online-Quelle. Jeder Wert wird vom Benutzer eingegeben.

`-root dir`
: Verwenden Sie *dir* als Arbeitsordner, anstatt es zu erkennen.

`-downloads dir`
: Legen Sie die Bewertungskopie in * dir * anstelle des Downloads-Ordners.

`-timeout duration`
: Halten Sie nach *Dauer* sicher an (z. B. `15m`). Maximum und Standard ist `25m`.

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

## FILES

`Downloads/numb-form-runs/<run>/<working folder>/`
: Die Überprüfungskopie.

`Downloads/numb-form-runs/<run>/numb-form.log`
: Log von allem, was der lauf gemacht hat.

`Downloads/numb-form-runs/<run>/manifest.json`
: Maschinenlesbare Aufzeichnung, die von numb-commit verwendet wird.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Log des Commits.

Kein Cache oder temporäre Dateien bleiben zurück. Netzwerkantworten werden nur im Speicher gespeichert, und Dokumente werden durch eine temporäre Datei geschrieben, die umbenannt oder entfernt wird. Ein Durchlauf schreibt höchstens 100 MB Dokumente neben der Kopie.

## EXIT STATUS

0 zum Erfolg, 1 zum Fehler, 2 zu einem Nutzungsfehler und 3 beim Stoppen durch das Zeitlimit, durch Strg + C oder durch den Benutzer.

## EXAMPLES

Überprüfen Sie die echte Vorlage und schreiben Sie ein Beispiel, um es anzusehen, ohne die Vorlage zu berühren:

```
numb-form check-template "Appendix A - Risk Assesment.docx" "%USERPROFILE%\Downloads\sample.docx"
```

Laufen, überprüfen, dann Commit (numb-commit wartet auf Bestätigung):

```
cd /d "S:\Kunden" && numb-form && numb-commit
```

Laufen ohne Netzwerkzugang:

```
cd /d "S:\Kunden" && numb-form -offline && numb-commit
```

## SEE ALSO

[install.md](install.md), [containers.md](containers.md), [../SECURITY.md](SECURITY.md), [../AGENTS.md](../../AGENTS.md)
