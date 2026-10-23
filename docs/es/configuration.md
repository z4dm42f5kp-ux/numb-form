# NUMB-FORM.CONF(5)

> Traducción automática (Argos Translate) del [original en inglés](../configuration.md). La versión en inglés es la que vale.

## NAME

**numb-form.conf** — numb-form configuration file

## DESCRIPTION

numb-form funciona sin un archivo de configuración; los defectos a continuación son el comportamiento documentado. Los archivos se leen en este orden, cada uno que anula el anterior:

1. El archivo del sistema, escrito por un administrador:

   | Sistema | Camino |
   | --- | --- |
   | Linux | `/etc/numb-form.conf` |
   | FreeBSD, OpenBSD, NetBSD | `/usr/local/etc/numb-form.conf`, más `/etc/numb-form.conf` |
   | macOS | `/Library/Application Support/numb-form/numb-form.conf` |
   | Windows | `%ProgramFiles%\numb-form\numb-form.conf` |

   En los sistemas Unix se ignora a menos que pertenezca a la raíz y no sea obra de grupo u otros.
2. El archivo por usuario:

   | Sistema | Camino |
   | --- | --- |
   | Linux, BSD | `$XDG_CONFIG_HOME/numb-form/numb-form.conf` (`~/.config/numb-form/`) |
   | macOS | `~/Library/Application Support/numb-form/numb-form.conf` |
   | Windows | `%APPDATA%\numb-form\numb-form.conf` |

3. Un archivo dado con `-f file`.

Cada línea es `key = value`. Las líneas que comienzan con `#` son comentarios, y ` #` comienza un comentario después de un valor. Los valores se pueden citar con `"`. Las claves desconocidas y los valores inválidos son errores, por lo que los errores de escritura se encuentran inmediatamente. Las claves de versiones anteriores (`confirm`, `address_sources`, `address_fallback`, `reps_sources`, `reps_min_trust`, `reps_fallback`, `caution`, `website`, `search`, `browser`, `guess_domains`) son ignoradas con una nota.

Guardar un solo ajuste sin editar el archivo:

```
numb-form config set person "Erika Mustermann"
```

Cree un archivo por usuario comentado con cada defecto y muestre la configuración efectiva:

```
numb-form config init && numb-form config
```

## KEYS

### Flujo de trabajo

`mode = direct` | `review`
: **direct** (default): carpetas y formas se crean en la carpeta de trabajo en sí, después de cada confirmación. **review**: toda la carpeta de trabajo se copia por primera vez a Downloads, todo el trabajo sucede en la copia, y `numb-commit` copia las carpetas aceptadas de vuelta.

`online = yes` | `no`
: Deje que `numb-form update` contacte con GitHub (por defecto sí). Las empresas nunca se buscan en línea.

`timeout = none`
: Deténgase con seguridad después de este largo, por ejemplo `30m` o `2h`; **none** (por defecto) no significa límite de tiempo.

`max_write_mb = 100`
: La mayoría de los megabytes de documentos que se ejecutan pueden escribir (1-100).

### Carpetas y archivos

`overview_dir = Übersicht`
: Pattern of the folder that holds the spreadsheet. Case, accents and punctuation are ignored, and umlauts match both spellings, so `Übersicht` matches `3. Übersichten` and `Uebersicht`.

`header = Customer Name`, `id_header = Customer ID`
: Cabeceras de columna de hoja de cálculo.

`template_name = Appendix A - Risk Assesment`
: Nombre de archivo de plantilla sin extensión; `Assessment` también coincide.

`downloads =`
: Modo de revisión: carpeta para la copia; vacío significa su carpeta Downloads.

`log_dir =`
: Modo directo: carpeta para registros y manifiestos; vacío significa la carpeta de estado por usuario (`%LOCALAPPDATA%\numb-form\runs`, `~/Library/Logs/numb-form` o `~/.local/state/numb-form/runs`).

### Valores

`person =`
: Nombre de la persona que procesa el formulario; medios vacíos piden una vez por ejecución.

`date_format = DD/MM/YYYY`
: Fecha de evaluación; fichas `DD`, `MM`, `YYYY`, `YY`.

`fuzzy_threshold = 0.85`
: Cómo similar debe ser un nombre de carpeta para ser ofrecido como un partido (0.5-1).

### Formatos

`soffice =`
: Camino a LibreOffice o Apache OpenOffice; búsqueda de medios vacíos.

`pages_converter = auto` Silencio `pages-app` Silencio `libreoffice` `numbers_converter = auto` Silencio `numbers-app` Silencio `libreoffice`
: Cómo se convierten los archivos `.pages` y `.numbers`. **auto** utiliza la aplicación de Apple en macOS cuando se instala, si no LibreOffice.

`pages_write = no`
: EXPERIMENTAL, macOS solamente: guardar formularios llenos como `.pages` a través de Apple Pages.

`legacy_formats = no`
: SYSTEM FILE ONLY: permitir plantillas `.doc` y hojas de cálculo `.xls`.

`legacy_output = docx` | `same`
: Las formas llenas de una plantilla `.doc` son `.docx`, o `.doc` de nuevo.

## EXAMPLES

Guarda tu nombre para que no se le pregunte en cada carrera:

```
person = Max Muster
```

Su flujo de trabajo anterior, con una copia de revisión en Downloads:

```
mode = review
```

Una hoja de cálculo alemana:

```
overview_dir = Übersicht
header = Firmenname
id_header = Kundennummer
```

## SEE ALSO

[usage.md](usage.md), [formats.md](formats.md)
