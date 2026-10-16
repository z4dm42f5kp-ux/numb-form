# Seguridad

> Traducción automática (Argos Translate) del [original en inglés](../../SECURITY.md). La versión en inglés es la que vale.

## Reporting a vulnerability

Usar el informe de vulnerabilidad privada de GitHub:
*Security* → *Informe una vulnerabilidad* en este repositorio.
No abra una cuestión pública.
No adjuntar los datos del cliente; describir el problema con los nombres inventados.

Sólo se admite la última versión.

## Lo que prometen los programas

numb-form maneja los documentos del cliente, por lo que las reglas del diseño son estrictas.

* La carpeta de trabajo real nunca está escrita por **numb-form**.
  Todo el árbol es copiado a Descargas primero y cada cambio se hace en
  la copia.
* **numb-commit** solo añade nuevas carpetas.
  Nunca sobrescribe o elimina, rechaza carpetas que aparecieron en el
  mientras tanto, copias a un nombre temporal, verifica cada archivo con SHA-256
  y luego renombres.
* La hoja de cálculo en `3. Übersichten` se abre sólo lectura y nunca es
  escrito.
  Su suma de comprobación SHA-256 se compara antes de la copia, después de la copia y
  al final de la carrera; una diferencia detiene la carrera.
* Los documentos no son reserializados.
  El XML solo se escanea para los offsets de byte y los valores se introducen en el
  bytes originales; todas las otras partes de la `.docx` son copiados byte para byte.
  Después de cada llenado el resultado se analiza de nuevo y se compara con el
  plantilla párrafo por párrafo.
  Los valores son compatibles con XML y no se admiten caracteres en XML.
  Se rechazan las plantillas con macros.
* Nunca se añade ningún formato.
  Los valores se unen a la propia edición de texto de la etiqueta, o una copia de esa carrera ya existente
  propiedades cuando la etiqueta es seguida por una pestaña.
* Cada documento se muestra a una persona, se lee de nuevo del disco, y debe ser
  aceptado.
  Los programas se niegan a funcionar sin un terminal interactivo.
* Una carrera se detiene seguro después de 25 minutos.
  Escribe a la mayoría de 100 MB de documentos además de la copia, y no deja
  caché o archivos temporales: los datos de red permanecen en memoria, y temporalmente
  Los archivos son renombrados o eliminados.

## Red

Los buscadores en línea envían sólo nombres de empresa y el sitio web de la empresa el usuario
confirma.
Van a la GLEIF API, Wikidata, el propio sitio web de la empresa y,
sólo cuando se establece una llave gratuita, UK Companies House.
Utilice `-offline` para enviar nada.

* HTTPS sólo, TLS 1.2 o posterior, a la mayoría de cinco HTTPS redirige.
* Sin un proxy, conexiones a loopback, privado, enlace-local y
  Se rechazan direcciones multicast (solicitud de guardia del servicio).
* Los cuerpos de respuesta están tapados a 2 MB, mantenidos en memoria y nunca escritos
  disco.
* Texto de la red está despojado de caracteres de control, escape ANSI
  secuencias y anulaciones bidireccionales antes de que se muestre, por lo que una página web
  no puede manipular el terminal.
* Los archivos se leen con límites de tamaño (guardiente de bomba de zip).
* Las claves de API vienen sólo del medio ambiente, se envían sólo a su propio
  servicio, y nunca se registran o almacenan.
* Los datos nunca se toman de una entidad madre, subsidiaria o similarmente nombrada.

## Cadena de suministro

**No hay módulos Go de terceros.**
Los programas utilizan sólo la biblioteca estándar Go.
`go.mod` no tiene líneas `require` y `scripts/check-deps.sh` falla
construir si aparece una dependencia, `go.sum` o `vendor/`, o si
se utiliza la importación no estándar.
Con nada que descargar, no hay nada que alucinar, tiposquat o
manipular.

**PGP y sus reemplazos. **
La autenticidad se verifica con los mecanismos que cada ecosistema en realidad
soportes.
PGP no se utiliza porque los ecosistemas involucrados ya no confían en él
(PyPI removió el soporte de firma PGP en 2023).

| Componente | Verificación |
| --- | --- |
| Go toolchain in CI | descargado por `actions/setup-go` de go.dev y chequeado contra el publicado SHA-256 |
| Herramientas en el CI (estático, govulncheck, gosec) | versiones exactas, verificadas por `go run` contra la base de datos Go checksum, un registro de transparencia pública |
| GitHub Actions | sujetado a SHAs de compromiso completo; Dependabot propone actualizaciones después de una refrigeración de 14 días |
| Python tools in CI (zizmor, Argos Translate) | todos los paquetes conectados con SHA-256 (`--require-hashes`), ruedas solamente (`--only-binary :all:`), sin resolución de dependencia en el tiempo de instalación (`--no-deps`) |
| Argos Traducir modelos | índice leído en un compromiso de pinned; el SHA-256 de cada modelo se fija en `.github/translate/models.sha256`, grabado en el primer uso y aplicado después |
| Imagen de base de contenedores | ninguna: las imágenes se construyen `FROM scratch` |

**Releases.**
Los binarios de liberación son construidos por GitHub Actions de una etiqueta, con
`-trimpath` y un timetamp fijo, por lo que son reproducibles.
Cada versión tiene:

* `SHA256SUMS` (formato GNU) y `SHA256` (formato BSD) sumas de comprobación;
* `SHA256SUMS.sigstore.json`, una firma sin llave de Sigstore
  cheques, vinculados a la identidad del flujo de trabajo de liberación;
* SLSA construye certificados de procedencia para cada archivo y binario
  (`gh attestation verify`);
* a imagen de contenedor en ghcr.io, firmado con cosign y certificado.

[docs/install.md](install.md) muestra cómo verificar cada uno de ellos.

## Gasoducto de pruebas de seguridad

Cada petición de empuje y tire corre:

| Check | Herramienta |
| --- | --- |
| No hay dependencias, formato, veterinario de seis sistemas operativos | `scripts/check-deps.sh`, `gofmt`, `go vet` |
| Análisis estadístico | Control estático |
| vulnerabilidades conocidas en código y en binarios construidos, para cada OS objetivo | Govulncheck |
| Forro de seguridad, reportado a escaneo de código | Gosec |
| Análisis semántico de Go y de los flujos de trabajo | CodeQL (`security-extended`, `security-and-quality`) |
| Fuzzing del llenador de documentos, lector de hojas de cálculo y matcher de nombre | Ir a fuzzing nativo, 2 minutos por objetivo |
| Seguridad del flujo de trabajo (inyección, credenciales, permisos, pins) | zizmor, pedantic |
| Archivos de cliente o credenciales cometidas por error | historia escaneada |
| Pruebas con el detector de carreras en Linux, macOS y Windows, con el más antiguo y nuevo Go | `go test -race` |
| Pruebas sobre FreeBSD, OpenBSD y NetBSD | máquinas virtuales |
| Container image construye y se niega a funcionar sin un terminal | Docker, podman |
| Nuevas dependencias en las solicitudes de exención | examen de la dependencia |

Los trabajos de vulnerabilidad y análisis también funcionan semanalmente, y OpenSSF
Scorecard revisa las prácticas del repositorio.

Los flujos de trabajo comienzan sin permisos y otorgan cada trabajo sólo lo que es
necesidades; los chequeos no mantienen credenciales.

## Containers and Kubernetes

La imagen contiene sólo los dos binarios, un paquete CA y la zona horaria
base de datos y funciona como UID 65532.
Los comandos y manifiestos documentados utilizan un sistema de archivos raíz sólo lectura,
soltar todas las capacidades, prohibir la escalada de privilegios, montar el trabajo
carpeta de sólo lectura para forma numb-form, montar sin cuenta de servicio token y establecer un
plazo.

## Este repositorio es público

Nunca comprometer documentos de clientes, hojas de cálculo, registros o carpetas de ejecución.
`.gitignore` los excluye y CI los rechaza.
