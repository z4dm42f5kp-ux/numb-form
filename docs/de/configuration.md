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

Erstellen Sie eine kommentierte Pro-Benutzer-Datei mit jedem Standard und zeigen Sie die effektiven Einstellungen an.

**Unix-like** (macOS, Linux, BSD), im Terminal:

```
numb-form config init && numb-form config
```

**Windows** (10, 11), in PowerShell:

```
numb-form config init; numb-form config
```

## KEYS

Die Programme, die einen Schlüssel verwenden, sind in Klammern benannt: **form** (numb-form), **commit** (numb-commit), **date** (Numb-date), **update** (numb-form Update). `numb-form config` listet jeden Schlüssel mit seinem effektiven Wert und der Datei auf, die ihn festgelegt hat.

### Auffinden der Dateien

`overview_dir = Übersicht` [form, Datum]
: Muster des Ordners, der die Tabellenkalkulation enthält. Fall, Akzente und Interpunktion werden ignoriert, und Umlaute stimmen mit beiden Schreibweisen überein, so dass `Übersicht` mit `3. Übersichten` und `Uebersicht` übereinstimmt.

`header = Customer Name` [form, Datum]
: Überschrift der Spalte Firmenname. numb-date versucht auch gemeinsame Namen (`Customer`, `Company Name`, `Kunde`, `Cliente` und andere) und findet die *Report created* Spalte von selbst.

`id_header = Customer ID` [form, Datum]
: Überschrift der Spalte Kundennummer, die neben jedem Unternehmen angezeigt wird.

`template_name = Appendix A - Risk Assesment` [form, Datum]
: Template-Dateiname ohne Erweiterung; `Assessment` stimmt ebenfalls überein.

### In das Formular eingeschriebene Werte

`person =` [form]
: Name der Person, die das Formular bearbeitet; leer bedeutet einmal pro Durchlauf fragen.

`date_format = DD/MM/YYYY` [form]
: Bewertungsdatum; Token `DD`, `MM`, `YYYY`, `YY`. numb-date schreibt immer `DD-MM-YYYY`.

### Wie ein Run funktioniert

`mode = direct` | `review` [form, commit]
: **direct** (Standard): Ordner und Formulare werden im Arbeitsordner selbst erstellt. **review**: Der gesamte Arbeitsordner wird zuerst in Downloads kopiert, alle Arbeiten finden in der Kopie statt und `numb-commit` kopiert die akzeptierten Ordner zurück.

`fuzzy_threshold = 0.85` [form, Datum]
: Wie ähnlich ein Ordnername sein muss, muss als Übereinstimmung angeboten werden (0,5-1).

`unicode = basic` | `full` [form, date]
: Wie Firmen- und Ordnernamen verglichen werden. **basic** (Standard): Fall, Abstand, Interpunktion und die Akzente west- und mitteleuropäischer Sprachen (`é`, `ç`, `ł`, `ü` als `u` oder `ue`) werden ignoriert. **full**: auch die Akzente und Sonderformen jedes anderen Alphabets, zum Beispiel `ș`, `ğ`, `ǎ`, Griechisch `ά`, Ligaturen wie `ﬁ` und Buchstaben voller Breite wie `Ａ`. Die zusätzliche Tabelle wird nur gelesen, wenn **full** gesetzt ist, so dass der Standard keine Arbeit hinzufügt.

`timeout = none` [form, commit, date]
: Stoppen Sie sicher nach diesem langen, zum Beispiel `30m` oder `2h`; **none** (Standard) bedeutet keine zeitliche Begrenzung.

`max_write_mb = 100` [form]
: Die meisten Megabytes von Dokumenten, die ein Lauf schreiben kann (1-100).

`downloads =` [form, commit]
: Überprüfungsmodus: Ordner für die Kopie; leer bedeutet Ihren Downloads-Ordner.

`log_dir =` [form, Datum]
: Ordner für Protokolle; leer bedeutet den Statusordner pro Benutzer (`%LOCALAPPDATA%\numb-form\runs`, `~/Library/Logs/numb-form` oder `~/.local/state/numb-form/runs`). Review-Modus-Logs bleiben bei der Kopie in Downloads.

`commit_from =` [commit]
: Die Arbeitskopie, die numb-commit vergleicht, zum Beispiel `C:\Users\you\Downloads\14. GWG`; siehe [numb-commit.md](numb-commit.md). Leer: numb-commit fragt.

`commit_to =` [commit]
: Der reale Arbeitsordner, mit dem numb-commit die Arbeitskopie vergleicht, zum Beispiel `F:\14. GWG`. Leer: numb-commit fragt.

`online = yes` | `no` [update]
: Lassen Sie `numb-form update` GitHub kontaktieren. Nichts anderes geht online.

### Formate

`soffice =` [form]
: Pfad zu LibreOffice oder Apache OpenOffice; leer bedeutet Suche.

`pages_converter = auto` | `pages-app` | `libreoffice`, `numbers_converter = auto` | `numbers-app` | `libreoffice` [form]
: Wie `.pages` und `.numbers` Dateien konvertiert werden. **auto** verwendet Apples App auf macOS, wenn installiert, sonst LibreOffice.

`pages_write = no` [form]
: EXPERIMENTAL, nur macOS: Speichern Sie ausgefüllte Formulare als `.pages` bis Apple Pages.

`legacy_formats = no` [form]
: SYSTEMDATEI NUR: `.doc`-Vorlagen und `.xls`-Tabellen zulassen.

`legacy_output = docx` | `same` [form]
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
