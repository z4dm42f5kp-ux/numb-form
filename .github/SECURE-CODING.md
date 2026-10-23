# Secure coding and audits

numb-form reads customer documents and writes into a shared drive.
It follows the practice of the Go project: the standard library only,
safe defaults, every parser fuzzed, and a written
[security policy](https://go.dev/doc/security/policy) with fixed
response times.
[SECURITY.md](../SECURITY.md) says what the programs promise; this file
says how the code keeps those promises and how they are audited.

## Rules for the code

Each rule names what enforces it.
A rule without an automatic check is checked in review.

1. **Standard library only.**
   No third-party modules (`scripts/check-deps.sh`).
2. **No cgo, no `math/rand`, `unsafe` only for Windows system calls.**
   Random values come from `crypto/rand` (`scripts/check-deps.sh`).
3. **Every input is untrusted.**
   Templates, forms, spreadsheets, configuration files, request files,
   release downloads and terminal input may be malformed or hostile.
   * Bound every read: sizes, zip member counts and uncompressed sizes.
   * Validate every path taken from a file: no absolute paths, no `..`,
     nothing outside the folder it belongs to.
   * Never panic on input; return an error that names the file.
4. **Every parser is fuzzed.**
   A new format or reader gets a fuzz target in `security.yml` in the
   same pull request (AGENTS.md).
5. **Write by copy, check and rename.**
   Write a temporary file beside the target, check its SHA-256, then
   rename it.
   Never write a file in place, never overwrite a file the program did
   not create in the same run (except the spreadsheet by numb-date, as
   SECURITY.md describes), and delete only what the run created.
6. **No shell.**
   Run other programs with `exec.Command` and a fixed argument list.
   Text from a file or a person is passed as an argument, never as
   script text (gosec G204).
7. **Network only in `numb-form update`.**
   HTTPS only, hosts on a fixed list, every download checked against
   `SHA256SUMS` before use.
8. **Terminal output is escaped.**
   Text from files goes through `ui.Safe`, so it cannot send terminal
   control sequences.
9. **Explicit permissions.**
   Files `0600` or `0644`, folders `0750`.
   Every deviation carries a `#nosec` comment with its reason (gosec).
10. **No secrets.**
    No keys, tokens or passwords in code, tests, logs or history
    (history scan).
11. **No customer data in the repository.**
    Tests and documentation use made-up names (`Foo GmbH`).
12. **Workflows have no permissions by default.**
    Each job is granted only what it needs, actions are pinned to full
    commit SHAs, checkouts do not keep credentials, and untrusted values
    reach scripts only through environment variables (zizmor, pedantic).

## Checks on every pull request

| What | Tool |
| --- | --- |
| dependencies, formatting, vet on six systems | `scripts/check-deps.sh`, `gofmt`, `go vet` |
| static analysis | staticcheck |
| security lint | gosec |
| known vulnerabilities | govulncheck |
| semantic analysis | CodeQL |
| workflow security | zizmor |
| tests with the race detector | `go test -race` |
| documentation follows the code | `docs_test.go`, `scripts/check-changes.sh` |

Fuzzing, the BSD systems, containers and LibreOffice formats run on
`main` and nightly (SECURITY.md lists them all).

## Review checklist

The reviewer of a pull request that reads files, writes files, runs
programs or uses the network answers each question:

* [ ] Is every size, count and path from the input bounded and checked?
* [ ] Can a malformed file cause a panic, a hang or unbounded memory?
* [ ] Is every write a copy, check and rename, of a file this run made?
* [ ] Are all errors handled, and do messages name the file?
* [ ] Is text from files escaped before it reaches the terminal?
* [ ] Does a new parser have a fuzz target?
* [ ] Does SECURITY.md still describe what the code does?

## Audits

**Every release.**
Before approving a release, check that ci and security passed on the
commit, that govulncheck reports nothing, and that SECURITY.md still
matches the code.
After the release, the daily `updates.yml` check scans the published
binaries with govulncheck and proposes a rebuild when a newer Go fixes
a vulnerability in them.

**Every three months, and before every major version.**
Audit the whole project and record the result in a GitHub issue titled
`Security audit YYYY-MM`:

1. Go through the rules above, file by file, for code changed since the
   last audit.
2. Read every `#nosec` comment and check that its reason still holds.
3. Read the workflows: permissions, pins, triggers, secrets.
4. Compare the OpenSSF Scorecard result with the last audit and fix
   regressions.
5. Check the project against the
   [OpenSSF Best Practices](https://www.bestpractices.dev/en/criteria/0)
   criteria (passing level) and note every criterion that is not met.
6. Review the fuzz targets: does every reader and filler have one?
7. Check that the vulnerability reporting address works and that open
   reports were answered within the times in SECURITY.md.

## How the project meets common criteria

| Criterion (OpenSSF Best Practices) | Where |
| --- | --- |
| vulnerability report process, private reports | [SECURITY.md](../SECURITY.md) |
| response time | SECURITY.md: acknowledged within 7 days |
| bug report process | issue forms in `.github/ISSUE_TEMPLATE` |
| documentation of the interface | the manual pages in `docs/` |
| release notes | [CHANGES.md](../CHANGES.md) |
| static analysis, warnings as errors | staticcheck, gosec, CodeQL, `go vet` |
| dynamic analysis | Go fuzzing, the race detector |
| automated tests on every change | ci workflow |
| unique version numbers, signed releases | semantic versions, Sigstore signatures, build provenance |
| delivery against man-in-the-middle | HTTPS and SHA-256 checksums |
| no known unfixed vulnerabilities | govulncheck on code and on published binaries |
