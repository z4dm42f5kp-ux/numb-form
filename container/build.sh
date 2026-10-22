#!/bin/sh
# Build the numb-form container image with docker or podman.
# usage: container/build.sh [image-name]
# The Go build cache is temporary and removed afterwards.
set -eu
image=${1:-numb-form:local}
engine=${ENGINE:-}
if [ -z "$engine" ]; then
	if command -v podman >/dev/null 2>&1; then engine=podman
	elif command -v docker >/dev/null 2>&1; then engine=docker
	else echo "build.sh: neither podman nor docker found" >&2; exit 1
	fi
fi
here=$(cd "$(dirname "$0")" && pwd)
ctx=$(mktemp -d)
GOCACHE=$(mktemp -d)
export GOCACHE
trap 'rm -rf "$ctx" "$GOCACHE"' EXIT INT TERM
arch=$(uname -m)
case $arch in
x86_64|amd64) goarch=amd64 ;;
aarch64|arm64) goarch=arm64 ;;
*) echo "build.sh: unsupported architecture $arch" >&2; exit 1 ;;
esac
cd "$here/.."
version=$(git describe --tags --always 2>/dev/null || echo dev)
for c in numb-form numb-commit numb-date; do
	CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH:-$goarch} go build -trimpath -mod=readonly \
		-tags timetzdata -ldflags "-s -w -X github.com/z4dm42f5kp-ux/numb-form/internal/app.Version=$version" \
		-o "$ctx/$c" "./cmd/$c"
done
for f in /etc/ssl/certs/ca-certificates.crt /etc/pki/tls/certs/ca-bundle.crt /etc/ssl/cert.pem; do
	if [ -s "$f" ]; then cp "$f" "$ctx/ca-certificates.crt"; break; fi
done
[ -s "$ctx/ca-certificates.crt" ] || { echo "build.sh: no CA bundle found" >&2; exit 1; }
cp "$here/Dockerfile" "$ctx/Dockerfile"
"$engine" build -t "$image" "$ctx"
echo "built $image with $engine"
