# NUMB-TYPO(1)

## NAME

**numb-typo** — correct misspelt or cut-off cities and countries in the
company names of the overview spreadsheet

## SYNOPSIS

```
numb-typo [-f conf] [-root dir] [-timeout duration]
```

## DESCRIPTION

Company names in the Excel list often end in a city or country that is
misspelt or cut off, for example `Foo S.A. de C.V., Ciudad de Me` or
`Foo SAS, Bogta, Colomiba`.
numb-typo finds them and, after you agree, writes the corrected names:
`Foo S.A. de C.V., Ciudad de México`, `Foo SAS, Bogotá, Colombia`.

Run it inside the working folder, like numb-form, before numb-form, so
that the new folders and forms get the corrected names.
The spreadsheet must be `.xlsx` or `.xlsm`, and closed in Excel and
LibreOffice.

## WHAT IS CORRECTED

Only the parts after a comma (and after ` - `) that look like a city or
a country:

* **Cut off**: the last part of a cell, when it is the beginning of a
  known place: `Ciudad de Me` becomes `Ciudad de México`, `Nairob`
  becomes `Nairobi`.
* **Typo**: a part that is one or two letters away from exactly one
  known place: `Colomiba` becomes `Colombia`, `Medelin` becomes
  `Medellín`, `Duabi` becomes `Dubái`.
  Parts of fewer than five letters are never corrected; an extra letter
  is corrected only from seven letters, two errors only from nine, and
  the first letter must match.
* When several places fit (`Santa` could be Santa Cruz, Santa Marta,
  …), numb-typo asks which one, or you leave it.

Capitals are kept as written: `GUAYAQIL` becomes `GUAYAQUIL`.

What is never changed:

* the company's name itself, before the first comma;
* legal forms (`S.A.`, `Ltda`, `LLC`), numbers and words that are not
  places;
* a place that is already right; missing accents (`Bogota`) are left as
  written;
* the names of companies that already have a folder: renaming them
  would part them from their folder and form, so they are only listed in
  the log.

The list of places covers the countries in Spanish, English, Portuguese,
French and German, and the main cities of Latin America, Africa, the
Gulf, Europe and North America.

## WHAT HAPPENS

1. numb-typo goes through the company names and shows each correction:

   ```
   Foo S.A. de C.V., Ciudad de Me  (Customer ID D1001, row 315)
     → Foo S.A. de C.V., Ciudad de México
   ```

2. It lists them all and asks
   `Write these 3 correction(s)? [a]ll, [o]ne by one, [n]one:`.
   With `o` it asks for each one.
3. It writes them, reads the spreadsheet back from disk, shows the cells
   and asks `Is this correct? [a]ccept, [u]ndo:`.

Only those cells of the company-name column change, with the same checks
as numb-date: every other cell of every sheet is compared before
writing, each cell keeps its style, SHA-256 guards every step, and
*undo* restores the exact original, which is kept in the log folder
until you accept.

## OPTIONS

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

`<log folder>/<date>_<time>_<id>_typo/numb-typo.log`
: Log of the run: every correction, every name left alone and why, and
  the SHA-256 of the spreadsheet before and after.

## EXIT STATUS

0 on success or when there is nothing to correct, 1 on an error, 2 on a
usage error, 3 when stopped with Ctrl+C or by the user.

## EXAMPLES

Correct the names, then fill the forms and the dates.

**Windows** (10, 11), in PowerShell, inside the working folder:

```
numb-typo; if ($?) { numb-form }; if ($?) { numb-date }
```

**Unix-like** (macOS, Linux, BSD), in Terminal:

```
numb-typo && numb-form && numb-date
```

## SEE ALSO

[numb-form](usage.md), [numb-date](numb-date.md),
[numb-commit](numb-commit.md)
