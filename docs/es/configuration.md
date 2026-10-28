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

Cree un archivo por usuario comentado con cada defecto y muestre la configuración efectiva.

**Unix-like** (macOS, Linux, BSD), en Terminal:

```
numb-form config init && numb-form config
```

**Windows** (10, 11), en PowerShell:

```
numb-form config init; numb-form config
```

## KEYS

The programs that use a key are named in brackets: **form** (numb-form), **commit** (numb-commit), **date** (numb-date), **update** (numb-form update). `numb-form config` lists every key with its effective value and the file that set it.

### Encontrar los archivos

`overview_dir = Übersicht` [form, date]
: Pattern of the folder that holds the spreadsheet. Case, accents and punctuation are ignored, and umlauts match both spellings, so `Übersicht` matches `3. Übersichten` and `Uebersicht`.

`header = Customer Name` [form, date]
: Dirección de la columna de nombre de empresa. numb-date también intenta nombres comunes (`Customer`, `Company Name`, `Kunde`, `Cliente` y otros) y encuentra la columna *Report created* por sí misma.

`id_header = Customer ID` [form, date]
: Dirección de la columna número cliente, mostrada junto a cada empresa.

`template_name = Appendix A - Risk Assesment` [form, date]
: Nombre de archivo de plantilla sin extensión; `Assessment` también coincide.

### Valores escritos en la forma

`person =` [form]
: Nombre de la persona que procesa el formulario; medios vacíos piden una vez por ejecución.

`date_format = DD/MM/YYYY` [form]
: Fecha de evaluación; fichas `DD`, `MM`, `YYYY`, `YY`. numb-date siempre escribe `DD-MM-YYYY`.

### Cómo funciona la carrera

`mode = direct` TENIDO `review` [form, commit]
: **direct** (default): carpetas y formas se crean en la propia carpeta de trabajo. **review**: toda la carpeta de trabajo se copia por primera vez a Downloads, todo el trabajo sucede en la copia, y `numb-commit` copia las carpetas aceptadas de vuelta.

`fuzzy_threshold = 0.85` [form, date]
: Cómo similar debe ser un nombre de carpeta para ser ofrecido como un partido (0.5-1).

`unicode = basic` TENIDO `full` [form, date]
: Cómo se comparan los nombres de las empresas y carpetas. **basic** (default): caso, espaciamiento, puntuación y los acentos de los idiomas de Europa occidental y central (`é`, `ç`, `ł`, `ü` como `u` o `ue`) son ignorados. **full**: también los acentos y formas especiales de cada otro alfabeto, por ejemplo `ș`, `ğ`, `ǎ` griego `ά`, ligaduras como `ﬁ` y letras de ancho completo como `Ａ`. La tabla extra se lee sólo cuando se establece **full**, por lo que el valor predeterminado no añade trabajo.

`timeout = none` [form, commit, date]
: Deténgase con seguridad después de este largo, por ejemplo `30m` o `2h`; **none** (por defecto) no significa límite de tiempo.

`max_write_mb = 100` [form]
: La mayoría de los megabytes de documentos que se ejecutan pueden escribir (1-100).

`downloads =` [form, commit]
: Modo de revisión: carpeta para la copia; vacío significa su carpeta Downloads.

`log_dir =` [form, date]
: Carpeta para troncos; vacía significa la carpeta de estado por usuario (`%LOCALAPPDATA%\numb-form\runs`, `~/Library/Logs/numb-form` o `~/.local/state/numb-form/runs`). Los registros de revisión se quedan con la copia en Downloads.

`commit_from =` [commit]
: La copia de trabajo que numb-commit compara, por ejemplo `C:\Users\you\Downloads\14. GWG`; véase [numb-commit.md](numb-commit.md). Vacío: pregunta numb-commit.

`commit_to =` [commit]
: La carpeta de trabajo real que numb-commit compara la copia de trabajo con, por ejemplo `F:\14. GWG`. Vacío: pregunta numb-commit.

`online = yes` Silencio `no` [update]
: Deje que `numb-form update` contacte a GitHub. Nada más va en línea.

### Formatos

`soffice =` [form]
: Camino a LibreOffice o Apache OpenOffice; búsqueda de medios vacíos.

`pages_converter = auto` TENIDO `pages-app` TENIDO `libreoffice`, `numbers_converter = auto` TENIDO `numbers-app` ANTE `libreoffice` [form]
: Cómo se convierten los archivos `.pages` y `.numbers`. **auto** utiliza la aplicación de Apple en macOS cuando se instala, si no LibreOffice.

`pages_write = no` [form]
: EXPERIMENTAL, macOS solamente: guardar formularios llenos como `.pages` a través de Apple Pages.

`legacy_formats = no` [form]
: SYSTEM FILE ONLY: permitir plantillas `.doc` y hojas de cálculo `.xls`.

`legacy_output = docx` Silencio `same` [form]
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
