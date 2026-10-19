# NUMB-FORM(1)

> Traducción automática (Argos Translate) del [original en inglés](../usage.md). La versión en inglés es la que vale.

## NAME

**numb-form**, **numb-commit** — fill risk assessment forms

## SYNOPSIS

```
numb-form [-f conf] [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form [-f conf] check-template template [sample]
numb-form [-f conf] config [init | set key value]
numb-form keys
numb-form [-f conf] update [stable | edge]
numb-commit [-f conf] [-run dir] [-downloads dir]
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
4. Los nombres de las empresas se leen desde la columna encabezada `Customer Name`, de arriba a abajo. Todas las demás columnas son ignoradas, incluyendo `Report created`. La hoja de cálculo nunca se escribe a; su suma de comprobación SHA-256 se revisa antes y después de la carrera.
5. Para cada empresa, las carpetas en la carpeta de trabajo y todos sus subcarpetas son buscados. Los nombres son comparados ignorando caso, acentos y umlauts (ü = u = ue), puntuación y espacios. Un partido significa que la forma ya existe, y la compañía está saltada. Un partido cercano se muestra al usuario para confirmar (o, con `confirm = none`, reportado como *needs review* y saltado). Las carpetas con una forma legal diferente nunca coinciden: `Foo GmbH` y `Foo S.A.` son diferentes empresas.
6. Para una nueva empresa, una carpeta llamada después de la creación de la empresa. La plantilla se copia en ella como `<template name> - <Company Name>.<ext>`, y la sección A se llena.
7. El documento se lee de nuevo del disco y se muestra. El usuario lo acepta, edita un campo o lo rechaza; se elimina una carpeta rechazada. Con `confirm = none` el read-back debe igualar los valores exactamente.

Sólo se llenan las líneas de la sección A, *Information About the Contracting Party*. La sección B y todo lo que está fuera de la sección A quedan byte para byte sin cambios, y no se añade formato.

| Campo | Valor |
| --- | --- |
| Name / Company Name of the Contracting Party | la hoja de cálculo literal |
| Legal form | encontrado en el nombre, también antes de una filial o nota de país (`Foo GmbH (Wien)`), escrito como en el nombre con capitales convencionales (`GMBH` da `GmbH`); preguntado sólo cuando el nombre no tiene ninguno |
| Headquarters / Address | elegido de los candidatos encontrados |
| Commercial Registry Number | siempre `No commercial registry number was found online consult internal sources for further information`; puede ser reemplazado por el número con *edit* |
| Authorized Representatives | `Full Name (Role 1, Role 2), Full Name (Role)`, roles en inglés |
| Date of the evaluation | hoy, `date_format` (predeterminado `DD/MM/YYYY`) |
| Person Processing the Form | desde la configuración o escribe al inicio |

Todo es buscado en un pase por compañía, sin preguntas. Los candidatos proceden, en orden de autoridad:

1. government registers: UK Companies House (only with a free key); EU VIES, for an EU VAT number printed on the company's website (name and address; Germany confirms only that the number is valid); Receita Federal, for a Brazilian CNPJ printed on the website (address, administrators and directors), read through the free BrasilAPI mirror; government gazettes and registers found by searching their websites for the company: BORME (Spain), Portal da Justiça (Portugal), Boletín Oficial (Argentina), Sunbiz (Florida), Companies House pages (United Kingdom); which one depends on the countries the legal form points to;
2. el registro GLEIF LEI (regulación de registro);
3. el propio sitio web de la empresa: su aviso legal (Impressum, Aviso legal, Mentions légales, Aviso legal) primero, luego sobre, páginas de contacto y privacidad;
4. Wikidata, y los fragmentos de los resultados de la búsqueda web (datos comunitarios, no verificados; los snippets se leen solamente cuando nada más se nombra a nadie).

El sitio web se encuentra automáticamente: desde Wikidata, desde un DuckDuckGo buscar el nombre exacto, y desde direcciones como `foobar.de` o `foobar.com.mx` construidas desde el nombre y los países de su forma legal. Un sitio se utiliza sólo si una de sus páginas llama a la empresa con su forma legal. En el aviso legal de un grupo que nombre varias empresas, sólo se lee la parte sobre la empresa.

Las etiquetas son reconocidas en alemán, inglés, español, portugués, francés, italiano, holandés, polaco, checo, eslovaco y nórdico, en cualquier lugar de una línea, en tablas y a pie, y se traducen a roles de inglés. Se eliminan las etiquetas que nombran a otras personas (papel de supervisión, oficial de protección de datos, persona responsable de contenido, contacto, agente registrado) y personas que se fueron (*Ceses/Dimisiones*, *ehemalig*, *former*).

Los números oficiales y las entradas de registro impresos en el sitio web (por ejemplo `HRB 12345` o `FN 123456 a`) se muestran antes de que el formulario esté escrito; el número se puede escribir en *Commercial Registry Number* con *edit*.

DuckDuckGo se le pide a la mayoría de unas cuantas veces por compañía, a unos segundos de distancia. Si pide un cheque humano, busca pausas por un minuto y las otras fuentes se utilizan mientras tanto.

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

`config set key value`
: Guardar una configuración en el archivo de configuración por usuario, por ejemplo `numb-form config set person "Erika Mustermann"`. El valor se verifica antes de que se guarde.

`keys`
: Explica las claves de API gratuitas opcionales.

`man` [*topic*] [`en` | `es` | `de`]
: Mostrar este manual en el terminal, una pantalla a la vez (Enter: siguiente pantalla; q y Enter: renunciar). Sin un tema se muestra todo el manual, comenzando por el guía. Temas: `guide`, `usage`, `configuration`, `install`, `formats`, `containers`, `security`; `numb-form man topics` los lista. El idioma sigue `LANG`, o se da. `numb-form man -install` hace que `man numb-form` funcione: en Linux, macOS y BSD escribe la página manual a `~/.local/share/man/man1` (refreshed by `numb-form update`); en Windows escribe `man.cmd` junto a `numb-form.exe`, para Command Prompt. Windows PowerShell mantiene `man` por su propia ayuda: allí, escriba `numb-form man`.

`update` [`stable` ← `edge`]
: Reemplazar el `numb-form` instalado y `numb-commit` con la nueva versión estable o la nueva construcción de desarrollo `edge`, como lo hacen los comandos de instalación: el archivo y `SHA256SUMS` se descargan sobre HTTPS de GitHub solamente, la suma de comprobación debe coincidir, y el nuevo `numb-form` debe comenzar antes de que todo sea reemplazado. Sin un argumento se utiliza el canal de la construcción de ejecución. Se mantienen los ajustes. Refusado con `-offline`, con `online = no`, y para construcciones de origen. En Windows el viejo programa es renombrado a `numb-form.old.exe` y eliminado al siguiente inicio.

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

0 sobre el éxito, 1 sobre error, 2 sobre un error de uso o configuración, y 3 cuando se detiene por el límite de tiempo, por Ctrl+C o por el usuario.

## EXAMPLES

Compruebe la plantilla real y escriba una muestra, sin tocar la plantilla:

```
numb-form check-template "Appendix A - Risk Assesment.docx" sample.docx
```

Llene los formularios directamente en la carpeta de trabajo:

```
cd /path/to/GWG && numb-form
```

Revisar primero, luego comprometer (numb-commit espera confirmación):

```
cd /path/to/GWG && numb-form -f review.conf && numb-commit
```

donde `review.conf` contiene `mode = review`.

Corre sin acceso a la red:

```
cd /path/to/GWG && numb-form -offline
```

## SEE ALSO

[configuration.md](configuration.md), [formats.md](formats.md), [install.md](install.md), [containers.md](containers.md), [../SECURITY.md](SECURITY.md), [../AGENTS.md](../../AGENTS.md)
