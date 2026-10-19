# Office suites and file formats

numb-form works the same with every office suite.
It edits the common open formats itself and uses office software already
on the computer only for formats it cannot edit safely.

## Templates

| Format | Written by | How numb-form fills it | Filled file |
| --- | --- | --- | --- |
| `.docx` | Microsoft Word, LibreOffice, Apache OpenOffice, Proton Docs, Apple Pages (export) | directly | `.docx` |
| `.odt` | LibreOffice, Apache OpenOffice, Collabora, Microsoft Word | directly | `.odt` |
| `.fodt` | LibreOffice (flat XML) | directly | `.fodt` |
| `.rtf` | every word processor | directly | `.rtf` |
| `.pages` | Apple Pages | converted, then filled | `.docx` (see below) |
| `.doc` | Word 97-2003 | legacy: converted, then filled | `.docx`, or `.doc` with `legacy_output = same` |

"Directly" means the file is never re-saved by another program: values
are inserted into the label's own text, and everything else stays byte
for byte as it was.
The result is read back and checked after every fill.

When the working folder holds the template in several formats, the
first one in the table wins.

## Spreadsheets

| Format | Written by | How numb-form reads it |
| --- | --- | --- |
| `.xlsx`, `.xlsm` | Microsoft Excel, LibreOffice Calc, Proton Sheets, Apple Numbers (export) | directly; macros are never run |
| `.ods`, `.fods` | LibreOffice Calc, Apache OpenOffice | directly |
| `.csv`, `.tsv` | every spreadsheet program | directly; UTF-8 or Windows-1252, `;` `,` or tab |
| `.numbers` | Apple Numbers | converted in memory |
| `.xls` | Excel 97-2003 | legacy: directly, with numb-form's own reader |

The spreadsheet is only read, never written.

## Converted formats

`.doc`, `.pages` and `.numbers` are converted with software already on
the computer:

* LibreOffice or Apache OpenOffice, on every system.
  numb-form searches the usual install locations; set `soffice = <path>`
  if it is elsewhere.
* On macOS, Apple Pages and Numbers, when installed
  (`pages_converter` and `numbers_converter`, default `auto`).

Conversion runs in a private temporary folder with a throwaway office
profile, and everything is deleted afterwards.
Nothing is downloaded.

### Apple Pages

Pages files use an undocumented format.
numb-form therefore reads a `.pages` template by converting it to Word
format, fills that, and saves the filled forms as `.docx`, which Pages
opens.
Check the layout of the first filled form.

On macOS with Pages installed, `pages_write = yes` saves the filled forms
as `.pages` through Pages itself.
This is **experimental**: it is tested, but off by default, and every
file is converted back and checked after saving.
If you do not need `.pages` files, leave it off.

### Legacy formats (.doc, .xls)

Legacy binary formats are disabled by default.
An administrator enables them in the system configuration file
(see [configuration.md](configuration.md)):

```
legacy_formats = yes
```

Runs that use a legacy format are always fully offline.
The `.xls` reader is numb-form's own and is fuzz-tested continuously.

## Office suites

* **Microsoft 365 / Office**: save the template as `.docx` and the
  overview as `.xlsx` (the defaults).
* **LibreOffice** and **Apache OpenOffice**: `.odt` and `.ods` work
  directly; so do `.docx` and `.xlsx`.
* **Proton Drive (Docs and Sheets)**: download the template as `.docx`
  and the overview as `.xlsx` into the working folder.
  This also works on systems without a desktop office suite, including
  the BSDs, from a web browser.
* **Apple Pages and Numbers**: use `.pages` and `.numbers` as described
  above, or export to `.docx` and `.xlsx` (File → Export To).
* **The BSDs**: LibreOffice is available as a package
  (`pkg install libreoffice`, `pkg_add libreoffice`,
  `pkgin install libreoffice`); Proton Docs in a browser needs no
  installation.
