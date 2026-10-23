# Changes

What changed in each release, newest first.
Every pull request that changes the programs adds its lines here, under
the version it will be released as
([.github/DOCUMENTATION.md](.github/DOCUMENTATION.md)).
**Action needed** lists what users must do after updating.
Releases up to v1.2.1 are described on the
[releases page](https://github.com/z4dm42f5kp-ux/numb-form/releases).

## v1.5.0

* numb-typo corrects misspelt or cut-off places at the end of company
  names in the overview spreadsheet, for example "Ciudad de Me" to
  "Ciudad de México", after you agree.
  It leaves the names of companies that already have a folder alone.
* numb-commit accepts numb-typo's corrections in the spreadsheet.
  Any other change to a name still stops it.
* `numb-form update` installs numb-typo with the other programs.

## v1.4.1

* numb-commit: the request file starts with a UTF-8 byte order mark,
  so its folder tree is readable in Windows PowerShell 5.1 and Notepad.
* numb-commit removes its staged copy in Downloads also when the window
  is closed.
  A staged copy left by a numb-commit that was killed is removed by the
  next run after one day.

## v1.4.0

* numb-commit copies the new forms from a working copy (for example in
  Downloads) into the real folder.
  It asks for both folders, or reads `-from`, `-to`, `commit_from` and
  `commit_to`.
  It stages the files to copy in Downloads, writes
  `numb-commit-request.txt` with the folder tree before and after, and
  copies only after you agree.
  It never overwrites a file, and skips a new folder of a company the
  real folder already has under its customer ID.
* `numb-commit review` reviews the request, by you or automatically
  until it is right.
* Each program has its own manual page: `numb-form man numb-commit`,
  `numb-form man numb-date`.
* **Action needed**: none. With `mode = review`, numb-commit commits
  review runs as before.

## v1.3.0

* numb-form: Enter on an empty address writes "No address was found
  online due to little to no information available", and Enter on the
  legal form of a name without one writes "No legal form was found
  online due to little to no information available".
  `-` still leaves the field blank.

## v1.2.4

* Documentation standard after OpenBSD and FreeBSD: a change to what a
  person sees updates its manual page in the same pull request, and is
  listed here.
  Tests fail when a flag or a configuration key is undocumented or a link
  in the documentation points nowhere.
* Problem reports: GitHub forms for bugs and for documentation errors,
  and a section of the guide on reporting a problem without customer
  data.
* Security: a disclosure process after the Go project's (tracks,
  response times, what is not a vulnerability), secure coding rules and
  an audit checklist in [.github/SECURE-CODING.md](.github/SECURE-CODING.md).
  `scripts/check-deps.sh` now also refuses `math/rand`, cgo, and
  `unsafe` outside the Windows system calls.
* The guide explains the words it uses.
* CI: the LibreOffice format tests cache LibreOffice's packages and run
  once per change, instead of waiting up to 45 minutes for a slow
  Ubuntu mirror twice.

## v1.2.3

* The AI prompt is rewritten so that it leaves nothing to interpret:
  the exact entity, sources ranked by trust with the registers that work
  for little-known companies by country, the format of every field,
  numb-form's standard texts, and no em dashes or semicolons.

## v1.2.2

* The AI prompt for finding a company's legal form, address and
  representatives is a separate download, `ai-prompt.txt`, with every
  release ([docs/ai-prompt.md](docs/ai-prompt.md)).
