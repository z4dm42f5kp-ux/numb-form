# Step-by-step guide

This guide is for everyone who fills in the risk assessment forms.
You do not need any technical knowledge.
You only copy, paste and press Enter.

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
3. fills in part A of the form (name, legal form, address,
   representatives, date and your name),
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

Each company starts with a line like `[3/11] Foo GmbH  (Sheet1!C5)`:
the third of eleven companies, and the cell it was taken from.

Then numb-form looks the company up by itself, which takes a few
seconds: official registers, the company's own website (it finds it
alone), official numbers printed there, government gazettes and a web
search.
It shows the website it used, any register entry printed there (for
example `HRB 12345`), and how many addresses and representatives it
found.
The legal form is taken from the company name without asking.
*Commercial Registry Number* always gets the sentence "No commercial
registry number was found online consult internal sources for further
information"; to put the number instead, choose `e` (edit) and field 4
at the end.

| numb-form asks | What it means | What you usually do |
| --- | --- | --- |
| `Full name of the person processing the form:` | Only asked if you skipped step 2 | Type your name; Enter at the next question uses it for every form in this run |
| `Is the existing folder "Foo Gmbh" this company's folder? (...) [y/N]` | A folder with a very similar name already exists | `y` if it is the same company (nothing is created), otherwise Enter |
| `Legal form (Enter = keep, '-' = blank) [GmbH]` | Asked only when the company name has no legal form such as `GmbH` or `S.A.`; the suggestion comes from the register or website | Enter |
| `Headquarters / Address candidates` and a numbered list, then `Choose a number, 'm' to type it, 'b' to leave blank [1]` | Addresses found, the most reliable (official registers) first; each shows its source | Enter takes number 1; type another number; `m` to type the address yourself; `b` to leave it blank |
| `Address (Enter = keep, or correct it) [...]` | The address you picked, so you can correct a typo | Enter, or type the corrected address |
| `Add the caution text to the address (not fully verified)? [y/N]` | Adds the sentence "Caution: This field might differ from internal sources ..." The same question can come for the representatives | Press Enter: the suggestion already depends on how reliable the source is |
| `Authorized representative candidates` and a numbered list, then `Numbers to include (e.g. 1,3), 'a' all, 'm' type them, 'n' none found [1,2]` | People found, each marked `[verified]` (official register) or `[unverified]` | Enter takes the verified ones; type numbers like `1,3`; `a` for all; `m` to type them yourself; `n` writes "No representatives were found online ..." |
| `Is this correct? [a]ccept, [e]dit a field, [r]eject:` | The form has been written; numb-form shows part A as the form now reads | Type `a` if it is right; `e` to change one line; `r` to throw it away. There is no default here: you must type a letter |

At the end numb-form shows a summary, for example
`Done: 3 accepted, 8 already had a folder, 0 rejected, 0 need review, 0 skipped, 0 errors.`
and where the log is.

You can stop at any time with **Ctrl+C**.
Forms you already accepted stay; the one you were working on is removed.

## 5. What changed on the drive

* One new folder per accepted company, with one filled form in it.
* Nothing else.
  No existing file or folder is changed or deleted, and the Excel list
  is never changed.
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
& { $ErrorActionPreference = 'Stop'; if (Get-Process numb-form, numb-commit -ErrorAction SilentlyContinue) { throw 'Close every numb-form window first, then run this again' }; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Push-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Remove-Item -Force -ErrorAction SilentlyContinue numb-form.exe, numb-commit.exe; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS; Pop-Location }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, reinstalled. Your settings are kept. Open a new PowerShell window' }
```

**Rolling version** (the newest changes, rebuilt automatically after
every update on GitHub; use this to try a fix right away):

```
& { $ErrorActionPreference = 'Stop'; if (Get-Process numb-form, numb-commit -ErrorAction SilentlyContinue) { throw 'Close every numb-form window first, then run this again' }; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Push-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Remove-Item -Force -ErrorAction SilentlyContinue numb-form.exe, numb-commit.exe; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS; Pop-Location }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, reinstalled. Your settings are kept. Open a new PowerShell window' }
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

Type `numb-form man` to read everything in the blue window, one screen
at a time (Enter: next screen; q and Enter: quit).
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

**`Time limit of 25m0s reached`**
: A run stops after 25 minutes for safety.
  Start it again; companies already done are skipped.

**Version 0.1.0 does not start inside GWG**
: Update (see [Reinstall or update](#reinstall-or-update)), or use
  `numb-form -root "H:\GWG"`.

**Running scripts or programs is blocked by your company**
: Ask your IT department to allow
  `%LOCALAPPDATA%\Programs\numb-form\numb-form.exe`.

For everything else, see the [reference manual](usage.md) and the
[settings](configuration.md).
