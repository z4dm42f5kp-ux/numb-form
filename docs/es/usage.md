# NUMB-FORM(1)

> Traducción automática (Argos Translate) del [original en inglés](../usage.md). La versión en inglés es la que vale.

## NAME

**numb-form**, **numb-commit**, **numb-date** — fill risk assessment forms

## SYNOPSIS

```
numb-form [-f conf] [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form [-f conf] check-template template [sample]
numb-form [-f conf] config [init | set key value]
numb-form [-f conf] update [stable | rolling | v1.2.3]
numb-form rollback
numb-commit [-f conf] [-run dir] [-downloads dir]
numb-date [-f conf] [-root dir] [-timeout duration]
```

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
| Legal form | found in the name: at the end, before a branch or country note (`Foo GmbH (Wien)`), before a city or country (`Foo SAS, Bogotá, Colombia`, `Foo S.A.C.I. Paraguay`), or first (`Ets Foo`); accents, dots and spaces may differ (`SOCIEDAD ANONIMA`, `S A`); written as in the name with conventional capitals (`GMBH` gives `GmbH`); asked only when the name has none |
| Headquarters / Address | escrito por el usuario; Introduzca las hojas en blanco |
| Commercial Registry Number | siempre `No commercial registry number was found online consult internal sources for further information`; puede ser reemplazado por el número con *edit* |
| Authorized Representatives | escrito por el usuario como `Full Name (Role 1, Role 2), Full Name (Role)`, roles en inglés; Enter escribe el texto no representativo a continuación, `-` deja en blanco |
| Date of the evaluation | hoy, `date_format` (predeterminado `DD/MM/YYYY`) |
| Person Processing the Form | desde la configuración o escribe al inicio |

numb-form no hace ninguna búsqueda en línea para las empresas. Sólo `numb-form update` contacta con GitHub.

Cuando los representantes quedan vacíos, el campo dice:

> No se encontraron representantes en línea, consulte otras fuentes internas para obtener información más precisa

**numb-commit** applies to review runs only. It copies the accepted folders of a run back into the real working folder. It only adds new folders, never overwrites or deletes anything, and refuses folders that appeared in the real tree in the meantime. Each folder is copied to a temporary name, verified with SHA-256 and then renamed. Documents edited during review are listed before the user confirms. Afterwards it offers to delete the review copy; the logs are kept.

**numb-date** llena la columna de la hoja de cálculo. Se inicia como numb-form, en o por encima de la carpeta de trabajo, y encuentra la misma hoja de cálculo, que debe ser `.xlsx` o `.xlsm`.

1. The company column is found by `header`, then by common names (`Customer Name`, `Customer`, `Company Name`, `Kunde`, `Cliente` and others, ignoring case); the date column in the same header row by `Report created AB`, `Report created`, `Report date`, `Created` and similar names, or any heading that mentions a report and its creation or date.
2. Las filas cuya celda contiene un dígito se saltan: ya tienen una fecha.
3. Para las otras filas, la carpeta de la empresa se encuentra como numb-form la encuentra (alrededor se piden cerillas). Los formularios `.docx`, `.odt`, `.fodt` o `.rtf` en la carpeta (y un nivel debajo) se leen, el formulario de evaluación del riesgo primero: el *Date of the evaluation* de la sección A, o, en los formularios establecidos de manera diferente, la fecha escrita después de una etiqueta de fecha. Se entienden fechas como `16/10/2026`, `16.10.2026`, `2026-10-16` y `16 October 2026`. Cuando no se encuentra ninguna fecha, el usuario lo escribe o salta la empresa.
4. Las fechas se enumeran, escritas como `DD-MM-YYYY` después de que el usuario acceda, lea de nuevo del disco y se muestra; el usuario las acepta o deshacer la escritura.

Sólo las células listadas cambian; vea [../SECURITY.md](SECURITY.md) para los cheques. Las formas sólo se leen. El archivo no debe abrirse en Excel o LibreOffice. El registro es `numb-date.log` en la carpeta de estado por usuario (o `log_dir`).

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
: Mostrar este manual en el terminal, una pantalla a la vez (Enter: siguiente pantalla; q y Enter: renunciar). Sin un tema se muestra todo el manual, comenzando por el guía. Temas: `guide`, `usage`, `configuration`, `install`, `formats`, `containers`, `security`; `numb-form man topics` los lista. El idioma sigue `LANG`, o se da. `numb-form man -install` hace que `man numb-form` funcione: en Linux, macOS y BSD escribe la página manual a `~/.local/share/man/man1` (refreshed by `numb-form update`); en Windows escribe `man.cmd` junto a `numb-form.exe`, para Command Prompt. Windows PowerShell mantiene `man` por su propia ayuda: allí, escriba `numb-form man`.

`update` [`stable` | `rolling` | *v1.2.3*]
: Sustitúyase el `numb-form` y `numb-commit`: **stable** con la última versión estable, **rolling** con la construcción de rodadura actual de `main` (reconstruido después de cada cambio; `edge` es aceptado como un nombre antiguo), o *v1.2.3*. La versión reemplazada se mantiene junto a los programas como `numb-form.previous` y `numb-commit.previous` para `rollback`. Funciona como lo hacen los comandos de instalación: el archivo y `SHA256SUMS` se descargan sobre HTTPS de GitHub solamente, la suma de comprobación debe coincidir, y el nuevo `numb-form` debe comenzar antes de que nada sea reemplazado. Sin un argumento se utiliza el canal de la construcción de ejecución. Se mantienen los ajustes. Refusado con `-offline`, con `online = no`, y para construcciones de origen. En Windows el viejo programa es renombrado a `numb-form.old.exe` y eliminado al siguiente inicio.

## OPTIONS

`-f file`
: Lea *file* después de los archivos de configuración del sistema y del usuario.

`-offline`
: Con `update`: negarse a contactar con GitHub.

`-root dir`
: Use *dir* como la carpeta de trabajo en lugar de detectarla.

`-downloads dir`
: Modo de revisión: ponga la copia en *dir*.

`-timeout duration`
: Stop safe after *duration*, for example `30m`; overrides a longer `timeout` in the settings. Sin ella no hay límite de tiempo.

`-run dir`
: (numb-commit) Envíe esta carpeta de ejecución en lugar de la más reciente sin compromiso.

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

`<log folder>/<run>_date/numb-date.log`
: Registro de una carrera numbruna, con la SHA-256 de la hoja de cálculo antes y después.

`Downloads/numb-form-runs/<run>/<working folder>/`
: Modo de revisión: la copia.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Modo de revisión: registro del commit.

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

[configuration.md](configuration.md), [formats.md](formats.md), [install.md](install.md), [containers.md](containers.md), [../SECURITY.md](SECURITY.md), [../AGENTS.md](../../AGENTS.md)
