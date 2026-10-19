# Sicherheit

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../../SECURITY.md). Maßgeblich ist die englische Fassung.

## Meldung einer Schwachstelle

Verwenden Sie GitHubs Private Vulnerability Reporting: *Security* → *Report a Vulnerability* on this repository. Öffnen Sie kein öffentliches Problem. Fügen Sie keine Kundendaten an; beschreiben Sie das Problem mit erfundenen Namen.

Nur das neueste Release wird unterstützt.

## Was die Programme versprechen

numb-form verarbeitet Kundendokumente, daher sind die Designregeln streng.

* **numb-form** fügt immer nur hinzu: einen neuen Ordner pro neuer Firma, mit einem ausgefüllten Formular darin. Es ändert oder löscht niemals eine vorhandene Datei oder einen vorhandenen Ordner und entfernt nur Ordner, die es selbst in derselben Ausführung erstellt hat (bei Ablehnung oder Fehler).
* Bei `mode = review` wird der eigentliche Arbeitsordner überhaupt nicht geschrieben: Der gesamte Baum wird zuerst in Downloads kopiert und jede Änderung wird in der Kopie vorgenommen.
* **numb-commit** (Review-Modus) fügt nur neue Ordner hinzu. Es überschreibt oder löscht niemals, lehnt Ordner ab, die in der Zwischenzeit erschienen sind, kopiert einen temporären Namen, überprüft jede Datei mit SHA-256 und benannt dann um.
* Die Tabelle in `3. Übersichten` wird schreibgeschützt geöffnet und niemals geschrieben. Seine SHA-256 Prüfsumme wird vor der Kopie, nach der Kopie und am Ende des Laufs verglichen; eine Differenz stoppt den Lauf.
* Dokumente werden nicht re-serialisiert. Das XML wird nur nach Byte-Offsets gescannt und Werte werden in die ursprünglichen Bytes gespleißt; alle anderen Teile des `.docx` werden Byte für Byte kopiert. Nach jedem Ausfüllen wird das Ergebnis erneut analysiert und Absatz für Absatz mit der Vorlage verglichen. Werte werden XML-escaped und von Zeichen entfernt, die in XML nicht erlaubt sind. Vorlagen mit Makros werden abgelehnt.
* Keine Formatierung wird jemals hinzugefügt. Werte werden dem eigenen Textlauf des Labels oder einer Kopie der vorhandenen Eigenschaften dieses Runs hinzugefügt, wenn dem Label ein Tab folgt.
* Jedes Dokument wird nach dem Schreiben von der Festplatte zurückgelesen. Mit `confirm = each` (Standard) sieht und akzeptiert eine Person es, und die Programme weigern sich, ohne ein interaktives Terminal zu laufen. Bei `confirm = none` muss das Zurücklesen genau den beabsichtigten Werten entsprechen, Near-Ordner-Namensübereinstimmungen werden nie automatisch entschieden und es werden nur die Quellen und das Vertrauensniveau in der Konfiguration verwendet.
* Ein Lauf stoppt sicher nach höchstens 25 Minuten. Es schreibt höchstens 100 MB Dokumente neben der Kopie und hinterlässt keinen Cache oder temporäre Dateien: Netzwerkdaten bleiben im Speicher, temporäre Dateien werden umbenannt oder entfernt, und Konverter laufen in einem privaten temporären Ordner mit einem Wegwerf-Office-Profil, danach gelöscht.

## Konfiguration und Formate

* Der Konfigurationsparser weist unbekannte Schlüssel und ungültige Werte zurück.
* Schlüssel, die erweitern, was numb-form akzeptiert (`legacy_formats`), werden nur in der Systemdatei geehrt. Unter Unix wird diese Datei ignoriert, es sei denn, sie gehört zu root und ist nicht von einer Gruppe oder anderen beschreibbar; unter Windows befindet sie sich unter Programmdateien.
* Legacy-Binärformate (.doc, .xls) sind standardmäßig deaktiviert, und Läufe, die sie verwenden, sind immer vollständig offline. Der .xls-Reader (Compound-Datei und BIFF8) ist numb-forms eigener, Bounds-Checks jeder Offset und Kette und wird kontinuierlich Fuzzed.
* .doc, .pages und .numbers werden von Office-Software bereits auf dem Computer (LibreOffice, Apache OpenOffice, Apple Pages oder Numbers) mit festen Argumenten konvertiert; Dateinamen werden als Argumente übergeben, niemals als Skripttext. Das Schreiben von .pages ist experimentell und standardmäßig ausgeschaltet.
* Vorlagen mit Makros werden abgelehnt; makrofähige Tabellenkalkulationen (.xlsm) werden nur für Zellenwerte gelesen.

## Netz

Online-Lookups senden nur Firmennamen und die Unternehmenswebsite, die der Benutzer bestätigt. Sie gehen auf die GLEIF API, Wikidata, die firmeneigene Website und, nur wenn ein freier Schlüssel gesetzt ist, UK Companies House. Verwenden Sie `-offline`, um nichts zu senden.

* Nur HTTPS, TLS 1.2 oder höher, höchstens fünf HTTPS-Weiterleitungen.
* Ohne Proxy werden Verbindungen zu Loopback-, Private-, Link-Local- und Multicast-Adressen abgelehnt (serverseitiger Request Forgery Guard).
* Antwortkörper sind auf 2 MB begrenzt, im Speicher gehalten und nie auf die Festplatte geschrieben.
* Der Text aus dem Netzwerk wird von Steuerzeichen, ANSI-Escape-Sequenzen und bidirektionalen Überschreibungen befreit, bevor er angezeigt wird, so dass eine Webseite das Terminal nicht manipulieren kann.
* Archive werden mit Größenbegrenzungen (Zip Bomb Guard) gelesen.
* API-Schlüssel kommen nur aus der Umgebung, werden nur an ihren eigenen Dienst gesendet und werden niemals protokolliert oder gespeichert.
* Daten werden niemals von einem Mutterunternehmen, einer Tochtergesellschaft oder einer ähnlich benannten Einheit übernommen.

`numb-form update` ist der einzige Befehl, der Programme herunterlädt. Es wird nur ausgeführt, wenn der Benutzer es eingibt, und niemals während eines Formularlaufs.

* Downloads kommt nur von `github.com` und `*.githubusercontent.com` über HTTPS, einschließlich jeder Weiterleitung.
* Das Archiv muss mit seiner Zeile im `SHA256SUMS` des Releases übereinstimmen; dies ist das gleiche Vertrauen wie die Installationsbefehle (die Sigstore-Signatur und die Build-Herkunft werden nicht vom Programm überprüft; überprüfen Sie sie von Hand, wie in [docs/install.md](install.md) beschrieben, falls erforderlich).
* Das Archiv (höchstens 64 MB) wird im Speicher gespeichert; nur `numb-form` und `numb-commit` auf seiner obersten Ebene werden daraus entnommen.
* Das neue `numb-form` muss seine Version starten und melden, bevor etwas ersetzt wird; bei einem Fehler bleiben die installierten Programme so, wie sie waren, und die inszenierten Dateien werden entfernt.
* Es wird mit `-offline`, mit `online = no` (die ein Administrator in der Systemkonfiguration festlegen kann) und für Builds von der Quelle abgelehnt.

## Lieferkette

**Keine Module von Drittanbietern Go.** Die Programme verwenden nur die Go-Standardbibliothek. `go.mod` hat keine `require`-Linien, und `scripts/check-deps.sh` schlägt den Build fehl, wenn eine Abhängigkeit, `go.sum` oder `vendor/` angezeigt wird oder wenn ein Nicht-Standard-Import verwendet wird. Da es nichts zum download gibt, gibt es nichts zu halluzinieren, typosquat oder zu manipulieren.

**PGP und seine Ersatzstoffe. ** Authentizität wird mit den Mechanismen überprüft, die jedes Ökosystem tatsächlich unterstützt. PGP wird nicht verwendet, weil die beteiligten Ökosysteme nicht mehr darauf angewiesen sind (PyPI entfernte PGP-Signaturunterstützung im Jahr 2023).

| Komponente | Überprüfung |
| --- | --- |
| Go Toolchain in CI | Heruntergeladen von `actions/setup-go` von go.dev und überprüft mit dem veröffentlichten SHA-256 |
| Go Tools in CI (staticcheck, govulncheck, gosec) | genaue Versionen, verifiziert von `go run` gegen die Go-Checksum-Datenbank, ein öffentliches Transparenzprotokoll |
| GitHub Aktionen | Angeheftet auf Full Commit SHAs; Dependabot schlägt Updates nach einer 14-tägigen Abklingzeit vor |
| Python-Tools in CI (zizmor, Argos Translate) | jedes Paket mit SHA-256 (`--require-hashes`), nur Rädern (`--only-binary :all:`), keine Abhängigkeitsauflösung zum Zeitpunkt der Installation (`--no-deps`) |
| Argos Translate Modelle | Index gelesen bei einem gepinnten Commit; SHA-256 jedes Modells wird in `.github/translate/models.sha256` gepinnt, bei der ersten Verwendung aufgezeichnet und danach durchgesetzt |
| Containerbasisbild | keine: Bilder sind gebaut `FROM scratch` |

**Releases.** Release-Binärdateien werden von GitHub Actions aus einem Tag mit `-trimpath` und einem festen Zeitstempel erstellt, so dass sie reproduzierbar sind. Jedes Release hat:

* Prüfsummen für `SHA256SUMS` (GNU-Format) und `SHA256` (BSD-Format);
* `SHA256SUMS.sigstore.json`, eine schlüssellose Sigstore-Signatur der Prüfsummen, gebunden an die Identität des Release-Workflows;
* SLSA Build Provenienzbescheinigungen für jedes Archiv und Binär (`gh attestation verify`);
* ein Containerbild auf ghcr.io, signiert mit cosign und bescheinigt.

[docs/install.md](install.md) zeigt, wie man jeden von ihnen überprüft.

## Pipeline für Sicherheitstests

Pull-Anfragen führen ein schnelles Gate aus (Formatierung, Tierarzt, Statikcheck, Tests unter Linux, macOS und Windows, govulncheck, gosec, zizmor, der Verlaufsscan und die Überprüfung der Abhängigkeit). Pushes zu `main` und Agentenzweige, und einen nächtlichen Zeitplan, führen Sie alles aus:

| Überprüfung | Werkzeug |
| --- | --- |
| Keine Abhängigkeiten, Formatierung, Tierarzt auf sechs Betriebssystemen | `scripts/check-deps.sh`, `gofmt`, `go vet` |
| Statische Analyse | Statikprüfung |
| Bekannte Sicherheitslücken im Code und in eingebauten Binärdateien für jedes Ziel-Betriebssystem | govulncheck |
| Security linting, gemeldet zum Code-Scannen | Gosec |
| Semantische Analyse von Go und der Workflows | CodeQL (`security-extended`, `security-and-quality`) |
| Fuzzing jedes Template-Füllers und Scanners (docx, odt, rtf), jedes Tabellenkalkulationslesegeräts (xlsx, ods, csv, legacy xls), des Konfigurationsparsers und des Namensmatchers | Go native fuzzing, 2 minuten pro ziel auf haupt, 10 nächtliche. |
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
