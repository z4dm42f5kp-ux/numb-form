# NUMB-FORM(1)

## NAME

**numb-form**, **numb-commit** — fill risk assessment forms and commit them

## SYNOPSIS

```
numb-form [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form check-template template.docx [sample.docx]
numb-form keys
numb-commit [-run dir] [-downloads dir]
```

## DESCRIPTION

**numb-form** is started in the folder *above* the working folder.
The working folder is the single folder found there.
If there are several, the one containing an overview folder is used;
otherwise the user is asked.

The working folder must contain:

* an overview folder whose name resembles `3. Übersichten`,
  holding exactly one `.xlsx` file, the spreadsheet;
* the template `Appendix A - Risk Assesment.docx` at its top level.

A run proceeds as follows:

1. The template is checked.
   The run stops if section A or one of its labels cannot be found.
2. The user types the name of the person processing the form, once.
3. The whole working folder is copied to
   `Downloads/numb-form-runs/<date>_<time>_<id>/`.
   Everything after this step happens in the copy.
4. Company names are read from the spreadsheet column headed
   `Company Name`, from top to bottom.
   All other columns are ignored, including `Report created`.
   The spreadsheet is never written to.
   Its SHA-256 checksum is checked before and after the run.
5. For each company, the folders in the working folder and all of its
   subfolders are searched.
   Names are compared ignoring case, accents and umlauts (ü = u = ue),
   punctuation and spaces.
   A match means the form already exists, and the company is skipped.
   A near match is shown to the user to confirm.
   Folders with a different legal form never match:
   `Foo GmbH` and `Foo S.A.` are different companies.
6. For a new company, a folder named after the company is created.
   The template is copied into it as
   `Appendix A - Risk Assesment - <Company Name>.docx`, and section A is filled.
7. The document is read back from disk and shown.
   The user accepts it, edits a field or rejects it.
   A rejected folder is deleted from the copy.

Only the lines of section A, *Information About the Contracting Party*,
are filled.
Section B and everything outside section A are left byte for byte
unchanged.
Values are appended to the label's own text run, so they take the
label's formatting; no new formatting is added.

| Field | Value |
| --- | --- |
| Name / Company Name of the Contracting Party | the literal spreadsheet cell |
| Legal form | the suffix of the name, as written (`GmbH`, `S.A.`, `C.A. de C.V.`) |
| Headquarters / Address | chosen by the user from the candidates found |
| Commercial Registry Number | always blank |
| Authorized Representatives | `Full Name (Role 1, Role 2), Full Name (Role)`, roles in English |
| Date of the evaluation | today, `DD/MM/YYYY` |
| Person Processing the Form | the name typed at the start |

Candidates are gathered from, in order of authority:

1. UK Companies House (government register; only with a free key);
2. the GLEIF LEI register (regulatory register; no key);
3. the company's own legal notice (Impressum, Aviso legal, Legal notice);
4. Wikidata (community-edited; used mainly to find the website).

People and addresses are never taken from a parent, subsidiary or
similarly named entity.
People found only on the website or only on Wikidata are marked
*unverified*, and by default this exact text is appended:

> Caution: This field might differ from internal sources or be innacurate due to insufficient information online

When no representative can be found, the field reads:

> No representatives were found online, please consult other internal sources for more accurate information

**numb-commit** copies the accepted folders of a run back into the real
working folder.
It only adds new folders.
It never overwrites or deletes anything.
It refuses folders that appeared in the real tree in the meantime.
Each folder is copied to a temporary name, verified with SHA-256 and then
renamed.
Documents edited during review are listed before the user confirms.
Afterwards it offers to delete the review copy; the logs are kept.

## OPTIONS

`-offline`
: Contact no online source. Every value is typed by the user.

`-root dir`
: Use *dir* as the working folder instead of detecting it.

`-downloads dir`
: Put the review copy in *dir* instead of the Downloads folder.

`-timeout duration`
: Stop safely after *duration* (for example `15m`). The maximum and
  default is `25m`.

`-run dir`
: (numb-commit) Commit this run folder instead of the newest
  uncommitted one.

`-allow-non-tty`
: Read answers from a pipe. For automated tests only.

`-version`
: Print the version.

## ENVIRONMENT

`NUMBFORM_DOWNLOADS`
: Downloads folder to use, for example a mounted volume in a container.

`NUMBFORM_COMPANIES_HOUSE_KEY`
: Optional free UK Companies House API key. Never logged or stored.

`HTTPS_PROXY`, `NO_PROXY`
: Proxy settings for online lookups.

## FILES

`Downloads/numb-form-runs/<run>/<working folder>/`
: The review copy.

`Downloads/numb-form-runs/<run>/numb-form.log`
: Log of everything the run did.

`Downloads/numb-form-runs/<run>/manifest.json`
: Machine-readable record used by numb-commit.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Log of the commit.

No cache or temporary files are left behind.
Network responses are held in memory only, and documents are written
through a temporary file that is renamed or removed.
A run writes at most 100 MB of documents besides the copy.

## EXIT STATUS

0 on success, 1 on error, 2 on a usage error, and 3 when stopped by the
time limit, by Ctrl+C or by the user.

## EXAMPLES

Check the real template and write a sample to look at, without touching
the template:

```
numb-form check-template "Appendix A - Risk Assesment.docx" "%USERPROFILE%\Downloads\sample.docx"
```

Run, review, then commit (numb-commit waits for confirmation):

```
cd /d "S:\Kunden" && numb-form && numb-commit
```

Run without any network access:

```
cd /d "S:\Kunden" && numb-form -offline && numb-commit
```

## SEE ALSO

[install.md](install.md), [containers.md](containers.md),
[../SECURITY.md](../SECURITY.md), [../AGENTS.md](../AGENTS.md)
