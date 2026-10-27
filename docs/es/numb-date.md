# NUMB-DATE(1)

> Traducción automática (Argos Translate) del [original en inglés](../numb-date.md). La versión en inglés es la que vale.

## NAME

**numb-date** — rellene la columna *Report created* de la hoja de cálculo general de las formas de las empresas

## SYNOPSIS

```
numb-date [-f conf] [-y] [-root dir] [-timeout duration]
```

## DESCRIPTION

numb-date se inicia como numb-form, dentro de la carpeta de trabajo (o un nivel por encima), generalmente justo después de numb-form. Es el único programa que escribe a la hoja de cálculo general, y no escribe nada más que reportar fechas. La hoja de cálculo debe ser `.xlsx` o `.xlsm`, y cerrado en Excel y LibreOffice.

1. The company column is found by `header`, then by common names (`Customer Name`, `Customer`, `Company Name`, `Kunde`, `Cliente` and others, ignoring case); the date column in the same header row by `Report created AB`, `Report created`, `Report date`, `Created` and similar names, or any heading that mentions a report and its creation or date.
2. Las filas cuya celda contiene un dígito se saltan: ya tienen una fecha.
3. Para las otras filas, la carpeta de la compañía se encuentra como numb-form la encuentra (cerca de los partidos se preguntan, o se toman con `-y`). Los formularios `.docx`, `.odt`, `.fodt` o `.rtf` en la carpeta (y un nivel debajo) se leen, el formulario de evaluación del riesgo primero: el *Date of the evaluation* de la sección A, o, en los formularios establecidos de manera diferente, la fecha escrita después de una etiqueta de fecha. Se entienden fechas como `16/10/2026`, `16.10.2026`, `2026-10-16` y `16 October 2026`. Cuando no se encuentra ninguna fecha, se escribe o se salta la empresa.
4. Las fechas se enumeran y se escriben como `DD-MM-YYYY` después de estar de acuerdo, luego se lee de nuevo del disco y se muestra; se aceptan o deshacer la escritura.

Las formas sólo se leen.

## WHAT CHANGES IN THE SPREADSHEET

Sólo las células listadas cambian. Mantienen su propio estilo celular y no tienen nuevo formato; cada otra parte del libro de trabajo es copiado byte para byte, y las células con fórmulas se quedan solas. Antes de escribir, cada celda de cada hoja del nuevo libro de trabajo se compara con el original: cualquier otra cosa que difiere detiene numb-date. SHA-256 guarda cada paso: el archivo todavía debe tener la suma de comprobación que tenía cuando se leía, la copia temporal al lado debe coincidir antes de que reemplace el archivo, y el archivo debe coincidir después. El original se guarda en la carpeta de registro hasta que usted acepta; *undo* lo pone de nuevo exactamente.

| numb-date pide | ¿Qué haces? |
| --- | --- |
| `Is the folder "…" this company's folder? (...) [y/N]` | `y` si es la misma empresa, de lo contrario Enter |
| `Date (DD-MM-YYYY, Enter = skip):` | El formulario no tiene fecha legible: escríbalo o entra para saltar |
| `Write these … date(s)? Nothing else in the spreadsheet changes. [y/N]` | `y` para escribirlos; Entrar no escribe nada |
| `Is this correct? [a]ccept, [u]ndo:` | `a` los mantiene, `u` restaura el archivo exactamente |

## OPTIONS

`-y`
: Tome una carpeta con un nombre similar como la carpeta de la empresa sin preguntar.

`-root dir`
: Use *dir* como la carpeta de trabajo en lugar de detectarla.

`-timeout duration`
: Parar con seguridad después de *duración*, por ejemplo `30m`.

`-f file`
: Lea *file* después de los archivos de configuración del sistema y del usuario.

`-allow-non-tty`
: Lea las respuestas de una tubería. Sólo para pruebas automatizadas.

`-version`
: Imprima la versión.

## FILES

`<log folder>/<date>_<time>_<id>_date/numb-date.log`
: Inicie la carrera, con la SHA-256 de la hoja de cálculo antes y después.

`<log folder>/<date>_<time>_<id>_date/before-<spreadsheet>`
: La hoja de cálculo original, guardada hasta que acepte o deshacer.

## EXIT STATUS

0 sobre el éxito o cuando no hay nada que escribir, 1 sobre un error, 2 sobre un error de uso, 3 cuando se detiene con Ctrl+C o por el usuario.

## EXAMPLES

**Windows** (10, 11), en PowerShell, dentro de la carpeta de trabajo:

```
numb-form; if ($?) { numb-date }
```

**Unix-like** (macOS, Linux, BSD), en Terminal:

```
numb-form && numb-date
```

## WHEN SOMETHING GOES WRONG

**`… is open in Excel or LibreOffice`**
: Cerrar el archivo y ejecutar numb-date de nuevo.

**`numb-date writes only .xlsx and .xlsm spreadsheets`**
: Guardar la vista general como `.xlsx` una vez.

**`no "Report created" column next to the company names`**
: El mensaje enumera los encabezados encontrados. Renombrar la columna de fecha rumbo a `Report created` o similar.

## SEE ALSO

[numb-form](usage.md), [numb-form.conf](configuration.md), [../SECURITY.md](SECURITY.md)
