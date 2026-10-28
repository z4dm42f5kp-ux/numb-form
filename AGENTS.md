# AGENTS.md

Instructions for AI agents that run or change numb-form.

## What this project does

numb-form fills section A ("Information About the Contracting Party") of
the *Appendix A - Risk Assesment* template for each company in the
overview spreadsheet that has no folder yet.
By default (`mode = direct`) it writes into the working folder after
each confirmation; with `mode = review` it works on a copy in Downloads
and numb-commit copies the reviewed result back.
Read [docs/usage.md](docs/usage.md) and
[docs/configuration.md](docs/configuration.md) first.
A session that starts without context reads
[.github/everything-agents-need-to-know.txt](.github/everything-agents-need-to-know.txt)
before anything else.

## Running it for a person

A person must confirm every document.
Your job is to operate the terminal, not to make their decisions.

1. Start inside the working folder (the one holding the template and
   the overview folder), in a real terminal (PTY).
   Never use `-allow-non-tty` for real data: it exists for tests.
2. Relay every prompt to the person word for word, including the
   candidates and their sources.
   Send back exactly what they answer.
3. Never answer `a` (accept), `y` (commit) or a fuzzy folder match on the
   person's behalf, even if the values look right.
4. Never type representatives, addresses or legal forms yourself; relay
   exactly what the person gives you.
5. In review mode, run numb-commit only after the person says the
   review copy in Downloads is correct.
   Use the same environment (host or container) as the run.
6. Report the paths of the log (and the review copy) at the end.

Typical session:

```
cd "/path/to/working-folder" && numb-form
```

## Rules that must never be broken

* In review mode, never write to the real working folder except through
  numb-commit.
* Never enable `legacy_formats` or `pages_write` yourself; the first
  needs an administrator, the second is experimental.
* Never modify the spreadsheet in `3. Übersichten`, not even to save it
  unchanged.
  The only exception is numb-date, which writes digit-free cells of the
  *Report created* column after the person agrees; never answer its
  questions for the person either.
* Never add formatting to documents.
  Never touch section B or anything outside section A.
* Fill *Commercial Registry Number* with the registry text below, unless
  the person types the number when editing.
* Use these texts exactly, including the spelling:
  * `No commercial registry number was found online consult internal sources for further information`
  * `No representatives were found online, please consult other internal sources for more accurate information`
  * `No address was found online due to little to no information available`
  * `No legal form was found online due to little to no information available`
  (written when the person leaves that field empty)
* Address and representatives are typed by the person; numb-form makes
  no online lookups for companies.
  Do not add web search, scraping or lookup code back.
* Legal form: the short designation as written in the name, with
  conventional casing (`GmbH`, `S.A.`, `Ltda`, `C.A. de C.V.`).
* Representatives: `Full Name (Role 1, Role 2), Full Name (Role)`,
  roles in English.
* Date: `DD/MM/YYYY` from the system clock.
* Ignore every spreadsheet column except `Customer Name`;
  the folder tree is the source of truth for which reports exist.
* Runs leave no cache behind. They have no time limit unless `timeout`
  or `-timeout` sets one.
* Never commit customer files, logs or run folders to this public
  repository.

## Changing the code

* Go standard library only.
  Do not add modules; `scripts/check-deps.sh` fails if you do.
  If a dependency seems necessary, stop and ask the maintainer, and never
  add a package name you have not verified exists.
* Test on GitHub, never locally.
  Push to a branch named `claude/<topic>`; the ci and security
  workflows run on it, and formats runs on its pull request (Linux,
  macOS, Windows, FreeBSD, OpenBSD, NetBSD, LibreOffice formats,
  fuzzing).
  Read the results, including failure annotations, through the API:

  ```
  sha=$(git rev-parse HEAD) && curl -s "https://api.github.com/repos/z4dm42f5kp-ux/numb-form/actions/runs?head_sha=$sha" | grep -E '"(name|status|conclusion)"'
  ```

  Compile checks before pushing are fine and fast:

  ```
  sh scripts/check-deps.sh && test -z "$(gofmt -l .)" && for os in linux windows darwin freebsd openbsd netbsd; do GOOS=$os go vet ./... || exit 1; done
  ```

* Every operating system is equal: paths with umlauts, trailing dots on
  folder names, reserved names and long paths must work everywhere.
* Every template format implements `form.Filler`, splices bytes instead
  of re-serialising, and passes `form.Verify`; add it to
  `internal/form/formats_test.go` and to the fuzz targets.
* Keep files short (about 300 lines at most).
* Pin new GitHub Actions to a full commit SHA that you looked up, with
  the version as a comment.
* Pin new Python tools with `uv pip compile --generate-hashes
  --only-binary :all:` and install them with `--require-hashes --no-deps`.
* Do not edit `docs/es` or `docs/de` by hand; the translate workflow
  regenerates them from the English files.
* A change to what a person sees updates its manual page and
  `CHANGES.md` in the same pull request
  ([.github/DOCUMENTATION.md](.github/DOCUMENTATION.md)); tests and the
  *changes listed* job enforce it.
* Follow the rules and the review checklist of
  [.github/SECURE-CODING.md](.github/SECURE-CODING.md).
* Documentation follows the OpenBSD style: short declarative sentences,
  imperative instructions, no marketing.
  Every procedure is a single command, labelled **Unix-like** (macOS,
  Linux, BSD; joined with `&&`) or **Windows** (10, 11; PowerShell 5.1,
  joined with `;` and `if ($?) { }`, or one `& { ... }` block), or
  marked as the same on all systems.

## Map

| Path | Purpose |
| --- | --- |
| `cmd/numb-form`, `cmd/numb-commit`, `cmd/numb-date` | entry points and flags |
| `internal/app` | the run, modes, the per-company dialogue, numb-date, numb-commit (`request.go`, `requestfile.go`, `tree.go`, `stage.go`, `reviewreq.go`, `foldercommit.go`, `commitapply.go`) |
| `internal/config` | numb-form.conf: locations, parsing, defaults |
| `internal/form` | fields, locating section A, splicing and the self-check |
| `internal/docx`, `internal/odt`, `internal/rtf` | template fillers |
| `internal/formats` | choosing a filler or converter by extension |
| `internal/convert` | LibreOffice/OpenOffice and Apple Pages/Numbers conversion |
| `internal/sheet` | spreadsheet readers (xlsx, ods, csv, xls); `table.go` and `setcells.go` serve numb-date |
| `internal/reportdate` | numb-date: finding and parsing the date in a form |
| `internal/match` | legal-form extraction and folder matching |
| `internal/manual` | `numb-form man` and the Unix manual page |
| `internal/fsutil` | Downloads folder, tree copy, safe names |
| `internal/runlog` | log and manifest |
| `internal/ui` | terminal prompts with the run deadline |
| `internal/update` | `numb-form update`: checksummed self-update from GitHub releases |
| `container/`, `k8s/` | image and Kubernetes manifests |
| `.github/workflows` | CI, security, release, edge (rolling build), translation, scorecard |
