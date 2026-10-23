# NUMB-FORM.CONF(5)

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../configuration.md). Maßgeblich ist die englische Fassung.

## NAME

**numb-form.conf** — numb-form configuration file

## DESCRIPTION

numb-form funktioniert ohne Konfigurationsdatei; die Standardwerte unten sind das dokumentierte Verhalten. Dateien werden in dieser Reihenfolge gelesen, jede überschreibt die vorherige:

1. Die Systemdatei, geschrieben von einem Administrator:

   | System | Wegstrecke |
   | --- | --- |
   | Linux | `/etc/numb-form.conf` |
   | FreeBSD, OpenBSD, NetBSD | `/usr/local/etc/numb-form.conf`, sonst `/etc/numb-form.conf` |
   | macOS | `/Library/Application Support/numb-form/numb-form.conf` |
   | Windows | `%ProgramFiles%\numb-form\numb-form.conf` |

   Auf Unix-Systemen wird es ignoriert, es sei denn, es gehört zu root und ist nicht von einer Gruppe oder anderen beschreibbar.
2. Die Per-User-Datei:

   | System | Wegstrecke |
   | --- | --- |
   | Linux, BSD | `$XDG_CONFIG_HOME/numb-form/numb-form.conf` (`~/.config/numb-form/`) |
   | macOS | `~/Library/Application Support/numb-form/numb-form.conf` |
   | Windows | `%APPDATA%\numb-form\numb-form.conf` |

3. Eine Datei mit `-f file`.

Jede Zeile ist `key = value`. Zeilen, die mit `#` beginnen, sind Kommentare, und ` #` beginnt einen Kommentar nach einem Wert. Werte können mit `"` angegeben werden. Unbekannte Schlüssel und ungültige Werte sind Fehler, so dass Tippfehler sofort gefunden werden. Schlüssel älterer Versionen (`confirm`, `address_sources`, `address_fallback`, `reps_sources`, `reps_min_trust`, `reps_fallback`, `caution`, `website`, `search`, `browser`, `guess_domains`) werden mit einer Notiz ignoriert.

Speichern Sie eine einzelne Einstellung, ohne die Datei zu bearbeiten:

```
numb-form config set person "Erika Mustermann"
```

Erstellen Sie eine kommentierte Pro-Benutzer-Datei mit jedem Standard und zeigen Sie die effektiven Einstellungen an:

```
numb-form config init && numb-form config
```

## KEYS

### Workflow

`mode = direct` | `review`
: **direct** (Standard): Ordner und Formulare werden nach jeder Bestätigung im Arbeitsordner selbst erstellt. **review**: Der gesamte Arbeitsordner wird zuerst in Downloads kopiert, alle Arbeiten finden in der Kopie statt und `numb-commit` kopiert die akzeptierten Ordner zurück.

`online = yes` | `no`
: Lassen Sie `numb-form update` GitHub kontaktieren (Standard ja). Unternehmen werden nie online nachgeschlagen.

`timeout = none`
: Stoppen Sie sicher nach diesem langen, zum Beispiel `30m` oder `2h`; **none** (Standard) bedeutet keine zeitliche Begrenzung.

`max_write_mb = 100`
: Die meisten Megabytes von Dokumenten, die ein Lauf schreiben kann (1-100).

### Ordner und Dateien

`overview_dir = Übersicht`
: Muster des Ordners, der die Tabellenkalkulation enthält. Fall, Akzente und Interpunktion werden ignoriert, und Umlaute stimmen mit beiden Schreibweisen überein, so dass `Übersicht` mit `3. Übersichten` und `Uebersicht` übereinstimmt.

`header = Customer Name`, `id_header = Customer ID`
: Tabellenüberschriften.

`template_name = Appendix A - Risk Assesment`
: Template-Dateiname ohne Erweiterung; `Assessment` stimmt ebenfalls überein.

`downloads =`
: Überprüfungsmodus: Ordner für die Kopie; leer bedeutet Ihren Downloads-Ordner.

`log_dir =`
: Direktmodus: Ordner für Protokolle und Manifeste; leer bedeutet den Statusordner pro Benutzer (`%LOCALAPPDATA%\numb-form\runs`, `~/Library/Logs/numb-form` oder `~/.local/state/numb-form/runs`).

### Werte

`person =`
: Name der Person, die das Formular bearbeitet; leer bedeutet einmal pro Durchlauf fragen.

`date_format = DD/MM/YYYY`
: Bewertungsdatum; Token `DD`, `MM`, `YYYY`, `YY`.

`fuzzy_threshold = 0.85`
: Wie ähnlich ein Ordnername sein muss, muss als Übereinstimmung angeboten werden (0,5-1).

### Formate

`soffice =`
: Pfad zu LibreOffice oder Apache OpenOffice; leer bedeutet Suche.

`pages_converter = auto` | `pages-app` | `libreoffice` `numbers_converter = auto` | `numbers-app` | `libreoffice`
: Wie `.pages` und `.numbers` Dateien konvertiert werden. **auto** verwendet Apples App auf macOS, wenn installiert, sonst LibreOffice.

`pages_write = no`
: EXPERIMENTAL, nur macOS: Speichern Sie ausgefüllte Formulare als `.pages` bis Apple Pages.

`legacy_formats = no`
: SYSTEMDATEI NUR: `.doc`-Vorlagen und `.xls`-Tabellen zulassen.

`legacy_output = docx` | `same`
: Ausgefüllte Formulare aus einer `.doc`-Vorlage sind wieder `.docx` oder `.doc`.

## EXAMPLES

Speichern Sie Ihren Namen, damit er nicht bei jedem Lauf gefragt wird:

```
person = Max Muster
```

Ihr früherer Workflow mit einer Überprüfungskopie in Downloads:

```
mode = review
```

Eine deutsche Tabelle:

```
overview_dir = Übersicht
header = Firmenname
id_header = Kundennummer
```

## SEE ALSO

[usage.md](usage.md), [formats.md](formats.md)
