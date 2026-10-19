# Guía paso a paso

> Traducción automática (Argos Translate) del [original en inglés](../guide.md). La versión en inglés es la que vale.

Esta guía es para todos los que llenan los formularios de evaluación de riesgos. No necesitas ningún conocimiento técnico. Sólo copia, pega y pulsa Enter.

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
3. rellena la parte A del formulario (nombre, forma legal, dirección, representantes, fecha y nombre),
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

Para actualizar numb-form más tarde, ejecute la misma línea de nuevo.

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

Cada empresa comienza con una línea como `[3/11] Foo GmbH  (Sheet1!C5)`: la tercera de once empresas, y la célula de la que fue tomado.

| numb-form pregunta | Lo que significa | Lo que normalmente haces |
| --- | --- | --- |
| `Full name of the person processing the form:` | Sólo preguntaba si saltaste el paso 2 | Escriba su nombre; Ingrese en la siguiente pregunta lo utiliza para cada forma en esta carrera |
| `Is the existing folder "Foo Gmbh" this company's folder? (...) [y/N]` | Una carpeta con un nombre muy similar ya existe | `y` si es la misma empresa (nada se crea), de lo contrario Enter |
| `Official company website to read its legal notice (...)` | El propio sitio web de la empresa; numb-form lee su aviso legal (Impressum) | Escriba la dirección, por ejemplo `foo.de`; o Enter, que toma el sitio web sugerido si se muestra y se salta de otra manera; `-` skips |
| `Legal form (Enter = keep, '-' = blank) [GmbH]` | Tomado desde el final del nombre de la empresa | Entra |
| `Headquarters / Address candidates` y una lista numerada, entonces `Choose a number, 'm' to type it, 'b' to leave blank [1]` | Discursos encontrados, los registros oficiales más fiables primero; cada uno muestra su fuente | Entrar toma el número 1; escriba otro número; `m` para escribir la dirección usted mismo; `b` para dejarlo en blanco |
| `Address (Enter = keep, or correct it) [...]` | La dirección que elegiste, para corregir un typo | Introduzca o escriba la dirección corregida |
| `Add the caution text to the address (not fully verified)? [y/N]` | Añade la frase "Caution: Este campo puede diferir de fuentes internas..." La misma pregunta puede plantearse a los representantes | Press Enter: la sugerencia ya depende de lo fiable que sea la fuente |
| `Authorized representative candidates` y una lista numerada, entonces `Numbers to include (e.g. 1,3), 'a' all, 'm' type them, 'n' none found [1,2]` | Personas encontradas, cada marcada `[verified]` (registro oficial) o `[unverified]` | Entrar toma los números verificados; números de tipo como `1,3`; `a` para todos; `m` para escribirlos usted mismo; `n` escribe "No se encontraron representantes en línea ..." |
| `Is this correct? [a]ccept, [e]dit a field, [r]eject:` | El formulario ha sido escrito; numb-form muestra la parte A como el formulario ahora lee | Tipo `a` si es correcto; `e` para cambiar una línea; `r` para tirarla. No hay defecto aquí: debe escribir una carta |

Al final numb-form muestra un resumen, por ejemplo `Done: 3 accepted, 8 already had a folder, 0 rejected, 0 need review, 0 skipped, 0 errors.` y dónde está el registro.

You can stop at any time with **Ctrl+C**. Forms you already accepted stay; the one you were working on is removed.

## 5. Lo que cambió en la unidad

* Una nueva carpeta por compañía aceptada, con una forma llena en ella.
* Nada más. Ningún archivo o carpeta existente se cambia o se elimina, y la lista de Excel nunca se cambia.
* Una forma rechazada no deja nada atrás.

Para deshacer una forma aceptada, eliminar su carpeta de empresa.

Un registro de todo lo que numb-form hizo (el registro) se mantiene en su propia carpeta de usuario, no en GWG: `%LOCALAPPDATA%\numb-form\runs`. La última línea de cada carrera muestra el archivo exacto.

## 6. Revise su plantilla de formulario (una vez, recomendado)

Antes de la primera ejecución real, deja que numb-form pruebe la plantilla. En una ventana PowerShell abierta en GWG (ver paso 3), escriba:

```
numb-form check-template "Appendix A - Risk Assesment.docx" "$env:USERPROFILE\Desktop\sample.docx"
```

Lee la plantilla sin cambiarla y guarda un ejemplo lleno en el escritorio. Abra `sample.docx` y compruebe que los valores están en los lugares correctos. Suprimirlo después.

## ¿Prefieres revisar todo antes de aterrizar en GWG?

Cambiar al modo de revisión una vez:

```
numb-form config set mode review
```

A partir de entonces, numb-form copias GWG en su carpeta **Downloads** y solo funciona en la copia. Mira a través de las nuevas carpetas allí. Cuando todo está bien, escriba `numb-commit`: copia las nuevas carpetas en la carpeta GWG real después de preguntarte una vez más. Vuelve con `numb-form config set mode direct`.

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

**`no sheet has a "Company Name" header cell`**
: numb-form busca una célula que dice `Company Name` y toma los nombres de abajo. Si su lista usa otro título, por ejemplo `Firmenname`, guárdalo una vez con `numb-form config set header "Firmenname"`.

**`Time limit of 25m0s reached`**
: Una carrera se detiene después de 25 minutos por seguridad. Comience de nuevo; las empresas ya hechas son saltadas.

**Version 0.1.0 does not start inside GWG**
: Actualizar ejecutando la línea de instalación desde el paso 1 otra vez, o utilizar `numb-form -root "H:\GWG"`.

**Running scripts o programas está bloqueado por su empresa**
: Pídale a su departamento de TI que permita `%LOCALAPPDATA%\Programs\numb-form\numb-form.exe`.

Para todo lo demás, consulte el manual [reference](usage.md) y el [settings](configuration.md).
