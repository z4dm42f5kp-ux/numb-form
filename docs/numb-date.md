# NUMB-DATE(1)

## NAME

**numb-date** — fill the *Report created* column of the overview
spreadsheet from the companies' forms

## SYNOPSIS

```
numb-date [-f conf] [-y] [-root dir] [-timeout duration]
```

## DESCRIPTION

numb-date is started like numb-form, inside the working folder (or one
level above it), usually right after numb-form.
It is the only program that writes to the overview spreadsheet, and it
writes nothing but report dates.
The spreadsheet must be `.xlsx` or `.xlsm`, and closed in Excel and
LibreOffice.

1. The company column is found by `header`, then by common names
   (`Customer Name`, `Customer`, `Company Name`, `Kunde`, `Cliente` and
   others, ignoring case); the date column in the same header row by
   `Report created AB`, `Report created`, `Report date`, `Created` and
   similar names, or any heading that mentions a report and its creation
   or date.
2. Rows whose date cell contains a digit are skipped: they already have
   a date.
3. For the other rows, the company's folder is found as numb-form finds
   it (near matches are asked, or taken with `-y`).
   The `.docx`, `.odt`, `.fodt` or `.rtf` forms in the folder (and one
   level below) are read, the risk assessment form first: the *Date of
   the evaluation* of section A, or, in forms laid out differently, the
   date written after a date label.
   Dates such as `16/10/2026`, `16.10.2026`, `2026-10-16` and
   `16 October 2026` are understood.
   When no date is found, you type it or skip the company.
4. The dates are listed and written as `DD-MM-YYYY` after you agree, then
   read back from disk and shown; you accept them or undo the write.

Forms are only read.

## WHAT CHANGES IN THE SPREADSHEET

Only the listed cells change.
They keep their own cell style and get no new formatting; every other
part of the workbook is copied byte for byte, and cells with formulas are
left alone.
Before writing, every cell of every sheet of the new workbook is compared
with the original: anything else that differs stops numb-date.
SHA-256 guards every step: the file must still have the checksum it had
when it was read, the temporary copy beside it must match before it
replaces the file, and the file must match afterwards.
The original is kept in the log folder until you accept; *undo* puts it
back exactly.

| numb-date asks | What you do |
| --- | --- |
| `Is the folder "…" this company's folder? (...) [y/N]` | `y` if it is the same company, otherwise Enter |
| `Date (DD-MM-YYYY, Enter = skip):` | The form has no readable date: type it, or Enter to skip |
| `Write these … date(s)? Nothing else in the spreadsheet changes. [y/N]` | `y` to write them; Enter writes nothing |
| `Is this correct? [a]ccept, [u]ndo:` | `a` keeps them, `u` restores the file exactly |

## OPTIONS

`-y`
: Take a folder with a similar name as the company's folder without
  asking.

`-root dir`
: Use *dir* as the working folder instead of detecting it.

`-timeout duration`
: Stop safely after *duration*, for example `30m`.

`-f file`
: Read *file* after the system and per-user configuration files.

`-allow-non-tty`
: Read answers from a pipe. For automated tests only.

`-version`
: Print the version.

## FILES

`<log folder>/<date>_<time>_<id>_date/numb-date.log`
: Log of the run, with the SHA-256 of the spreadsheet before and after.

`<log folder>/<date>_<time>_<id>_date/before-<spreadsheet>`
: The original spreadsheet, kept until you accept or undo.

## EXIT STATUS

0 on success or when there is nothing to write, 1 on an error, 2 on a
usage error, 3 when stopped with Ctrl+C or by the user.

## EXAMPLES

**Windows** (10, 11), in PowerShell, inside the working folder:

```
numb-form; if ($?) { numb-date }
```

**Unix-like** (macOS, Linux, BSD), in Terminal:

```
numb-form && numb-date
```

## WHEN SOMETHING GOES WRONG

**`… is open in Excel or LibreOffice`**
: Close the file and run numb-date again.

**`numb-date writes only .xlsx and .xlsm spreadsheets`**
: Save the overview as `.xlsx` once.

**`no "Report created" column next to the company names`**
: The message lists the headings found. Rename the date column's
  heading to `Report created` or similar.

## SEE ALSO

[numb-form](usage.md), [numb-form.conf](configuration.md),
[../SECURITY.md](../SECURITY.md)
