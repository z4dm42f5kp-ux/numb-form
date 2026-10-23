# NUMB-COMMIT(1)

## NAME

**numb-commit** — copy the new forms into the real working folder,
after showing and checking exactly what changes

## SYNOPSIS

```
numb-commit [-f conf] [-from dir] [-to dir]
numb-commit [-f conf] [-from dir] [-to dir] review
numb-commit [-f conf] [-run dir] [-downloads dir]
```

## DESCRIPTION

numb-form and numb-date work on a folder on your own computer.
numb-commit brings the result into the real working folder, usually on
a network drive, and nothing else.

There are two ways to work; numb-commit handles both.

**A working copy** (recommended for a network drive).
You copy the real folder, for example `F:\14. GWG`, to your computer,
for example `C:\Users\you\Downloads\14. GWG`, run numb-form and
numb-date there, and then numb-commit inside the copy.
Tell numb-commit once where the real folder is:

```
numb-form config set commit_to "F:\14. GWG"
```

**A review run** (`mode = review`).
numb-form makes the copy itself in `Downloads\numb-form-runs`, and
numb-commit, started anywhere, commits the newest run.
This is the older way; see [usage.md](usage.md).
It is used when `commit_to` is not set and neither `-to` nor `review` is
given.

## WHAT NUMB-COMMIT DOES WITH A WORKING COPY

1. It compares the working copy with the real folder, file by file.
   Files with the same size and time are the same; for the others, the
   SHA-256 checksum decides.
2. It writes the list of changes to **`numb-commit-request.txt`** in the
   working copy, and shows it.
3. It reviews the list (see REVIEW) and shows the verdict.
   When the review fails, it stops: nothing is copied.
4. It asks `Copy this into F:\14. GWG now? [y/N]`.
   Only `y` copies.
5. It copies, checking every file's SHA-256 again, and marks the request
   as applied.
6. It reviews the real folder once more: every file must be there with
   its checksum.

What is copied:

* **New folders** (normally one per new company): built under a
  temporary name in the real folder and renamed when complete, so a
  folder appears whole or not at all.
* **New files** in folders that already exist.
* **The overview spreadsheet**, when numb-date added report dates.
  Its old version is kept in the log folder first.

What is never copied:

* the template (`Appendix A - Risk Assesment.docx`), even if it changed;
* logs, `manifest.json`, `numb-form-runs` folders and the request file;
* Office and LibreOffice lock files (`~$…`, `.~lock.…#`), hidden and
  temporary files, `Thumbs.db`, `desktop.ini`;
* any other file that differs from the real folder: numb-commit never
  overwrites a file except the spreadsheet, and lists the others under
  *Left alone*.

Nothing in the real folder is ever deleted.

## THE REQUEST FILE

`numb-commit-request.txt` lies in the working copy, never in the real
folder.
It says, in plain words, exactly what changes:

```
numb-commit request: what numb-commit copies into the real folder

From:   C:\Users\you\Downloads\14. GWG
To:     F:\14. GWG
Made:   23/10/2026 14:05
Status: waiting

New folder  Foo GmbH
  + Foo GmbH\Appendix A - Risk Assesment - Foo GmbH.docx  (18 KB, sha256 4be1…)

Replaced  3. Übersichten\Kundenliste.xlsx
  sha256 9f02… in the real folder now
  sha256 c771… afterwards
  the replaced version is kept in the log folder

Left alone
  - Appendix A - Risk Assesment.docx  (not copied: the template)
  - Foo GmbH\~$pendix A - Risk Assesment - Foo GmbH.docx  (not copied: an Office lock file)
```

*Status* is `waiting` until the copy is done, then `applied`.
Below the readable part, a data section repeats the list for
numb-commit; do not edit it.
A copy of the applied request is kept in the log folder.

## REVIEW

`numb-commit review` checks the request without copying anything.
numb-commit runs the same review itself before it asks.
It is meant for tight deadlines and for colleagues who would rather not
read checksums: it reads the request so you do not have to.

Before the request is applied, the review checks that:

* every file to copy still has the checksum in the request (nothing was
  edited in the working copy since);
* none of them exists in the real folder yet;
* the real spreadsheet still has the checksum in the request (nobody
  else changed it in the meantime);
* the spreadsheet only gains report dates: only empty or digit-free
  cells of the *Report created* column change, and each new value looks
  like `DD-MM-YYYY`;
* each new folder belongs to a company in the Excel list and holds a
  risk assessment form whose section A names that company and has no
  blank line;
* nothing that must stay behind (template, logs, lock files) is listed.

After the request is applied, the review checks that every listed file
is in the real folder with its checksum.

Each line is marked `ok`, `check` (look at it before committing) or
`FAIL`, followed by the verdict:

| Verdict | Meaning |
| --- | --- |
| `OK to commit.` | Every check passed. |
| `commit after checking the lines marked check.` | Nothing is wrong with the files, but something deserves a look, for example a blank address. |
| `do not commit.` | Something changed or does not belong; numb-commit refuses to copy. Run numb-commit again after fixing it. |

## OPTIONS

`-from dir`
: The working copy. Default: the current folder.

`-to dir`
: The real working folder. Default: `commit_to` in the configuration.

`review`
: Only review the request in the working copy; copy nothing.

`-run dir`, `-downloads dir`
: Review runs only: the run to commit, and where `numb-form-runs` is.

`-f file`
: Read *file* after the system and per-user configuration files.

`-allow-non-tty`
: Read answers from a pipe. For automated tests only.

`-version`
: Print the version.

## FILES

`<working copy>/numb-commit-request.txt`
: The request: what is (or was) copied.

`<log folder>/<date>_<time>_<id>_commit/numb-commit.log`
: Log of the copy, with every checksum.

`<log folder>/<date>_<time>_<id>_commit/replaced/`
: The real folder's previous version of every replaced file.

`<log folder>/<date>_<time>_<id>_commit/numb-commit-request.txt`
: The applied request.

The log folder is `log_dir`, by default `%LOCALAPPDATA%\numb-form\runs`
(Windows), `~/Library/Logs/numb-form` (macOS) or
`~/.local/state/numb-form/runs` (Linux, BSD).

## EXIT STATUS

0 when the copy is done or there is nothing to do, 1 on an error or a
failed review, 2 on a usage error, 3 when stopped with Ctrl+C or by the
user.

## EXAMPLES

Once: say where the real folder is (all systems):

```
numb-form config set commit_to "F:\14. GWG"
```

Every time, inside the working copy.

**Windows** (10, 11), in PowerShell:

```
cd "$env:USERPROFILE\Downloads\14. GWG"; numb-form; if ($?) { numb-date }; if ($?) { numb-commit }
```

**Unix-like** (macOS, Linux, BSD), in Terminal:

```
cd ~/Downloads/14.\ GWG && numb-form && numb-date && numb-commit
```

Let numb-commit check a request, for example one a colleague prepared,
without copying (all systems, inside the working copy):

```
numb-commit review
```

## WHEN SOMETHING GOES WRONG

**`the real folder F:\14. GWG is not reachable`**
: Connect the network drive (open it once in File Explorer) and run
  numb-commit again.

**`… changed after the request was made; run numb-commit again`**
: A file was edited after the request was written. Run numb-commit
  again: it writes a new request.

**`… in the real folder changed after the request was made`**
: Someone else changed the spreadsheet on the network drive.
  numb-commit does not overwrite their work. Copy the real spreadsheet
  into the working copy again, run numb-date there, then numb-commit.

**`… already exists in the real folder`**
: A colleague created the same folder meanwhile. Compare the two and
  remove the one in the working copy if it is the same company.

**`spreadsheet cell … changed; only report dates may change`**
: The working copy's spreadsheet was edited by hand. numb-commit only
  carries report dates; make other changes in the real spreadsheet.

## SEE ALSO

[numb-form](usage.md), [numb-date](numb-date.md),
[configuration.md](configuration.md), [the guide](guide.md)
