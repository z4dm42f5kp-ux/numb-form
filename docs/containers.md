# Running numb-form in containers

The image holds only the two static binaries, a CA bundle and the time
zone database.
No base image is used.
It runs as an unprivileged user (UID 65532).

Two volumes are needed:

* `/data`: the folder *above* the working folder.
  In the default direct mode numb-form writes the new folders into it.
  In review mode numb-form only reads it, so it can be mounted read-only,
  and numb-commit later writes the accepted folders into it.
* `/downloads`: logs and manifests (direct mode, via `log_dir`) or the
  review copy (review mode).

With `confirm = each` (the default) every document must be confirmed, so
the container runs interactively (`-it`).
A configuration file is passed with `-f`, for example from a third
volume.
In review mode, run numb-commit the same way you ran numb-form, with the
same mounts, because the manifest records container paths.

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

On Linux, macOS and the BSDs, started inside the working folder (or the
folder above it):

```
mkdir -p "$HOME/Downloads/numb-form-logs" && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -e XDG_STATE_HOME=/downloads/numb-form-logs -v "$PWD:/data" -v "$HOME/Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

On Windows with Docker Desktop, in Windows PowerShell replace `%cd%` with
`${PWD}` and `%USERPROFILE%` with `$env:USERPROFILE`.
In Command Prompt:

```
docker run --rm -it --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -e XDG_STATE_HOME=/downloads/numb-form-logs -v "%cd%:/data" -v "%USERPROFILE%\Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Podman

Rootless; `--userns=keep-id` keeps your own file ownership and `:Z`
relabels the volumes on SELinux systems such as Fedora:

```
mkdir -p "$HOME/Downloads/numb-form-logs" && podman run --rm -it --userns=keep-id --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -e XDG_STATE_HOME=/downloads/numb-form-logs -v "$PWD:/data:Z" -v "$HOME/Downloads:/downloads:Z" ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Review mode in a container

Mount the working folder read-only for the run, read-write for the
commit, and pass `mode = review` in a configuration file:

```
mkdir -p "$HOME/Downloads" && printf 'mode = review\n' > "$HOME/Downloads/review.conf" && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "$PWD:/data:ro" -v "$HOME/Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest -f /downloads/review.conf && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -v "$PWD:/data" -v "$HOME/Downloads:/downloads" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

Podman takes the same options plus `--userns=keep-id` and `:Z`.

## Kubernetes

[`k8s/`](../k8s) has example volume claims and two interactive pods.
Fill `numb-form-data` with the working folder (or the folder above it),
then:

```
kubectl apply -f k8s/pvc.yaml && kubectl apply -f k8s/numb-form-pod.yaml && kubectl wait --for=condition=Ready pod/numb-form --timeout=120s && kubectl attach -it numb-form && kubectl delete pod numb-form
```

In review mode, review the copy in the `numb-form-downloads` volume,
then commit:

```
kubectl apply -f k8s/numb-commit-pod.yaml && kubectl wait --for=condition=Ready pod/numb-commit --timeout=120s && kubectl attach -it numb-commit && kubectl delete pod numb-commit
```

If the first prompt was printed before you attached, press Enter to see
it again.
The pods use a read-only root file system, drop all capabilities, mount
no service account token and stop after 30 minutes.
