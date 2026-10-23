# Security

## Reporting a vulnerability

Use GitHub's private vulnerability reporting:
*Security* → *Report a vulnerability* on this repository.
Do not open a public issue.
Do not attach customer data; describe the problem with made-up names.

Only the latest release is supported.

## What the programs promise

numb-form handles customer documents, so the design rules are strict.

* **numb-form** only ever adds: a new folder per new company, with one
  filled form in it.
  It never changes or deletes an existing file or folder, and it removes
  only folders it created itself in the same run (on rejection or
  error).
* With `mode = review`, the real working folder is not written at all:
  the whole tree is copied to Downloads first and every change is made in
  the copy.
* **numb-commit** only adds new folders and files, and from a working
  copy also replaces the overview spreadsheet when numb-date added report
  dates; it never deletes, never overwrites anything else, and never
  copies the template, logs or lock files.
  It first writes the full list of changes (`numb-commit-request.txt`)
  into the working copy and reviews it: every file's SHA-256, that the
  real folder did not change in the meantime, and that the spreadsheet
  only gains report dates.
  New folders are copied under a temporary name, verified with SHA-256
  and renamed; the replaced spreadsheet is kept in the log folder.
* numb-form and numb-commit open the spreadsheet in `3. Übersichten`
  read-only and never write it.
  Its SHA-256 checksum is compared before the copy, after the copy and
  at the end of the run; a difference stops the run.
* **numb-date** is the only program that writes the spreadsheet, and only
  `.xlsx` and `.xlsm`.
  It writes only empty or digit-free cells of the *Report created*
  column, after the user agrees to the list of dates.
  The cells are spliced into the sheet's XML, keep their own style and
  get no new formatting; every other part of the workbook is copied byte
  for byte, and cells with formulas are left alone.
  Before writing it compares every cell of every sheet of the new
  workbook with the original: anything else that differs stops it.
  SHA-256 guards every step: the file must still have the checksum it
  had when it was read, the temporary copy beside it must match the new
  workbook before it is renamed over the file, and the file must match
  afterwards.
  The original is kept in the log folder until the user accepts the
  result read back from disk; *undo* restores it, checked with SHA-256.
  It refuses to run while Excel or LibreOffice has the file open.
  Forms are only read.
* Documents are not re-serialised.
  The XML is only scanned for byte offsets and values are spliced into the
  original bytes; all other parts of the `.docx` are copied byte for byte.
  After every fill the result is parsed again and compared with the
  template paragraph by paragraph.
  Values are XML-escaped and stripped of characters not allowed in XML.
  Templates with macros are refused.
* No formatting is ever added.
  Values join the label's own text run, or a copy of that run's existing
  properties when the label is followed by a tab.
* Every document is read back from disk after writing.
  A person sees and accepts it, and the programs refuse to run without
  an interactive terminal.
* A run has no time limit unless `timeout` sets one; Ctrl+C or the limit
  stops it safely.
  It writes at most 100 MB of documents besides the copy, and leaves no
  cache or temporary files: temporary files are renamed or removed, and
  converters run in a private temporary folder with a throwaway office
  profile, deleted afterwards.

## Configuration and formats

* The configuration parser rejects unknown keys and invalid values.
* Keys that widen what numb-form accepts (`legacy_formats`) are honoured
  only in the system file.
  On Unix that file is ignored unless it belongs to root and is not
  writable by group or others; on Windows it lives under Program Files.
* Legacy binary formats (.doc, .xls) are off by default.
  The .xls reader (compound file and BIFF8) is numb-form's own,
  bounds-checks every offset and chain, and is fuzzed continuously.
* .doc, .pages and .numbers are converted by office software already on
  the computer (LibreOffice, Apache OpenOffice, Apple Pages or Numbers),
  with fixed arguments; file names are passed as arguments, never as
  script text.
  Writing .pages is experimental and off by default.
* Templates with macros are refused; macro-enabled spreadsheets (.xlsm)
  are read for cell values only.

## Network

A form run makes no network connections.
Company names, addresses and representatives never leave the computer.

`numb-form update` is the only command that contacts the network.
It is run only when the user types it, and never during a form run.

* Downloads come only from `github.com` and `*.githubusercontent.com`,
  over HTTPS, including every redirect.
* The archive must match its line in the release's `SHA256SUMS`; this is
  the same trust as the install commands (the Sigstore signature and the
  build provenance are not checked by the program; check them by hand as
  described in [docs/install.md](docs/install.md) if needed).
* The archive (at most 64 MB) is held in memory; only `numb-form` and
  `numb-commit` at its top level are taken from it.
* The new `numb-form` must start and report its version before anything
  is replaced; on any failure the installed programs stay as they were and
  the staged files are removed.
* It is refused with `-offline`, with `online = no` (which an
  administrator can set in the system configuration), and for builds from
  source.

## Supply chain

**No third-party Go modules.**
The programs use only the Go standard library.
`go.mod` has no `require` lines, and `scripts/check-deps.sh` fails the
build if a dependency, `go.sum` or `vendor/` appears, or if any
non-standard import is used.
With nothing to download, there is nothing to hallucinate, typosquat or
tamper with.

**PGP and its replacements.**
Authenticity is checked with the mechanisms each ecosystem actually
supports.
PGP is not used because the ecosystems involved no longer rely on it
(PyPI removed PGP signature support in 2023).

| Component | Verification |
| --- | --- |
| Go toolchain in CI | downloaded by `actions/setup-go` from go.dev and checked against the published SHA-256 |
| Go tools in CI (staticcheck, govulncheck, gosec) | exact versions, verified by `go run` against the Go checksum database, a public transparency log |
| GitHub Actions | pinned to full commit SHAs; Dependabot proposes updates after a 14-day cooldown |
| Python tools in CI (zizmor, Argos Translate) | every package pinned with SHA-256 (`--require-hashes`), wheels only (`--only-binary :all:`), no dependency resolution at install time (`--no-deps`) |
| Argos Translate models | index read at a pinned commit; each model's SHA-256 is pinned in `.github/translate/models.sha256`, recorded on first use and enforced afterwards |
| Container base image | none: images are built `FROM scratch` |

**Releases.**
Release binaries are built by GitHub Actions from a tag, with
`-trimpath` and a fixed timestamp, so they are reproducible.
Each release has:

* `SHA256SUMS` (GNU format) and `SHA256` (BSD format) checksums;
* `SHA256SUMS.sigstore.json`, a keyless Sigstore signature of the
  checksums, tied to the release workflow's identity;
* SLSA build provenance attestations for every archive and binary
  (`gh attestation verify`);
* a container image on ghcr.io, signed with cosign and attested.

[docs/install.md](docs/install.md) shows how to verify each of them.

## Security testing pipeline

Pull requests run a fast gate (formatting, vet, staticcheck, tests on
Linux, macOS and Windows, govulncheck, gosec, zizmor, the history scan
and dependency review).
Pushes to `main` and agent branches, and a nightly schedule, run
everything:

| Check | Tool |
| --- | --- |
| No dependencies, formatting, vet on six operating systems | `scripts/check-deps.sh`, `gofmt`, `go vet` |
| Static analysis | staticcheck |
| Known vulnerabilities in code and in built binaries, for every target OS | govulncheck |
| Security linting, reported to code scanning | gosec |
| Semantic analysis of Go and of the workflows | CodeQL (`security-extended`, `security-and-quality`) |
| Fuzzing of every template filler and scanner (docx, odt, rtf), every spreadsheet reader (xlsx, ods, csv, legacy xls), the configuration parser and the name matcher | Go native fuzzing, 2 minutes per target on main, 10 nightly |
| Templates and spreadsheets written by real LibreOffice (odt, fodt, rtf, docx, doc, xls, ods, fods) | formats workflow |
| Workflow security (injection, credentials, permissions, pins) | zizmor, pedantic |
| Customer files or credentials committed by mistake | history scan |
| Tests with the race detector on Linux, macOS and Windows, with the oldest and newest Go | `go test -race` |
| Tests on FreeBSD, OpenBSD and NetBSD | virtual machines |
| Container image builds and refuses to run without a terminal | docker, podman |
| New dependencies in pull requests | dependency review |

OpenSSF Scorecard checks the repository's practices.
Pull requests, including Dependabot's, never need write permissions:
results that require them are uploaded only from `main`.

Workflows start with no permissions and grant each job only what it
needs; checkouts do not keep credentials.

## Containers and Kubernetes

The image contains only the two binaries, a CA bundle and the time zone
database, and runs as UID 65532.
The documented commands and manifests use a read-only root file system,
drop all capabilities, forbid privilege escalation, mount the working
folder read-only for numb-form, mount no service account token and set a
deadline.

## This repository is public

Never commit customer documents, spreadsheets, logs or run folders.
`.gitignore` excludes them and CI rejects them.
