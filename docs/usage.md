# NUMB-FORM(1)

## NAME

**numb-form**, **numb-commit** — fill risk assessment forms (with
**numb-date**)

## SYNOPSIS

```
numb-form [-f conf] [-y] [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form [-f conf] check-template template [sample]
numb-form [-f conf] config [init | set key value]
numb-form [-f conf] update [stable | rolling | v1.2.3]
numb-form rollback
numb-commit [-f conf] [-run dir] [-downloads dir]
```

numb-date has its own page: [numb-date.md](numb-date.md).

## DESCRIPTION

**numb-form** is started inside the working folder, the folder that
holds the template and the overview folder (for example `GWG`).
It may also be started one level above it: the working folder is then
the single folder found there, or the one containing an overview folder,
or the user is asked.
`-root` names the working folder explicitly.

Office staff: see the [step-by-step guide](guide.md).

The working folder must contain:

* an overview folder whose name contains `overview_dir` (default
  `Übersicht`, matching `3. Übersichten`), holding exactly one
  spreadsheet;
* the template, named `template_name` (default
  `Appendix A - Risk Assesment`), at its top level.

Supported formats are listed in [formats.md](formats.md).
Settings are described in [configuration.md](configuration.md).

A run proceeds as follows:

1. The template is checked with a trial fill of every field.
   The run stops if section A, or any label other than *Commercial
   Registry Number*, cannot be found or is not empty.
2. The name of the person processing the form is taken from the
   configuration, or typed once.
3. With `mode = review`, the whole working folder is copied to
   `Downloads/numb-form-runs/<date>_<time>_<id>/` and everything after
   this step happens in the copy.
   With `mode = direct` (the default), the working folder itself is used
   and the log goes to the per-user state folder.
4. Company names are read from the column headed `Customer Name`, from
   top to bottom.
   All other columns are ignored, including `Report created`.
   numb-form never writes the spreadsheet; its SHA-256 checksum is checked
   before and after the run.
5. For each company, the folders in the working folder and all of its
   subfolders are searched.
   Names are compared ignoring case, accents and umlauts (ü = u = ue),
   punctuation and spaces.
   A match means the form already exists, and the company is skipped.
   A near match is shown to the user to confirm.
   Folders with a different legal form never match:
   `Foo GmbH` and `Foo S.A.` are different companies.
6. For a new company, the user types the headquarters address and the
   authorized representatives.
   Everything else is filled without questions.
7. A folder named after the company is created.
   The template is copied into it as
   `<template name> - <Company Name>.<ext>`, and section A is filled.
8. The document is read back from disk and shown.
   The user accepts it, edits a field or rejects it; a rejected folder is
   deleted.

Only the lines of section A, *Information About the Contracting Party*,
are filled.
Section B and everything outside section A are left byte for byte
unchanged, and no formatting is added.

| Field | Value |
| --- | --- |
| Name / Company Name of the Contracting Party | the literal spreadsheet cell |
| Legal form | found in the name: at the end, before a branch or country note (`Foo GmbH (Wien)`), before a city or country (`Foo SAS, Bogotá, Colombia`, `Foo S.A.C.I. Paraguay`), or first (`Ets Foo`); accents, dots and spaces may differ (`SOCIEDAD ANONIMA`, `S A`); written as in the name with conventional capitals (`GMBH` gives `GmbH`); asked only when the name has none, where Enter writes the no-legal-form text below and `-` leaves it blank |
| Headquarters / Address | typed by the user; Enter writes the no-address text below, `-` leaves it blank |
| Commercial Registry Number | always `No commercial registry number was found online consult internal sources for further information`; it can be replaced with the number with *edit* |
| Authorized Representatives | typed by the user as `Full Name (Role 1, Role 2), Full Name (Role)`, roles in English; Enter writes the no-representatives text below, `-` leaves it blank |
| Date of the evaluation | today, `date_format` (default `DD/MM/YYYY`) |
| Person Processing the Form | from the configuration or typed at the start |

numb-form makes no online lookups for companies.
Only `numb-form update` contacts GitHub.

When a value is left empty (Enter alone), the field reads:

| Field | Text |
| --- | --- |
| Legal form | `No legal form was found online due to little to no information available` |
| Headquarters / Address | `No address was found online due to little to no information available` |
| Authorized Representatives | `No representatives were found online, please consult other internal sources for more accurate information` |

**numb-commit** applies to review runs only.
It copies the accepted folders of a run back into the real working
folder.
It only adds new folders, never overwrites or deletes anything, and
refuses folders that appeared in the real tree in the meantime.
Each folder is copied to a temporary name, verified with SHA-256 and then
renamed.
Documents edited during review are listed before the user confirms.
Afterwards it offers to delete the review copy; the logs are kept.

**numb-date** fills the *Report created* column of the spreadsheet from
the forms: see [numb-date.md](numb-date.md).

## COMMANDS

`check-template template [sample]`
: Check that a template can be filled, and optionally write a filled
  sample to *sample*. The template is only read.

`config`
: Show the configuration files and effective settings.

`config init`
: Write a commented per-user configuration file with every default.

`config set key value`
: Save one setting in the per-user configuration file, for example
  `numb-form config set person "Erika Mustermann"`.
  The value is checked before it is saved.

`rollback`
: Switch back to the version that the last `update` replaced.
  The version rolled back from is kept in its place, so a second
  `rollback` returns to it.
  Without a kept version, install a release by number instead:
  `numb-form update v0.1.1`.

`man` [*topic*] [`en` | `es` | `de`]
: Show this manual in the terminal, like less(1): keys act at once,
  without Enter.
  Space or PageDown: next page; b or PageUp: previous page; Enter, j or
  Down: next line; k or Up: previous line; g or Home: start; G or End:
  end; /text and Enter: search, n and N: next and previous match; h:
  help; q: quit.
  On macOS, Linux and the BSDs the installed `less` is used when there
  is one.
  Without a topic the whole manual is shown, starting with the guide.
  Topics: `guide`, `usage`, `numb-date`, `configuration`, `install`,
  `formats`, `containers`, `security`; `numb-form man topics` lists
  them.
  The language follows `LANG`, or is given.
  `numb-form man -install` makes `man numb-form` work: on Linux, macOS
  and the BSDs it writes the manual page to `~/.local/share/man/man1`
  (refreshed by `numb-form update`); on Windows it writes `man.cmd` next
  to `numb-form.exe`, for Command Prompt.
  Windows PowerShell keeps `man` for its own help: there, type
  `numb-form man`.

`update` [`stable` | `rolling` | *v1.2.3*]
: Replace the installed `numb-form` and `numb-commit`:
  **stable** with the latest stable release, **rolling** with the
  current rolling build of `main` (rebuilt after every change; `edge` is
  accepted as an older name), or *v1.2.3* with exactly that release.
  The version replaced is kept next to the programs as
  `numb-form.previous` and `numb-commit.previous` for `rollback`.
  It works as the install commands do: the archive and `SHA256SUMS` are downloaded over HTTPS from
  GitHub only, the checksum must match, and the new `numb-form` must start
  before anything is replaced.
  Without an argument the channel of the running build is used.
  Settings are kept.
  Refused with `-offline`, with `online = no`, and for builds from source.
  On Windows the old program is renamed to `numb-form.old.exe` and removed
  at the next start.

## OPTIONS

`-f file`
: Read *file* after the system and per-user configuration files.

`-y`
: Take a folder with a similar name as the
  company's folder instead of asking "Is the existing folder … this
  company's folder?".
  The folder is named on screen and in the log, and the summary says how
  many folders were taken this way.
  Every document is still shown and must be accepted.
  `-Y` is the same.

`-offline`
: With `update`: refuse to contact GitHub.

`-root dir`
: Use *dir* as the working folder instead of detecting it.

`-downloads dir`
: Review mode: put the copy in *dir*.

`-timeout duration`
: Stop safely after *duration*, for example `30m`; overrides a longer
  `timeout` in the configuration. Without it there is no time limit.

`-run dir`
: (numb-commit) Commit this run folder instead of the newest uncommitted
  one.

`-allow-non-tty`
: Read answers from a pipe. For automated tests only.

`-version`
: Print the version.

## ENVIRONMENT

`NUMBFORM_DOWNLOADS`
: Downloads folder to use, for example a mounted volume in a container.

`HTTPS_PROXY`, `NO_PROXY`
: Proxy settings for `numb-form update`.

`TZ`
: Time zone of the evaluation date.

## FILES

`<log folder>/<run>/numb-form.log`
: Log of everything the run did.

`<log folder>/<run>/manifest.json`
: Machine-readable record of the run.

`Downloads/numb-form-runs/<run>/<working folder>/`
: Review mode: the copy.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Review mode: log of the commit.

The files of numb-date are listed on its page.

No cache or temporary files are left behind.
Documents are written through
a temporary file that is renamed or removed, and converters work in a
private temporary folder that is deleted.

## EXIT STATUS

0 on success, 1 on error, 2 on a usage or configuration error, and 3
when stopped by the time limit, by Ctrl+C or by the user.

## EXAMPLES

Commands are shown for both kinds of system.
Windows PowerShell 5.1 has no `&&`: commands are separated with `;`, and
`if ($?) { ... }` runs the next one only when the first succeeded.

Check the real template and write a sample, without touching the
template (all systems):

```
numb-form check-template "Appendix A - Risk Assesment.docx" sample.docx
```

Fill the forms directly in the working folder, then the report dates.

**Unix-like** (macOS, Linux, BSD), in Terminal:

```
cd /path/to/GWG && numb-form && numb-date
```

**Windows** (10, 11), in PowerShell:

```
cd "H:\GWG"; numb-form; if ($?) { numb-date }
```

Review first, then commit (numb-commit waits for confirmation), where
`review.conf` contains `mode = review`.

**Unix-like** (macOS, Linux, BSD), in Terminal:

```
cd /path/to/GWG && numb-form -f review.conf && numb-commit
```

**Windows** (10, 11), in PowerShell:

```
cd "H:\GWG"; numb-form -f review.conf; if ($?) { numb-commit }
```

## SEE ALSO

[numb-date.md](numb-date.md),
[configuration.md](configuration.md), [formats.md](formats.md),
[install.md](install.md), [containers.md](containers.md),
[../SECURITY.md](../SECURITY.md), [../AGENTS.md](../AGENTS.md)
