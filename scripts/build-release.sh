#!/bin/sh
# Build reproducible release archives into dist/.
# usage: scripts/build-release.sh v1.2.3
set -eu
version=${1:?usage: build-release.sh version}
cd "$(dirname "$0")/.."
rm -rf dist
mkdir -p dist/bin
SOURCE_DATE_EPOCH=${SOURCE_DATE_EPOCH:-$(git log -1 --format=%ct)}
export SOURCE_DATE_EPOCH CGO_ENABLED=0
targets="linux/amd64 linux/arm64 linux/386 linux/armv7
darwin/amd64 darwin/arm64
windows/amd64 windows/arm64 windows/386
freebsd/amd64 freebsd/arm64
openbsd/amd64 openbsd/arm64
netbsd/amd64 netbsd/arm64"
ldflags="-s -w -buildid= -X github.com/z4dm42f5kp-ux/numb-form/internal/app.Version=$version"
for t in $targets; do
	os=${t%/*}
	arch=${t#*/}
	goarch=$arch
	goarm=
	if [ "$arch" = armv7 ]; then goarch=arm; goarm=7; fi
	ext=
	[ "$os" = windows ] && ext=.exe
	out="dist/bin/${os}_${arch}"
	mkdir -p "$out"
	for c in numb-form numb-commit numb-date; do
		GOOS=$os GOARCH=$goarch GOARM=$goarm go build -trimpath -mod=readonly -tags timetzdata \
			-ldflags "$ldflags" -o "$out/$c$ext" "./cmd/$c"
	done
	cp README.md LICENSE "$out/"
	touch -d "@$SOURCE_DATE_EPOCH" "$out"/* 2>/dev/null || true
	name="numb-form_${os}_${arch}"
	if [ "$os" = windows ]; then
		(cd "$out" && zip -X -q "../../$name.zip" numb-form.exe numb-commit.exe numb-date.exe README.md LICENSE)
	else
		tar --sort=name --mtime="@$SOURCE_DATE_EPOCH" --owner=0 --group=0 --numeric-owner \
			-C "$out" -cf - numb-form numb-commit numb-date README.md LICENSE | gzip -n -9 > "dist/$name.tar.gz"
	fi
done
cd dist
sha256sum ./*.tar.gz ./*.zip | sed 's| \./| |' > SHA256SUMS
: > SHA256
for f in *.tar.gz *.zip; do
	printf 'SHA256 (%s) = %s\n' "$f" "$(sha256sum "$f" | cut -d' ' -f1)" >> SHA256
done
cat SHA256SUMS
