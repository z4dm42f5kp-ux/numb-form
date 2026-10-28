# Sicherheit

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../../SECURITY.md). Maßgeblich ist die englische Fassung.

## Meldung einer Schwachstelle

Melden Sie es privat auf GitHub: Öffnen Sie die Registerkarte *Security* des Repositorys und wählen Sie *Report a vulnerability* (<https://github.com/z4dm42f5kp-ux/numb-form/security/advisories/new>). Öffnen Sie kein öffentliches Problem für eine Schwachstelle.

Den Bericht kurz halten:

* Was das Problem ist und welches Programm und welche Version (`numb-form -version`) es hat.
* Wie es ausgenutzt werden kann.
* Eine kleine Datei, Befehl oder Test, die es zeigt.

Senden Sie ein Problem pro Bericht. Fügen Sie niemals Kundendaten an: Beschreiben Sie das Problem mit erfundenen Namen (`Foo GmbH`) und ersetzen Sie echte Dateien durch kleine erfundene.

Berichte, die von KI-Tools geschrieben wurden, sind nur willkommen, nachdem eine Person überprüft hat, dass jeder Befund real ist und ein Sicherheitsproblem darstellt.

## Was als nächstes passiert

Der Prozess folgt dem [Go-Projekt ](https://go.dev/doc/security/policy).

* Der Bericht wird innerhalb von 7 Tagen bestätigt.
* Es wird innerhalb von 90 Tagen danach behoben oder veröffentlicht.
* Jede bestätigte Schwachstelle erhält eine GitHub-Sicherheitsempfehlung und, wenn sie qualifiziert ist, eine CVE. Der Reporter wird gutgeschrieben, es sei denn, er bittet darum, es nicht zu sein.

Jeder Bericht wird auf eine von drei Spuren gesetzt:

| Spurweite | Wann | Wie es festgelegt ist |
| --- | --- | --- |
| **URGENT** | ausgenutzt oder Kundendaten können ohne Aktion der Person, die das Programm ausführt, auslaufen oder zerstört werden | privat, sofort als Patch-Release veröffentlicht |
| **PRIVATE** | Ausnutzbar mit einer handgefertigten Datei oder Download, aber braucht die Person, um es zu verwenden | privat, veröffentlicht als Patch-Release |
| **PUBLIC** | schwer auszubeuten oder zu härten | in einem öffentlichen Pull Request, im nächsten Release |

Nur das neueste Release wird unterstützt. Update ist `numb-form update`.

## Was ist keine Vulnerabilität

Diese liegen außerhalb des Bedrohungsmodells:

* Jemand, der bereits die Programme, die Konfigurationsdateien oder den Arbeitsordner auf dem Computer ändern kann.
* Die Person, die die Programme ausführt, bestätigt ein falsches Formular oder ein falsches Datum. Jede Veränderung wird gezeigt und muss akzeptiert werden.
* Antworten eines KI-Assistenten, der mit der [AI prompt](ai-prompt.md) verwendet wird. Die Eingabeaufforderung bittet um eine Überprüfung mit internen Quellen.
* Office-Software, die numb-form für Konvertierungen startet, die sich so verhalten, wie sie dokumentiert ist.

Bugs, die keine Sicherheitsprobleme darstellen, gehen in das Bug-Berichtsformular (*Issues* → *New Issue*).

## Sichere Codierung und Audits

Die Regeln, denen der Code folgt, die Überprüfungs-Checkliste und die Audit-Routine sind in [.github/SECURE-CODING.md](../../.github/SECURE-CODING.md).

## Was die Programme versprechen

numb-form verarbeitet Kundendokumente, daher sind die Designregeln streng.

* **numb-form** only ever adds: a new folder per new company, with one filled form in it. It never changes or deletes an existing file or folder, and it removes only folders it created itself in the same run (on rejection or error).
* Bei `mode = review` wird der eigentliche Arbeitsordner überhaupt nicht geschrieben: Der gesamte Baum wird zuerst in Downloads kopiert und jede Änderung wird in der Kopie vorgenommen.
* **numb-commit** (review mode) only adds new folders. It never overwrites or deletes, refuses folders that appeared in the meantime, copies to a temporary name, verifies every file with SHA-256 and then renames.
* numb-form und numb-commit öffnen die Tabelle in `3. Übersichten` schreibgeschützt und schreiben sie niemals. Seine SHA-256 Prüfsumme wird vor der Kopie, nach der Kopie und am Ende des Laufs verglichen; eine Differenz stoppt den Lauf.
* **numb-date** ist das einzige Programm, das die Tabellenkalkulation schreibt, und nur `.xlsx` und `.xlsm`. Es schreibt nur leere oder ziffernfreie Zellen der Spalte *Report created*, nachdem der Benutzer der Liste der Daten zugestimmt hat. Die Zellen werden in das XML des Blattes gespleißt, behalten ihren eigenen Stil bei und erhalten keine neue Formatierung; jeder andere Teil der Arbeitsmappe wird Byte für Byte kopiert, und Zellen mit Formeln werden in Ruhe gelassen. Vor dem Schreiben vergleicht es jede Zelle jedes Blatts der neuen Arbeitsmappe mit dem Original: Alles andere, was sich unterscheidet, stoppt es. SHA-256 schützt jeden Schritt: Die Datei muss immer noch die Prüfsumme haben, die sie beim Lesen hatte, die temporäre Kopie daneben muss mit der neuen Arbeitsmappe übereinstimmen, bevor sie in die Datei umbenannt wird, und die Datei muss danach übereinstimmen. Das Original wird im Protokollordner aufbewahrt, bis der Benutzer das von der Festplatte zurückgelesene Ergebnis akzeptiert; *undo* stellt es wieder her, überprüft mit SHA-256. Es weigert sich zu laufen, während Excel oder LibreOffice die Datei geöffnet hat. Formulare werden nur gelesen.
* Dokumente werden nicht re-serialisiert. Das XML wird nur nach Byte-Offsets gescannt und Werte werden in die ursprünglichen Bytes gespleißt; alle anderen Teile des `.docx` werden Byte für Byte kopiert. Nach jedem Ausfüllen wird das Ergebnis erneut analysiert und Absatz für Absatz mit der Vorlage verglichen. Werte werden XML-escaped und von Zeichen entfernt, die in XML nicht erlaubt sind. Vorlagen mit Makros werden abgelehnt.
* Keine Formatierung wird jemals hinzugefügt. Werte werden dem eigenen Textlauf des Labels oder einer Kopie der vorhandenen Eigenschaften dieses Runs hinzugefügt, wenn dem Label ein Tab folgt.
* Jedes Dokument wird nach dem Schreiben von der Festplatte zurückgelesen. Eine Person sieht und akzeptiert es, und die Programme weigern sich, ohne ein interaktives Terminal zu laufen.
* Ein Lauf hat kein Zeitlimit, es sei denn, `timeout` setzt eins; Ctrl+C oder das Limit stoppt es sicher. Es schreibt höchstens 100 MB Dokumente neben der Kopie und hinterlässt keinen Cache oder temporäre Dateien: temporäre Dateien werden umbenannt oder entfernt, und Konverter laufen in einem privaten temporären Ordner mit einem Wegwerf-Office-Profil, das danach gelöscht wird.

## Konfiguration und Formate

* Der Konfigurationsparser weist unbekannte Schlüssel und ungültige Werte zurück.
* Schlüssel, die erweitern, was numb-form akzeptiert (`legacy_formats`), werden nur in der Systemdatei geehrt. Unter Unix wird diese Datei ignoriert, es sei denn, sie gehört zu root und ist nicht von einer Gruppe oder anderen beschreibbar; unter Windows befindet sie sich unter Programmdateien.
* Legacy-Binärformate (.doc, .xls) sind standardmäßig deaktiviert. Der .xls-Reader (Compound-Datei und BIFF8) ist numb-forms eigener, Bounds-Checks jeder Offset und Kette und wird kontinuierlich Fuzzed.
* .doc, .pages und .numbers werden von Office-Software bereits auf dem Computer (LibreOffice, Apache OpenOffice, Apple Pages oder Numbers) mit festen Argumenten konvertiert; Dateinamen werden als Argumente übergeben, niemals als Skripttext. Das Schreiben von .pages ist experimentell und standardmäßig ausgeschaltet.
* Vorlagen mit Makros werden abgelehnt; makrofähige Tabellenkalkulationen (.xlsm) werden nur für Zellenwerte gelesen.

## Netz

Ein Formularlauf macht keine Netzwerkverbindungen. Firmennamen, Adressen und Vertreter verlassen niemals den Computer.

`numb-form update` ist der einzige Befehl, der das Netzwerk kontaktiert. Es wird nur ausgeführt, wenn der Benutzer es eingibt, und niemals während eines Formularlaufs.

* Downloads kommt nur von `github.com` und `*.githubusercontent.com` über HTTPS, einschließlich jeder Weiterleitung.
* Das Archiv muss mit seiner Zeile im `SHA256SUMS` des Releases übereinstimmen; dies ist das gleiche Vertrauen wie die Installationsbefehle (die Sigstore-Signatur und die Build-Herkunft werden nicht vom Programm überprüft; überprüfen Sie sie von Hand, wie in [docs/install.md](install.md) beschrieben, falls erforderlich).
* Das Archiv (höchstens 64 MB) wird im Speicher gespeichert; nur `numb-form` und `numb-commit` auf seiner obersten Ebene werden daraus entnommen.
* Das neue `numb-form` muss seine Version starten und melden, bevor etwas ersetzt wird; bei einem Fehler bleiben die installierten Programme so, wie sie waren, und die inszenierten Dateien werden entfernt.
* Es wird mit `-offline`, mit `online = no` (die ein Administrator in der Systemkonfiguration festlegen kann) und für Builds von der Quelle abgelehnt.

## Lieferkette

**Keine Module von Drittanbietern Go.** Die Programme verwenden nur die Go-Standardbibliothek. `go.mod` hat keine `require`-Linien, und `scripts/check-deps.sh` schlägt den Build fehl, wenn eine Abhängigkeit, `go.sum` oder `vendor/` angezeigt wird oder wenn ein Nicht-Standard-Import verwendet wird. Da es nichts zum download gibt, gibt es nichts zu halluzinieren, typosquat oder zu manipulieren.

**PGP und seine Ersatzstoffe.** Authentizität wird mit den Mechanismen überprüft, die jedes Ökosystem tatsächlich unterstützt. PGP wird nicht verwendet, weil die beteiligten Ökosysteme nicht mehr darauf angewiesen sind (PyPI entfernte PGP-Signaturunterstützung im Jahr 2023).

| Komponente | Überprüfung |
| --- | --- |
| Go Toolchain in CI | das neueste stabile Go, das von `actions/setup-go` von go.dev heruntergeladen und mit dem veröffentlichten SHA-256 verglichen wurde; ein täglicher Check schlägt einen Umbau des neuesten Releases vor, wenn ein neuerer Go behebt eine Schwachstelle in seinen Binärdateien |
| Go Tools in CI (staticcheck, govulncheck, gosec) | genaue Versionen, verifiziert durch `go run` gegen die Go-Checksum-Datenbank, ein öffentliches Transparenzprotokoll; eine tägliche Überprüfung schlägt neue stabile Versionen nach 14 Tagen vor |
| GitHub Aktionen | Angeheftet auf vollständige Commit-SHAs; Dependabot prüft täglich und schlägt stabile Updates nach einer 14-tägigen Abklingzeit vor, Sicherheitsupdates sofort |
| Python-Tools in CI (zizmor, Argos Translate) | jedes Paket mit SHA-256 (`--require-hashes`), nur Rädern (`--only-binary :all:`), keine Abhängigkeitsauflösung zum Zeitpunkt der Installation (`--no-deps`) |
| Argos Translate Modelle | Index gelesen bei einem gepinnten Commit; SHA-256 jedes Modells wird in `.github/translate/models.sha256` gepinnt, bei der ersten Verwendung aufgezeichnet und danach durchgesetzt |
| Containerbasisbild | keine: Bilder sind gebaut `FROM scratch` |

**Releases.** Release-Binärdateien werden von GitHub Actions aus einem Tag mit `-trimpath` und einem festen Zeitstempel erstellt, so dass sie reproduzierbar sind. Ein Release wird erst veröffentlicht, nachdem der Besitzer es in der `release`-Umgebung genehmigt hat (siehe [.github/RELEASING.md](../../.github/RELEASING.md)). Jedes Release hat:

* Prüfsummen für `SHA256SUMS` (GNU-Format) und `SHA256` (BSD-Format);
* `SHA256SUMS.sigstore.json`, eine schlüssellose Sigstore-Signatur der Prüfsummen, gebunden an die Identität des Release-Workflows;
* SLSA Build Provenienzbescheinigungen für jedes Archiv und Binär (`gh attestation verify`);
* ein Containerbild auf ghcr.io, signiert mit cosign und bescheinigt.

[docs/install.md](install.md) zeigt, wie man jeden von ihnen überprüft.

## Pipeline für Sicherheitstests

Pull-Anfragen führen ein schnelles Gate aus (Formatierung, Tierarzt, Statikcheck, Tests unter Linux, macOS und Windows, govulncheck, gosec, zizmor, der Verlaufsscan und die Überprüfung der Abhängigkeit). Pushes zu `main` und Agent Branchs und einen nächtlichen Zeitplan, führen Sie alles aus (die LibreOffice Formattests laufen einmal pro Pull Request Change und nicht für jeden Push):

| Überprüfung | Werkzeug |
| --- | --- |
| Keine Abhängigkeiten, Formatierung, Tierarzt auf sechs Betriebssystemen | `scripts/check-deps.sh`, `gofmt`, `go vet` |
| Statische Analyse | Statikprüfung |
| Bekannte Sicherheitslücken im Code und in eingebauten Binärdateien für jedes Ziel-Betriebssystem | govulncheck |
| Security linting, gemeldet zum Code-Scannen | Gosec |
| Semantische Analyse von Go und der Workflows | CodeQL (`security-extended`, `security-and-quality`) |
| Fuzzing jedes Template-Füllers und Scanners (docx, odt, rtf), jedes Tabellenkalkulationslesegeräts (xlsx, ods, csv, legacy xls), des Konfigurationsparsers, des numb-commit-Requestreaders und des Namensmatchers | Go native fuzzing, 2 minuten pro ziel auf haupt, 10 nächtliche. |
| Vorlagen und Tabellenkalkulationen, geschrieben von real LibreOffice (odt, fodt, rtf, docx, doc, xls, ods, fods) | Formate Workflow |
| Workflow-Sicherheit (Injection, Anmeldeinformationen, Berechtigungen, Pins) | Zizmor, pedantisch |
| Kundendateien oder Anmeldeinformationen aus Versehen | Historienscan |
| Tests mit dem Renndetektor unter Linux, macOS und Windows, mit dem ältesten und neuesten Go | `go test -race` |
| Tests auf FreeBSD, OpenBSD und NetBSD | Virtuelle Maschinen |
| Container-Image baut und weigert sich, ohne Terminal zu laufen | Docker, Podman |
| Neue Abhängigkeiten in Pull Requests | Überprüfung der Abhängigkeit |

OpenSSF Scorecard überprüft die Praktiken des Repository. Pull-Anfragen, einschließlich Dependabots, benötigen niemals Schreibberechtigungen: Ergebnisse, die sie erfordern, werden nur von `main` hochgeladen.

Workflows beginnen ohne Berechtigungen und gewähren jedem Job nur das, was er braucht; Checkouts behalten keine Anmeldeinformationen.

## Container und Kubernetes

Das Bild enthält nur die beiden Binärdateien, ein CA-Bundle und die Zeitzonendatenbank und läuft als UID 65532. Die dokumentierten Befehle und Manifeste verwenden ein schreibgeschütztes Root-Dateisystem, lassen alle Funktionen fallen, verbieten die Eskalation von Privilegien, montieren den Arbeitsordner schreibgeschützt für numb-form, montieren kein Dienstkonto-Token und legen eine Frist fest.

## Dieses Repository ist öffentlich

Legen Sie niemals Kundendokumente, Tabellenkalkulationen, Protokolle oder ausgeführte Ordner fest. `.gitignore` schließt sie aus und CI lehnt sie ab.
