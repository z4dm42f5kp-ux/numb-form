# Changes

What changed in each release, newest first.
Every pull request that changes the programs adds its lines here, under
the version it will be released as
([.github/DOCUMENTATION.md](.github/DOCUMENTATION.md)).
**Action needed** lists what users must do after updating.
Releases up to v1.2.1 are described on the
[releases page](https://github.com/z4dm42f5kp-ux/numb-form/releases).

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
