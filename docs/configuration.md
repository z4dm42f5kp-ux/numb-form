# NUMB-FORM.CONF(5)

## NAME

**numb-form.conf** — numb-form configuration file

## DESCRIPTION

numb-form works without a configuration file; the defaults below are the
documented behaviour.
Files are read in this order, each one overriding the previous:

1. The system file, written by an administrator:

   | System | Path |
   | --- | --- |
   | Linux | `/etc/numb-form.conf` |
   | FreeBSD, OpenBSD, NetBSD | `/usr/local/etc/numb-form.conf`, else `/etc/numb-form.conf` |
   | macOS | `/Library/Application Support/numb-form/numb-form.conf` |
   | Windows | `%ProgramFiles%\numb-form\numb-form.conf` |

   On Unix systems it is ignored unless it belongs to root and is not
   writable by group or others.
2. The per-user file:

   | System | Path |
   | --- | --- |
   | Linux, BSD | `$XDG_CONFIG_HOME/numb-form/numb-form.conf` (`~/.config/numb-form/`) |
   | macOS | `~/Library/Application Support/numb-form/numb-form.conf` |
   | Windows | `%APPDATA%\numb-form\numb-form.conf` |

3. A file given with `-f file`.

Each line is `key = value`.
Lines starting with `#` are comments, and ` #` starts a comment after a
value.
Values may be quoted with `"`.
Unknown keys and invalid values are errors, so typing mistakes are found
at once.
Keys of older versions (`confirm`, `address_sources`, `address_fallback`,
`reps_sources`, `reps_min_trust`, `reps_fallback`, `caution`, `website`,
`search`, `browser`, `guess_domains`) are ignored with a note.

Save a single setting without editing the file:

```
numb-form config set person "Erika Mustermann"
```

Create a commented per-user file with every default, and show the
effective settings:

```
numb-form config init && numb-form config
```

## KEYS

### Workflow

`mode = direct` | `review`
: **direct** (default): folders and forms are created in the working
  folder itself, after each confirmation.
  **review**: the whole working folder is first copied to Downloads, all
  work happens in the copy, and `numb-commit` copies the accepted
  folders back.

`online = yes` | `no`
: Let `numb-form update` contact GitHub (default yes).
  Companies are never looked up online.

`timeout = none`
: Stop safely after this long, for example `30m` or `2h`; **none**
  (default) means no time limit.

`max_write_mb = 100`
: Most megabytes of documents one run may write (1-100).

### Folders and files

`overview_dir = Übersicht`
: Pattern of the folder that holds the spreadsheet.
  Case, accents and punctuation are ignored, and umlauts match both
  spellings, so `Übersicht` matches `3. Übersichten` and `Uebersicht`.

`header = Customer Name`, `id_header = Customer ID`
: Spreadsheet column headers.

`template_name = Appendix A - Risk Assesment`
: Template file name without extension; `Assessment` also matches.

`downloads =`
: Review mode: folder for the copy; empty means your Downloads folder.

`log_dir =`
: Direct mode: folder for logs and manifests; empty means the per-user
  state folder (`%LOCALAPPDATA%\numb-form\runs`, `~/Library/Logs/numb-form`
  or `~/.local/state/numb-form/runs`).

### Values

`person =`
: Name of the person processing the form; empty means ask once per run.

`date_format = DD/MM/YYYY`
: Evaluation date; tokens `DD`, `MM`, `YYYY`, `YY`.

`fuzzy_threshold = 0.85`
: How similar a folder name must be to be offered as a match (0.5-1).

### Formats

`soffice =`
: Path to LibreOffice or Apache OpenOffice; empty means search.

`pages_converter = auto` | `pages-app` | `libreoffice`
`numbers_converter = auto` | `numbers-app` | `libreoffice`
: How `.pages` and `.numbers` files are converted.
  **auto** uses Apple's app on macOS when installed, else LibreOffice.

`pages_write = no`
: EXPERIMENTAL, macOS only: save filled forms as `.pages` through Apple
  Pages.

`legacy_formats = no`
: SYSTEM FILE ONLY: allow `.doc` templates and `.xls` spreadsheets.

`legacy_output = docx` | `same`
: Filled forms from a `.doc` template are `.docx`, or `.doc` again.

## EXAMPLES

Save your name so it is not asked at every run:

```
person = Max Muster
```

Your previous workflow, with a review copy in Downloads:

```
mode = review
```

A German spreadsheet:

```
overview_dir = Übersicht
header = Firmenname
id_header = Kundennummer
```

## SEE ALSO

[usage.md](usage.md), [formats.md](formats.md)
