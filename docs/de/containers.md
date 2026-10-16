# numb-form in Containern

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../containers.md). Maßgeblich ist die englische Fassung.

Das Bild enthält nur die beiden statischen Binärdateien, ein CA-Bundle und die Zeitzonendatenbank. Es wird kein Basisbild verwendet. Es läuft als unprivilegierter Benutzer (UID 65532).

Es werden zwei Bände benötigt:

* `/data`: der Ordner *über* dem Arbeitsordner. numb-form liest es nur, also mounte es nur schreibgeschützt. numb-commit schreibt die akzeptierten Ordner hinein.
* `/downloads`: wo die Überprüfungskopie, Protokolle und Manifest gehen.

Jedes Dokument muss bestätigt werden, so dass Container immer interaktiv laufen (`-it`). Führen Sie numb-commit auf die gleiche Weise aus, wie Sie numb-form mit den gleichen Halterungen ausgeführt haben, da das Manifest Containerpfade aufzeichnet.

Stellen Sie `TZ` so ein, dass das Bewertungsdatum die lokale Zeit verwendet.

## Das Bild bekommen

Ziehen Sie das signierte Release-Bild:

```
docker pull ghcr.io/z4dm42f5kp-ux/numb-form:latest && cosign verify ghcr.io/z4dm42f5kp-ux/numb-form:latest --certificate-identity-regexp '^https://github.com/z4dm42f5kp-ux/numb-form/\.github/workflows/release\.yml@refs/tags/v' --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

Oder bauen Sie es lokal von einer Kasse (muss Go und Docker oder Podman):

```
git clone --depth 1 https://github.com/z4dm42f5kp-ux/numb-form.git && cd numb-form && container/build.sh numb-form:local
```

## Docker

Unter Linux, macOS und den BSDs, gestartet im Ordner über dem Arbeitsordner:

```
mkdir -p "$HOME/Downloads" && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "$PWD:/data:ro" -v "$HOME/Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -v "$PWD:/data" -v "$HOME/Downloads:/downloads" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

Unter Windows mit Docker Desktop, in Command Prompt:

```
docker run --rm -it --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "%cd%:/data:ro" -v "%USERPROFILE%\Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest && docker run --rm -it --read-only --cap-drop ALL --security-opt no-new-privileges -v "%cd%:/data" -v "%USERPROFILE%\Downloads:/downloads" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Podman

Rootless; `--userns=keep-id` behält Ihren eigenen Dateibesitz und `:Z` benennt die Volumes auf SELinux-Systemen wie Fedora um:

```
mkdir -p "$HOME/Downloads" && podman run --rm -it --userns=keep-id --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "$PWD:/data:ro,Z" -v "$HOME/Downloads:/downloads:Z" ghcr.io/z4dm42f5kp-ux/numb-form:latest && podman run --rm -it --userns=keep-id --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -v "$PWD:/data:Z" -v "$HOME/Downloads:/downloads:Z" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Kubernetes

[`k8s/`](../../k8s) hat Beispielvolumenansprüche und zwei interaktive Pods. Füllen Sie `numb-form-data` mit dem Ordner über dem Arbeitsordner aus, dann:

```
kubectl apply -f k8s/pvc.yaml && kubectl apply -f k8s/numb-form-pod.yaml && kubectl wait --for=condition=Ready pod/numb-form --timeout=120s && kubectl attach -it numb-form && kubectl delete pod numb-form
```

Überprüfen Sie die Kopie im `numb-form-downloads`-Volume und verpflichten Sie dann:

```
kubectl apply -f k8s/numb-commit-pod.yaml && kubectl wait --for=condition=Ready pod/numb-commit --timeout=120s && kubectl attach -it numb-commit && kubectl delete pod numb-commit
```

Wenn die erste Aufforderung vor dem Anfügen gedruckt wurde, drücken Sie Enter, um sie erneut zu sehen. Die Pods verwenden ein schreibgeschütztes Root-Dateisystem, lassen alle Funktionen fallen, montieren kein Service-Account-Token und stoppen nach 30 Minuten. Für den optionalen Companies House-Schlüssel müssen Sie den `env`-Eintrag unkommentieren und das Geheimnis erstellen:

```
kubectl create secret generic numb-form --from-literal=companies-house='your-free-key'
```
