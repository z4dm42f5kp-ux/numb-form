# numb-form

Fills section A of the *Appendix A - Risk Assesment* form for every
company in the overview spreadsheet that has no folder yet.

It fills the name, legal form, registry sentence, date and your name by
itself, and asks you for the address and the representatives.
You confirm every document.
numb-form never changes the spreadsheet; afterwards, `numb-date` writes
the date of each new form into its *Report created* cell, after you
agree, and changes nothing else.

Español: [docs/es/README.md](docs/es/README.md) ·
Deutsch: [docs/de/README.md](docs/de/README.md)

**New here? Read the [step-by-step guide](docs/guide.md).**
It needs no technical knowledge.

* Install: [docs/install.md](docs/install.md)
* Use: [numb-form](docs/usage.md), [numb-commit](docs/numb-commit.md),
  [numb-date](docs/numb-date.md)
* Configure: [docs/configuration.md](docs/configuration.md)
* Office suites and formats: [docs/formats.md](docs/formats.md)
* Containers: [docs/containers.md](docs/containers.md)
* Security: [SECURITY.md](SECURITY.md)
* AI agents: [AGENTS.md](AGENTS.md)

Start it inside the working folder (the one that contains
`3. Übersichten` and the template).

**Unix-like** (macOS, Linux, BSD), in Terminal:

```
cd /path/to/GWG && numb-form
```

**Windows** (10, 11): open the folder in File Explorer, type `powershell` in the
address bar, press Enter, then type:

```
numb-form
```
