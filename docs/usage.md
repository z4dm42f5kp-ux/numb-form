# NUMB-FORM(1)

## NAME

**numb-form**, **numb-commit**, **numb-date** — fill risk assessment forms

## SYNOPSIS

```
numb-form [-f conf] [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form [-f conf] check-template template [sample]
numb-form [-f conf] config [init | set key value]
numb-form [-f conf] update [stable | rolling | v1.2.3]
numb-form rollback
numb-commit [-f conf] [-run dir] [-downloads dir]
numb-date [-f conf] [-root dir] [-timeout duration]
```

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
| Legal form | found in the name: at the end, before a branch or country note (`Foo GmbH (Wien)`), before a city or country (`Foo SAS, Bogotá, Colombia`, `Foo S.A.C.I. Paraguay`), or first (`Ets Foo`); accents, dots and spaces may differ (`SOCIEDAD ANONIMA`, `S A`); written as in the name with conventional capitals (`GMBH` gives `GmbH`); asked only when the name has none |
| Headquarters / Address | typed by the user; Enter leaves it blank |
| Commercial Registry Number | always `No commercial registry number was found online consult internal sources for further information`; it can be replaced with the number with *edit* |
| Authorized Representatives | typed by the user as `Full Name (Role 1, Role 2), Full Name (Role)`, roles in English; Enter writes the no-representatives text below, `-` leaves it blank |
| Date of the evaluation | today, `date_format` (default `DD/MM/YYYY`) |
| Person Processing the Form | from the configuration or typed at the start |

numb-form makes no online lookups for companies.
Only `numb-form update` contacts GitHub.

When the representatives are left empty, the field reads:

> No representatives were found online, please consult other internal sources for more accurate information

**numb-commit** applies to review runs only.
It copies the accepted folders of a run back into the real working
folder.
It only adds new folders, never overwrites or deletes anything, and
refuses folders that appeared in the real tree in the meantime.
Each folder is copied to a temporary name, verified with SHA-256 and then
renamed.
Documents edited during review are listed before the user confirms.
Afterwards it offers to delete the review copy; the logs are kept.

**numb-date** fills the report-date column of the spreadsheet.
It is started like numb-form, in or above the working folder, and finds
the same overview spreadsheet, which must be `.xlsx` or `.xlsm`.

1. The company column is found by `header`, then by common names
   (`Customer Name`, `Customer`, `Company Name`, `Kunde`, `Cliente` and
   others, ignoring case); the date column in the same header row by
   `Report created AB`, `Report created`, `Report date`, `Created` and
   similar names, or any heading that mentions a report and its creation
   or date.
2. Rows whose date cell contains a digit are skipped: they already have
   a date.
3. For the other rows, the company's folder is found as numb-form finds
   it (near matches are asked).
   The `.docx`, `.odt`, `.fodt` or `.rtf` forms in the folder (and one
   level below) are read, the risk assessment form first: the *Date of
   the evaluation* of section A, or, in forms laid out differently, the
   date written after a date label.
   Dates such as `16/10/2026`, `16.10.2026`, `2026-10-16` and
   `16 October 2026` are understood.
   When no date is found, the user types it or skips the company.
4. The dates are listed, written as `DD-MM-YYYY` after the user agrees,
   read back from disk and shown; the user accepts them or undoes the
   write.

Only the listed cells change; see [../SECURITY.md](../SECURITY.md) for
the checks.
Forms are only read.
The file must not be open in Excel or LibreOffice.
The log is `numb-date.log` in the per-user state folder (or `log_dir`).

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
: Show this manual in the terminal, one screen at a time (Enter: next
  screen; q and Enter: quit).
  Without a topic the whole manual is shown, starting with the guide.
  Topics: `guide`, `usage`, `configuration`, `install`, `formats`,
  `containers`, `security`; `numb-form man topics` lists them.
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

`<log folder>/<run>_date/numb-date.log`
: Log of a numb-date run, with the SHA-256 of the spreadsheet before
  and after.

`Downloads/numb-form-runs/<run>/<working folder>/`
: Review mode: the copy.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Review mode: log of the commit.

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

[configuration.md](configuration.md), [formats.md](formats.md),
[install.md](install.md), [containers.md](containers.md),
[../SECURITY.md](../SECURITY.md), [../AGENTS.md](../AGENTS.md)
