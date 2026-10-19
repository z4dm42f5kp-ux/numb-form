# Oficina suites y formatos de archivo

> Traducción automática (Argos Translate) del [original en inglés](../formats.md). La versión en inglés es la que vale.

numb-form funciona igual con cada suite de oficina. Edita los formatos abiertos comunes y utiliza el software de oficina ya en el ordenador sólo para formatos que no puede editar con seguridad.

## Plantillas

| Formato | Escrito por | Cómo numb-form lo llena | Archivo lleno |
| --- | --- | --- | --- |
| `.docx` | Microsoft Word, LibreOffice, Apache OpenOffice, Proton Docs, Apple Pages (exportación) | directamente | `.docx` |
| `.odt` | LibreOffice, Apache OpenOffice, Collabora, Microsoft Word | directamente | `.odt` |
| `.fodt` | LibreOffice (flat XML) | directamente | `.fodt` |
| `.rtf` | cada palabra procesador | directamente | `.rtf` |
| `.pages` | Apple Pages | convertido, luego lleno | `.docx` (ver más abajo) |
| `.doc` | Word 97-2003 | legado: convertido, luego lleno | `.docx`, o `.doc` con `legacy_output = same` |

"Directly" significa que el archivo nunca es re-salvado por otro programa: los valores se insertan en el propio texto de la etiqueta, y todo lo demás permanece byte for byte as it was. El resultado es leído y revisado después de cada llenado.

Cuando la carpeta de trabajo sostiene la plantilla en varios formatos, la primera en la tabla gana.

## Hojas de cálculo

| Formato | Escrito por | Cómo numb-form lo lee |
| --- | --- | --- |
| `.xlsx`, `.xlsm` | Microsoft Excel, LibreOffice Calc, Proton Sheets, Apple Numbers (exportación) | directamente; las macros nunca se ejecutan |
| `.ods`, `.fods` | LibreOffice Calc, Apache OpenOffice | directamente |
| `.csv`, `.tsv` | cada programa de hojas de cálculo | directamente; UTF-8 o Windows-1252, `;` `,` o pestaña |
| `.numbers` | Apple Numbers | convertido en memoria |
| `.xls` | Excel 97-2003 | legado: directamente, con el propio lector de numb-form |

La hoja de cálculo sólo es leída, nunca escrita.

## Formatos convertidos

`.doc`, `.pages` y `.numbers` se convierten con software ya en el equipo:

* LibreOffice o Apache OpenOffice, en cada sistema. numb-form busca los lugares de instalación habituales; establece `soffice = <path>` si es en otro lugar.
* En macOS, Apple Pages y Numbers, cuando se instala (`pages_converter` y `numbers_converter`, `auto` predeterminado).

La conversión se ejecuta en una carpeta temporal privada con un perfil de oficina desechable, y todo se elimina después. Nada se descarga.

### Apple Pages

Los archivos Pages usan un formato indocumentado. numb-form por lo tanto lee una plantilla `.pages` convirtiéndola en formato Word, llena eso y guarda los formularios llenos como `.docx`, que Pages abre. Revise el diseño de la primera forma llena.

En macOS con Pages instalado, `pages_write = yes` ahorra las formas llenas como `.pages` a través de Pages en sí. Este es **experimental**: se prueba, pero se apaga por defecto, y cada archivo se convierte de nuevo y se verifica después de guardar. Si no necesita archivos `.pages`, déjelo apagado.

### Formatos de Legado (.doc, .xls)

Los formatos binarios de Legacy están deshabilitados por defecto. Un administrador los permite en el archivo de configuración del sistema (ver [configuration.md](configuration.md)):

```
legacy_formats = yes
```

Corres que usan un formato legado siempre están completamente fuera de línea. El lector `.xls` es el propio numb-form y es probado continuamente.

## Suites de oficina

* **Microsoft 365 / Office**: guardar la plantilla como `.docx` y la vista general como `.xlsx` (los predeterminados).
* **LibreOffice** and **Apache OpenOffice**: `.odt` and `.ods` work directly; so do `.docx` and `.xlsx`.
* **Proton Drive (Docs and Sheets)**: descargar la plantilla como `.docx` y la vista general como `.xlsx` en la carpeta de trabajo. Esto también funciona en sistemas sin una suite de oficina de escritorio, incluyendo los BSD, desde un navegador web.
* **Apple Pages and Numbers**: use `.pages` and `.numbers` as described above, or export to `.docx` and `.xlsx` (File → Export To).
* **El BSDs**: LibreOffice está disponible como un paquete (`pkg install libreoffice`, `pkg_add libreoffice`, `pkgin install libreoffice`); Proton Docs en un navegador no necesita ninguna instalación.
