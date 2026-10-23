# Guía paso a paso

> Traducción automática (Argos Translate) del [original en inglés](../guide.md). La versión en inglés es la que vale.

Esta guía es para todos los que llenan los formularios de evaluación de riesgos en **Windows** (10, 11). No necesitas ningún conocimiento técnico. Sólo copia, pega y pulsa Enter. En **macOS, Linux o BSD**, instalar con [install.md](install.md); las preguntas y respuestas son las mismas, y los comandos unidos con `;` aquí se unen con `&&` allí.

## Qué hace numb-form

Your team keeps a working folder, for example **GWG**. It looks like this:

```
GWG
├── 3. Übersichten
│   └── the Excel list of companies
├── Appendix A - Risk Assesment.docx      (the empty form)
├── Foo GmbH                              (one folder per company
└── Bar S.A.                               that already has a form)
```

numb-form pasa por la lista de Excel de arriba a abajo. Para cada empresa que aún no tiene una carpeta, es:

1. crea una carpeta con el nombre de la empresa,
2. pone una copia del formulario en él, llamado `Appendix A - Risk Assesment - <company>.docx`,
3. llena la parte A del formulario: el nombre, la forma legal, la sentencia del registro, la fecha y su nombre por sí mismo, y la dirección y los representantes que escriba,
4. te muestra lo que escribió y espera tu OK.

Las empresas que ya tienen una carpeta se saltan. La lista de Excel nunca se cambia. La parte B de la forma nunca se toca.

## 1. Instala (una vez)

1. Open the **Start** menu, type `PowerShell` and open **Windows PowerShell**. A blue window appears.
2. Copiar toda esta línea, pegarlo en la ventana azul (pasos de clic derecho) y presione **Enter**:

   ```
   & { $ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Set-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; 'OK, open a new PowerShell window' }
   ```

3. Espera hasta que diga `OK, open a new PowerShell window`. Cierra la ventana.

No necesita derechos de administrador. Si dice `CHECKSUM MISMATCH`, la descarga fue dañada o cambiada: no continuar, y decir a quien se ocupa de la herramienta.

Para actualizar numb-form más tarde, consulte [Reinstalar o actualizar](#reinstall-or-update).

## 2. Guarda tu nombre (una vez)

Tu nombre entra en el campo *Person Processing the Form*. Abra Windows PowerShell de nuevo y escriba, con su propio nombre:

```
numb-form config set person "Erika Mustermann"
```

Responde a `Saved person = Erika Mustermann`. Si salta este paso, numb-form pide su nombre cada vez.

## 3. Rellene los formularios (cada vez)

1. Open the **GWG** folder in File Explorer, the folder that contains `3. Übersichten`.
2. Haga clic una vez en el **address bar** en la parte superior de la ventana (donde se muestra la ruta de la carpeta), escriba `powershell` y pulse **Enter**. Una ventana azul abre que ya está "en" la carpeta GWG.
3. Tipo `numb-form` y pulse **Enter**.

numb-form muestra primero lo que encontró:

```
Working folder: H:\GWG
Spreadsheet:    3. Übersichten\Kundenliste.xlsx (read only)
Template:       Appendix A - Risk Assesment.docx
```

Si eso es correcto, pasa por las empresas una a una.

## 4. Las preguntas y qué responder

Muchas preguntas muestran una sugerencia entre corchetes, como `[GmbH]` o `[y/N]`. **Presing Enter acepta la sugerencia** (para `[y/N]` que es *no*, para `[Y/n]` es *yes*). Para dar una respuesta diferente, escriba y presione Enter.

Cada empresa comienza con una línea como `[3/11] Foo GmbH  (Customer ID D1001, row 5)`: la tercera de once empresas, su número de cliente y su fila en la lista de Excel.

La forma legal se toma del nombre de la empresa sin preguntar. numb-form le pide la dirección y los representantes; no busca nada en línea. *Commercial Registry Number* siempre obtiene la frase "No se encontró ningún número de registro comercial en línea consulta fuentes internas para más información"; para poner el número en su lugar, seleccione `e` (edita) y campo 4 al final.

| numb-form pregunta | Lo que significa | Lo que normalmente haces |
| --- | --- | --- |
| `Full name of the person processing the form:` | Sólo preguntaba si saltaste el paso 2 | Escriba su nombre; Ingrese en la siguiente pregunta lo utiliza para cada forma en esta carrera |
| `Is the existing folder "Foo Gmbh" this company's folder? (...) [y/N]` | Una carpeta con un nombre muy similar ya existe | `y` si es la misma empresa (nada se crea), de lo contrario Enter |
| `Legal form (the name has none; Enter = blank):` | Se le preguntó solamente cuando el nombre de la empresa no tiene forma legal como `GmbH` o `S.A.` | Escribelo o entra para dejarlo en blanco |
| `Headquarters / Address (one line, Enter = blank):` | Dirección de la empresa, de sus propias fuentes | Escribe o pega en una línea y pulsa Enter |
| `Representatives (Enter = "No representatives were found…", '-' = blank):` | La gente que puede firmar para la empresa, escribió `Full Name (Role 1, Role 2), Full Name (Role)` con los roles en inglés | Escribe o pega; Entra solo escribe "No se encontraron representantes en línea ..."; `-` deja el campo en blanco |
| `Is this correct? [a]ccept, [e]dit a field, [r]eject:` | El formulario ha sido escrito; numb-form muestra la parte A como el formulario ahora lee | Tipo `a` si es correcto; `e` para cambiar una línea; `r` para tirarlo lejos. No hay defecto aquí: debe escribir una carta |

Al final numb-form muestra un resumen, por ejemplo `Done: 3 accepted, 8 already had a folder, 0 rejected, 0 skipped, 0 errors.` y dónde está el registro.

You can stop at any time with **Ctrl+C**. Forms you already accepted stay; the one you were working on is removed.

## 5. Lo que cambió en la unidad

* Una nueva carpeta por compañía aceptada, con una forma llena en ella.
* Nada más. Ningún archivo o carpeta existente se cambia o se elimina, y la lista de Excel nunca se cambia por numb-form (ver paso 7 para numb-date).
* Una forma rechazada no deja nada atrás.

Para deshacer una forma aceptada, eliminar su carpeta de empresa.

Un registro de todo lo que numb-form hizo (el registro) se mantiene en su propia carpeta de usuario, no en GWG: `%LOCALAPPDATA%\numb-form\runs`. La última línea de cada carrera muestra el archivo exacto.

## 6. Revise su plantilla de formulario (una vez, recomendado)

Antes de la primera ejecución real, deja que numb-form pruebe la plantilla. En una ventana PowerShell abierta en GWG (ver paso 3), escriba:

```
numb-form check-template "Appendix A - Risk Assesment.docx" "$env:USERPROFILE\Desktop\sample.docx"
```

Lee la plantilla sin cambiarla y guarda un ejemplo lleno en el escritorio. Abra `sample.docx` y compruebe que los valores están en los lugares correctos. Suprimirlo después.

## 7. Llenar las fechas del informe (después de numb-form)

La lista de Excel tiene una columna como *Report created AB*. `numb-date` te lo llena. En la misma ventana PowerShell en GWG, escriba:

```
numb-date
```

Pasa por la lista. Las empresas cuya celda *Report created* ya contiene una fecha se saltan. Para los demás abre el formulario de la empresa (sólo para leerlo), toma el *Date of the evaluation* y las listas, por ejemplo:

```
Dates to write into Kundenliste.xlsx (column "Report created AB"):
  F13    16-10-2026  Foo Ltda
```

| numb-date pide | ¿Qué haces? |
| --- | --- |
| `Is the folder "Fooo Inc" this company's folder? (...) [y/N]` | `y` si es la misma empresa, de lo contrario Enter |
| `Date (DD-MM-YYYY, Enter = skip):` | El formulario no tiene fecha legible: escribirlo, o Entrar para saltar la empresa |
| `Write these 1 date(s)? Nothing else in the spreadsheet changes. [y/N]` | `y` para escribirlos; Entrar no escribe nada |
| `Is this correct? [a]ccept, [u]ndo:` | numb-date muestra las células como el archivo ahora lee; `a` para mantenerlas, `u` para poner el archivo exactamente como era |

**Cerrar la lista de Excel primero. ** numb-date se detiene si el archivo está abierto. Sólo las células *Report created* cambian: ninguna otra célula, ningún formato.

## ¿Prefieres revisar todo antes de aterrizar en GWG?

Cambiar al modo de revisión una vez:

```
numb-form config set mode review
```

A partir de entonces, numb-form copias GWG en su carpeta **Downloads** y solo funciona en la copia. Mira a través de las nuevas carpetas allí. Cuando todo está bien, escriba `numb-commit`: copia las nuevas carpetas en la carpeta GWG real después de preguntarte una vez más. Vuelve con `numb-form config set mode direct`.

## Reinstalar o actualizar

La manera fácil: abrir Windows PowerShell y escribir uno de estos:

| Tipo | Lo que consigues |
| --- | --- |
| `numb-form update stable` | la última versión estable (utiliza esto para el trabajo real) |
| `numb-form update rolling` | la construcción de rodaje actual: los cambios más recientes, reconstruidos automáticamente después de cada actualización en GitHub |
| `numb-form update` | la versión más reciente del tipo que ya tiene |
| `numb-form update v0.1.1` | exactamente ese lanzamiento (cualquier número de la página de lanzamientos) |
| `numb-form rollback` | volver a la versión que tenía antes de la última actualización; escribir de nuevo para volver |

Se descarga la versión, se comprueba, reemplaza el programa e imprime la nueva versión, por ejemplo `Updated to numb-form v0.1.2 (stable)`. Los ajustes guardados (su nombre, la partida, el modo) se mantienen.

Versiones 0.1.1 y mayores no tienen `numb-form update`. Para aquellos, o si `numb-form update` no puede conectarse, use una de las dos líneas siguientes. No escriba ellos: Abra esta página en su navegador, haga clic en el icono de copia en el borde derecho de la caja gris, y haga clic con el botón derecho en la ventana PowerShell para pegar. Cierre cada ventana numb-form primero.

**Stable version** (el nuevo lanzamiento publicado; utilizar esto para el trabajo real):

```
& { $ErrorActionPreference = 'Stop'; if (Get-Process numb-form, numb-commit, numb-date -ErrorAction SilentlyContinue) { throw 'Close every numb-form window first, then run this again' }; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Push-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Remove-Item -Force -ErrorAction SilentlyContinue numb-form.exe, numb-commit.exe, numb-date.exe; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS; Pop-Location }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, reinstalled. Your settings are kept. Open a new PowerShell window' }
```

**Rolling version** (los cambios más recientes, reconstruidos automáticamente después de cada actualización en GitHub; use esto para probar una solución de inmediato):

```
& { $ErrorActionPreference = 'Stop'; if (Get-Process numb-form, numb-commit, numb-date -ErrorAction SilentlyContinue) { throw 'Close every numb-form window first, then run this again' }; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $u = 'https://github.com/z4dm42f5kp-ux/numb-form/releases/download/edge'; $z = 'numb-form_windows_amd64.zip'; $d = Join-Path $env:LOCALAPPDATA 'Programs\numb-form'; New-Item -ItemType Directory -Force $d | Out-Null; Push-Location $d; try { Invoke-WebRequest -UseBasicParsing "$u/$z" -OutFile $z; Invoke-WebRequest -UseBasicParsing "$u/SHA256SUMS" -OutFile SHA256SUMS; $h = (Get-FileHash $z -Algorithm SHA256).Hash.ToLower(); if (-not (Select-String -Path SHA256SUMS -SimpleMatch "$h  $z" -Quiet)) { throw 'CHECKSUM MISMATCH' }; Remove-Item -Force -ErrorAction SilentlyContinue numb-form.exe, numb-commit.exe, numb-date.exe; Expand-Archive -Force $z -DestinationPath $d } finally { Remove-Item -ErrorAction SilentlyContinue $z, SHA256SUMS; Pop-Location }; $p = [Environment]::GetEnvironmentVariable('Path', 'User'); if (-not $p) { $p = '' }; if (($p -split ';') -notcontains $d) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $d).TrimStart(';'), 'User') }; & (Join-Path $d 'numb-form.exe') -version; 'OK, reinstalled. Your settings are kept. Open a new PowerShell window' }
```

Ambos imprimen la versión que instalaron, por ejemplo `v0.1.1` (estable) o `rolling-3f976cc` (rolling; construcciones de rodamiento mayores dicen `edge-…`), y terminan con `OK, reinstalled`. Abra una nueva ventana PowerShell después.

Si PowerShell responde `Unexpected token` o `Missing ...` en rojo, la línea fue cambiada mientras se copia (normalmente se perdió un personaje): copiarlo de nuevo con el icono de copia.

Si la línea de rodaje dice `CHECKSUM MISMATCH`, se estaba subiendo una nueva construcción en ese momento: esperar cinco minutos y ejecutarla una vez más. Si sucede de nuevo, o con la línea estable, deténgase y dígale a quien se ocupa de la herramienta.

## Lea este manual en la ventana

Escriba `numb-form man` para leer todo en la ventana azul, una pantalla a la vez (Enter: siguiente pantalla; q y Enter: dejar). `numb-form man guide` sólo muestra esta guía; añadir `de` o `es` para alemán o español: `numb-form man guide de`. En Command Prompt (la ventana negra) `man numb-form` funciona también, después de ejecutar `numb-form man -install` una vez.

## Cuando algo sale mal

**`numb-form` no es reconocido**
: Cierra la ventana y abre una nueva. Si todavía sucede, instale de nuevo (paso 1).

**`... is not a working folder`** or **`several folders could be the working folder`**
: La ventana azul no se abrió en GWG. Repita el paso 3 desde dentro de la carpeta GWG. También puede nombrar la carpeta directamente, por ejemplo `numb-form -root "H:\GWG"`.

**`template "Appendix A - Risk Assesment" ... not found`**
: La forma vacía debe estar directamente en GWG, no en una subcarpeta, y su nombre debe comenzar con `Appendix A - Risk Assesment`.

**`no folder whose name contains "Übersicht" found`**
: GWG debe contener la carpeta `3. Übersichten` (cualquier nombre que contenga `Übersicht` funciona).

**`2 spreadsheets in ...; expected exactly one`**
: `3. Übersichten` debe contener exactamente un archivo Excel. Muevan a los demás.

**`no sheet has a "Customer Name" header cell`**
: numb-form busca el encabezado `Customer Name` y toma los nombres de abajo. El mensaje enumera los encabezados que encontró. Si su lista usa otro título, por ejemplo `Firmenname`, guárdalo una vez con `numb-form config set header "Firmenname"`. Versiones 0.1.1 y búsqueda más antigua para `Company Name` en su lugar; allí, ejecutar `numb-form config set header "Customer Name"` una vez.

**`Time limit of 30m0s reached`**
: Se estableció un límite de tiempo con `timeout` en la configuración o con `-timeout`; no hay ninguno por defecto. Comience de nuevo; las empresas ya hechas son saltadas.

**Version 0.1.0 does not start inside GWG**
: Actualizar (ver [Reinstalar o actualizar](#reinstall-or-update)), o utilizar `numb-form -root "H:\GWG"`.

**Running scripts o programas está bloqueado por su empresa**
: Pídale a su departamento de TI que permita `%LOCALAPPDATA%\Programs\numb-form\numb-form.exe`.

Para todo lo demás, consulte el manual [reference](usage.md) y el [settings](configuration.md).
