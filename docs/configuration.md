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

`confirm = each` | `none`
: **each** (default): every value and every document is shown and must
  be accepted.
  **none**: runs without questions; each document must still read back
  exactly as written.
  Near folder-name matches are never decided automatically: the company
  is skipped and reported as *needs review*.
  Requires `person`.

`online = yes` | `no`
: Look up registers and websites (default yes).

`timeout = 25m`
: Stop safely after this long; at most `25m`.

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

### Sources

Sources are `companies-house`, `vies`, `brasil-cnpj`, `official`
(government gazettes and registers found by search), `gleif`, `website`,
`directory` (company-data websites that republish register data, such as
northdata, Moneyhouse, Racius or OpenCorporates), `search` (web search
snippets; search also finds the website), `osm` (OpenStreetMap, for an
address) and `wikidata`.
Lists are separated by spaces or commas; their order is the order of
preference.

`address_sources = companies-house vies brasil-cnpj official gleif website directory search osm wikidata`
: Sources for *Headquarters / Address*.

`address_fallback = ask` | `blank`
: When no address is found. With `confirm = none`, `ask` acts as
  `blank`.

`reps_sources = companies-house vies brasil-cnpj official gleif website directory search osm wikidata`
: Sources for *Authorized Representatives*.

`reps_min_trust = any` | `website` | `register`
: Least trusted source accepted for representatives: `register` accepts
  only government or regulatory registers.

`reps_fallback = no-reps-text` | `blank`
: When nobody is found: the exact no-representatives text, or blank.

`caution = auto` | `always` | `never`
: The caution text.
  **auto**: suggested for values not confirmed by a register (asked when
  `confirm = each`).

`website = auto` | `ask` | `never`
: The company website read for its legal notice.
  **auto** (default): found automatically and used only if it names the
  company.
  **ask**: you type the address for every company (Enter finds it
  automatically).
  **never**: no website is read.
  With `confirm = none`, `ask` acts as `auto`.
  The older value `wikidata` means `auto`.

`search = duckduckgo mojeek yahoo brave` | `none`
: Web search engines used to find websites, government gazette entries
  and, as a last resort, representatives in result snippets, in order of
  preference: when one refuses, it rests for a minute and the next one is
  asked.
  Only the company name is sent, with words such as `impressum` or
  `site:boe.es`.
  **none** turns searching off; registers and guessed addresses are still
  used.

`browser = auto` | `none` | *path*
: The browser started invisibly (headless, no window) for websites built
  with JavaScript and for searching when every engine refuses.
  **auto**: Google Chrome; without it, Chromium or Microsoft Edge.
  A *path* names another Chromium-based browser.
  Every start uses a new, empty profile that is deleted right after.

`guess_domains = yes` | `no`
: Also try addresses built from the name, such as `foobar.de` or
  `foobar.com.br`, chosen by the countries of the legal form.

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
  Runs that use them are always offline.

`legacy_output = docx` | `same`
: Filled forms from a `.doc` template are `.docx`, or `.doc` again.

## EXAMPLES

Unattended runs that accept only register data:

```
confirm = none
person = Max Muster
address_sources = companies-house gleif
reps_sources = companies-house gleif
reps_min_trust = register
caution = auto
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
