# NUMB-TYPO(1)

> Traducción automática (Argos Translate) del [original en inglés](../numb-typo.md). La versión en inglés es la que vale.

## NAME

**numb-typo** — correctas ciudades y países de exclusión o corte en los nombres de la empresa de la hoja de cálculo

## SYNOPSIS

```
numb-typo [-f conf] [-root dir] [-timeout duration]
```

## DESCRIPTION

Los nombres de las empresas en la lista de Excel a menudo terminan en una ciudad o país que es misspelt o cortado, por ejemplo `Foo S.A. de C.V., Ciudad de Me` o `Foo SAS, Bogta, Colomiba`. numb-typo los encuentra y, después de estar de acuerdo, escribe los nombres corregidos: `Foo S.A. de C.V., Ciudad de México`, `Foo SAS, Bogotá, Colombia`.

Ejecutarlo dentro de la carpeta de trabajo, como numb-form, antes de numb-form, para que las nuevas carpetas y formas obtengan los nombres corregidos. La hoja de cálculo debe ser `.xlsx` o `.xlsm`, y cerrado en Excel y LibreOffice.

## WHAT IS CORRECTED

Sólo las partes después de una coma (y después de ` - `) que parecen una ciudad o un país:

* **Cut off**: la última parte de una célula, cuando es el comienzo de un lugar conocido: `Ciudad de Me` se convierte en `Ciudad de México`, `Nairob` se convierte en `Nairobi`.
* **Typo**: una parte que está a una o dos letras de exactamente un lugar conocido: `Colomiba` se convierte en `Colombia`, `Medelin` se convierte en `Medellín`, `Duabi` se convierte en `Dubái`. Partes de menos de cinco letras nunca se corrige; una carta extra se corrige sólo de siete letras, dos errores sólo de nueve, y la primera letra debe coincidir.
* Cuando caben varios lugares (`Santa` podría ser Santa Cruz, Santa Marta, ...), numb-typo pregunta cuál, o usted lo deja.

Las capitales se guardan como escrito: `GUAYAQIL` se convierte en `GUAYAQUIL`.

Lo que nunca se cambia:

* el nombre de la compañía, antes de la primera coma;
* formularios legales (`S.A.`, `Ltda`, `LLC`), números y palabras que no son lugares;
* un lugar que ya está a la derecha; los acentos perdidos (`Bogota`) quedan como escrito;
* los nombres de las empresas que ya tienen una carpeta: renombrarlas los separaría de su carpeta y forma, por lo que sólo se enumeran en el registro.

La lista de lugares abarca los países en español, inglés, portugués, francés y alemán, y las principales ciudades de América Latina, África, el Golfo, Europa y América del Norte.

## WHAT HAPPENS

1. numb-typo pasa por los nombres de la empresa y muestra cada corrección:

   ```
   Foo S.A. de C.V., Ciudad de Me  (Customer ID D1001, row 315)
     → Foo S.A. de C.V., Ciudad de México
   ```

2. Los enumera a todos y pregunta `Write these 3 correction(s)? [a]ll, [o]ne by one, [n]one:`. Con `o` pide cada uno.
3. Los escribe, lee la hoja de cálculo de nuevo del disco, muestra las células y pregunta `Is this correct? [a]ccept, [u]ndo:`.

Sólo las células de la columna de nombre de empresa cambian, con los mismos cheques que numb-date: cada otra célula de cada hoja se compara antes de escribir, cada célula mantiene su estilo, SHA-256 guarda cada paso, y *undo* restaura el original exacto, que se mantiene en la carpeta de registro hasta que usted acepta.

## OPTIONS

`-root dir`
: Use *dir* como la carpeta de trabajo en lugar de detectarla.

`-timeout duration`
: Parar con seguridad después de *duración*, por ejemplo `30m`.

`-f file`
: Lea *file* después de los archivos de configuración del sistema y del usuario.

`-allow-non-tty`
: Lea las respuestas de una tubería. Sólo para pruebas automatizadas.

`-version`
: Imprima la versión.

## FILES

`<log folder>/<date>_<time>_<id>_typo/numb-typo.log`
: Registro de la carrera: cada corrección, cada nombre dejó solo y por qué, y la SHA-256 de la hoja de cálculo antes y después.

## EXIT STATUS

0 en el éxito o cuando no hay nada que corregir, 1 en un error, 2 en un error de uso, 3 cuando se detuvo con Ctrl+C, cerrando la ventana o por el usuario.

## EXAMPLES

Corregir los nombres, luego llenar los formularios y las fechas.

**Windows** (10, 11), en PowerShell, dentro de la carpeta de trabajo:

```
numb-typo; if ($?) { numb-form }; if ($?) { numb-date }
```

**Unix-like** (macOS, Linux, BSD), en Terminal:

```
numb-typo && numb-form && numb-date
```

## SEE ALSO

[numb-form](usage.md), [numb-date](numb-date.md), [numb-commit](numb-commit.md)
