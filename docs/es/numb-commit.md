# NUMB-COMMIT(1)

> Traducción automática (Argos Translate) del [original en inglés](../numb-commit.md). La versión en inglés es la que vale.

## NAME

**numb-commit** — copy the new forms into the real working folder, after showing and checking exactly what changes

## SYNOPSIS

```
numb-commit [-f conf] [-from dir] [-to dir]
numb-commit [-f conf] [-from dir] [-to dir] review
numb-commit [-f conf] [-run dir] [-downloads dir]
```

## DESCRIPTION

numb-form y trabajo numb-date en una carpeta en su propio equipo o en la propia carpeta de trabajo real. numb-commit trae el resultado a la carpeta de trabajo real, generalmente en una unidad de red, y nada más.

Hay dos maneras de trabajar; numb-commit maneja ambos.

Copia de trabajo **A** (recomendado para una unidad de red). Usted copia la carpeta real, por ejemplo `F:\14. GWG`, a su computadora, por ejemplo `C:\Users\you\Downloads\14. GWG`, y ejecuta numb-form y numb-date allí. Luego comienza numb-commit; pide las dos carpetas:

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

**A review run** (`mode = review`). numb-form makes the copy itself in `Downloads\numb-form-runs`, and numb-commit copies the accepted folders of the newest run back into the real working folder. It only adds new folders, never overwrites or deletes anything, and refuses folders that appeared in the real folder in the meantime. Each folder is copied to a temporary name, verified with SHA-256 and then renamed. Documents edited during the review are listed before you confirm. Afterwards it offers to delete the review copy; the logs are kept. numb-commit does this when `mode = review` is set or review runs are waiting, and neither `-from`, `-to`, `commit_from`, `commit_to` nor `review` names a folder; `-run` always does it.

## WHAT NUMB-COMMIT DOES WITH A WORKING COPY

1. Se compara la copia de trabajo con el objetivo, archivo por archivo. Los archivos con el mismo tamaño y tiempo son los mismos; para los demás, el checksum SHA-256 decide.
2. Copia exactamente los archivos que se comprometerán en una copia **staged** en su carpeta Downloads (`numb-commit staged copy <date> <time> <id>`), establecida como en el objetivo, verifica cada suma de comprobación, y los hace sólo lectura. Nada toca el objetivo todavía.
3. It writes **`numb-commit-request.txt`** into the working copy and shows it: the list of changes and the target's folder tree before and after (see THE REQUEST FILE).
4. Pregunta cómo revisar:

   ```
   Review [m]anually, or let numb-commit [r]eview and correct it until it is right?
   ```

   `m`: numb-commit ejecuta sus propios cheques (ver REVIEW) y se niega a continuar cuando uno falla; usted mira la copia escenificada y la solicitud usted mismo.

   `r`: numb-commit revisa la solicitud, corrige lo que puede por sí mismo y vuelve a revisar, redondeada tras ronda, hasta que la revisión pase, algo necesita una persona, o cinco rondas han pasado.

5. Pregunta `Copy this into F:\14. GWG now? [y/N]`. Sólo copias `y`.
6. Copia los archivos escenificados en el objetivo, revisa cada suma de comprobación de nuevo, y marca la solicitud como se aplica.
7. Revisa el objetivo una vez más: cada archivo debe estar allí con su chequesum.
8. Se elimina la copia escenificada, también cuando responde no, cuando un cheque falla, cuando pulsa Ctrl+C o cuando cierra la ventana. Una copia escenificada izquierda por un numb-commit que fue asesinado (por ejemplo, por un corte de energía) se elimina por el siguiente numb-commit después de un día.

Lo que se copia:

* **Nuevas carpetas** (normalmente uno por compañía nueva): construido bajo un nombre temporal en el objetivo y renombrado cuando está completo, por lo que una carpeta aparece entera o no en absoluto.
* **Nuevos archivos** en carpetas que ya existen.
* **La hoja de cálculo general**, cuando se fechan las fechas del informe añadido. Su versión antigua se guarda en la carpeta de registro primero.

Lo que nunca se copia:

* **A nueva carpeta de una empresa que el objetivo ya tiene.** En el objetivo, las carpetas de la empresa a menudo comienzan con el ID del cliente, por ejemplo `D10629 Foo S.A.`. Cuando la copia de trabajo carece de esa carpeta, numb-form hace una nueva `Foo S.A.` con sólo la evaluación de riesgo. numb-commit no lo copia cuando el objetivo tiene una carpeta que comienza con el ID de cliente de la compañía (de la lista Excel), o cuyo nombre después de un ID de cliente es el de la empresa, o que tiene el nombre de la empresa en otra ortografía (caso, acentos, puntuación). Si tal carpeta tiene más que el formulario de numb-form, la revisión le pide que copie esos archivos a mano.
* **A archivo que existe en el objetivo con otro contenido**: numb-commit nunca sobrescribe un archivo excepto la hoja de cálculo. Los enumera como diferentes, así que vea lo que cambió en la copia de trabajo.
* la plantilla (`Appendix A - Risk Assesment.docx`), incluso si cambió;
* registros, carpetas `manifest.json`, `numb-form-runs` y el archivo de solicitud;
* Archivos de bloqueo de Office y LibreOffice (`~$…`, `.~lock.…#`), archivos ocultos y temporales, `Thumbs.db`, `desktop.ini`.

Nunca se elimina nada en el objetivo.

## FORMS MADE BY NUMB-FORM

numb-commit tells which forms numb-form made and nobody changed since. numb-form writes no mark into the documents. It records the SHA-256 checksum of every form it writes in the `manifest.json` of its run: in `Downloads/numb-form-runs` for a review run, in the log folder for a direct run. numb-commit compares each form's checksum with those records. A form is labelled:

* `made by numb-form, unchanged` cuando una ejecución registró su suma de comprobación;
* `a form edited after filling, made by hand or filled on another computer` de lo contrario;
* nada, para otros archivos.

Esto funciona para cada formato de plantilla.

## THE REQUEST FILE

`numb-commit-request.txt` se encuentra en la copia de trabajo, nunca en el objetivo. Dice, en palabras claras, exactamente lo que cambia:

```
numb-commit request: what numb-commit copies into the real folder

From:   C:\Users\you\Downloads\14. GWG
To:     F:\14. GWG
Staged: C:\Users\you\Downloads\numb-commit staged copy 2026-10-23 14.05.12 3fa1
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
  - Foo S.A\Appendix A - Risk Assesment - Foo S.A.docx  (not copied: the real folder already has "D10629 Foo S.A." for Foo S.A. (it starts with the customer id D10629); made by numb-form, unchanged)

Nothing in the real folder is deleted, and no file there is overwritten except the ones under Replaced.

Before: F:\14. GWG
├── 3. Übersichten
│   └── Kundenliste.xlsx
├── D10629 Foo S.A.
└── … 1204 more, unchanged

After the commit:
├── 3. Übersichten
│   └── Kundenliste.xlsx                                ~ replaced
├── Bar Ltda                                            + new folder
│   └── Appendix A - Risk Assesment - Bar Ltda.docx     + new, made by numb-form, unchanged
├── D10629 Foo S.A.                                     = unchanged; "Foo S.A" of the working copy is the same company, not copied
└── … 1204 more, unchanged
```

Los árboles muestran sólo lo que la solicitud toca; la última línea cuenta las otras carpetas y archivos en la parte superior del objetivo. En el árbol después del compromiso:

| Mark | Significado |
| --- | --- |
| `+` | nuevo: copiado en el objetivo |
| `~` | reemplazado: la hoja de cálculo con sus nuevas fechas de informe |
| `≠` | difiere en la copia de trabajo; la versión del objetivo permanece |
| `=` | la carpeta del objetivo de una empresa cuya nueva carpeta no se copia |

*Status* es `waiting` hasta que se haga la copia, entonces `applied`. Debajo de la parte legible, una sección de datos repite la lista para numb-commit; no editarla. Una copia de la solicitud aplicada se guarda en la carpeta de registro.

## REVIEW

La revisión verifica que:

* cada archivo a copiar tiene la suma de comprobación en la solicitud, en la copia escalonada y en la copia de trabajo (nada se editó desde entonces);
* ninguno de ellos existe todavía en el objetivo;
* la hoja de cálculo del objetivo todavía tiene la suma de comprobación en la solicitud (nadie más lo cambió mientras tanto);
* la hoja de cálculo sólo gana las fechas del informe: sólo las células vacías o libres de dígitos del cambio de columna *Report created*, y cada nuevo valor parece `DD-MM-YYYY`;
* cada nueva carpeta pertenece a una empresa en la lista de Excel y tiene un formulario de evaluación de riesgos cuya sección A nombres que empresa y no tiene línea en blanco, y el objetivo no tiene carpeta similar para ella;
* nada que debe permanecer detrás (templa, registros, archivos de bloqueo) está listado.

Después de la solicitud se aplica, comprueba que cada archivo listado está en el objetivo con su suma de comprobación.

Cada línea está marcada `ok`, `check` (mira antes de cometer) o `FAIL`, seguido del veredicto:

| Vered | Significado |
| --- | --- |
| `OK to commit.` | Cada cheque pasó. |
| `commit after checking the lines marked check.` | Nada está mal con los archivos, pero algo merece una mirada, por ejemplo una dirección en blanco. |
| `do not commit.` | Algo cambió o no pertenece; numb-commit se niega a copiar. |

Con `r`, numb-commit corrige estos fallos por sí mismo y vuelve a revisar:

| Fallo | Corrección |
| --- | --- |
| un archivo cambiado en la copia de trabajo o en la copia escenificada | comparación y etapa de nuevo |
| un archivo apareció en el objetivo mientras | lo deja fuera: nunca está sobrescrito |
| un archivo que debe permanecer detrás se enumera | Déjalo. |

Se detiene, sin copiar, en cualquier otra cosa, por ejemplo una celda de hoja de cálculo fuera de las fechas del informe o una hoja de cálculo de objetivos que alguien más cambió: los que necesitan una persona.

`numb-commit review` ejecuta la revisión sobre la solicitud en la copia de trabajo en cualquier momento, sin copiar, por ejemplo a petición de un colega preparado.

## OPTIONS

`-from dir`
: La copia de trabajo. Predeterminado: `commit_from` en la configuración; cuando está vacío, pregunta numb-commit.

`-to dir`
: El objetivo, la verdadera carpeta de trabajo. Predeterminado: `commit_to` en la configuración; cuando está vacío, pregunta numb-commit.

`review`
: Sólo revise la solicitud en la copia de trabajo; no copie nada.

`-run dir`
: Corres de revisión: comprometer esta carpeta de ejecución en lugar de la más reciente no comprometida.

`-downloads dir`
: Donde `numb-form-runs` y la copia escenificada están. Predeterminado: `downloads` en la configuración, de lo contrario su carpeta Downloads.

`-f file`
: Lea *file* después de los archivos de configuración del sistema y del usuario.

`-allow-non-tty`
: Lea las respuestas de una tubería. Sólo para pruebas automatizadas.

`-version`
: Imprima la versión.

## FILES

`<working copy>/numb-commit-request.txt`
: La solicitud: qué es (o fue) copiado. UTF-8 con una marca de orden byte, por lo que Notepad y Windows PowerShell 5.1 (`Get-Content`) muestran el árbol correctamente.

`<Downloads>/numb-commit staged copy <date> <time> <id>/`
: Los archivos a copiar, mientras numb-commit funciona. Retirada al final.

`<log folder>/<date>_<time>_<id>_commit/numb-commit.log`
: Registro de la copia, con cada suma de comprobación.

`<log folder>/<date>_<time>_<id>_commit/replaced/`
: La versión anterior del objetivo de cada archivo reemplazado.

`<log folder>/<date>_<time>_<id>_commit/numb-commit-request.txt`
: La solicitud aplicada.

La carpeta de registro es `log_dir`, por defecto `%LOCALAPPDATA%\numb-form\runs` (Windows), `~/Library/Logs/numb-form` (macOS) o `~/.local/state/numb-form/runs` (Linux, BSD).

## EXIT STATUS

0 cuando se hace la copia, no hay nada que hacer o usted respondió no; 1 en un error o una revisión fallida; 2 en un error de uso; 3 cuando se detuvo con Ctrl+C o cerrando la ventana.

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

Deje que numb-commit revise una solicitud, por ejemplo un colega preparado, sin copiar (todos los sistemas):

```
numb-commit review
```

## WHEN SOMETHING GOES WRONG

**`F:\14. GWG is not reachable (is the drive connected?)`**
: Conecte la unidad de red (abráelo una vez en File Explorer) y escriba el camino de nuevo.

**`… changed after the request was made`**
: Un archivo fue editado después de que fue escenificado. Con `r`, numb-commit lo vuelve a poner por sí mismo; con `m`, ejecute numb-commit de nuevo.

**`… in the real folder changed after the request was made`**
: Alguien más cambió la hoja de cálculo en la unidad de red. numb-commit no sobrescribe su trabajo. Copia la hoja de cálculo real en la copia de trabajo de nuevo, ejecutar numb-date allí, luego numb-commit.

**`… already exists in the real folder and is never overwritten`**
: Un colega creó el mismo archivo mientras tanto. Con `r`, numb-commit lo deja fuera; compare los dos a mano.

**`spreadsheet cell … changed; only report dates may change`**
: La hoja de cálculo de la copia de trabajo fue editada a mano. numb-commit sólo lleva fechas de informe; hacer otros cambios en la hoja de cálculo real.

**`… is not a working folder: it has no template and "Übersicht" folder`**
: Dar la carpeta que contiene la plantilla y la carpeta de visión general, no una carpeta dentro de ella.

**`no request to review`**
: Ejecutar `numb-commit` primero; `numb-commit review` comprueba la solicitud que escribe.

## SEE ALSO

[numb-form](usage.md), [numb-date](numb-date.md), [configuration.md](configuration.md), [the guide](guide.md)
