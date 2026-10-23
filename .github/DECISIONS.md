# Decisions for open issues

Recorded 2026-10-16 from the maintainer's answers.
Status: implemented on branch `claude/issue-decisions`; the two GitHub
settings under #6 remain for the maintainer.

## #6 Tests on Dependabot pull requests

* The Dependabot PRs (#1-#5) were opened on a9a3b89, before the CI fixes.
  Comment `@dependabot rebase` on each so they run against current `main`.
* `dependency-review` fails because the Dependency graph is off:
  Settings > Code security > Dependency graph > Enable.
* Translation dependencies were never tested on PRs.
  Fixed in this branch: `translate.yml` now runs on pull requests that touch
  `.github/translate/**` (translate only, no commit).
* Then require the checks before merging: Settings > Branches > `main` >
  Require status checks.

## #11 Downloads copy

* Default `mode = direct`: each confirmed document is written straight into
  the real folder. The spreadsheet stays read-only; the log is kept.
* `mode = review` (opt-in): today's flow, Downloads copy + `numb-commit`.
* Docs: describe both; direct first.

## #9 Configuration file

* Per-user location: `%APPDATA%\numb-form\numb-form.conf`,
  `~/.config/numb-form/numb-form.conf`,
  `~/Library/Application Support/numb-form/numb-form.conf`; `-f file` overrides.
* OpenBSD-style `key = value`, comments with `#`, defaults equal to the
  current documented behaviour.
* Modular keys, at least: `mode`, `overview_dir` (default pattern
  "Übersicht"), `template_name`, `timeout`, `confirm` (each | none),
  per-field source lists and fallbacks, e.g.
  `address_sources = companies-house gleif website wikidata`,
  `address_fallback = blank`, `reps_min_trust = register | website | any`,
  `reps_fallback = no-reps-text`, `caution = auto | always | never`,
  `online = yes | no`.

## #7 Agnostic design, formats, smaller files

* No platform is first-class; docs list all systems equally.
* Split files to fewer lines each (one format per package).
* Templates: .docx, .odt, .fodt, .rtf; .pages read-only by default
  (filled output as .docx). Native .pages writing only behind an
  experimental conf flag, with an explicit warning; tested in CI, not
  shipped as default. Default `pages_write = no` (read .pages) when not on
  macOS.
* Spreadsheets: .xlsx, .xlsm (read only, no macros), .ods, .fods, .csv,
  .tsv, .numbers (read).
* Legacy .doc / .xls: behind `legacy_formats = yes` in the conf file, off by
  default, enabling requires administrator rights (system-wide conf), always
  offline, fuzzed in the security workflow.
* Docs: LibreOffice, OpenOffice, Proton Docs/Sheets (export .docx/.xlsx),
  Apple Pages/Numbers. For BSD, LibreOffice runs natively; add Proton Docs
  (web) as the no-install option.

## #10 Installation without administrator rights

* Every platform gets a user-only path: release binary into the user's own
  folder, user PATH change only; Go from the official tarball into the home
  folder for source builds; Windows without winget admin (user scope, or
  zip only).

## #8 Internal testing only on GitHub

* Applies to development, not users: AGENTS.md tells agents not to run
  tests locally; push a branch and read CI.
* Remove `go test` from user install commands.
* Fast PR gate (lint, unit tests on 3 OSes, govulncheck); BSD VMs, long
  fuzzing and containers on `main` and nightly; skip Go jobs for
  docs-only changes.

## Rolling builds (release tag edge)

Releases are created by hand in the GitHub web page, which is too slow
while the tool changes daily.
`edge.yml` rebuilds the archives after every push to `main` (after the
tests and govulncheck pass) and replaces the files of the `edge`
pre-release, signed and attested like a stable release.
It is never marked latest, so `releases/latest` keeps meaning stable.
Archives are uploaded before the checksums, so a download during an
update fails the checksum check rather than installing a mismatched file.

## More search engines and an invisible browser (#13)

DuckDuckGo refuses programs after a few searches from busy networks.
Mojeek, Yahoo and Brave Search are asked in turn, each resting a minute
after a refusal.
Go cannot run JavaScript, so websites built by scripts and searches that
every engine refuses are loaded in the locally installed Google Chrome
(else Chromium or Edge), started headless with `--dump-dom` and a
throwaway profile. No browser library or protocol client is needed, so
the program stays standard-library only.
`numb-form update stable` and `update rolling` name the two channels;
`edge` stays accepted and is still the release tag of the rolling build.

## No online lookups

Automatic lookups never found addresses or representatives for the
small Latin American, African and Gulf companies the team handles, and
cost more time than they saved.
They are removed with everything that served them: web search, the
headless browser, registers, API keys, `numb-form keys`, the live test,
`confirm = none` and the source settings.
numb-form now fills the name, legal form, registry sentence, date and
person by itself and asks for the address and the representatives.
Old configuration keys are ignored with a note.
`numb-form update`, `rollback` and `numb-commit` are unchanged.

## numb-date writes the report dates

numb-date is the only program allowed to write the overview
spreadsheet, and only digit-free cells of the *Report created* column.
The date is read from each company's form (*Date of the evaluation*),
which also covers forms filled by hand, so numb-form writes no
`for-numb-date.txt` list: a second source of the same date could only
disagree with the form.
Cells are spliced into the sheet XML like values in templates; every
write is checked with SHA-256 and a cell-by-cell comparison, and the
user accepts the result read back from disk or undoes it.
Only `.xlsx` and `.xlsm` are written.

## No time limit by default; legal forms in real names

The 25-minute limit stopped long runs that a person was working through
by hand, so `timeout` is now `none` by default; it can still be set.
Legal forms were missed in names written with their city or country
(`Foo SAS, Bogotá, Colombia`), without accents, with a form first
(`Ets Foo`), or with forms that were not listed (free-zone, Gulf,
Colombian BIC and E.S.P., Argentine and Paraguayan combinations).
Trailing places are now removed before the form is looked for, accents
are ignored when matching, and the list was extended.

## -y and a less-like manual

`-y` (numb-form, numb-date) takes a similar existing folder as the
company's folder without asking; documents are still confirmed one by
one, and the choice is printed and logged.
`numb-form man` reads single keys like less(1): the console is put into
raw mode (SetConsoleMode on Windows 10 and later, stty elsewhere) and
restored afterwards; the installed less is still preferred on Unix.

## The rolling build is never behind stable

Translations are committed by the translate workflow with the workflow
token, and such pushes start no other workflow, so the rolling build
skipped them; a stable release tagged on the translation commit was then
newer than rolling.
The rolling build now always builds the newest `main`, and runs after a
release is published and after translations are committed.

## The manual viewer wraps long lines itself

Code blocks keep their long commands on one line; a terminal wraps such
a line over several rows, so a page overflowed the window and the view
jumped.
The viewer now cuts every line to the window width (continuing on the
next row) and redraws by overwriting rows instead of clearing the
screen, which also stops the flicker.

## Releases wait for approval

Merging a pull request titled `vX.Y.Z: ...` proposes that release once ci
passes on main; it is tagged and published only after the owner approves
the `release` environment (one tap in the mail or on the run page).
A tag created with the workflow token starts no workflow, so the approval
starts `release.yml` on the tag explicitly; the signing identity stays
`release.yml@refs/tags/v...`, so the verification commands are unchanged.

## Daily update checks

Dependabot runs daily (stable versions after 14 days, security updates
at once). Go and the pinned Go tools have no Dependabot support here:
`updates.yml` checks them daily, proposes newer tools as a pull request,
and, when a newer Go fixes a vulnerability in the newest release's
binaries, proposes a patch release of the same code built with that Go.

## The AI prompt is a separate download

numb-form makes no online lookups, so the address and the
representatives come from the person.
The team finds them with an AI assistant and a fixed prompt.
The prompt is kept in the repository (`docs/ai-prompt.txt`) and shipped
with every release as its own file, checksummed in `SHA256SUMS`, so
that everyone uses the same version.
It is not built into the programs.

## The AI prompt leaves nothing to interpret

The prompt now defines the entity, ranks the sources (official registers
first, aggregators and news only as leads), names the registers that
work for little-known companies in Germany, Austria, Switzerland, Spain
and Latin America, and fixes the format of every field.
It matches numb-form: the legal form in its conventional casing
(`GmbH`, `S.A.`) rather than in capitals, the representatives as
`Full Name (Role)` with the roles in English, and numb-form's standard
texts when nothing is found.
The assistant writes no em dashes, en dashes or semicolons.
"innacurate" in the caution became "inaccurate".

## Documentation is written down like the BSDs

The manual is part of the program, as on OpenBSD and FreeBSD: a change
to what a person sees updates its manual page in the same pull request,
and `CHANGES.md` (FreeBSD's RELNOTES and UPDATING) says what changed and
what users must do.
Tests check what a program can: every flag and configuration key is
documented and every link resolves.

## Security and problem reports are written down like Go and OpenBSD

Problem reports use forms that ask for what OpenBSD asks for: exact
steps, exact output and the version, without customer data.
The security policy follows the Go project's (private reports, tracks,
7 days to acknowledge, 90 days to fix or disclose), and
`.github/SECURE-CODING.md` turns SECURITY.md's promises into rules,
a review checklist and an audit every three months.

## Standard texts for a missing address or legal form

Enter on an empty address, or on the legal form of a name that has
none, now writes a standard text like the one for representatives:
"No address was found online due to little to no information available"
and "No legal form was found online due to little to no information
available". `-` still leaves the field blank.

## numb-commit from a working copy, with a request file

The real GWG folder is on a network drive, and people work on a copy in
Downloads.
numb-commit (with `commit_to` or `-to`) compares the copy with the real
folder and writes every change to `numb-commit-request.txt` in the
copy: new folders and files, and the spreadsheet when only report dates
changed.
The template, logs and lock files are never listed; nothing is deleted.
The review is a subcommand (`numb-commit review`) rather than a fourth
program, so installs and updates stay as they are.
The first release only previews: it writes and reviews the request and
copies nothing, so the comparison can be checked on real folders before
numb-commit writes to a network drive.
Review runs (`mode = review`) work as before.

## numb-commit stages and copies after a yes

numb-commit stages exactly the files to copy in Downloads, read-only,
and touches the target only after the person agrees.
A new folder is built under a temporary name and renamed when complete;
no existing file is overwritten, and the replaced spreadsheet is kept in
the log folder.
numb-form marks nothing in its documents (nothing outside section A may
change); numb-commit recognises its forms by filling the template with
the values a form holds and comparing.

## numb-commit skips known companies and reviews in rounds

A new folder of a company the target already has, typically under its
customer ID (`D10629 Foo S.A.`), is not copied.
The person chooses a manual review or an automatic one that corrects
what it can and reviews again, at most five rounds.

## numb-typo corrects places in company names

Customer names often end in a misspelt or cut-off city or country
("Ciudad de Me").
numb-typo corrects only those parts after the first comma, from a list
of countries and cities with their accents, never the name itself or
its legal form.
A cut-off part is completed only at the end of a cell; a typo is
corrected only when one place is clearly closest (no corrections below
five letters, extra letters only from seven, two errors only from nine;
the first letter must match), and several candidates are asked.
Many real mid-size cities are listed so that they are never "corrected".
Names of companies that already have a folder are left alone, since
renaming them would part them from their forms.
It writes through the same checked path as numb-date.

## Install commands are pinned

An install command that downloads `releases/latest` and checks it
against a `SHA256SUMS` from the same place proves only that the download
is intact: whoever can replace the files can replace the checksums.
The commands now name one release and carry its checksums, so a
replaced file fails the check.
They are written by `scripts/install` from templates and
`scripts/install/pins.txt`; a test fails when the documentation differs.
After each stable release, `install-docs.yml` checks the release's
Sigstore signature, pins it and opens a pull request; it also runs
daily.

## Linux is tested before minor releases

The Linux distributions the team uses are tested in containers with
their own Go and the documented commands, only for minor releases
(pull requests titled `vX.Y.0:`), because the check takes long.
Patch releases, `main` and the rolling build are not held up by it.
The images are pinned by digest rather than letting tests run on
whatever `latest` is.
