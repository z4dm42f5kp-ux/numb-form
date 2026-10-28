# NUMB-COMMIT(1)

## NAME

**numb-commit** — copy reviewed forms into the real working folder, and
show what a working copy would add to it

## SYNOPSIS

```
numb-commit [-f conf] [-from dir] [-to dir]
numb-commit [-f conf] [-from dir] [-to dir] review
numb-commit [-f conf] [-run dir] [-downloads dir]
```

## DESCRIPTION

numb-form and numb-date work on a folder on your own computer or on the
real working folder itself.
numb-commit deals with the result.
There are two ways to work.

**A working copy** (for a network drive).
You copy the real folder, for example `F:\14. GWG`, to your computer,
for example `C:\Users\you\Downloads\14. GWG`, and run numb-form and
numb-date there.
Then you start numb-commit; it asks for the two folders:

```
Which directory should be committed (insert absolute path)?: C:\Users\you\Downloads\14. GWG
Which directory is the target?: F:\14. GWG
```

Paths copied with *Copy as path* (with quotes) are accepted.
To answer once and for all, set them in the configuration:

```
numb-form config set commit_from "C:\Users\you\Downloads\14. GWG"
numb-form config set commit_to "F:\14. GWG"
```

`-from` and `-to` answer them for one run.

This version of numb-commit only compares the two folders, writes the
request and reviews it.
It copies nothing into the target.
Copy the files listed under *New folder* and *Replaced* yourself.

**A review run** (`mode = review`).
numb-form makes the copy itself in `Downloads\numb-form-runs`, and
numb-commit copies the accepted folders of the newest run back into the
real working folder.
It only adds new folders, never overwrites or deletes anything, and
refuses folders that appeared in the real folder in the meantime.
Each folder is copied to a temporary name, verified with SHA-256 and then
renamed.
Documents edited during the review are listed before you confirm.
Afterwards it offers to delete the review copy; the logs are kept.

numb-commit commits a review run when `mode = review` is set or review
runs are waiting, and neither `-from`, `-to`, `commit_from`,
`commit_to` nor `review` names a folder; `-run` always does it.

## WHAT NUMB-COMMIT DOES WITH A WORKING COPY

1. It compares the working copy with the target, file by file.
   Files with the same size and time are the same; for the others, the
   SHA-256 checksum decides.
2. It writes **`numb-commit-request.txt`** into the working copy and
   shows it: the list of changes and the target's folder tree before and
   after (see THE REQUEST FILE).
3. It reviews the request (see REVIEW) and prints a verdict.
4. It copies nothing: the target is not touched.

What the request lists as copied:

* **New folders** (normally one per new company) and **new files** in
  folders that already exist.
* **The overview spreadsheet**, when numb-date added report dates.

What is never copied:

* **A file that exists in the target with other content**: it is never
  overwritten, except the spreadsheet.
  The request lists these as differing, so you see what changed in the
  working copy.
* the template (`Appendix A - Risk Assesment.docx`), even if it changed;
* logs, `manifest.json`, `numb-form-runs` folders and the request file;
* Office and LibreOffice lock files (`~$…`, `.~lock.…#`), hidden and
  temporary files, `Thumbs.db`, `desktop.ini`.

## FORMS MADE BY NUMB-FORM

numb-commit tells which forms numb-form made and nobody changed since.
numb-form writes no mark into the documents.
It records the SHA-256 checksum of every form it writes in the
`manifest.json` of its run: in `Downloads/numb-form-runs` for a review
run, in the log folder for a direct run.
numb-commit compares each form's checksum with those records.
A form is labelled:

* `made by numb-form, unchanged` when a run recorded its checksum;
* `a form edited after filling, made by hand or filled on another computer`
  otherwise;
* nothing, for other files.

This works for every template format.

## THE REQUEST FILE

`numb-commit-request.txt` lies in the working copy, never in the target.
It says, in plain words, exactly what would change:

```
numb-commit request: what numb-commit copies into the real folder

From:   C:\Users\you\Downloads\14. GWG
To:     F:\14. GWG
Made:   23/10/2026 14:05
Status: waiting

New folder  Bar Ltda
  + Bar Ltda\Appendix A - Risk Assesment - Bar Ltda.docx  (18 KB, made by numb-form, unchanged, sha256 4be1…)

Replaced  3. Übersichten\Kundenliste.xlsx
  sha256 9f02… in the real folder now
  sha256 c771… afterwards
  the replaced version is kept in the log folder

Left alone
  - Appendix A - Risk Assesment.docx  (not copied: the template)

Nothing in the real folder is deleted, and no file there is overwritten except the ones under Replaced.

Before: F:\14. GWG
├── 3. Übersichten
│   └── Kundenliste.xlsx
└── … 1204 more, unchanged

After the commit:
├── 3. Übersichten
│   └── Kundenliste.xlsx                                ~ replaced
├── Bar Ltda                                            + new folder
│   └── Appendix A - Risk Assesment - Bar Ltda.docx     + new, made by numb-form, unchanged
└── … 1204 more, unchanged
```

The trees show only what the request touches; the last line counts the
other folders and files at the top of the target.
In the tree after the commit:

| Mark | Meaning |
| --- | --- |
| `+` | new: to be copied into the target |
| `~` | replaced: the spreadsheet with its new report dates |
| `≠` | differs in the working copy; the target's version stays |

*Status* is `waiting`.
Below the readable part, a data section repeats the list for
numb-commit; do not edit it.

## REVIEW

The review checks that:

* every file to copy still has the checksum in the request (nothing was
  edited since);
* none of them exists in the target yet;
* the target's spreadsheet still has the checksum in the request
  (nobody else changed it in the meantime);
* the spreadsheet only gains report dates: only empty or digit-free
  cells of the *Report created* column change, and each new value looks
  like `DD-MM-YYYY`;
* each new folder belongs to a company in the Excel list and holds a
  risk assessment form whose section A names that company and has no
  blank line, and the target has no similar folder for it;
* nothing that must stay behind (template, logs, lock files) is listed.

Each line is marked `ok`, `check` (look at it) or `FAIL`, followed by
the verdict:

| Verdict | Meaning |
| --- | --- |
| `OK to commit.` | Every check passed. |
| `commit after checking the lines marked check.` | Nothing is wrong with the files, but something deserves a look, for example a blank address. |
| `do not commit.` | Something changed or does not belong. |

`numb-commit review` runs the review again on the request in the
working copy at any time, for example on a request a colleague
prepared.

## OPTIONS

`-from dir`
: The working copy. Default: `commit_from` in the configuration; when
  that is empty, numb-commit asks.

`-to dir`
: The target, the real working folder. Default: `commit_to` in the
  configuration; when that is empty, numb-commit asks.

`review`
: Only review the request in the working copy.

`-run dir`
: Review runs: commit this run folder instead of the newest uncommitted
  one.

`-downloads dir`
: Review runs: the folder that holds `numb-form-runs`.
  Default: `downloads` in the configuration, else your Downloads folder.

`-f file`
: Read *file* after the system and per-user configuration files.

`-allow-non-tty`
: Read answers from a pipe. For automated tests only.

`-version`
: Print the version.

## FILES

`<working copy>/numb-commit-request.txt`
: The request: what would be copied. UTF-8.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Review runs: log of the commit.

## EXIT STATUS

0 when the review passes or there is nothing to do; 1 on an error or a
failed review; 2 on a usage error; 3 when stopped with Ctrl+C.

## EXAMPLES

Once: say where both folders are (all systems):

```
numb-form config set commit_from "C:\Users\you\Downloads\14. GWG"; numb-form config set commit_to "F:\14. GWG"
```

Every time.

**Windows** (10, 11), in PowerShell:

```
cd "$env:USERPROFILE\Downloads\14. GWG"; numb-form; if ($?) { numb-date }; if ($?) { numb-commit }
```

**Unix-like** (macOS, Linux, BSD), in Terminal:

```
cd ~/Downloads/14.\ GWG && numb-form && numb-date && numb-commit
```

Check a request again, for example one a colleague prepared (all
systems):

```
numb-commit review
```

## WHEN SOMETHING GOES WRONG

**`F:\14. GWG is not reachable (is the drive connected?)`**
: Connect the network drive (open it once in File Explorer) and type the
  path again.

**`… is not a working folder: it has no template and "Übersicht" folder`**
: Give the folder that holds the template and the overview folder, not
  a folder inside it.

**`… changed after the request was made`**
: A file was edited after the request was written. Run numb-commit
  again.

**`… in the real folder changed after the request was made`**
: Someone else changed the spreadsheet on the network drive.
  Copy the real spreadsheet into the working copy again, run numb-date
  there, then numb-commit.

**`… already exists in the real folder and is never overwritten`**
: A colleague created the same file meanwhile. Compare the two by hand.

**`spreadsheet cell … changed; only report dates may change`**
: The working copy's spreadsheet was edited by hand. Make other changes
  in the real spreadsheet.

**`no request to review`**
: Run `numb-commit` first; `numb-commit review` checks the request it
  writes.

## SEE ALSO

[numb-form](usage.md), [numb-date](numb-date.md),
[configuration.md](configuration.md), [the guide](guide.md)
