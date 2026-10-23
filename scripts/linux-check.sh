#!/bin/sh
# Linux compatibility check (.github/workflows/linux.yml): installs Go
# and the tools with the commands that docs/install.md shows, runs the
# tests with the distribution's own Go, and runs the documented install
# commands (the pinned release binary and the build from source).
# usage: linux-check.sh prepare|test|install DISTRO
# The repository is at /src (read-only) or, for guix, the current folder.
set -eu
step=${1:?usage: linux-check.sh prepare|test|install DISTRO}
distro=${2:?usage: linux-check.sh prepare|test|install DISTRO}
# The nixos/nix image has no sed or grep: run in a shell that has them.
if [ "$distro" = nixos ] && [ -z "${IN_NIX_SHELL:-}" ]; then
	exec nix-shell -p gnused gnugrep coreutils --run "sh '$0' '$step' '$distro'"
fi
src=/src
[ -d "$src/scripts" ] || src=$(pwd)
t=$src/scripts/install/templates

# row FILE LABEL: the command of the table row whose first cell starts
# with LABEL, run as root in a container without a person: no sudo, doas
# or --ask, and pacman without its question.
row() {
	# shellcheck disable=SC2016 # the backquotes are Markdown, not shell
	grep -F "| $2" "$t/$1" | head -n 1 | sed 's/^[^`]*`//; s/`[^`]*$//; s/\\|/|/g; s/sudo //g; s/doas //g; s/ --ask//g; s/pacman -S /pacman -S --noconfirm /g'
}
# doc NAME: the generated command NAME in docs/install.md.
doc() {
	sed -n "/<!-- install:$1 -->/,/<!-- \/install:$1 -->/p" "$src/docs/install.md" | grep -v '^<!--' | grep -v '^```'
}
label() {
	case $distro in
	debian) echo "Debian 13" ;;
	fedora) echo "Fedora 44" ;;
	arch) echo "Arch Linux" ;;
	gentoo) echo "Gentoo" ;;
	void) echo "Void Linux" ;;
	opensuse-*) echo "openSUSE" ;;
	alpine) echo "Alpine Linux" ;;
	*) echo "" ;;
	esac
}
# inside CMD: run CMD where Go and git exist.
inside() {
	case $distro in
	nixos) nix-shell -p go git cacert curl --run "$1" ;;
	guix) guix shell go git nss-certs coreutils curl tar gzip -- sh -c "export SSL_CERT_DIR=\"\$GUIX_ENVIRONMENT/etc/ssl/certs\" SSL_CERT_FILE=\"\$GUIX_ENVIRONMENT/etc/ssl/certs/ca-certificates.crt\" GIT_SSL_CAINFO=\"\$GUIX_ENVIRONMENT/etc/ssl/certs/ca-certificates.crt\" && $1" ;;
	*) sh -c "$1" ;;
	esac
}

case $step in
prepare)
	# What a minimal container lacks and a real installation has.
	case $distro in
	debian) export DEBIAN_FRONTEND=noninteractive ;;
	arch) pacman -Syu --noconfirm ;;
	gentoo)
		emerge-webrsync -q
		getuto
		export FEATURES="getbinpkg -sandbox -usersandbox -ipc-sandbox -network-sandbox -pid-sandbox" ;;
	void) xbps-install -Syu xbps && xbps-install -Syu ;;
	opensuse-*) zypper -n refresh ;;
	alpine) apk update ;;
	esac
	case $distro in
	nixos | guix) ;;
	*)
		sh -c "$(row linux-tools.md "$(label)")"
		sh -c "$(row go-packages.md "$(label)")"
		;;
	esac
	;;
test)
	inside "rm -rf /tmp/src && cp -R '$src' /tmp/src && cd /tmp/src && export HOME=/tmp/home GOCACHE=/tmp/gocache GOPATH=/tmp/gopath CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-mod=readonly && mkdir -p /tmp/home && go version && go vet ./... && go test -count=1 ./..."
	;;
install)
	want=$(sed -n 's/^version //p' "$src/scripts/install/pins.txt")
	home=$(mktemp -d)
	inside "export HOME='$home' && $(doc linux)" | tee /tmp/bin.log
	grep -q "$want" /tmp/bin.log || { echo "the release binary did not report $want" >&2; exit 1; }
	home=$(mktemp -d)
	case $distro in
	nixos) cmd=$(doc source-nixos) ;;
	guix) cmd=$(doc source-guix) ;;
	*) cmd=$(doc source-unix) ;;
	esac
	case $distro in
	nixos | guix) HOME=$home sh -c "$cmd" | tee /tmp/src.log ;;
	*) inside "export HOME='$home' && $cmd" | tee /tmp/src.log ;;
	esac
	grep -q "dev" /tmp/src.log || { echo "the build from source did not report dev" >&2; exit 1; }
	echo "OK: $distro"
	;;
*)
	echo "unknown step $step" >&2
	exit 2
	;;
esac
