# Sicherheit

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../../SECURITY.md). Maßgeblich ist die englische Fassung.

## Meldung einer Schwachstelle

Verwenden Sie GitHubs private Vulnerability Reporting:
*Sicherheit* → * Melden Sie eine Sicherheitslücke* in diesem Repository.
Öffnen Sie kein öffentliches Problem.
Fügen Sie keine Kundendaten an; beschreiben Sie das Problem mit erfundenen Namen.

Nur das neueste Release wird unterstützt.

## Was die Programme versprechen

numb-form verarbeitet Kundendokumente, daher sind die Designregeln streng.

* Der eigentliche Arbeitsordner wird niemals von **numb-form** geschrieben.
  Der gesamte Baum wird zuerst in Downloads kopiert und jede Änderung wird in
  die Kopie.
* **numb-commit** fügt nur neue Ordner hinzu.
  Es nie überschreibt oder löscht, lehnt Ordner, die in der
  Kopien auf einen temporären Namen, überprüft jede Datei mit SHA-256
  und dann Umbenennungen.
* Die Tabelle in `3. Übersichten` wird schreibgeschützt geöffnet und ist niemals
  schriftlich.
  Seine SHA-256 Prüfsumme wird vor der Kopie, nach der Kopie und
  Am Ende des Laufs; eine Differenz stoppt den Lauf.
* Dokumente werden nicht re-serialisiert.
  Das XML wird nur nach Byte-Offsets gescannt und Werte werden in die
  Originalbytes; alle anderen Teile des `.docx` werden Byte für Byte kopiert.
  Nach jedem Ausfüllen wird das Ergebnis erneut analysiert und mit dem
  Vorlage Absatz für Absatz.
  Werte werden XML-escaped und von Zeichen entfernt, die in XML nicht erlaubt sind.
  Vorlagen mit Makros werden abgelehnt.
* Keine Formatierung wird jemals hinzugefügt.
  Werte schließen sich dem eigenen Textlauf des Labels oder einer Kopie des vorhandenen Ablaufs an
  Eigenschaften, wenn dem Label ein Tab folgt.
* Jedes Dokument wird einer Person gezeigt, von der Festplatte zurückgelesen und muss
  akzeptiert.
  Die Programme weigern sich, ohne ein interaktives Terminal zu laufen.
* Ein Lauf stoppt sicher nach höchstens 25 Minuten.
  Es schreibt höchstens 100 MB Dokumente neben der Kopie und hinterlässt keine
  Cache oder temporäre Dateien: Netzwerkdaten bleiben im Speicher und vorübergehend
  Dateien werden umbenannt oder entfernt.

## Netz

Online-Lookups senden nur Firmennamen und die Website des Unternehmens
bestätigt.
Sie gehen auf die GLEIF API, Wikidata, die eigene Website des Unternehmens und,
Nur wenn ein freier Schlüssel gesetzt ist, UK Companies House.
Verwenden Sie `-offline`, um nichts zu senden.

* Nur HTTPS, TLS 1.2 oder höher, höchstens fünf HTTPS-Weiterleitungen.
* Ohne Proxy, Verbindungen zu Loopback, privat, link-lokal und
  Multicast-Adressen werden abgelehnt (serverseitiger Fälschungsschutz).
* Response Bodys sind auf 2 MB begrenzt, im Speicher gehalten und nie geschrieben
  Scheibe.
* Text aus dem Netzwerk wird von Kontrollzeichen entfernt, ANSI entkommen
  Sequenzen und bidirektionale Überschreibungen, bevor es angezeigt wird, also eine Webseite
  Das Terminal kann nicht manipuliert werden.
* Archive werden mit Größenbegrenzungen (Zip Bomb Guard) gelesen.
* API-Schlüssel kommen nur aus der Umgebung, werden nur an ihre eigenen gesendet
  und werden niemals protokolliert oder gespeichert.
* Daten werden niemals von einem Mutterunternehmen, einer Tochtergesellschaft oder einer ähnlich benannten Einheit übernommen.

## Lieferkette

**Keine Module von Drittanbietern Go.**
Die Programme verwenden nur die Go-Standardbibliothek.
`go.mod` hat keine `require`-Linien, und `scripts/check-deps.sh` schlägt die
Build, wenn eine Abhängigkeit, `go.sum` oder `vendor/` erscheint, oder falls vorhanden
Nicht-Standard-Import verwendet wird.
Mit nichts zum Download gibt es nichts zu halluzinieren, typosquat oder
Manipulation.

**PGP und seine Ersatzstoffe. **
Authentizität wird mit den Mechanismen jedes Ökosystems tatsächlich überprüft
Stützen.
PGP wird nicht verwendet, weil die beteiligten Ökosysteme nicht mehr darauf angewiesen sind
(PyPI entfernt PGP Unterschrift Unterstützung im Jahr 2023).

| Komponente | Überprüfung |
| --- | --- |
| Go Toolchain in CI | Heruntergeladen von `actions/setup-go` von go.dev und überprüft mit dem veröffentlichten SHA-256 |
| Go Tools in CI (staticcheck, govulncheck, gosec) | genaue Versionen, verifiziert von `go run` gegen die Go-Checksum-Datenbank, ein öffentliches Transparenzprotokoll |
| GitHub Aktionen | Angeheftet auf Full Commit SHAs; Dependabot schlägt Updates nach einer 14-tägigen Abklingzeit vor |
| Python-Tools in CI (zizmor, Argos Translate) | jedes Paket mit SHA-256 (`--require-hashes`), nur Rädern (`--only-binary :all:`), keine Abhängigkeitsauflösung zum Zeitpunkt der Installation (`--no-deps`) |
| Argos Übersetzen Modelle | Index gelesen bei einem gepinnten Commit; SHA-256 jedes Modells wird in `.github/translate/models.sha256` gepinnt, bei der ersten Verwendung aufgezeichnet und danach durchgesetzt |
| Containerbasisbild | keine: Bilder sind gebaut `FROM scratch` |

**Releases.**
Release-Binärdateien werden von GitHub-Aktionen aus einem Tag mit
`-trimpath` und einem festen Zeitstempel, so dass sie reproduzierbar sind.
Jedes Release hat:

* Prüfsummen für `SHA256SUMS` (GNU-Format) und `SHA256` (BSD-Format);
* `SHA256SUMS.sigstore.json`, eine schlüssellose Sigstore Signatur des
  Prüfsummen, die an die Identität des Release-Workflows gebunden sind;
* SLSA Build Provenienzbescheinigungen für jedes Archiv und Binär
  (`gh attestation verify`);
* ein Containerbild auf ghcr.io, mit cosign signiert und bescheinigt.

[docs/install.md](install.md) zeigt, wie man jeden von ihnen überprüft.

## Pipeline für Sicherheitstests

Jede Push- und Pull-Anforderung läuft:

| Überprüfung | Werkzeug |
| --- | --- |
| Keine Abhängigkeiten, Formatierung, Tierarzt auf sechs Betriebssystemen | `scripts/check-deps.sh`, `gofmt`, `go vet` |
| Statische Analyse | Statikprüfung |
| Bekannte Sicherheitslücken im Code und in eingebauten Binärdateien für jedes Ziel-Betriebssystem | govulncheck |
| Security linting, gemeldet zum Code-Scannen | Gosec |
| Semantische Analyse von Go und der Workflows | CodeQL (`security-extended`, `security-and-quality`) |
| Fuzzing des Dokumentenfüllers, des Tabellenkalkulationslesers und des Namensmatchers | Go native fuzzing, 2 Minuten pro Ziel |
| Workflow-Sicherheit (Injection, Anmeldeinformationen, Berechtigungen, Pins) | Zizmor, pedantisch |
| Kundendateien oder Anmeldeinformationen aus Versehen | Historienscan |
| Tests mit dem Renndetektor unter Linux, macOS und Windows, mit dem ältesten und neuesten Go | `go test -race` |
| Tests auf FreeBSD, OpenBSD und NetBSD | Virtuelle Maschinen |
| Container-Image baut und weigert sich, ohne Terminal zu laufen | Docker, Podman |
| Neue Abhängigkeiten in Pull Requests | Überprüfung der Abhängigkeit |

Die Schwachstellen- und Analysejobs laufen ebenfalls wöchentlich, und OpenSSF
Scorecard überprüft die Praktiken des Repository.

Workflows beginnen ohne Berechtigungen und gewähren jedem Job nur das, was er
Bedürfnisse; Checkouts behalten keine Anmeldeinformationen.

## Container und Kubernetes

Das Bild enthält nur die beiden Binärdateien, ein CA-Bundle und die Zeitzone
Datenbank und läuft als UID 65532.
Die dokumentierten Befehle und Manifeste verwenden ein schreibgeschütztes Root-Dateisystem,
alle Fähigkeiten fallen lassen, Privileg Eskalation verbieten, montieren die Arbeit
Ordner Read-only für numb-form, mount no service account token und legen Sie eine
Frist.

## Dieses Repository ist öffentlich

Legen Sie niemals Kundendokumente, Tabellenkalkulationen, Protokolle oder ausgeführte Ordner fest.
`.gitignore` schließt sie aus und CI lehnt sie ab.
