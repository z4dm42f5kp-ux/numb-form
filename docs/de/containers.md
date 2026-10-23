# numb-form in Containern

> Maschinelle Übersetzung (Argos Translate) des [englischen Originals](../containers.md). Maßgeblich ist die englische Fassung.

Das Bild enthält nur die beiden statischen Binärdateien, ein CA-Bundle und die Zeitzonendatenbank. Es wird kein Basisbild verwendet. Es läuft als unprivilegierter Benutzer (UID 65532).

Es werden zwei Bände benötigt:

* `/data`: der Ordner *über* dem Arbeitsordner. Im Standard-Direktmodus schreibt numb-form die neuen Ordner hinein. Im Überprüfungsmodus liest numb-form es nur, so dass es schreibgeschützt montiert werden kann, und numb-commit schreibt später die akzeptierten Ordner hinein.
* `/downloads`: Logs und Manifeste (direkter Modus, über `log_dir`) oder die Überprüfungskopie (Review-Modus).

Mit `confirm = each` (Standard) muss jedes Dokument bestätigt werden, sodass der Container interaktiv läuft (`-it`). Eine Konfigurationsdatei wird mit `-f` übergeben, beispielsweise aus einem dritten Volume. Führen Sie im Überprüfungsmodus numb-commit auf die gleiche Weise aus, wie Sie numb-form mit den gleichen Halterungen ausgeführt haben, da das Manifest Containerpfade aufzeichnet.

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

**Unix-like** (macOS, Linux, BSD) begann im Terminal im Arbeitsordner (oder im Ordner darüber):

```
mkdir -p "$HOME/Downloads/numb-form-logs" && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -e XDG_STATE_HOME=/downloads/numb-form-logs -v "$PWD:/data" -v "$HOME/Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

**Windows** (10, 11) mit Docker Desktop, in Command Prompt (in PowerShell schreiben Sie `${PWD}` für `%cd%` und `$env:USERPROFILE` für `%USERPROFILE%`):

```
docker run --rm -it --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -e XDG_STATE_HOME=/downloads/numb-form-logs -v "%cd%:/data" -v "%USERPROFILE%\Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Podman

Rootless; `--userns=keep-id` behält Ihren eigenen Dateibesitz und `:Z` benennt die Volumes auf SELinux-Systemen wie Fedora neu. **Unix-like** (Linux, macOS, BSD), im Terminal:

```
mkdir -p "$HOME/Downloads/numb-form-logs" && podman run --rm -it --userns=keep-id --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -e XDG_STATE_HOME=/downloads/numb-form-logs -v "$PWD:/data:Z" -v "$HOME/Downloads:/downloads:Z" ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

## Überprüfungsmodus in einem Container

Mounten Sie den Arbeitsordner Read-only für den Lauf, Read-Write für den Commit und übergeben Sie `mode = review` in einer Konfigurationsdatei. **Unix-like** (Linux, macOS, BSD), im Terminal:

```
mkdir -p "$HOME/Downloads" && printf 'mode = review\n' > "$HOME/Downloads/review.conf" && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -e TZ=Europe/Berlin -v "$PWD:/data:ro" -v "$HOME/Downloads:/downloads" ghcr.io/z4dm42f5kp-ux/numb-form:latest -f /downloads/review.conf && docker run --rm -it --user "$(id -u):$(id -g)" --read-only --cap-drop ALL --security-opt no-new-privileges -v "$PWD:/data" -v "$HOME/Downloads:/downloads" --entrypoint /numb-commit ghcr.io/z4dm42f5kp-ux/numb-form:latest
```

Podman nimmt die gleichen Optionen sowie `--userns=keep-id` und `:Z`.

## Kubernetes

[`k8s/`](../../k8s) hat Beispielvolumenansprüche und zwei interaktive Pods. Füllen Sie `numb-form-data` mit dem Arbeitsordner (oder dem Ordner darüber), dann:

```
kubectl apply -f k8s/pvc.yaml && kubectl apply -f k8s/numb-form-pod.yaml && kubectl wait --for=condition=Ready pod/numb-form --timeout=120s && kubectl attach -it numb-form && kubectl delete pod numb-form
```

Überprüfen Sie im Überprüfungsmodus die Kopie im `numb-form-downloads`-Volume und verpflichten Sie dann:

```
kubectl apply -f k8s/numb-commit-pod.yaml && kubectl wait --for=condition=Ready pod/numb-commit --timeout=120s && kubectl attach -it numb-commit && kubectl delete pod numb-commit
```

Wenn die erste Aufforderung vor dem Anfügen gedruckt wurde, drücken Sie Enter, um sie erneut zu sehen. Die Pods verwenden ein schreibgeschütztes Root-Dateisystem, lassen alle Funktionen fallen, montieren kein Service-Account-Token und stoppen nach 30 Minuten.
