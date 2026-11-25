# Ejecutando numb-form en contenedores

> Traducción automática (Argos Translate) del [original en inglés](../containers.md). La versión en inglés es la que vale.

La imagen contiene sólo los tres programas estáticos (numb-form, numb-commit y numb-date, cada uno con la base de datos de zona horaria construida en) y un paquete CA. No se utiliza imagen de base. Funciona como un usuario no privilegiado (UID 65532).

Se necesitan dos volúmenes:

* `/data`: la carpeta *sobre* la carpeta de trabajo. En el modo directo predeterminado numb-form escribe las nuevas carpetas en él. En modo de revisión numb-form sólo lo lee, por lo que puede ser montado sólo lectura, y numb-commit más tarde escribe las carpetas aceptadas en él.
* `/downloads`: registros y manifiestos (modo directo, vía `log_dir`) o la copia de revisión (modo de revisión).

Con `confirm = each` (el predeterminado) cada documento debe ser confirmado, por lo que el contenedor se ejecuta de forma interactiva (`-it`). Un archivo de configuración se transmite con `-f`, por ejemplo desde un tercer volumen. En modo de revisión, ejecute numb-commit de la misma manera que corrió numb-form, con las mismas monturas, porque las rutas de contenedores de registros manifiestos.

Establecer `TZ` para que la fecha de evaluación use la hora local.

## Consiguiendo la imagen

Saque la imagen de liberación firmada:

```
docker pull ghcr.io/z4dm42f5kp-ux/numb-form:latest && cosign verify ghcr.io/z4dm42f5kp-ux/numb-form:latest --certificate-identity-regexp '^https://github.com/z4dm42f5kp-ux/numb-form/\.github/workflows/release\.yml@refs/tags/v' --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

O construirlo localmente desde un checkout (necesidades Ir y docker o podman):

```
git clone --depth 1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && container/build.sh numb-form:local
```

## Docker

**Unix-like** (macOS, Linux, BSD), en Terminal, comenzó dentro de la carpeta de trabajo (o la carpeta encima):

```
mkdir -p "$HOME/Downloads/numb-form-logs" && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -e XDG_STATE_HOME=/downloads/numb-form-logs -v "$PWD:/data" -v "$HOME/Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

**Windows** (10, 11) con Docker Desktop, en Command Prompt (en PowerShell, escriba `${PWD}` para `%cd%` y `$env:USERPROFILE` para `%USERPROFILE%`):

```
docker run --rm -it --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -e XDG_STATE_HOME=/downloads/numb-form-logs -v "%cd%:/data" -v "%USERPROFILE%\Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Podman

Rootless; `--userns=keep-id` mantiene su propia propiedad de archivos y `:Z` relabels los volúmenes en sistemas SELinux como Fedora. **Unix-like** (Linux, macOS, BSD), en Terminal:

```
mkdir -p "$HOME/Downloads/numb-form-logs" && podman run --rm -it --userns=keep-id --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -e XDG_STATE_HOME=/downloads/numb-form-logs -v "$PWD:/data:Z" -v "$HOME/Downloads:/downloads:Z" ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Modo de revisión en un contenedor

Montar la carpeta de trabajo solo para ejecutar, leer-escribir para el commit, y pasar `mode = review` en un archivo de configuración. **Unix-like** (Linux, macOS, BSD), en Terminal:

```
mkdir -p "$HOME/Downloads" && printf 'mode = review\n' > "$HOME/Downloads/review.conf" && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "$PWD:/data:ro" -v "$HOME/Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest -f /downloads/review.conf && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -v "$PWD:/data" -v "$HOME/Downloads:/downloads" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

Podman toma las mismas opciones más `--userns=keep-id` y `:Z`.

## numb-date en un contenedor

numb-date escribe la hoja de cálculo, por lo que la carpeta de trabajo se monta leer-escritura. **Unix-like** (Linux, macOS, BSD), en Terminal:

```
docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -e XDG_STATE_HOME=/downloads/numb-form-logs -v "$PWD:/data" -v "$HOME/Downloads:/downloads" --entrypoint /numb-date ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

Podman toma las mismas opciones más `--userns=keep-id` y `:Z`.

## Kubernetes

[`k8s/`](../../k8s) has example volume claims and two interactive pods. Fill `numb-form-data` with the working folder (or the folder above it), then:

```
kubectl apply -f k8s/pvc.yaml && kubectl apply -f k8s/numb-form-pod.yaml && kubectl wait --for=condition=Ready pod/numb-form --timeout=120s && kubectl attach -it numb-form && kubectl delete pod numb-form
```

En modo de revisión, revise la copia en el volumen `numb-form-downloads`, y luego comprometa:

```
kubectl apply -f k8s/numb-commit-pod.yaml && kubectl wait --for=condition=Ready pod/numb-commit --timeout=120s && kubectl attach -it numb-commit && kubectl delete pod numb-commit
```

Si el primer aviso fue impreso antes de adjuntar, presione Enter para verlo de nuevo. Las cápsulas utilizan un sistema de archivos raíz de sólo lectura, bajan todas las capacidades, montan sin cuenta de servicio token y paran después de 30 minutos.

## La cárcel de FreeBSD

Docker no funciona en los BSD. En FreeBSD, el script `numb-form-jail` ejecuta los programas instalados en una cárcel, el propio contenedor de FreeBSD, que es parte del sistema base; no se necesita software de contenedor. La cárcel está vacía y sólo lectura excepto para dos carpetas: la carpeta actual como `/data` y su carpeta Downloads como `/downloads`. No tiene red, ejecuta el programa como usted, no como raíz, y se elimina con cada montaje cuando el programa termina, también después de Ctrl+C. Crear una cárcel necesita raíz, por lo que el script se inicia con `doas` (o `sudo`). El script viene en el archivo de liberación de FreeBSD, comprobado como los programas. Instalar los programas primero ([install.md](install.md)), luego instalar el script una vez, propiedad de root (en el brazo64, escriba `arm64` para `amd64`).

**Unix-like** (FreeBSD), en Terminal:

```
cd "$(mktemp -d)" && fetch https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/numb-form_freebsd_amd64.tar.gz && fetch https://github.com/z4dm42f5kp-ux/numb-form/releases/latest/download/SHA256SUMS && sha256 -c "$(awk '$2=="numb-form_freebsd_amd64.tar.gz"{print $1}' SHA256SUMS)" numb-form_freebsd_amd64.tar.gz && tar -xzf numb-form_freebsd_amd64.tar.gz jail.sh && doas install -o root -g wheel -m 0555 jail.sh /usr/local/sbin/numb-form-jail && cd / && rm -rf "$OLDPWD" && echo OK
```

Luego, dentro de la carpeta de trabajo (o la carpeta encima de ella):

```
doas numb-form-jail numb-form
```

Escriba `numb-commit` o `numb-date` en lugar de `numb-form` para ejecutarlos; las opciones siguen el nombre del programa. Los programas se toman desde `~/.local/bin`; para otra carpeta, u otra carpeta Downloads, comience como `doas env NUMBFORM_BIN=folder NUMBFORM_DOWNLOADS=folder numb-form-jail numb-form`.

## OpenBSD y NetBSD

OpenBSD y NetBSD no tienen un sistema de contenedores que ejecute estos programas. Ninguno es necesario: los programas son archivos estáticos únicos que no utilizan ningún otro software, leen solamente la carpeta de trabajo, la hoja de cálculo y la plantilla, y contacten con la red sólo para `numb-form update`. Instalarlos como se describe en [install.md](install.md), establecer `online = no` en el archivo de configuración del sistema para apagar `numb-form update`, y ejecutarlos como un usuario ordinario, nunca como root.
