# NUMB-COMMIT(1)

> Traducción automática (Argos Translate) del [original en inglés](../numb-commit.md). La versión en inglés es la que vale.

## NAME

**numb-commit** — copy reviewed forms into the real working folder, and show what a working copy would add to it

## SYNOPSIS

```
numb-commit [-f conf] [-from dir] [-to dir]
numb-commit [-f conf] [-from dir] [-to dir] review
numb-commit [-f conf] [-run dir] [-downloads dir]
```

## DESCRIPTION

numb-form y trabajo numb-date en una carpeta en su propio equipo o en la propia carpeta de trabajo real. numb-commit se ocupa del resultado. Hay dos formas de trabajar.

**A copia de trabajo** (para una unidad de red). Usted copia la carpeta real, por ejemplo `F:\14. GWG`, a su computadora, por ejemplo `C:\Users\you\Downloads\14. GWG`, y ejecuta numb-form y numb-date allí. Luego comienza numb-commit; pide las dos carpetas:

```
Which directory should be committed (insert absolute path)?: C:\Users\you\Downloads\14. GWG
Which directory is the target?: F:\14. GWG
```

Se aceptan caminos copiados con *Copy como camino* (con comillas). Para responder de una vez por todas, establecerlos en la configuración:

```
numb-form config set commit_from "C:\Users\you\Downloads\14. GWG"
numb-form config set commit_to "F:\14. GWG"
```

`-from` y `-to` responden por una sola vez.

Esta versión de numb-commit solo compara las dos carpetas, escribe la solicitud y la revisa. No copia nada al objetivo. Copia los archivos listados bajo *Nueva carpeta* y *Reemplazado* usted mismo.

**A review run** (`mode = review`). numb-form hace la copia en `Downloads\numb-form-runs`, y numb-commit copia las carpetas aceptadas de la más reciente ejecutada de nuevo en la carpeta de trabajo real. Sólo añade nuevas carpetas, nunca sobrescribe o elimina nada, y rechaza carpetas que aparecieron en la carpeta real mientras tanto. Cada carpeta es copiada a un nombre temporal, verificada con SHA-256 y luego renombrada. Los documentos editados durante la revisión se enumeran antes de confirmar. Después ofrece eliminar la copia de revisión; los registros se guardan.

numb-commit se compromete a ejecutar una reseña cuando `mode = review` está programada o las carreras de revisión están esperando, y ni `-from`, `-to`, `commit_from`, `commit_to` ni `review` nombres de carpeta; `-run` siempre lo hace.

## WHAT NUMB-COMMIT DOES WITH A WORKING COPY

1. Se compara la copia de trabajo con el objetivo, archivo por archivo. Los archivos con el mismo tamaño y tiempo son los mismos; para los demás, el checksum SHA-256 decide.
2. It writes **`numb-commit-request.txt`** into the working copy and shows it: the list of changes and the target's folder tree before and after (see THE REQUEST FILE).
3. Revisa la solicitud (ver REVIEW) e imprime un veredicto.
4. No copia nada: el objetivo no se toca.

Lo que la solicitud enumera como copiado:

* **Nuevas carpetas** (normalmente uno por compañía nueva) y archivos **new** en carpetas que ya existen.
* **La hoja de cálculo general**, cuando se fechan las fechas del informe añadido.

Lo que nunca se copia:

* **A archivo que existe en el objetivo con otro contenido**: nunca es sobrescrito, excepto la hoja de cálculo. La solicitud enumera estos como diferentes, así que vea lo que cambió en la copia de trabajo.
* la plantilla (`Appendix A - Risk Assesment.docx`), incluso si cambió;
* registros, carpetas `manifest.json`, `numb-form-runs` y el archivo de solicitud;
* Archivos de bloqueo de Office y LibreOffice (`~$…`, `.~lock.…#`), archivos ocultos y temporales, `Thumbs.db`, `desktop.ini`.

## FORMS MADE BY NUMB-FORM

numb-commit tells which forms numb-form made and nobody changed since. numb-form writes no mark into the documents. It records the SHA-256 checksum of every form it writes in the `manifest.json` of its run: in `Downloads/numb-form-runs` for a review run, in the log folder for a direct run. numb-commit compares each form's checksum with those records. A form is labelled:

* `made by numb-form, unchanged` cuando una ejecución registró su suma de comprobación;
* `a form edited after filling, made by hand or filled on another computer` de lo contrario;
* nada, para otros archivos.

Esto funciona para cada formato de plantilla.

## THE REQUEST FILE

`numb-commit-request.txt` se encuentra en la copia de trabajo, nunca en el objetivo. Dice, en palabras claras, exactamente lo que cambiaría:

```
numb-commit request: what numb-commit copies into the real folder

From:   C:\Users\you\Downloads\14. GWG
To:     F:\14. GWG
Made:   23/10/2026 14:05
Status: waiting

New folder  Bar Ltda
  + Bar Ltda\Appendix A - Risk Assesment - Bar Ltda.docx  (18 KB, made by numb-form, unchanged, sha256 4be1…)

Replaced  3. Übersichten\Kundenliste.xlsx
  sha256 9f02… in the real folder now
  sha256 c771… afterwards
  the replaced version is kept in the log folder

Left alone
  - Appendix A - Risk Assesment.docx  (not copied: the template)

Nothing in the real folder is deleted, and no file there is overwritten except the ones under Replaced.

Before: F:\14. GWG
├── 3. Übersichten
│   └── Kundenliste.xlsx
└── … 1204 more, unchanged

After the commit:
├── 3. Übersichten
│   └── Kundenliste.xlsx                                ~ replaced
├── Bar Ltda                                            + new folder
│   └── Appendix A - Risk Assesment - Bar Ltda.docx     + new, made by numb-form, unchanged
└── … 1204 more, unchanged
```

Los árboles muestran sólo lo que la solicitud toca; la última línea cuenta las otras carpetas y archivos en la parte superior del objetivo. En el árbol después del compromiso:

| Mark | Significado |
| --- | --- |
| `+` | nuevo: ser copiado en el objetivo |
| `~` | reemplazado: la hoja de cálculo con sus nuevas fechas de informe |
| `≠` | difiere en la copia de trabajo; la versión del objetivo permanece |

*Status* es `waiting`. Debajo de la parte legible, una sección de datos repite la lista para numb-commit; no editarla.

## REVIEW

La revisión verifica que:

* cada archivo a copiar todavía tiene la suma de comprobación en la solicitud (nada fue editado desde entonces);
* ninguno de ellos existe todavía en el objetivo;
* la hoja de cálculo del objetivo todavía tiene la suma de comprobación en la solicitud (nadie más lo cambió mientras tanto);
* la hoja de cálculo sólo gana las fechas del informe: sólo las células vacías o libres de dígitos del cambio de columna *Report created*, y cada nuevo valor parece `DD-MM-YYYY`;
* cada nueva carpeta pertenece a una empresa en la lista de Excel y tiene un formulario de evaluación de riesgos cuya sección A nombres que empresa y no tiene línea en blanco, y el objetivo no tiene carpeta similar para ella;
* nada que debe permanecer detrás (templa, registros, archivos de bloqueo) está listado.

Cada línea está marcada `ok`, `check` (mira) o `FAIL`, seguido del veredicto:

| Vered | Significado |
| --- | --- |
| `OK to commit.` | Cada cheque pasó. |
| `commit after checking the lines marked check.` | Nada está mal con los archivos, pero algo merece una mirada, por ejemplo una dirección en blanco. |
| `do not commit.` | Algo cambió o no pertenece. |

`numb-commit review` vuelve a revisar la solicitud en la copia de trabajo en cualquier momento, por ejemplo a petición de un colega preparado.

## OPTIONS

`-from dir`
: La copia de trabajo. Predeterminado: `commit_from` en la configuración; cuando está vacío, pregunta numb-commit.

`-to dir`
: El objetivo, la verdadera carpeta de trabajo. Predeterminado: `commit_to` en la configuración; cuando está vacío, pregunta numb-commit.

`review`
: Sólo revise la solicitud en la copia de trabajo.

`-run dir`
: Corres de revisión: comprometer esta carpeta de ejecución en lugar de la más reciente no comprometida.

`-downloads dir`
: Corres de revisión: la carpeta que contiene `numb-form-runs`. Predeterminado: `downloads` en la configuración, de lo contrario su carpeta Downloads.

`-f file`
: Lea *file* después de los archivos de configuración del sistema y del usuario.

`-allow-non-tty`
: Lea las respuestas de una tubería. Sólo para pruebas automatizadas.

`-version`
: Imprima la versión.

## FILES

`<working copy>/numb-commit-request.txt`
: La solicitud: qué sería copiado. UTF-8.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Corres de revisión: registro del commit.

## EXIT STATUS

0 cuando la revisión pasa o no hay nada que hacer; 1 en un error o una revisión fallida; 2 en un error de uso; 3 cuando se detiene con Ctrl+C.

## EXAMPLES

Una vez: diga dónde están ambas carpetas (todos los sistemas):

```
numb-form config set commit_from "C:\Users\you\Downloads\14. GWG"; numb-form config set commit_to "F:\14. GWG"
```

Cada vez.

**Windows** (10, 11), en PowerShell:

```
cd "$env:USERPROFILE\Downloads\14. GWG"; numb-form; if ($?) { numb-date }; if ($?) { numb-commit }
```

**Unix-like** (macOS, Linux, BSD), en Terminal:

```
cd ~/Downloads/14.\ GWG && numb-form && numb-date && numb-commit
```

Revise una solicitud de nuevo, por ejemplo un colega preparado (todos los sistemas):

```
numb-commit review
```

## WHEN SOMETHING GOES WRONG

**`F:\14. GWG is not reachable (is the drive connected?)`**
: Conecte la unidad de red (abráelo una vez en File Explorer) y escriba el camino de nuevo.

**`… is not a working folder: it has no template and "Übersicht" folder`**
: Dar la carpeta que contiene la plantilla y la carpeta de visión general, no una carpeta dentro de ella.

**`… changed after the request was made`**
: Un archivo fue editado después de la solicitud fue escrita. Corre numb-commit otra vez.

**`… in the real folder changed after the request was made`**
: Alguien más cambió la hoja de cálculo en la unidad de red. Copia la hoja de cálculo real en la copia de trabajo de nuevo, ejecutar numb-date allí, luego numb-commit.

**`… already exists in the real folder and is never overwritten`**
: Un colega creó el mismo archivo mientras tanto. Compare los dos a mano.

**`spreadsheet cell … changed; only report dates may change`**
: La hoja de cálculo de la copia de trabajo fue editada a mano. Haz otros cambios en la hoja de cálculo real.

**`no request to review`**
: Ejecutar `numb-commit` primero; `numb-commit review` comprueba la solicitud que escribe.

## SEE ALSO

[numb-form](usage.md), [numb-date](numb-date.md), [configuration.md](configuration.md), [the guide](guide.md)
