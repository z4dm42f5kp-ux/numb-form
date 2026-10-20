# NUMB-FORM(1)

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../usage.md). Maßgeblich ist die englische Fassung.

## NAME

**numb-form**, **numb-commit** — Formulare für die Risikobewertung ausfüllen

## SYNOPSIS

```
numb-form [-f conf] [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form [-f conf] check-template template [sample]
numb-form [-f conf] config [init | set key value]
numb-form keys
numb-form [-f conf] update [stable | rolling | v1.2.3]
numb-form rollback
numb-commit [-f conf] [-run dir] [-downloads dir]
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
4. Firmennamen werden aus der Spalte mit der Überschrift `Customer Name` von oben nach unten gelesen. Alle anderen Spalten werden ignoriert, einschließlich `Report created`. Die Tabelle wird nie geschrieben; seine SHA-256 Prüfsumme wird vor und nach dem Lauf überprüft.
5. Für jedes Unternehmen werden die Ordner im Arbeitsordner und alle Unterordner durchsucht. Namen werden ohne Fall, Akzente und Umlaute (ü = u = ue), Interpunktion und Leerzeichen verglichen. Ein Match bedeutet, dass das Formular bereits existiert und das Unternehmen übersprungen wird. Ein Beinahe-Match wird dem Benutzer zur Bestätigung angezeigt (oder mit `confirm = none` als *needs review* gemeldet und übersprungen). Ordner mit einer anderen Rechtsform stimmen nie überein: `Foo GmbH` und `Foo S.A.` sind unterschiedliche Unternehmen.
6. Für ein neues Unternehmen wird ein nach dem Unternehmen benannter Ordner erstellt. Die Vorlage wird als `<template name> - <Company Name>.<ext>` kopiert und Abschnitt A wird ausgefüllt.
7. Das Dokument wird von der Festplatte zurückgelesen und angezeigt. Der Benutzer akzeptiert es, bearbeitet ein Feld oder lehnt es ab; ein abgelehnter Ordner wird gelöscht. Bei `confirm = none` muss das Rücklesen den Werten genau entsprechen.

Nur die Zeilen von Abschnitt A, *Information About the Contracting Party*, sind gefüllt. Abschnitt B und alles außerhalb von Abschnitt A werden Byte für Byte unverändert gelassen und es wird keine Formatierung hinzugefügt.

| Feld | Wert |
| --- | --- |
| Name / Company Name of the Contracting Party | die wörtliche Tabellenkalkulationszelle |
| Legal form | gefunden im Namen, auch vor einer Zweig- oder Ländernote (`Foo GmbH (Wien)`), geschrieben wie im Namen mit herkömmlichen Großbuchstaben (`GMBH` gibt `GmbH`); nur gefragt, wenn der Name keine hat |
| Headquarters / Address | ausgewählt aus den gefundenen Kandidaten |
| Commercial Registry Number | immer `No commercial registry number was found online consult internal sources for further information`; es kann durch die Nummer mit *edit* ersetzt werden |
| Authorized Representatives | `Full Name (Role 1, Role 2), Full Name (Role)`, Rollen in Englisch |
| Date of the evaluation | heute `date_format` (Standard `DD/MM/YYYY`) |
| Person Processing the Form | aus der Konfiguration oder zu Beginn eingegeben |

Alles wird in einem Durchgang pro Unternehmen nachgeschlagen, ohne Fragen. Die Kandidaten kommen aus, in der Reihenfolge der Autorität:

1. Regierungsregister: UK Companies House (nur mit einem kostenlosen Schlüssel); EU VIES, für eine EU-Mehrwertsteuernummer, die auf der Website des Unternehmens gedruckt wird (Name und Anschrift; Deutschland bestätigt nur, dass die Nummer gültig ist); Receita Federal, für eine brasilianische CNPJ, die auf der Website gedruckt wird (Adresse, Administratoren und Direktoren), lesen Sie den kostenlosen BrasilAPI-Spiegel; Regierungsblatter und Register, die bei der Suche nach ihren Websites für das Unternehmen gefunden werden: BORME (Spanien), Portal da Justiça (Portugal), Boletín Oficial (Argentinien), Sunbiz (Florida), Companies House-Seiten (Vereinigtes Königreich); welche von den Ländern abhängt, auf die die Rechtsform hinweist;
2. das GLEIF LEI-Register (Regulierungsregister);
3. die eigene Website des Unternehmens: seine rechtlichen Hinweise (Impressum, Aviso legal, Erwähnungen légales, Rechtliche Hinweise) zuerst, dann über, Kontakt- und Datenschutzseiten;
4. Community-Daten, die mit dem Vorsichtstext nicht verifiziert angezeigt werden: Unternehmensdatenverzeichnisse, die Registerdaten neu veröffentlichen (northdata, firmenwissen, Moneyhouse, Racius, Empresite, cnpj.biz, OpenCorporates, Dun & Bradstreet und ähnliches; höchstens drei Seiten pro Unternehmen), die Ausschnitte der Web-Suchergebnisse, OpenStreetMap (für eine Adresse) und Wikidata.

Wenn von einer offiziellen Quelle bestätigte Vertreter oder eine Adresse aus einem Register oder der Website des Unternehmens noch fehlen, führt numb-form eine weitere Suche nach den Vertretern (mit dem lokalen Wort, wie *Geschäftsführer* oder *representante legal*) und eine nach der Adresse durch und liest die Verzeichnisseiten, die sie finden. Eine Person, die durch zwei verschiedene Arten von Quellen benannt wurde (z. B. die Website des Unternehmens und ein Verzeichnis), zählt als verifiziert.

Die Website wird automatisch gefunden: von Wikidata, von einer DuckDuckGo-Suche nach dem genauen Namen und von Adressen wie `foobar.de` oder `foobar.com.mx`, die aus dem Namen und den Ländern ihrer Rechtsform erstellt wurden. Eine Website wird nur verwendet, wenn eine ihrer Seiten das Unternehmen mit seiner Rechtsform benennt. Auf dem rechtlichen Hinweis einer Gruppe, der mehrere Unternehmen benennt, wird nur der Teil über das Unternehmen gelesen.

Repräsentative Labels werden in deutscher, englischer, spanischer, portugiesischer, französischer, italienischer, niederländischer, polnischer, tschechischer, slowakischer und nordischer Sprache überall in einer Zeile, in Tabellen und Fußzeilen anerkannt und in englische Rollen übersetzt. Labels, die andere Personen benennen (Aufsichtsrat, Datenschutzbeauftragter, Inhaltsverantwortlicher, Kontakt, registrierter Agent) und Personen, die gegangen sind (*Ceses/Dimisiones*, *ehemalig*, *ehemalig*), werden übersprungen.

Offizielle Nummern und Registereinträge, die auf der Website gedruckt werden (z. B. `HRB 12345` oder `FN 123456 a`), werden vor dem Abfassen des Formulars angezeigt; die Nummer kann mit *Bearbeiten* in *Commercial Registry Number* eingegeben werden.

Die Suche verwendet wiederum schlüssellose Engines: DuckDuckGo, dann Mojeek, Yahoo und Brave Search (`search` in der Konfiguration). Jeder wird höchstens ein paar Mal pro Unternehmen gefragt, ein paar Sekunden auseinander. Ein Motor, der sich weigert (ein menschlicher Scheck, ein Tariflimit), ruht eine Minute lang und der nächste wird gefragt.

Wenn Google Chrome installiert ist (oder, ohne es, Chromium oder Microsoft Edge), startet numb-form es unsichtbar, ohne ein Fenster (headless), für zwei Dinge: Lesen von Firmenwebsites, deren Inhalt von JavaScript erstellt wird oder die Programme wegschalten, höchstens drei Seiten pro Site; und Suchen, wenn jede Engine ablehnt. Jeder Start verwendet ein neues, leeres Profil im temporären Ordner, das direkt danach gelöscht wird; der Browser, die Historie und die Anmeldungen des Benutzers werden nie verwendet. `browser = none` schaltet dies aus.

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

`config set key value`
: Speichern Sie eine Einstellung in der Konfigurationsdatei pro Benutzer, z. B. `numb-form config set person "Erika Mustermann"`. Der Wert wird überprüft, bevor er gespeichert wird.

`keys`
: Erklären Sie die optionalen kostenlosen API-Schlüssel.

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

Laufen ohne Netzwerkzugang:

```
cd /path/to/GWG && numb-form -offline
```

## SEE ALSO

[configuration.md](configuration.md), [formats.md](formats.md), [install.md](install.md), [containers.md](containers.md), [../SECURITY.md](SECURITY.md), [../AGENTS.md](../../AGENTS.md)
