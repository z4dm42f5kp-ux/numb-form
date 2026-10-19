# NUMB-FORM(1)

> Traducción automática (Argos Translate) del [original en inglés](../usage.md). La versión en inglés es la que vale.

## NAME

**numb-form**, **numb-commit** — fill risk assessment forms

## SYNOPSIS

```
numb-form [-f conf] [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form [-f conf] check-template template [sample]
numb-form [-f conf] config [init]
numb-form keys
numb-commit [-f conf] [-run dir] [-downloads dir]
```

## DESCRIPTION

**numb-form** is started in the folder *above* the working folder. The working folder is the single folder found there. If there are several, the one containing an overview folder is used; otherwise the user is asked.

La carpeta de trabajo debe contener:

* una carpeta de visión general cuyo nombre contiene `overview_dir` (predeterminado `Übersicht`, que coincide con `3. Übersichten`), conteniendo exactamente una hoja de cálculo;
* la plantilla, llamada `template_name` (predeterminado `Appendix A - Risk Assesment`), en su nivel superior.

Los formatos compatibles se enumeran en [formats.md](formats.md). Los ajustes se describen en [configuration.md](configuration.md).

A run proceeds as follows:

1. La plantilla se revisa con un llenado de ensayo de cada campo. La carrera se detiene si la sección A, o cualquier etiqueta que no sea *Commercial Registry Number*, no se puede encontrar o no está vacía.
2. El nombre de la persona que procesa el formulario se toma de la configuración, o se escribe una vez.
3. Con `mode = review`, toda la carpeta de trabajo se copia a `Downloads/numb-form-runs/<date>_<time>_<id>/` y todo después de este paso sucede en la copia. Con `mode = direct` (el predeterminado), la carpeta de trabajo en sí se utiliza y el registro va a la carpeta de estado por usuario.
4. Los nombres de las empresas se leen desde la columna encabezada `Company Name`, de arriba a abajo. Todas las demás columnas son ignoradas, incluyendo `Report created`. La hoja de cálculo nunca se escribe a; su suma de comprobación SHA-256 se revisa antes y después de la carrera.
5. Para cada empresa, las carpetas en la carpeta de trabajo y todos sus subcarpetas son buscados. Los nombres son comparados ignorando caso, acentos y umlauts (ü = u = ue), puntuación y espacios. Un partido significa que la forma ya existe, y la compañía está saltada. Un partido cercano se muestra al usuario para confirmar (o, con `confirm = none`, reportado como *needs review* y saltado). Las carpetas con una forma legal diferente nunca coinciden: `Foo GmbH` y `Foo S.A.` son diferentes empresas.
6. Para una nueva empresa, una carpeta llamada después de la creación de la empresa. La plantilla se copia en ella como `<template name> - <Company Name>.<ext>`, y la sección A se llena.
7. El documento se lee de nuevo del disco y se muestra. El usuario lo acepta, edita un campo o lo rechaza; se elimina una carpeta rechazada. Con `confirm = none` el read-back debe igualar los valores exactamente.

Sólo se llenan las líneas de la sección A, *Information About the Contracting Party*. La sección B y todo lo que está fuera de la sección A quedan byte para byte sin cambios, y no se añade formato.

| Campo | Valor |
| --- | --- |
| Name / Company Name of the Contracting Party | la hoja de cálculo literal |
| Legal form | el sufijo del nombre, como está escrito (`GmbH`, `S.A.`, `C.A. de C.V.`) |
| Headquarters / Address | elegido de los candidatos encontrados |
| Commercial Registry Number | siempre en blanco |
| Authorized Representatives | `Full Name (Role 1, Role 2), Full Name (Role)`, roles en inglés |
| Date of the evaluation | hoy, `date_format` (predeterminado `DD/MM/YYYY`) |
| Person Processing the Form | desde la configuración o escribe al inicio |

Los candidatos son recogidos, en orden de autoridad:

1. UK Companies House (registro gubernamental; sólo con una llave gratuita);
2. el registro GLEIF LEI (regulación de registro; no llave);
3. el propio aviso legal de la empresa (Impressum, Aviso legal, Aviso legal);
4. Wikidata (editado comunitario; utilizado principalmente para encontrar el sitio web).

Las fuentes alimentan qué campo, la fuente menos confiable aceptada para los representantes y los inconvenientes se establecen en la configuración. Las personas y direcciones nunca son tomadas de una entidad madre, subsidiaria o similarmente nombrada. Los valores no verificados obtienen este texto exacto cuando `caution` lo dice:

> Precaución: Este campo puede diferir de fuentes internas o ser inexacto debido a la insuficiente información en línea

Cuando no se encuentra representante, el campo dice:

> No se encontraron representantes en línea, consulte otras fuentes internas para obtener información más precisa

**numb-commit** applies to review runs only. It copies the accepted folders of a run back into the real working folder. It only adds new folders, never overwrites or deletes anything, and refuses folders that appeared in the real tree in the meantime. Each folder is copied to a temporary name, verified with SHA-256 and then renamed. Documents edited during review are listed before the user confirms. Afterwards it offers to delete the review copy; the logs are kept.

## COMMANDS

`check-template template [sample]`
: Compruebe que una plantilla se puede rellenar, y opcionalmente escriba una muestra llena a *sample*. La plantilla solo se lee.

`config`
: Mostrar los archivos de configuración y los ajustes efectivos.

`config init`
: Escriba un archivo de configuración por usuario comentado con cada defecto.

`keys`
: Explica las claves de API gratuitas opcionales.

## OPTIONS

`-f file`
: Lea *file* después de los archivos de configuración del sistema y del usuario.

`-offline`
: Contacte sin fuente en línea.

`-root dir`
: Use *dir* como la carpeta de trabajo en lugar de detectarla.

`-downloads dir`
: Modo de revisión: ponga la copia en *dir*.

`-timeout duration`
: Parar con seguridad después de *duración*; en la mayoría de `25m`.

`-run dir`
: (numb-commit) Envíe esta carpeta de ejecución en lugar de la más reciente sin compromiso.

`-allow-non-tty`
: Lea las respuestas de una tubería. Sólo para pruebas automatizadas.

`-version`
: Imprima la versión.

## ENVIRONMENT

`NUMBFORM_DOWNLOADS`
: Carpeta Downloads para utilizar, por ejemplo, un volumen montado en un contenedor.

`NUMBFORM_COMPANIES_HOUSE_KEY`
: Opcional gratuito UK Companies House API key. Nunca conectado o almacenado.

`HTTPS_PROXY`, `NO_PROXY`
: Ajustes proxy para búsquedas en línea.

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

No hay caché ni archivos temporales detrás. Las respuestas de la red sólo se realizan en memoria, los documentos se escriben a través de un archivo temporal que se renombra o se elimina, y los convertidores trabajan en una carpeta temporal privada que se elimina.

## EXIT STATUS

0 sobre el éxito, 1 sobre el error, 2 sobre un error de uso o configuración, y 3 cuando se detiene por el límite de tiempo, por Ctrl+C o por el usuario.

## EXAMPLES

Compruebe la plantilla real y escriba una muestra, sin tocar la plantilla:

```
numb-form check-template "Appendix A - Risk Assesment.docx" sample.docx
```

Llene los formularios directamente en la carpeta de trabajo:

```
cd /path/to/folder-above && numb-form
```

Revisar primero, luego comprometer (numb-commit espera confirmación):

```
cd /path/to/folder-above && numb-form -f review.conf && numb-commit
```

donde `review.conf` contiene `mode = review`.

Corre sin acceso a la red:

```
cd /path/to/folder-above && numb-form -offline
```

## SEE ALSO

[configuration.md](configuration.md), [formats.md](formats.md), [install.md](install.md), [containers.md](containers.md), [../SECURITY.md](SECURITY.md), [../AGENTS.md](../../AGENTS.md)
