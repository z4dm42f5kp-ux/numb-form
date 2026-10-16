#!/usr/bin/env python3
"""Translate the English documentation to Spanish and German with Argos
Translate, without touching code blocks, inline code, commands or URLs.

Run by .github/workflows/translate.yml. The English files are
authoritative; the translations are regenerated from them.

Supply chain:
* Python packages are installed with --require-hashes from requirements.txt.
* The Argos package index is read at a pinned commit.
* Each model's SHA-256 is pinned in models.sha256. A model that is not
  pinned yet is recorded on first use (trust on first use, visible in the
  commit that adds it); a model whose hash differs from its pin is refused.
"""

import hashlib
import json
import os
import pathlib
import re
import sys
import tempfile
import urllib.request

ROOT = pathlib.Path(__file__).resolve().parents[2]
HERE = pathlib.Path(__file__).resolve().parent
INDEX = ("https://raw.githubusercontent.com/argosopentech/argospm-index/"
         "ff90de60728f7c1338ff6b75974e4c89b2442d22/index.json")
PINS = HERE / "models.sha256"
LANGS = {"es": "Español", "de": "Deutsch"}
SOURCES = ["README.md", "SECURITY.md", "docs/install.md", "docs/usage.md", "docs/containers.md"]
NOTE = {
    "es": "> Traducción automática (Argos Translate) del [original en inglés]({src}). "
          "La versión en inglés es la que vale.",
    "de": "> Maschinelle Übersetzung (Argos Translate) des [englischen Originals]({src}). "
          "Maßgeblich ist die englische Fassung.",
}
MAX_MODEL = 500 * 1024 * 1024


def fetch(url, limit):
    if not url.startswith("https://"):
        raise SystemExit(f"refusing non-HTTPS URL {url}")
    req = urllib.request.Request(url, headers={"User-Agent": "numb-form-docs/1"})
    with urllib.request.urlopen(req, timeout=120) as r:  # noqa: S310 (https only)
        data = r.read(limit + 1)
    if len(data) > limit:
        raise SystemExit(f"{url} is larger than {limit} bytes")
    return data


def load_pins():
    pins = {}
    if PINS.exists():
        for line in PINS.read_text().splitlines():
            line = line.strip()
            if line and not line.startswith("#"):
                digest, name = line.split(None, 1)
                pins[name] = digest
    return pins


def install_models():
    import argostranslate.package

    index = json.loads(fetch(INDEX, 10 * 1024 * 1024))
    pins = load_pins()
    changed = False
    for lang in LANGS:
        pkg = next(p for p in index if p.get("from_code") == "en" and p.get("to_code") == lang)
        name = f"{pkg['code']}-{pkg['package_version']}"
        url = next(u for u in pkg["links"] if u.startswith("https://"))
        data = fetch(url, MAX_MODEL)
        digest = hashlib.sha256(data).hexdigest()
        if name in pins and pins[name] != digest:
            raise SystemExit(f"{name}: SHA-256 {digest} does not match pinned {pins[name]}")
        if name not in pins:
            print(f"::warning::{name}: not pinned yet, recording SHA-256 {digest}")
            pins[name] = digest
            changed = True
        with tempfile.TemporaryDirectory() as d:
            path = pathlib.Path(d) / f"{name}.argosmodel"
            path.write_bytes(data)
            argostranslate.package.install_from_path(path)
        del data
    if changed:
        PINS.write_text("# SHA-256 of Argos Translate models, pinned on first use.\n" +
                        "".join(f"{v}  {k}\n" for k, v in sorted(pins.items())))


TOKEN = "ZQX{}QXZ"
# Names that must never be translated: program names, literal file and
# folder names, spreadsheet headers and form labels.
GLOSSARY = sorted([
    "numb-form", "numb-commit", "NUMB-FORM", "Appendix A - Risk Assesment", "Appendix A",
    "3. Übersichten", "Company Name", "Report created", "Customer ID", "Downloads",
    "Information About the Contracting Party", "Name / Company Name of the Contracting Party",
    "Date of the evaluation", "Person Processing the Form", "Español", "Deutsch", "Commercial Registry Number",
    "Authorized Representatives", "Legal form", "Headquarters / Address",
    "Impressum", "Aviso legal", "GLEIF", "Wikidata", "Companies House", "Argos Translate",
    "Sigstore", "cosign", "Docker", "Podman", "Kubernetes", "Command Prompt", "PowerShell",
], key=len, reverse=True)
PROTECT = re.compile("|".join(re.escape(g) for g in GLOSSARY) + r"|`[^`]*`|\[[^\]\s]*\.[A-Za-z0-9]+\]\([^)]*\)|\]\([^)]*\)|https?://\S+|\*\*|\[|\]|<[^>]+>")
MARKER = re.compile(r"^(\s*(?:#{1,6}\s+|[-*+]\s+|\d+\.\s+|>\s*|:\s+)?)(.*)$")


def translate_text(text, tr):
    """Translate prose, protecting code, links and emphasis markers."""
    if not re.search(r"[a-z]{2}", text):
        return text  # nothing to translate (headings like NUMB-FORM(1), symbols)
    saved = []

    def keep(m):
        saved.append(m.group(0))
        return TOKEN.format(len(saved) - 1)

    masked = PROTECT.sub(keep, text)
    if not re.sub(r"ZQX\d+QXZ", "", masked).strip():
        return text
    out = tr.translate(masked)
    for i, s in enumerate(saved):
        t = TOKEN.format(i)
        if t not in out:
            return text  # a protected part was lost; keep the English line
        out = out.replace(t, s, 1)
    return out


def rewrite_links(line, src, dst, translated):
    """Point relative links at the translated copy when one exists."""
    def fix(m):
        target = m.group(1)
        if re.match(r"^[a-z]+:|^#", target):
            return m.group(0)
        path, _, frag = target.partition("#")
        absolute = (ROOT / src).parent.joinpath(path).resolve()
        rel_to_root = absolute.relative_to(ROOT).as_posix() if absolute.is_relative_to(ROOT) else None
        if rel_to_root in translated:
            absolute = ROOT / translated[rel_to_root]
        new = os.path.relpath(absolute, (ROOT / dst).parent).replace(os.sep, "/")
        return "](" + new + ("#" + frag if frag else "") + ")"
    return re.sub(r"\]\(([^)\s]+)\)", fix, line)


STRUCT = re.compile(r"^\s*(?:#{1,6}\s|[-*+]\s|\d+\.\s|>|:\s|\|)")


def translate_file(src, dst, lang, tr, translated):
    lines = (ROOT / src).read_text(encoding="utf-8").splitlines()
    out = []
    block = []  # lines of the current paragraph or list item
    fence = False
    note_done = False

    def flush():
        nonlocal note_done
        if not block:
            return
        first = block[0]
        m = MARKER.match(first)
        prefix = m.group(1)
        text = " ".join([m.group(2)] + [b.strip() for b in block[1:]])
        line = prefix + translate_text(text, tr)
        out.append(rewrite_links(line, src, dst, translated))
        block.clear()
        if not note_done and out[-1].startswith("# "):
            rel = os.path.relpath(ROOT / src, (ROOT / dst).parent).replace(os.sep, "/")
            out.extend(["", NOTE[lang].format(src=rel)])
            note_done = True

    for line in lines:
        if line.lstrip().startswith("```"):
            flush()
            fence = not fence
            out.append(line)
            continue
        if fence:
            out.append(line)
            continue
        if not line.strip():
            flush()
            out.append(line)
            continue
        if line.lstrip().startswith("|"):
            flush()
            cells = re.split(r"(?<!\\)\|", line)
            if all(re.fullmatch(r"\s*:?-+:?\s*", c) or not c.strip() for c in cells):
                out.append(line)
                continue
            line = "|".join(c if not c.strip() else " " + translate_text(c.strip(), tr) + " " for c in cells)
            out.append(rewrite_links(line, src, dst, translated))
            continue
        heading = re.match(r"^\s*#{1,6}\s", line)
        if STRUCT.match(line) or heading:
            flush()
            block.append(line)
            if heading:
                flush()
            continue
        block.append(line)  # continuation of a paragraph or list item
    flush()
    text = "\n".join(out) + "\n"
    (ROOT / dst).parent.mkdir(parents=True, exist_ok=True)
    (ROOT / dst).write_text(text, encoding="utf-8")


def main():
    install_models()
    import argostranslate.translate

    for lang in LANGS:
        tr = argostranslate.translate.get_translation_from_codes("en", lang)
        translated = {s: f"docs/{lang}/{pathlib.Path(s).name}" for s in SOURCES}
        for src in SOURCES:
            translate_file(src, translated[src], lang, tr, translated)
            print(f"translated {src} -> {translated[src]}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
