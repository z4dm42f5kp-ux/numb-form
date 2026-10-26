# Changes

What changed in each release, newest first.
Every pull request that changes the programs adds its lines here, under
the version it will be released as
([.github/DOCUMENTATION.md](.github/DOCUMENTATION.md)).
**Action needed** lists what users must do after updating.
Releases up to v1.2.1 are described on the
[releases page](https://github.com/z4dm42f5kp-ux/numb-form/releases).

## v1.6.1

* numb-date: the check after writing lists each cell with the company
  and where its date came from (the form, or `typed`), for example
  `F3  16-10-2026  (Foo Ltd, typed)`.
  The message for an open spreadsheet now reads `… is open in Excel or
  LibreOffice; close it and start again`.
* Internal: the checked spreadsheet writer of numb-date (self-check,
  SHA-256, original kept until you accept, undo) is shared, so that a
  later program can write other cells the same way.

## v1.6.0

* numb-commit does not copy a new folder of a company the real folder
  already has, for example `Foo S.A.` when the real folder has
  `D10629 Foo S.A.`: it starts with the company's customer ID, has the
  company's name after a customer ID, or has the company's name in
  another spelling.
  The request lists it under *Left alone*, the tree marks the real
  folder with `=`, and the review asks you to copy other files from it
  by hand.
* numb-commit asks `Review [m]anually, or let numb-commit [r]eview and
  correct it until it is right?`.
  `r` stages changed files again and leaves out files that must not be
  copied, reviewing again until the review passes, at most five rounds.

## v1.5.3

* Internal: numb-commit's automatic review, which corrects what it can
  by itself (stages a file again when it changed, leaves out a file
  that must not be copied) and reviews again, at most five rounds, and
  stops at once on what needs a person.
  Nothing uses it yet. No change in behaviour.

## v1.5.2

* Internal: numb-commit can mark the new folder of a company that the
  real folder already has as not to be copied, and its review explains
  each such folder and asks for a look when it holds more than
  numb-form's form.
  Nothing marks folders yet. No change in behaviour.

## v1.5.1

* Internal: finding the folder that the real working folder already has
  for a company: one that starts with its customer ID (for example
  `D10629 Foo S.A.`), one with its name after a customer ID, or one with
  its name in another spelling.
  Nothing uses it yet. No change in behaviour.

## v1.5.0

* numb-commit copies a working copy into the real working folder.
  It stages the files to copy in Downloads, writes
  `numb-commit-request.txt`, reviews it, and asks
  `Copy this into F:\14. GWG now? [y/N]`; only `y` copies.
  New folders appear whole, every file is checked with SHA-256, no
  existing file is overwritten, and the replaced Excel list is kept in
  the log folder.
  The real folder is reviewed again afterwards.
* numb-commit removes its staged copy at the end, and staged copies
  that a killed numb-commit left in Downloads more than a day ago.

## v1.4.5

* numb-commit stops cleanly when its window is closed, as with Ctrl+C,
  instead of being killed in the middle of a step.
* Internal: removing staged copies that a killed numb-commit left in
  Downloads more than a day ago.
  Nothing uses it yet.

## v1.4.4

* Internal: numb-commit's replacement of the overview spreadsheet in the
  real folder: only when the real spreadsheet still has the checksum it
  had when the request was made, after keeping the old version in the
  log folder, through a checked temporary copy.
  Both versions of the spreadsheet are read with the 256 MB limit.
  Nothing uses it yet. No change in behaviour.
* numb-date's temporary copy beside the spreadsheet is now named
  `.numb-replace-…` instead of `.numb-date-tmp-…`, as numb-commit will
  use the same code.

## v1.4.3

* Internal: numb-commit's copy of new folders and files from the staged
  copy into the real folder.
  A new folder is built under a temporary name and renamed when
  complete, every file is checked with SHA-256, and nothing is
  overwritten.
  Nothing uses it yet. No change in behaviour.

## v1.4.2

* Internal: numb-commit's staged copy, a folder in Downloads that holds
  exactly the files to copy, each checked with SHA-256 and read-only,
  and that numb-commit removes again.
  Nothing uses it yet. No change in behaviour.

## v1.4.1

* numb-commit: `numb-commit-request.txt` starts with a UTF-8 byte order
  mark, so Windows PowerShell 5.1 (`Get-Content`) and Notepad show the
  folder tree's lines and umlauts correctly instead of `â”œâ”€â”€`.

## v1.4.0

* numb-commit compares a working copy with the real working folder:
  `numb-commit` asks "Which directory should be committed (insert
  absolute path)?" and "Which directory is the target?", or reads
  `-from` and `-to`, or `commit_from` and `commit_to` in the
  configuration.
  It writes `numb-commit-request.txt` into the working copy, with every
  file it would add or replace and the real folder's tree before and
  after, and reviews it.
  It copies nothing yet.
  See [docs/numb-commit.md](docs/numb-commit.md).
* `numb-commit review` reviews the request again.
* numb-commit commits a review run as before when `mode = review` is
  set, review runs are waiting in Downloads, or `-run` is given.
* New configuration keys `commit_from` and `commit_to`.
* numb-commit prints the notes about the configuration, as numb-form
  does.
* numb-commit has its own manual page: `numb-form man numb-commit`.

**Action needed:** in direct mode with no review runs waiting,
`numb-commit` now asks for the two folders instead of saying "There is
nothing to commit."
Press **Ctrl+C** to leave.

## v1.3.6

* Internal: numb-commit's questions for the two folders, "Which
  directory should be committed (insert absolute path)?" and "Which
  directory is the target?".
  They accept paths copied with "Copy as path" and ask again until both
  are different working folders that are not inside each other.
  Nothing uses them yet. No change in behaviour.

## v1.3.5

* Internal: numb-commit's review of a request: every file to copy still
  has its checksum and is not in the real folder yet, the spreadsheet
  only gains report dates in cells without one, every new company
  folder holds a filled form, and residue and the template stay behind.
  Nothing uses it yet. No change in behaviour.
* Internal: the request reader refuses files to copy without a SHA-256
  checksum and paths with empty, `.` or `..` parts or a backslash; the
  comparison refuses file names with a backslash; forms and the
  spreadsheet are read with a size limit (50 MB and 256 MB).

## v1.3.4

* Internal: the numb-commit request file, `numb-commit-request.txt`:
  the list of changes and the real folder's tree before and after, for
  people, followed by the same data for numb-commit.
  The reader refuses files over 16 MB and paths that leave the folder,
  and is fuzzed.
  Nothing uses it yet. No change in behaviour.

## v1.3.3

* Internal: numb-commit's comparison of a working copy with the real
  working folder: which files are new, whether the spreadsheet changed,
  what is never copied (residue, the template, changed existing files)
  and which forms numb-form made unchanged.
  Nothing uses it yet. No change in behaviour.

## v1.3.2

* Internal: numb-form can tell a form it filled and nobody edited since
  from one changed by hand, by filling the template again with the
  values the form holds and comparing.
  Nothing uses it yet. No change in behaviour.

## v1.3.1

* Documentation: numb-date has its own manual page,
  [docs/numb-date.md](docs/numb-date.md), shown by
  `numb-form man numb-date` (or `numb-form man date`).
  It lists every question, the files it writes and what to do when
  something goes wrong.

## v1.3.0

* numb-form: Enter on an empty address writes "No address was found
  online due to little to no information available", and Enter on the
  legal form of a name without one writes "No legal form was found
  online due to little to no information available".
  `-` still leaves the field blank.

## v1.2.8

* numb-form: `-` leaves the address and the legal form blank, as it
  already did for the representatives; Enter still leaves them blank.
  Before, `-` was written into the form.

## v1.2.7

* Internal: the standard texts for a missing address and legal form, and
  one helper for questions where Enter writes a standard text and `-`
  leaves the field blank, used for the representatives as before.
  No change in behaviour.

## v1.2.6

* CI: the LibreOffice format tests run side by side in Alpine Linux
  3.24 (LibreOffice 25.8) and Debian 13 (LibreOffice 25.2), the two
  release lines people install. Both install from their project's CDN,
  once per change, instead of waiting up to 45 minutes for a slow Ubuntu
  mirror twice.

## v1.2.5

* Problem reports: GitHub forms for bugs and for documentation errors,
  and a section of the guide on reporting a problem without customer
  data.
* Security: a disclosure process after the Go project's (tracks,
  response times, what is not a vulnerability), secure coding rules and
  an audit checklist in [.github/SECURE-CODING.md](.github/SECURE-CODING.md).
  `scripts/check-deps.sh` now also refuses `math/rand`, cgo, and
  `unsafe` outside the Windows system calls.
* The guide explains the words it uses.

## v1.2.4

* Documentation standard after OpenBSD and FreeBSD: a change to what a
  person sees updates its manual page in the same pull request, and is
  listed here.
  Tests fail when a flag or a configuration key is undocumented or a link
  in the documentation points nowhere.
* `numb-form man changes` shows this file.

## v1.2.3

* The AI prompt is rewritten so that it leaves nothing to interpret:
  the exact entity, sources ranked by trust with the registers that work
  for little-known companies by country, the format of every field,
  numb-form's standard texts, and no em dashes or semicolons.

## v1.2.2

* The AI prompt for finding a company's legal form, address and
  representatives is a separate download, `ai-prompt.txt`, with every
  release ([docs/ai-prompt.md](docs/ai-prompt.md)).
