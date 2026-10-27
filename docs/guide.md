# Step-by-step guide

This guide is for everyone who fills in the risk assessment forms on
**Windows** (10, 11).
You do not need any technical knowledge.
You only copy, paste and press Enter.
On **macOS, Linux or BSD**, install with [install.md](install.md); the
questions and answers are the same, and commands joined with `;` here
are joined with `&&` there.

## Words used in this guide

**working folder**
: The team's folder with the empty form, the Excel list and one folder
  per company, for example `F:\14. GWG`.
  The guide calls it GWG.

**template**
: The empty form, `Appendix A - Risk Assesment.docx`, directly in GWG.

**Excel list**
: The spreadsheet in the folder `3. Übersichten`, with one company per
  row in the column *Customer Name*.

**PowerShell**
: The blue window where you type commands.
  Open it in a folder by typing `powershell` in the address bar of File
  Explorer and pressing **Enter**.

**command**
: A line you type or paste into PowerShell and confirm with **Enter**,
  for example `numb-form`.

**path**
: Where a file or folder is, written out in full, for example
  `C:\Users\you\Downloads\14. GWG`.
  In File Explorer, hold **Shift**, right-click the folder and choose
  *Copy as path*.

**log**
: A text file in which the programs note what they did.
  They print where it is at the end, after `Log:`.

## What numb-form does

Your team keeps a working folder, for example **GWG**.
It looks like this:

```
GWG
├── 3. Übersichten
│   └── the Excel list of companies
├── Appendix A - Risk Assesment.docx      (the empty form)
├── Foo GmbH                              (one folder per company
└── Bar S.A.                               that already has a form)
```

numb-form goes through the Excel list from top to bottom.
For every company that does not have a folder yet, it:

1. creates a folder with the company's name,
2. puts a copy of the form in it, named
   `Appendix A - Risk Assesment - <company>.docx`,
3. fills in part A of the form: the name, legal form, registry
   sentence, date and your name by itself, and the address and
   representatives you type,
4. shows you what it wrote and waits for your OK.

Companies that already have a folder are skipped.
The Excel list is never changed.
Part B of the form is never touched.

## 1. Install it (once)

1. Open the **Start** menu, type `PowerShell` and open
   **Windows PowerShell**.
   A blue window appears.
2. Copy this whole line, paste it into the blue window (right-click
   pastes) and press **Enter**:

   ```
   & { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Set-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; 'OK, open a new PowerShell window' }
   ```

3. Wait until it says `OK, open a new PowerShell window`.
   Close the window.

You do not need administrator rights.
If it says `CHECKSUM MISMATCH`, the download was damaged or changed:
do not continue, and tell whoever looks after the tool.

To update numb-form later, see [Reinstall or update](#reinstall-or-update).

## 2. Save your name (once)

Your name goes into the field *Person Processing the Form*.
Open Windows PowerShell again and type, with your own name:

```
numb-form config set person "Erika Mustermann"
```

It answers `Saved person = Erika Mustermann`.
If you skip this step, numb-form asks for your name every time.

## 3. Fill in the forms (every time)

1. Open the **GWG** folder in File Explorer, the folder that contains
   `3. Übersichten`.
2. Click once into the **address bar** at the top of the window (where
   the folder's path is shown), type `powershell` and press **Enter**.
   A blue window opens that is already "in" the GWG folder.
3. Type `numb-form` and press **Enter**.

numb-form first shows what it found:

```
Working folder: H:\GWG
Spreadsheet:    3. Übersichten\Kundenliste.xlsx (read only)
Template:       Appendix A - Risk Assesment.docx
```

If that is right, it goes through the companies one by one.

## 4. The questions, and what to answer

Many questions show a suggestion in square brackets, like `[GmbH]` or
`[y/N]`.
**Pressing Enter accepts the suggestion** (for `[y/N]` that is *no*, for
`[Y/n]` it is *yes*).
To give a different answer, type it and press Enter.

Each company starts with a line like
`[3/11] Foo GmbH  (Customer ID D1001, row 5)`: the third of eleven
companies, its customer number and its row in the Excel list.

The legal form is taken from the company name without asking.
numb-form then asks you for the address and the representatives; it
does not look anything up online.
To find them, give an AI assistant [the AI prompt](ai-prompt.md) and
the company's name, and check its answer against your internal sources.
*Commercial Registry Number* always gets the sentence "No commercial
registry number was found online consult internal sources for further
information"; to put the number instead, choose `e` (edit) and field 4
at the end.

| numb-form asks | What it means | What you usually do |
| --- | --- | --- |
| `Full name of the person processing the form:` | Only asked if you skipped step 2 | Type your name; Enter at the next question uses it for every form in this run |
| `Is the existing folder "Foo Gmbh" this company's folder? (...) [y/N]` | A folder with a very similar name already exists | `y` if it is the same company (nothing is created), otherwise Enter. Start with `numb-form -y` to answer `y` to all of these: each one is shown as `already has a folder: … (similar name, -y)` |
| `Legal form (the name has none; Enter = "No legal form was found…", '-' = blank):` | Asked only when the company name has no legal form such as `GmbH` or `S.A.` | Type it; Enter alone writes "No legal form was found online ..."; `-` leaves it blank |
| `Headquarters / Address (one line, Enter = "No address was found…", '-' = blank):` | The company's address, from your own sources | Type or paste it on one line; Enter alone writes "No address was found online ..."; `-` leaves it blank |
| `Representatives (Enter = "No representatives were found…", '-' = blank):` | The people who may sign for the company, written `Full Name (Role 1, Role 2), Full Name (Role)` with the roles in English | Type or paste them; Enter alone writes "No representatives were found online ..."; `-` leaves the field blank |
| `Is this correct? [a]ccept, [e]dit a field, [r]eject:` | The form has been written; numb-form shows part A as the form now reads | Type `a` if it is right; `e` to change one line; `r` to throw it away. There is no default here: you must type a letter |

At the end numb-form shows a summary, for example
`Done: 3 accepted, 8 already had a folder, 0 rejected, 0 skipped, 0 errors.`
and where the log is.

You can stop at any time with **Ctrl+C**.
Forms you already accepted stay; the one you were working on is removed.

## 5. What changed on the drive

* One new folder per accepted company, with one filled form in it.
* Nothing else.
  No existing file or folder is changed or deleted, and the Excel list
  is never changed by numb-form (see step 7 for numb-date).
* A rejected form leaves nothing behind.

To undo an accepted form, delete its company folder.

A record of everything numb-form did (the log) is kept in your own user
folder, not in GWG: `%LOCALAPPDATA%\numb-form\runs`.
The last line of every run shows the exact file.

## 6. Check your form template (once, recommended)

Before the first real run, let numb-form test the template.
In a PowerShell window opened in GWG (see step 3), type:

```
numb-form check-template "Appendix A - Risk Assesment.docx" "$env:USERPROFILE\Desktop\sample.docx"
```

It reads the template without changing it and saves a filled example on
your desktop.
Open `sample.docx` and check that the values are in the right places.
Delete it afterwards.

## 7. Fill in the report dates (after numb-form)

The Excel list has a column such as *Report created AB*.
`numb-date` fills it for you.
In the same PowerShell window in GWG, type:

```
numb-date
```

It goes through the list.
Companies whose *Report created* cell already contains a date are
skipped.
For the others it opens the company's form (only to read it), takes the
*Date of the evaluation* and lists, for example:

```
Dates to write into Kundenliste.xlsx (column "Report created AB"):
  1  F13    16-10-2026  Foo Ltda  (Foo Ltda\Appendix A - Risk Assesment - Foo Ltda.docx)
```

It never asks you for a date.
Companies without a folder, a form or a date in the form are skipped, and
listed with the reason, for example
`Foo GmbH — no form in the folder Foo GmbH`, so you can fill them by hand.

| numb-date asks | What you do |
| --- | --- |
| `Is the folder "Fooo Inc" this company's folder? (...) [y/N]` | `y` if it is the same company, otherwise Enter. Start with `numb-date -y` to answer `y` to all of these |
| `Write these 1 date(s)? Nothing else in the spreadsheet changes. [a]ccept, [e]dit a date, [r]eject:` | `a` writes them; `e` changes one date (or leaves a company out with `-`); `r` writes nothing |
| `Is this correct? [a]ccept, [u]ndo:` | numb-date shows the cells as the file now reads; `a` to keep them, `u` to put the file back exactly as it was |

**Close the Excel list first.**
numb-date stops if the file is open.
Only the *Report created* cells change: no other cell, no formatting.

## 8. GWG on a network drive: work on a copy, then check it

When GWG lives on a network drive, for example `F:\14. GWG`, you can
work on a copy on your computer.
numb-commit then shows exactly what the copy adds to the real folder.

Once, in PowerShell, tell numb-commit where the real folder is:

```
numb-form config set commit_to "F:\14. GWG"
```

Every time:

1. In File Explorer, copy `F:\14. GWG` into your **Downloads** folder.
2. Open the copy (`Downloads\14. GWG`), type `powershell` in the address
   bar and press Enter.
3. Run numb-form and numb-date there as usual.
4. Type `numb-commit` and press Enter.

numb-commit asks for the two folders (unless `commit_from` and
`commit_to` are set), compares the copy with `F:\14. GWG` and writes
**`numb-commit-request.txt`** into the copy: the list of exactly what
is new, with the folder tree before and after.
It then checks the list and ends with a verdict:

* `Verdict: OK to commit.` Everything is in order.
* `Verdict: commit after checking the lines marked check.` Read the
  lines marked `check`, for example a form with a blank address.
* `Verdict: do not commit.` The lines marked `FAIL` say what is wrong.

numb-commit does not copy anything yet.
Copy the new company folders listed under *New folder* into
`F:\14. GWG` yourself, and the Excel list if it is listed under
*Replaced*.
Never copy the template, logs or Office lock files.
`numb-commit review` repeats the check at any time, for example on a
copy a colleague prepared.
The [numb-commit manual](numb-commit.md) explains every message.

## Prefer to review everything before it lands in GWG?

Switch to review mode once:

```
numb-form config set mode review
```

From then on, numb-form copies GWG into your **Downloads** folder and
works only on the copy.
Look through the new folders there.
When everything is right, type `numb-commit`: it copies the new folders
into the real GWG folder after asking you once more.
Switch back with `numb-form config set mode direct`.

## Reinstall or update

The easy way: open Windows PowerShell and type one of these:

| Type | What you get |
| --- | --- |
| `numb-form update stable` | the latest stable release (use this for real work) |
| `numb-form update rolling` | the current rolling build: the newest changes, rebuilt automatically after every update on GitHub |
| `numb-form update` | the newest version of the kind you already have |
| `numb-form update v0.1.1` | exactly that release (any number from the Releases page) |
| `numb-form rollback` | back to the version you had before the last update; type it again to return |

It downloads the version, checks it, replaces the program and prints
the new version, for example `Updated to numb-form v0.1.2 (stable)`.
Your saved settings (your name, the heading, the mode) are kept.

Versions 0.1.1 and older do not have `numb-form update`.
For those, or if `numb-form update` cannot connect, use one of the two
lines below.
Do not type them: open this page in your browser, click the copy icon at
the right edge of the grey box, and right-click in the PowerShell window
to paste.
Close every numb-form window first.

**Stable version** (the newest published release; use this for real work):

```
& { $ErrorActionPreference = 'Stop'; if (Get-Process numb-form, numb-commit, numb-date -ErrorAction SilentlyContinue) { throw 'Close every numb-form window first, then run this again' }; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Push-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Remove-Item -Force -ErrorAction SilentlyContinue numb-form.exe, numb-commit.exe, numb-date.exe; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS; Pop-Location }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, reinstalled. Your settings are kept. Open a new PowerShell window' }
```

**Rolling version** (the newest changes, rebuilt automatically after
every update on GitHub; use this to try a fix right away):

```
& { $ErrorActionPreference = 'Stop'; if (Get-Process numb-form, numb-commit, numb-date -ErrorAction SilentlyContinue) { throw 'Close every numb-form window first, then run this again' }; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Push-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Remove-Item -Force -ErrorAction SilentlyContinue numb-form.exe, numb-commit.exe, numb-date.exe; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS; Pop-Location }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, reinstalled. Your settings are kept. Open a new PowerShell window' }
```

Both print the version they installed, for example `v0.1.1` (stable) or
`rolling-3f976cc` (rolling; older rolling builds say `edge-…`), and end
with `OK, reinstalled`.
Open a new PowerShell window afterwards.

If PowerShell answers `Unexpected token` or `Missing ...` in red, the
line was changed while copying (usually a character was lost): copy it
again with the copy icon.

If the rolling line says `CHECKSUM MISMATCH`, a new build was being
uploaded at that moment: wait five minutes and run it once more.
If it happens again, or with the stable line, stop and tell whoever looks
after the tool.

## Read this manual in the window

Type `numb-form man` to read everything in the blue window.
Keys work at once, without Enter: **Space** next page, **b** previous
page, **↑ ↓** one line, **/** search, **q** quit, **h** all keys.
`numb-form man guide` shows only this guide; add `de` or `es` for German
or Spanish: `numb-form man guide de`.
In Command Prompt (the black window) `man numb-form` works too, after
running `numb-form man -install` once.

## When something goes wrong

**`numb-form` is not recognized**
: Close the window and open a new one.
  If it still happens, install again (step 1).

**`... is not a working folder`** or **`several folders could be the working folder`**
: The blue window was not opened in GWG.
  Repeat step 3 from inside the GWG folder.
  You can also name the folder directly, for example
  `numb-form -root "H:\GWG"`.

**`template "Appendix A - Risk Assesment" ... not found`**
: The empty form must be directly in GWG, not in a subfolder, and its
  name must start with `Appendix A - Risk Assesment`.

**`no folder whose name contains "Übersicht" found`**
: GWG must contain the folder `3. Übersichten` (any name containing
  `Übersicht` works).

**`2 spreadsheets in ...; expected exactly one`**
: `3. Übersichten` must contain exactly one Excel file.
  Move the others elsewhere.

**`no sheet has a "Customer Name" header cell`**
: numb-form looks for the heading `Customer Name` and takes the names
  below it.
  The message lists the headings it did find.
  If your list uses another heading, for example `Firmenname`, save it
  once with `numb-form config set header "Firmenname"`.
  Versions 0.1.1 and older look for `Company Name` instead; there, run
  `numb-form config set header "Customer Name"` once.

**`Time limit of 30m0s reached`**
: A time limit was set with `timeout` in the settings or with
  `-timeout`; there is none by default.
  Start it again; companies already done are skipped.

**Version 0.1.0 does not start inside GWG**
: Update (see [Reinstall or update](#reinstall-or-update)), or use
  `numb-form -root "H:\GWG"`.

**Running scripts or programs is blocked by your company**
: Ask your IT department to allow
  `%LOCALAPPDATA%\Programs\numb-form\numb-form.exe`.

For everything else, see the [reference manual](usage.md) and the
[settings](configuration.md).

## Reporting a problem

If something goes wrong and this guide does not help, report it on
GitHub: *Issues*, *New issue*, *Bug report*.
A wrong or unclear sentence in this guide is a problem too: choose
*Documentation error*.

The form asks for:

1. the version: type `numb-form -version` and copy what it prints,
2. every step you took, with what you typed,
3. the text in the window, copied, including every error message,
4. what you expected instead,
5. the end of the log, if the program printed one.

The reports are public.
Before you send one, replace every customer name, person's name and
address with a made-up one, for example `Foo GmbH`.
Never attach a form or the Excel list.

A problem that could expose or damage customer data is a security
problem: report it privately as [SECURITY.md](../SECURITY.md) says.
