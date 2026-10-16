# Funcionamiento de forma de tuerca en contenedores

> Traducción automática (Argos Translate) del [original en inglés](../containers.md). La versión en inglés es la que vale.

La imagen contiene sólo los dos binarios estáticos, un paquete de CA y el tiempo
base de datos de zona.
No se utiliza imagen de base.
Funciona como un usuario no privilegiado (UID 65532).

Se necesitan dos volúmenes:

* `/data`: la carpeta *sobre* la carpeta de trabajo.
  numb-form sólo lo lee, así que montarlo solo.
  numb-commit escribe las carpetas aceptadas en ella.
* `/downloads`: donde va la copia de revisión, registros y manifiesto.

Cada documento debe ser confirmado, así que los contenedores siempre funcionan de forma interactiva
(`-it`).
Ejecutar numb-commit de la misma manera que corriste numb-form, con los mismos montes,
porque el manifiesto registra las rutas de contenedores.

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

En Linux, macOS y BSD, comenzó en la carpeta por encima del trabajo
carpeta:

```
mkdir -p "$HOME/Downloads" && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "$PWD:/data:ro" -v "$HOME/Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -v "$PWD:/data" -v "$HOME/Downloads:/downloads" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

En Windows con Docker Desktop, en Command Prompt:

```
docker run --rm -it --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "%cd%:/data:ro" -v "%USERPROFILE%\Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest && docker run --rm -it --read-only --cap-drop ALL --security-opt no-new-privileges -v "%cd%:/data" -v "%USERPROFILE%\Downloads:/downloads" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Podman

Rootless; `--userns=keep-id` mantiene su propia propiedad de archivo y `:Z`
reetiqueta los volúmenes en sistemas SELinux como Fedora:

```
mkdir -p "$HOME/Downloads" && podman run --rm -it --userns=keep-id --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "$PWD:/data:ro,Z" -v "$HOME/Downloads:/downloads:Z" ghcr.io/z4dm42f5kp-ux/numb-form:latest && podman run --rm -it --userns=keep-id --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -v "$PWD:/data:Z" -v "$HOME/Downloads:/downloads:Z" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Kubernetes

[`k8s/`](../../k8s) has example volume claims and two interactive pods.
Llene `numb-form-data` con la carpeta por encima de la carpeta de trabajo, luego:

```
kubectl apply -f k8s/pvc.yaml && kubectl apply -f k8s/numb-form-pod.yaml && kubectl wait --for=condition=Ready pod/numb-form --timeout=120s && kubectl attach -it numb-form && kubectl delete pod numb-form
```

Revisa la copia en el volumen `numb-form-downloads`, y luego compromete:

```
kubectl apply -f k8s/numb-commit-pod.yaml && kubectl wait --for=condition=Ready pod/numb-commit --timeout=120s && kubectl attach -it numb-commit && kubectl delete pod numb-commit
```

Si el primer aviso fue impreso antes de adjuntar, presione Enter para ver
otra vez.
Las cápsulas utilizan un sistema de archivos raíz sólo lectura, desplega todas las capacidades, montaje
sin cuenta de servicio ficha y para después de 30 minutos.
Para la llave opcional de la casa de las empresas, descomiendo la entrada `env` y
crear el secreto:

```
kubectl create secret generic numb-form --from-literal=companies-house='your-free-key'
```
