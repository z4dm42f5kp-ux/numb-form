# Running numb-form in containers

The image holds only the two static binaries, a CA bundle and the time
zone database.
No base image is used.
It runs as an unprivileged user (UID 65532).

Two volumes are needed:

* `/data`: the folder *above* the working folder.
  numb-form only reads it, so mount it read-only.
  numb-commit writes the accepted folders into it.
* `/downloads`: where the review copy, logs and manifest go.

Every document must be confirmed, so containers always run interactively
(`-it`).
Run numb-commit the same way you ran numb-form, with the same mounts,
because the manifest records container paths.

Set `TZ` so the evaluation date uses local time.

## Getting the image

Pull the signed release image:

```
docker pull ghcr.io/z4dm42f5kp-ux/numb-form:latest && cosign verify ghcr.io/z4dm42f5kp-ux/numb-form:latest --certificate-identity-regexp '^https://github.com/z4dm42f5kp-ux/numb-form/\.github/workflows/release\.yml@refs/tags/v' --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

Or build it locally from a checkout (needs Go and docker or podman):

```
git clone --depth 1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && container/build.sh numb-form:local
```

## Docker

On Linux, macOS and the BSDs, started in the folder above the working
folder:

```
mkdir -p "$HOME/Downloads" && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "$PWD:/data:ro" -v "$HOME/Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -v "$PWD:/data" -v "$HOME/Downloads:/downloads" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

On Windows with Docker Desktop, in Command Prompt:

```
docker run --rm -it --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "%cd%:/data:ro" -v "%USERPROFILE%\Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest && docker run --rm -it --read-only --cap-drop ALL --security-opt no-new-privileges -v "%cd%:/data" -v "%USERPROFILE%\Downloads:/downloads" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Podman

Rootless; `--userns=keep-id` keeps your own file ownership and `:Z`
relabels the volumes on SELinux systems such as Fedora:

```
mkdir -p "$HOME/Downloads" && podman run --rm -it --userns=keep-id --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "$PWD:/data:ro,Z" -v "$HOME/Downloads:/downloads:Z" ghcr.io/z4dm42f5kp-ux/numb-form:latest && podman run --rm -it --userns=keep-id --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -v "$PWD:/data:Z" -v "$HOME/Downloads:/downloads:Z" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Kubernetes

[`k8s/`](../k8s) has example volume claims and two interactive pods.
Fill `numb-form-data` with the folder above the working folder, then:

```
kubectl apply -f k8s/pvc.yaml && kubectl apply -f k8s/numb-form-pod.yaml && kubectl wait --for=condition=Ready pod/numb-form --timeout=120s && kubectl attach -it numb-form && kubectl delete pod numb-form
```

Review the copy in the `numb-form-downloads` volume, then commit:

```
kubectl apply -f k8s/numb-commit-pod.yaml && kubectl wait --for=condition=Ready pod/numb-commit --timeout=120s && kubectl attach -it numb-commit && kubectl delete pod numb-commit
```

If the first prompt was printed before you attached, press Enter to see
it again.
The pods use a read-only root file system, drop all capabilities, mount
no service account token and stop after 30 minutes.
For the optional Companies House key, uncomment the `env` entry and
create the secret:

```
kubectl create secret generic numb-form --from-literal=companies-house='your-free-key'
```
