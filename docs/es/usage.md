# NUMB-FORM(1)

> Traducción automática (Argos Translate) del [original en inglés](../usage.md). La versión en inglés es la que vale.

## NAME

**numb-form** — fill risk assessment forms (with **numb-commit** and **numb-date**)

## SYNOPSIS

```
numb-form [-f conf] [-y] [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form [-f conf] check-template template [sample]
numb-form [-f conf] config [init | set key value]
numb-form [-f conf] update [stable | rolling | v1.2.3]
numb-form rollback
```

numb-commit y numb-date tienen sus propias páginas: [numb-commit.md](numb-commit.md), [numb-date.md](numb-date.md).

## DESCRIPTION

**numb-form** is started inside the working folder, the folder that holds the template and the overview folder (for example `GWG`). It may also be started one level above it: the working folder is then the single folder found there, or the one containing an overview folder, or the user is asked. `-root` names the working folder explicitly.

Personal de oficina: vea la guía [step-by-step](guide.md).

La carpeta de trabajo debe contener:

* una carpeta de visión general cuyo nombre contiene `overview_dir` (predeterminado `Übersicht`, que coincide con `3. Übersichten`), conteniendo exactamente una hoja de cálculo;
* la plantilla, llamada `template_name` (predeterminado `Appendix A - Risk Assesment`), en su nivel superior.

Los formatos compatibles se enumeran en [formats.md](formats.md). Los ajustes se describen en [configuration.md](configuration.md).

A run proceeds as follows:

1. La plantilla se revisa con un llenado de ensayo de cada campo. La carrera se detiene si la sección A, o cualquier etiqueta que no sea *Commercial Registry Number*, no se puede encontrar o no está vacía.
2. El nombre de la persona que procesa el formulario se toma de la configuración, o se escribe una vez.
3. Con `mode = review`, toda la carpeta de trabajo se copia a `Downloads/numb-form-runs/<date>_<time>_<id>/` y todo después de este paso sucede en la copia. Con `mode = direct` (el predeterminado), la carpeta de trabajo en sí se utiliza y el registro va a la carpeta de estado por usuario.
4. Los nombres de las empresas se leen desde la columna encabezada `Customer Name`, de arriba a abajo. Todas las demás columnas son ignoradas, incluyendo `Report created`. numb-form nunca escribe la hoja de cálculo; su suma de comprobación SHA-256 se revisa antes y después de la carrera.
5. Para cada empresa, las carpetas en la carpeta de trabajo y todos sus subcarpetas son buscados. Los nombres son comparados ignorando caso, acentos y umlauts (ü = u = ue), puntuación y espacios. Un partido significa que la forma ya existe, y la compañía está saltada. Un partido cercano se muestra al usuario para confirmar. Las carpetas con una forma legal diferente nunca coinciden: `Foo GmbH` y `Foo S.A.` son diferentes empresas.
6. Para una nueva empresa, el usuario escribe la dirección de la sede y los representantes autorizados. Todo lo demás está lleno sin preguntas.
7. Una carpeta llamada después de la creación de la empresa. La plantilla se copia en ella como `<template name> - <Company Name>.<ext>`, y la sección A se llena.
8. El documento se lee de nuevo del disco y se muestra. El usuario lo acepta, edita un campo o lo rechaza; se elimina una carpeta rechazada.

Sólo se llenan las líneas de la sección A, *Information About the Contracting Party*. La sección B y todo lo que está fuera de la sección A quedan byte para byte sin cambios, y no se añade formato.

| Campo | Valor |
| --- | --- |
| Name / Company Name of the Contracting Party | la hoja de cálculo literal |
| Legal form | found in the name: at the end, before a branch or country note (`Foo GmbH (Wien)`), before a city or country (`Foo SAS, Bogotá, Colombia`, `Foo S.A.C.I. Paraguay`), or first (`Ets Foo`); accents, dots and spaces may differ (`SOCIEDAD ANONIMA`, `S A`); written as in the name with conventional capitals (`GMBH` gives `GmbH`); asked only when the name has none, where Enter writes the no-legal-form text below and `-` leaves it blank |
| Headquarters / Address | escrito por el usuario; Enter escribe el texto de no dirección a continuación, `-` deja en blanco |
| Commercial Registry Number | siempre `No commercial registry number was found online consult internal sources for further information`; puede ser reemplazado por el número con *edit* |
| Authorized Representatives | escrito por el usuario como `Full Name (Role 1, Role 2), Full Name (Role)`, roles en inglés; Enter escribe el texto no representativo a continuación, `-` deja en blanco |
| Date of the evaluation | hoy, `date_format` (predeterminado `DD/MM/YYYY`) |
| Person Processing the Form | desde la configuración o escribe al inicio |

numb-form no hace ninguna búsqueda en línea para las empresas. Sólo `numb-form update` contacta con GitHub.

Cuando un valor queda vacío (Enter solo), el campo lee:

| Campo | Texto |
| --- | --- |
| Legal form | `No legal form was found online due to little to no information available` |
| Headquarters / Address | `No address was found online due to little to no information available` |
| Authorized Representatives | `No representatives were found online, please consult other internal sources for more accurate information` |

**numb-commit** copies the accepted folders of a review run into the real working folder, and shows and checks what a working copy would add to it: see [numb-commit.md](numb-commit.md).

**numb-date** llena la columna *Report created* de la hoja de cálculo de los formularios: ver [numb-date.md](numb-date.md).

## COMMANDS

`check-template template [sample]`
: Compruebe que una plantilla se puede rellenar, y opcionalmente escriba una muestra llena a *sample*. La plantilla solo se lee.

`config`
: Mostrar los archivos de configuración y los ajustes efectivos.

`config init`
: Escriba un archivo de configuración por usuario comentado con cada defecto.

`config set key value`
: Guardar una configuración en el archivo de configuración por usuario, por ejemplo `numb-form config set person "Erika Mustermann"`. El valor se verifica antes de que se guarde.

`rollback`
: Switch back to the version that the last `update` replaced. The version rolled back from is kept in its place, so a second `rollback` returns to it. Without a kept version, install a release by number instead: `numb-form update v0.1.1`.

`man` [*topic*] [`en` | `es` | `de`]
: Mostrar este manual en el terminal, como menos(1): las teclas actúan inmediatamente, sin Enter. Espacio o páginaAbajo: siguiente página; b o PageUp: página anterior; Enter, j o Down: siguiente línea; k o Up: línea anterior; g o Home: start; G o End: end; /text and Enter: search, n y N: siguiente y último partido; h: help; q: quit. En macOS, Linux y los BSD se utiliza el `less` instalado cuando hay uno. Sin un tema se muestra todo el manual, comenzando por el guía. Temas: `guide`, `usage`, `numb-commit`, `numb-date`, `configuration`, `install`, `formats`, `containers`, `security`; `numb-form man topics` los lista. El idioma sigue `LANG`, o se da. `numb-form man -install` hace que `man numb-form` funcione: en Linux, macOS y BSD escribe la página manual a `~/.local/share/man/man1` (refreshed by `numb-form update`); en Windows escribe `man.cmd` junto a `numb-form.exe`, para Command Prompt. Windows PowerShell mantiene `man` por su propia ayuda: allí, escriba `numb-form man`.

`update` [`stable` | `rolling` | *v1.2.3*]
: Sustitúyase el `numb-form` y `numb-commit`: **stable** con la última versión estable, **rolling** con la construcción de rodadura actual de `main` (reconstruido después de cada cambio; `edge` es aceptado como un nombre antiguo), o *v1.2.3*. La versión reemplazada se mantiene junto a los programas como `numb-form.previous` y `numb-commit.previous` para `rollback`. Funciona como lo hacen los comandos de instalación: el archivo y `SHA256SUMS` se descargan sobre HTTPS de GitHub solamente, la suma de comprobación debe coincidir, y el nuevo `numb-form` debe comenzar antes de que nada sea reemplazado. Sin un argumento se utiliza el canal de la construcción de ejecución. Se mantienen los ajustes. Refusado con `-offline`, con `online = no`, y para construcciones de origen. En Windows el viejo programa es renombrado a `numb-form.old.exe` y eliminado al siguiente inicio.

## OPTIONS

`-f file`
: Lea *file* después de los archivos de configuración del sistema y del usuario.

`-y`
: Tome una carpeta con un nombre similar como la carpeta de la empresa en lugar de preguntar "¿Es la carpeta existente ... la carpeta de esta empresa?". La carpeta se llama en pantalla y en el registro, y el resumen dice cuántas carpetas fueron tomadas de esta manera. Cada documento se muestra y debe ser aceptado. `-Y` es el mismo.

`-offline`
: Con `update`: negarse a contactar con GitHub.

`-root dir`
: Use *dir* como la carpeta de trabajo en lugar de detectarla.

`-downloads dir`
: Modo de revisión: ponga la copia en *dir*.

`-timeout duration`
: Stop safe after *duration*, for example `30m`; overrides a longer `timeout` in the settings. Sin ella no hay límite de tiempo.

`-allow-non-tty`
: Lea las respuestas de una tubería. Sólo para pruebas automatizadas.

`-version`
: Imprima la versión.

## ENVIRONMENT

`NUMBFORM_DOWNLOADS`
: Carpeta Downloads para utilizar, por ejemplo, un volumen montado en un contenedor.

`HTTPS_PROXY`, `NO_PROXY`
: Ajustes proxy para `numb-form update`.

`TZ`
: Zona horaria de la fecha de evaluación.

## FILES

`<log folder>/<run>/numb-form.log`
: Lograr todo lo que hizo la carrera.

`<log folder>/<run>/manifest.json`
: Grabación legible a máquina de la carrera.

`Downloads/numb-form-runs/<run>/<working folder>/`
: Modo de revisión: la copia.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Modo de revisión: registro del commit.

Los archivos de numb-commit y numb-date se enumeran en sus páginas.

No hay caché ni archivos temporales detrás. Los documentos se escriben a través de un archivo temporal que se renombra o elimina, y los convertidores trabajan en una carpeta temporal privada que se elimina.

## EXIT STATUS

0 sobre el éxito, 1 sobre error, 2 sobre un error de uso o configuración, y 3 cuando se detiene por el límite de tiempo, por Ctrl+C o por el usuario.

## EXAMPLES

Los comandos se muestran para ambos tipos de sistema. Windows PowerShell 5.1 no tiene `&&`: los comandos están separados con `;`, y `if ($?) { ... }` ejecuta el siguiente sólo cuando el primero tuvo éxito.

Compruebe la plantilla real y escriba una muestra, sin tocar la plantilla (todos los sistemas):

```
numb-form check-template "Appendix A - Risk Assesment.docx" sample.docx
```

Llene los formularios directamente en la carpeta de trabajo, luego el informe data.

**Unix-like** (macOS, Linux, BSD), en Terminal:

```
cd /path/to/GWG && numb-form && numb-date
```

**Windows** (10, 11), en PowerShell:

```
cd "H:\GWG"; numb-form; if ($?) { numb-date }
```

Revisar primero, luego comprometer (numb-commit espera confirmación), donde `review.conf` contiene `mode = review`.

**Unix-like** (macOS, Linux, BSD), en Terminal:

```
cd /path/to/GWG && numb-form -f review.conf && numb-commit
```

**Windows** (10, 11), en PowerShell:

```
cd "H:\GWG"; numb-form -f review.conf; if ($?) { numb-commit }
```

## SEE ALSO

[numb-commit.md](numb-commit.md), [numb-date.md](numb-date.md), [configuration.md](configuration.md), [formats.md](formats.md), [install.md](install.md), [containers.md](containers.md), [../SECURITY.md](SECURITY.md), [../AGENTS.md](../../AGENTS.md)
