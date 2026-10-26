# Changes

What changed in each release, newest first.
Every pull request that changes the programs adds its lines here, under
the version it will be released as
([.github/DOCUMENTATION.md](.github/DOCUMENTATION.md)).
**Action needed** lists what users must do after updating.
Releases up to v1.2.1 are described on the
[releases page](https://github.com/z4dm42f5kp-ux/numb-form/releases).

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
