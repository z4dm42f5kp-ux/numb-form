#!/bin/sh
# Run numb-form, numb-commit, numb-date or numb-typo in a throwaway
# FreeBSD jail: FreeBSD's own containers, part of the base system, for
# machines where Docker and Podman are not available.
#
# In plain words: the program runs in an empty, read-only mini system
# that sees only the working folder and the Downloads folder, has no
# network, runs as you (not as root), and is removed afterwards.
#
# usage: doas numb-form-jail [program [options]]
#        (installed from the FreeBSD release archive, docs/containers.md;
#        or sudo; program is numb-form, numb-commit, numb-date or numb-typo)
#
# Run it inside the working folder (or the folder above it), as with
# numb-form itself. The programs are taken from the folder of the
# installed numb-form (or from NUMBFORM_BIN), and must be the static
# release programs. Everything the jail needs is made in a temporary
# folder, and every mount is removed again, also after Ctrl+C.
set -eu
[ "$(uname -s)" = FreeBSD ] || { echo "jail.sh: jails exist only on FreeBSD" >&2; exit 1; }
[ "$(id -u)" = 0 ] || { echo "jail.sh: start it with doas or sudo" >&2; exit 1; }
prog=${1:-numb-form}
[ $# -gt 0 ] && shift
case $prog in
numb-form | numb-commit | numb-date | numb-typo) ;;
*) echo "jail.sh: unknown program $prog" >&2; exit 2 ;;
esac
user=${DOAS_USER:-${SUDO_USER:-}}
[ -n "$user" ] && [ "$user" != root ] || { echo "jail.sh: start it with doas or sudo as your own user" >&2; exit 1; }
uid=$(id -u "$user")
gid=$(id -g "$user")
home=$(getent passwd "$user" | cut -d: -f6)
bin=${NUMBFORM_BIN:-$(dirname "$(command -v numb-form || echo "$home/.local/bin/numb-form")")}
[ -x "$bin/$prog" ] || { echo "jail.sh: $bin/$prog not found; set NUMBFORM_BIN" >&2; exit 1; }
data=$(pwd)
downloads=${NUMBFORM_DOWNLOADS:-$home/Downloads}
if [ ! -d "$downloads" ]; then
	mkdir -p "$downloads"
	chown "$uid:$gid" "$downloads"
fi

stage=$(mktemp -d /tmp/numb-form-jail.XXXXXX)
root="$stage/root"
cleanup() {
	for m in downloads data dev; do
		umount "$root/$m" 2>/dev/null || true
	done
	umount "$root" 2>/dev/null || true
	# Never remove the folder while the working folder or Downloads is
	# still mounted inside it: that would delete their files. rm -x also
	# never crosses into another file system.
	if mount -p | awk '{print $2}' | grep -qF "$stage"; then
		echo "jail.sh: $stage is still mounted; unmount it and remove it by hand" >&2
	else
		rm -rfx "$stage"
	fi
}
trap cleanup EXIT INT TERM

# The jail's files: the programs, a user entry, the time zone.
tree="$stage/tree"
mkdir -p "$root" "$tree/etc" "$tree/data" "$tree/downloads" "$tree/dev" "$tree/tmp"
for p in numb-form numb-commit numb-date numb-typo; do
	[ -x "$bin/$p" ] && install -m 0555 "$bin/$p" "$tree/$p"
done
printf '%s:*:%s:%s::0:0:numb-form:/downloads:/nonexistent\n' "$user" "$uid" "$gid" > "$tree/etc/master.passwd"
printf '%s:*:%s:\n' "$user" "$gid" > "$tree/etc/group"
pwd_mkdb -p -d "$tree/etc" "$tree/etc/master.passwd"
for f in /etc/localtime /etc/login.conf /etc/login.conf.db; do
	[ -e "$f" ] && cp "$f" "$tree$f"
done

# Read-only root, the two folders read-write, a minimal /dev.
mount -t nullfs -o ro "$tree" "$root"
mount -t nullfs "$data" "$root/data"
mount -t nullfs "$downloads" "$root/downloads"
mount -t devfs devfs "$root/dev"
devfs -m "$root/dev" ruleset 4
devfs -m "$root/dev" rule applyset

# No IP addresses: the jail has no network. Nothing is allowed beyond the
# defaults, and the jail ends with the program.
env -i HOME=/downloads NUMBFORM_DOWNLOADS=/downloads XDG_STATE_HOME=/downloads/numb-form-logs \
	TERM="${TERM:-vt100}" ${TZ:+"TZ=$TZ"} PATH=/ \
	/usr/sbin/jail -c path="$root" name="numbform$$" host.hostname=numb-form \
	ip4=disable ip6=disable securelevel=3 children.max=0 \
	exec.jail_user="$user" exec.system_jail_user=false \
	command="/$prog" "$@"
