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
