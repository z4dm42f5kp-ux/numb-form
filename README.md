# numb-form

Fills section A of the *Appendix A - Risk Assesment* form for every
company in the `3. Übersichten` spreadsheet that has no folder yet.

It works on a copy in your Downloads folder.
You confirm every document.
`numb-commit` copies the reviewed folders back.

* Install: [docs/install.md](docs/install.md)
* Use: [docs/usage.md](docs/usage.md)
* Security: [SECURITY.md](SECURITY.md)

```
cd "C:\path\to\folder-above-the-working-folder" && numb-form && numb-commit
```

`numb-commit` asks before it writes anything.
Review the copy in Downloads before you answer.
