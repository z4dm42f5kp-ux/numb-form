# Seguridad

> Traducción automática (Argos Translate) del [original en inglés](../../SECURITY.md). La versión en inglés es la que vale.

## Reporting a vulnerability

Utilice la información de vulnerabilidad privada de GitHub: *Security* → *Informe una vulnerabilidad* en este repositorio. No abra una cuestión pública. No adjuntar los datos del cliente; describir el problema con los nombres inventados.

Sólo se admite la última versión.

## Lo que prometen los programas

numb-form maneja documentos de cliente, por lo que las reglas de diseño son estrictas.

* **numb-form** only ever adds: a new folder per new company, with one filled form in it. It never changes or deletes an existing file or folder, and it removes only folders it created itself in the same run (on rejection or error).
* Con `mode = review`, la carpeta de trabajo real no está escrita en absoluto: todo el árbol es copiado a Downloads primero y cada cambio se hace en la copia.
* **numb-commit** (review mode) only adds new folders. It never overwrites or deletes, refuses folders that appeared in the meantime, copies to a temporary name, verifies every file with SHA-256 and then renames.
* La hoja de cálculo en `3. Übersichten` se abre sólo lectura y nunca se escribe. Su suma de comprobación SHA-256 se compara antes de la copia, después de la copia y al final de la carrera; una diferencia para la ejecución.
* Los documentos no son reserializados. El XML solo se escanea para los offsets de byte y los valores se introducen en los bytes originales; todas las otras partes del `.docx` son copiadas por byte para byte. Después de cada llenado el resultado se analiza de nuevo y se compara con el párrafo de plantilla por párrafo. Los valores son compatibles con XML y no se admiten caracteres en XML. Se rechazan las plantillas con macros.
* Nunca se añade ningún formato. Los valores se unen a la propia edición de texto de la etiqueta, o una copia de las propiedades existentes de esa carrera cuando la etiqueta es seguida por una pestaña.
* Cada documento se lee de nuevo del disco después de escribir. Con `confirm = each` (por defecto) una persona lo ve y lo acepta, y los programas se niegan a funcionar sin un terminal interactivo. Con `confirm = none` el read-back debe igualar exactamente los valores previstos, los partidos cercanos al nombre de carpeta nunca se deciden automáticamente, y sólo se utilizan las fuentes y el nivel de confianza en la configuración.
* Una carrera se detiene seguro después de 25 minutos. Escribe a la mayoría de 100 MB de documentos además de la copia, y no deja ningún caché o archivos temporales: los datos de red se quedan en memoria, los archivos temporales se renombran o eliminan, y los convertidores se ejecutan en una carpeta temporal privada con un perfil de oficina desechable, eliminados después.

## Configuración y formatos

* El analizador de configuración rechaza claves desconocidas y valores inválidos.
* Las claves que ensanchan lo que numb-form acepta (`legacy_formats`) se honran sólo en el archivo del sistema. En Unix ese archivo es ignorado a menos que pertenezca a root y no es obra de grupo u otros; en Windows vive bajo Archivos del programa.
* Los formatos binarios Legacy (.doc, .xls) están apagados por defecto, y las carreras que los usan siempre están completamente fuera de línea. El lector de .xls (archivo completo y BIFF8) es el propio numb-form controla cada offset y cadena, y se fusiona continuamente.
* .doc, .pages y .numbers son convertidos por software de oficina ya en el equipo (LibreOffice, Apache OpenOffice, Apple Pages o Numbers), con argumentos fijos; nombres de archivos se transmiten como argumentos, nunca como texto de script. Escribir .pages es experimental y apagado por defecto.
* Se rechazan las plantillas con macros; se leen hojas de cálculo (.xlsm) macróficas para valores celulares solamente.

## Red

Los buscadores en línea envían sólo nombres de empresa y el sitio web de la empresa el usuario confirma. Ellos van a la API GLEIF, Wikidata, el propio sitio web de la empresa y, sólo cuando se establece una llave gratuita, UK Companies House. Utilice `-offline` para enviar nada.

* HTTPS sólo, TLS 1.2 o posterior, a la mayoría de cinco HTTPS redirige.
* Sin un proxy, se rechazan las conexiones con el bucleback, las direcciones privadas, locales de enlace y multicast (solicitud de guardia de falsificación del lado del servidor).
* Los cuerpos de respuesta están tapados a 2 MB, mantenidos en memoria y nunca escritos en disco.
* El texto de la red está despojado de caracteres de control, secuencias de escape ANSI y anulaciones bidireccionales antes de que se muestre, por lo que una página web no puede manipular el terminal.
* Los archivos se leen con límites de tamaño (guardiente de bomba de zip).
* Las claves de API vienen sólo del medio ambiente, se envían sólo a su propio servicio, y nunca se registran o almacenan.
* Los datos nunca se toman de una entidad madre, subsidiaria o similarmente nombrada.

## Cadena de suministro

**No hay módulos Go de terceros.** Los programas utilizan sólo la biblioteca estándar Go. `go.mod` no tiene líneas `require`, y `scripts/check-deps.sh` falla la construcción si aparece una dependencia, `go.sum` o `vendor/`, o si se utiliza cualquier importación no estándar. Con nada que descargar, no hay nada que alucinar, tiposquat o manipular.

**PGP y sus reemplazos. ** La autenticidad se verifica con los mecanismos que cada ecosistema realmente apoya. PGP no se utiliza porque los ecosistemas involucrados ya no dependen de él (PyPI removió el soporte de firma PGP en 2023).

| Componente | Verificación |
| --- | --- |
| Go toolchain in CI | descargado por `actions/setup-go` de go.dev y chequeado contra el publicado SHA-256 |
| Herramientas en el CI (estático, govulncheck, gosec) | versiones exactas, verificadas por `go run` contra la base de datos Go checksum, un registro de transparencia pública |
| GitHub Actions | sujetado a SHAs de compromiso completo; Dependabot propone actualizaciones después de una refrigeración de 14 días |
| Python tools in CI (zizmor, Argos Translate) | todos los paquetes conectados con SHA-256 (`--require-hashes`), ruedas solamente (`--only-binary :all:`), sin resolución de dependencia en el tiempo de instalación (`--no-deps`) |
| Modelos Argos Translate | índice leído en un compromiso de pinned; el SHA-256 de cada modelo se fija en `.github/translate/models.sha256`, grabado en el primer uso y aplicado después |
| Imagen de base de contenedores | ninguna: las imágenes se construyen `FROM scratch` |

**Releases.** Los binarios de liberación son construidos por GitHub Actions de una etiqueta, con `-trimpath` y un timetamp fijo, por lo que son reproducibles. Cada versión tiene:

* `SHA256SUMS` (formato GNU) y `SHA256` (formato BSD) sumas de comprobación;
* `SHA256SUMS.sigstore.json`, una firma Sigstore sin llave de las sumas de comprobación, vinculada a la identidad del flujo de trabajo de liberación;
* SLSA build provenance attestations for every archive and binary (`gh attestation verify`);
* a imagen de contenedor en ghcr.io, firmado con cosign y certificado.

[docs/install.md](install.md) muestra cómo verificar cada uno de ellos.

## Gasoducto de pruebas de seguridad

Pull requests run a fast gate (formatting, vet, staticcheck, tests en Linux, macOS y Windows, govulncheck, gosec, zizmor, the history scan and dependncy review). Empuja a `main` y ramas de agente, y un horario nocturno, ejecutar todo:

| Check | Herramienta |
| --- | --- |
| No hay dependencias, formato, veterinario de seis sistemas operativos | `scripts/check-deps.sh`, `gofmt`, `go vet` |
| Análisis estadístico | Control estático |
| vulnerabilidades conocidas en código y en binarios construidos, para cada OS objetivo | Govulncheck |
| Forro de seguridad, reportado a escaneo de código | Gosec |
| Análisis semántico de Go y de los flujos de trabajo | CodeQL (`security-extended`, `security-and-quality`) |
| Fuzzing of every template filler and scanner (docx, odt, rtf), cada lector de hojas de cálculo (xlsx, ods, csv, xls legado), el parser de configuración y el matcher de nombre | Ir a fuzzing nativo, 2 minutos por objetivo en principal, 10 noches |
| Plantillas y hojas de cálculo escritas por LibreOffice real (odt, fodt, rtf, docx, doc, xls, ods, fods) | formatos de trabajo |
| Seguridad del flujo de trabajo (inyección, credenciales, permisos, pins) | zizmor, pedantic |
| Archivos de cliente o credenciales cometidas por error | historia escaneada |
| Pruebas con el detector de carreras en Linux, macOS y Windows, con el más antiguo y nuevo Go | `go test -race` |
| Pruebas sobre FreeBSD, OpenBSD y NetBSD | máquinas virtuales |
| Container image construye y se niega a funcionar sin un terminal | Docker, podman |
| Nuevas dependencias en las solicitudes de exención | examen de la dependencia |

OpenSSF Scorecard revisa las prácticas del repositorio. Pull requests, including Dependabot's, never need write permissions: results that require them are uploaded only from `main`.

Los flujos de trabajo comienzan sin permisos y conceden cada trabajo sólo lo que necesita; los cheques no guardan credenciales.

## Contenedores y Kubernetes

La imagen contiene sólo los dos binarios, un paquete CA y la base de datos de zona horaria, y funciona como UID 65532. Los comandos y manifiestos documentados utilizan un sistema de archivos raíz de sólo lectura, bajan todas las capacidades, prohiben la escalada de privilegios, montan la carpeta de trabajo sólo para numb-form, no montan ninguna cuenta de servicio y establecen un plazo.

## Este repositorio es público

Nunca comprometer documentos de clientes, hojas de cálculo, registros o carpetas de ejecución. `.gitignore` los excluye y CI los rechaza.
