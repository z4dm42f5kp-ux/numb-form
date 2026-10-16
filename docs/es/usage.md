# NUMB-FORM(1)

> Traducción automática (Argos Translate) del [original en inglés](../usage.md). La versión en inglés es la que vale.

## NAME

**numb-form**, **numb-commit** — fill risk assessment forms and commit them

## SYNOPSIS

```
numb-form [-offline] [-root dir] [-downloads dir] [-timeout duration]
numb-form check-template template.docx [sample.docx]
numb-form keys
numb-commit [-run dir] [-downloads dir]
```

## DESCRIPTION

**numb-form** is started in the folder *above* the working folder. The working folder is the single folder found there. If there are several, the one containing an overview folder is used; otherwise the user is asked.

La carpeta de trabajo debe contener:

* una carpeta de visión general cuyo nombre se asemeja a `3. Übersichten`, conteniendo exactamente un archivo `.xlsx`, la hoja de cálculo;
* la plantilla `Appendix A - Risk Assesment.docx` en su nivel superior.

A run proceeds as follows:

1. La plantilla está marcada. La carrera se detiene si no se puede encontrar la sección A o una de sus etiquetas.
2. El usuario escribe el nombre de la persona que procesa el formulario, una vez.
3. Toda la carpeta de trabajo se copia a `Downloads/numb-form-runs/<date>_<time>_<id>/`. Todo después de este paso sucede en la copia.
4. Los nombres de las empresas se leen desde la columna de hoja de cálculo encabezada `Company Name`, de arriba a abajo. Todas las demás columnas son ignoradas, incluyendo `Report created`. La hoja de cálculo nunca está escrita. Su suma de comprobación SHA-256 se revisa antes y después de la carrera.
5. Para cada empresa, las carpetas en la carpeta de trabajo y todos sus subcarpetas son buscados. Los nombres son comparados ignorando caso, acentos y umlauts (ü = u = ue), puntuación y espacios. Un partido significa que la forma ya existe, y la compañía está saltada. Un partido cercano se muestra al usuario para confirmar. Las carpetas con una forma legal diferente nunca coinciden: `Foo GmbH` y `Foo S.A.` son diferentes empresas.
6. Para una nueva empresa, una carpeta llamada después de la creación de la empresa. La plantilla se copia en ella como `Appendix A - Risk Assesment - <Company Name>.docx`, y la sección A se llena.
7. El documento se lee de nuevo del disco y se muestra. El usuario lo acepta, edita un campo o lo rechaza. Una carpeta rechazada se elimina de la copia.

Sólo se llenan las líneas de la sección A, *Information About the Contracting Party*. La sección B y todo lo que está fuera de la sección A quedan byte para byte sin cambios. Los valores se anexan a la propia edición de texto de la etiqueta, por lo que toman el formato de la etiqueta; no se añade ningún nuevo formato.

| Campo | Valor |
| --- | --- |
| Name / Company Name of the Contracting Party | la hoja de cálculo literal |
| Legal form | el sufijo del nombre, como está escrito (`GmbH`, `S.A.`, `C.A. de C.V.`) |
| Headquarters / Address | elegido por el usuario de los candidatos encontrados |
| Commercial Registry Number | siempre en blanco |
| Authorized Representatives | `Full Name (Role 1, Role 2), Full Name (Role)`, roles en inglés |
| Date of the evaluation | hoy, `DD/MM/YYYY` |
| Person Processing the Form | el nombre escrito al principio |

Los candidatos son recogidos, en orden de autoridad:

1. UK Companies House (registro gubernamental; sólo con una llave gratuita);
2. el registro GLEIF LEI (regulación de registro; no llave);
3. el propio aviso legal de la empresa (Impressum, Aviso legal, Aviso legal);
4. Wikidata (editado comunitario; utilizado principalmente para encontrar el sitio web).

Las personas y direcciones nunca son tomadas de una entidad madre, subsidiaria o similarmente nombrada. Las personas encontradas sólo en el sitio web o sólo en Wikidata están marcadas *unverified*, y por defecto este texto exacto se adjunta:

> Precaución: Este campo puede diferir de fuentes internas o ser inexacto debido a la insuficiente información en línea

Cuando no se puede encontrar representante, el campo dice:

> No se encontraron representantes en línea, consulte otras fuentes internas para obtener información más precisa

**numb-commit** copies the accepted folders of a run back into the real working folder. It only adds new folders. It never overwrites or deletes anything. It refuses folders that appeared in the real tree in the meantime. Each folder is copied to a temporary name, verified with SHA-256 and then renamed. Documents edited during review are listed before the user confirms. Afterwards it offers to delete the review copy; the logs are kept.

## OPTIONS

`-offline`
: Contacte sin fuente en línea. Cada valor es escrito por el usuario.

`-root dir`
: Use *dir* como la carpeta de trabajo en lugar de detectarla.

`-downloads dir`
: Ponga la copia de la reseña en *dir* en lugar de la carpeta Downloads.

`-timeout duration`
: Stop safely after *duration* (for example `15m`). The maximum and default is `25m`.

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

## FILES

`Downloads/numb-form-runs/<run>/<working folder>/`
: La copia de revisión.

`Downloads/numb-form-runs/<run>/numb-form.log`
: Lograr todo lo que hizo la carrera.

`Downloads/numb-form-runs/<run>/manifest.json`
: Grabación legible a máquina utilizada por numb-commit.

`Downloads/numb-form-runs/<run>/numb-commit.log`
: Registro del compromiso.

No hay caché ni archivos temporales detrás. Las respuestas de la red sólo se llevan a cabo en memoria, y los documentos se escriben a través de un archivo temporal que se renombra o elimina. Una carrera escribe a la mayoría de 100 MB de documentos además de la copia.

## EXIT STATUS

0 sobre el éxito, 1 sobre el error, 2 sobre un error de uso, y 3 cuando se detiene por el límite de tiempo, por Ctrl+C o por el usuario.

## EXAMPLES

Compruebe la plantilla real y escriba una muestra para ver, sin tocar la plantilla:

```
numb-form check-template "Appendix A - Risk Assesment.docx" "%USERPROFILE%\Downloads\sample.docx"
```

Correr, revisar, luego comprometer (numb-commit espera confirmación):

```
cd /d "S:\Kunden" && numb-form && numb-commit
```

Corre sin acceso a la red:

```
cd /d "S:\Kunden" && numb-form -offline && numb-commit
```

## SEE ALSO

[install.md](install.md), [containers.md](containers.md), [../SECURITY.md](SECURITY.md), [../AGENTS.md](../../AGENTS.md)
