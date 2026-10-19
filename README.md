# numb-form

Fills section A of the *Appendix A - Risk Assesment* form for every
company in the overview spreadsheet that has no folder yet.

You confirm every document, unless you configure it to run unattended.
The spreadsheet is never changed.

Español: [docs/es/README.md](docs/es/README.md) ·
Deutsch: [docs/de/README.md](docs/de/README.md)

* Install: [docs/install.md](docs/install.md)
* Use: [docs/usage.md](docs/usage.md)
* Configure: [docs/configuration.md](docs/configuration.md)
* Office suites and formats: [docs/formats.md](docs/formats.md)
* Containers: [docs/containers.md](docs/containers.md)
* Security: [SECURITY.md](SECURITY.md)
* AI agents: [AGENTS.md](AGENTS.md)

Start it in the folder above the working folder:

```
cd /path/to/folder-above && numb-form
```

On Windows, in Command Prompt:

```
cd /d "S:\path\to\folder-above" && numb-form
```

In Windows PowerShell:

```
Set-Location 'S:\path\to\folder-above'; if ($?) { numb-form }
```
