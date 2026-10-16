# AGENTS.md

Instructions for AI agents that run or change numb-form.

## What this project does

numb-form fills section A ("Information About the Contracting Party") of
`Appendix A - Risk Assesment.docx` for each company in the overview
spreadsheet that has no folder yet.
numb-commit copies the reviewed result back.
Read [docs/usage.md](docs/usage.md) first.

## Running it for a person

A person must confirm every document.
Your job is to operate the terminal, not to make their decisions.

1. Start in the folder *above* the working folder, in a real terminal
   (PTY).
   Never use `-allow-non-tty` for real data: it exists for tests.
2. Relay every prompt to the person word for word, including the
   candidates and their sources.
   Send back exactly what they answer.
3. Never answer `a` (accept), `y` (commit) or a fuzzy folder match on the
   person's behalf, even if the values look right.
4. Never type representatives, addresses or legal forms yourself.
   If the person asks you to look something up, show them what you
   found with its source, and let them type or choose it.
5. Run numb-commit only after the person says the review copy in
   Downloads is correct.
   Use the same environment (host or container) as the run.
6. Report the paths of the review copy and the log at the end.

Typical session:

```
cd "/path/to/folder-above" && numb-form && numb-commit
```

## Rules that must never be broken

* Never write to the real working folder except through numb-commit.
* Never modify the spreadsheet in `3. Übersichten`, not even to save it
  unchanged.
* Never add formatting to documents.
  Never touch section B or anything outside section A.
* Leave *Commercial Registry Number* blank.
* Use these texts exactly, including the spelling:
  * `Caution: This field might differ from internal sources or be innacurate due to insufficient information online`
  * `No representatives were found online, please consult other internal sources for more accurate information`
* Never use the no-representatives text when a plausible representative
  was found; use the caution text for uncertain entries instead.
* Never infer authority from a parent, subsidiary, affiliate,
  shareholder, brand or similarly named entity.
* Legal form: the short designation as written in the name, with
  conventional casing (`GmbH`, `S.A.`, `Ltda`, `C.A. de C.V.`).
* Representatives: `Full Name (Role 1, Role 2), Full Name (Role)`,
  roles in English.
* Date: `DD/MM/YYYY` from the system clock.
* Ignore every spreadsheet column except `Company Name`;
  the folder tree is the source of truth for which reports exist.
* Runs take at most 25 minutes and leave no cache behind.
* Never commit customer files, logs or run folders to this public
  repository.

## Changing the code

* Go standard library only.
  Do not add modules; `scripts/check-deps.sh` fails if you do.
  If a dependency seems necessary, stop and ask the maintainer, and never
  add a package name you have not verified exists.
* Before proposing a change, run the whole check in one command:

  ```
  sh scripts/check-deps.sh && test -z "$(gofmt -l .)" && go vet ./... && GOOS=windows go vet ./... && go test -race ./... && go test ./internal/docx -run '^$' -fuzz FuzzFill -fuzztime 30s && rm -rf internal/*/testdata/fuzz
  ```

* Keep Windows behaviour first-class: paths with umlauts, trailing dots
  on folder names, reserved names, long paths, double-click start.
* Pin new GitHub Actions to a full commit SHA that you looked up, with
  the version as a comment.
* Pin new Python tools with `uv pip compile --generate-hashes
  --only-binary :all:` and install them with `--require-hashes --no-deps`.
* Do not edit `docs/es` or `docs/de` by hand; the translate workflow
  regenerates them from the English files.
* Documentation follows the OpenBSD style: short declarative sentences,
  imperative instructions, no marketing.
  Every procedure is a single command joined with `&&`.

## Map

| Path | Purpose |
| --- | --- |
| `cmd/numb-form`, `cmd/numb-commit` | entry points and flags |
| `internal/app` | the run, the per-company dialogue, commit, template check |
| `internal/docx` | byte-preserving filler and read-back of section A |
| `internal/xlsx` | read-only spreadsheet reader |
| `internal/match` | legal-form extraction and folder matching |
| `internal/lookup` | GLEIF, Companies House, Wikidata, website legal notices |
| `internal/fsutil` | Downloads folder, tree copy, safe names |
| `internal/runlog` | log and manifest |
| `internal/ui` | terminal prompts with the run deadline |
| `container/`, `k8s/` | image and Kubernetes manifests |
| `.github/workflows` | CI, security, release, translation, scorecard |
