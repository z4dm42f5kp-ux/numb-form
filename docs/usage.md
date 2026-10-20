# NUMB-FORM(1)

## NAME

**numb-form**, **numb-commit** — fill risk assessment forms

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
   The spreadsheet is never written to; its SHA-256 checksum is checked
   before and after the run.
5. For each company, the folders in the working folder and all of its
   subfolders are searched.
   Names are compared ignoring case, accents and umlauts (ü = u = ue),
   punctuation and spaces.
   A match means the form already exists, and the company is skipped.
   A near match is shown to the user to confirm (or, with
   `confirm = none`, reported as *needs review* and skipped).
   Folders with a different legal form never match:
   `Foo GmbH` and `Foo S.A.` are different companies.
6. For a new company, a folder named after the company is created.
   The template is copied into it as
   `<template name> - <Company Name>.<ext>`, and section A is filled.
7. The document is read back from disk and shown.
   The user accepts it, edits a field or rejects it; a rejected folder is
   deleted.
   With `confirm = none` the read-back must equal the values exactly.

Only the lines of section A, *Information About the Contracting Party*,
are filled.
Section B and everything outside section A are left byte for byte
unchanged, and no formatting is added.

| Field | Value |
| --- | --- |
| Name / Company Name of the Contracting Party | the literal spreadsheet cell |
| Legal form | found in the name, also before a branch or country note (`Foo GmbH (Wien)`), written as in the name with conventional capitals (`GMBH` gives `GmbH`); asked only when the name has none |
| Headquarters / Address | chosen from the candidates found |
| Commercial Registry Number | always `No commercial registry number was found online consult internal sources for further information`; it can be replaced with the number with *edit* |
| Authorized Representatives | `Full Name (Role 1, Role 2), Full Name (Role)`, roles in English |
| Date of the evaluation | today, `date_format` (default `DD/MM/YYYY`) |
| Person Processing the Form | from the configuration or typed at the start |

Everything is looked up in one pass per company, without questions.
Candidates come from, in order of authority:

1. government registers:
   UK Companies House (only with a free key);
   EU VIES, for an EU VAT number printed on the company's website
   (name and address; Germany confirms only that the number is valid);
   Receita Federal, for a Brazilian CNPJ printed on the website (address,
   administrators and directors), read through the free BrasilAPI mirror;
   government gazettes and registers found by searching their websites
   for the company: BORME (Spain), Portal da Justiça (Portugal),
   Boletín Oficial (Argentina), Sunbiz (Florida), Companies House pages
   (United Kingdom); which one depends on the countries the legal form
   points to;
2. the GLEIF LEI register (regulatory register);
3. the company's own website: its legal notice (Impressum, Aviso legal,
   Mentions légales, Legal notice) first, then about, contact and privacy
   pages;
4. community data, shown unverified with the caution text: company-data
   directories that republish register data (northdata, firmenwissen,
   Moneyhouse, Racius, Empresite, cnpj.biz, OpenCorporates, Dun &
   Bradstreet and similar; at most three pages per company), the
   snippets of web search results, OpenStreetMap (for an address) and
   Wikidata.

When representatives confirmed by an official source, or an address from
a register or the company's website, are still missing, numb-form runs
one more search for the representatives (with the local word, such as
*Geschäftsführer* or *representante legal*) and one for the address, and
reads the directory pages they find.
A person named by two different kinds of source (for example the
company's website and a directory) counts as verified.

The website is found automatically: from Wikidata, from a DuckDuckGo
search for the exact name, and from addresses such as `foobar.de` or
`foobar.com.mx` built from the name and the countries of its legal form.
A site is used only if one of its pages names the company with its legal
form.
On a group's legal notice that names several companies, only the part
about the company is read.

Representative labels are recognised in German, English, Spanish,
Portuguese, French, Italian, Dutch, Polish, Czech, Slovak and the Nordic
languages, anywhere in a line, in tables and in footers, and translated
to English roles.
Labels that name other people (supervisory board, data protection
officer, person responsible for content, contact, registered agent) and
people who left (*Ceses/Dimisiones*, *ehemalig*, *former*) are skipped.

Official numbers and register entries printed on the website (for
example `HRB 12345` or `FN 123456 a`) are shown before the form is
written; the number can be typed into *Commercial Registry Number* with
*edit*.

Searching uses keyless engines in turn: DuckDuckGo, then Mojeek, Yahoo
and Brave Search (`search` in the configuration).
Each is asked at most a few times per company, a few seconds apart.
An engine that refuses (a human check, a rate limit) rests for a minute
and the next one is asked.

When Google Chrome is installed (or, without it, Chromium or Microsoft
Edge), numb-form starts it invisibly, without a window (headless), for
two things: reading company websites whose content is built by
JavaScript or that turn programs away, at most three pages per site;
and searching when every engine refuses.
Each start uses a new, empty profile in the temporary folder that is
deleted right after; the user's own browser, history and sign-ins are
never used.
`browser = none` turns this off.

Which sources feed which field, the least trusted source accepted for
representatives and the fallbacks are set in the configuration.
People and addresses are never taken from a parent, subsidiary or
similarly named entity.
Unverified values get this exact text when `caution` says so:

> Caution: This field might differ from internal sources or be innacurate due to insufficient information online

When no representative is found, the field reads:

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

`keys`
: Explain the optional free API keys.

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
: Contact no online source.

`-root dir`
: Use *dir* as the working folder instead of detecting it.

`-downloads dir`
: Review mode: put the copy in *dir*.

`-timeout duration`
: Stop safely after *duration*; at most `25m`.

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

`NUMBFORM_COMPANIES_HOUSE_KEY`
: Optional free UK Companies House API key. Never logged or stored.

`HTTPS_PROXY`, `NO_PROXY`
: Proxy settings for online lookups.

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

No cache or temporary files are left behind.
Network responses are held in memory only, documents are written through
a temporary file that is renamed or removed, and converters work in a
private temporary folder that is deleted.

## EXIT STATUS

0 on success, 1 on error, 2 on a usage or configuration error, and 3
when stopped by the time limit, by Ctrl+C or by the user.

## EXAMPLES

Check the real template and write a sample, without touching the
template:

```
numb-form check-template "Appendix A - Risk Assesment.docx" sample.docx
```

Fill the forms directly in the working folder:

```
cd /path/to/GWG && numb-form
```

Review first, then commit (numb-commit waits for confirmation):

```
cd /path/to/GWG && numb-form -f review.conf && numb-commit
```

where `review.conf` contains `mode = review`.

Run without any network access:

```
cd /path/to/GWG && numb-form -offline
```

## SEE ALSO

[configuration.md](configuration.md), [formats.md](formats.md),
[install.md](install.md), [containers.md](containers.md),
[../SECURITY.md](../SECURITY.md), [../AGENTS.md](../AGENTS.md)
