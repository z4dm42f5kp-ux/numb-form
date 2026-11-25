#!/bin/sh
# Daily operating system tracking (.github/workflows/updates.yml, job
# systems). Run from the repository root.
#
# 1. Every container image in the workflows is pinned as
#    name:tag@sha256:digest. When the tag now points to another digest
#    (security updates of Debian, Fedora, openSUSE and the rolling
#    distributions), the digest is replaced in place.
# 2. When a fixed-release distribution has a newer stable release
#    (name:latest no longer matches the pinned tag), or the BSD virtual
#    machine action supports newer FreeBSD, OpenBSD or NetBSD versions
#    than ci.yml tests, it is written to systems-report.txt, for a person
#    to update the tests and the documentation.
set -eu
report=systems-report.txt
: > "$report"

# digest IMAGE TAG: the current digest of IMAGE:TAG (Docker Hub or ghcr.io).
digest() {
	case $1 in
	ghcr.io/*)
		repo=${1#ghcr.io/}
		token=$(curl -fsS "https://ghcr.io/token?scope=repository:$repo:pull" | python3 -c 'import json,sys; print(json.load(sys.stdin)["token"])')
		curl -fsSI -H "Authorization: Bearer $token" \
			-H "Accept: application/vnd.oci.image.index.v1+json, application/vnd.docker.distribution.manifest.list.v2+json" \
			"https://ghcr.io/v2/$repo/manifests/$2" | tr -d '\r' | sed -n 's/^[Dd]ocker-[Cc]ontent-[Dd]igest: //p'
		;;
	*)
		repo=$1
		case $repo in */*) ;; *) repo=library/$repo ;; esac
		curl -fsS "https://hub.docker.com/v2/repositories/$repo/tags/$2" | python3 -c 'import json,sys; print(json.load(sys.stdin)["digest"])'
		;;
	esac
}

grep -rhoE '[a-z0-9./-]+:[A-Za-z0-9._-]+@sha256:[0-9a-f]{64}' .github/workflows | sort -u | while read -r ref; do
	image=${ref%%:*}
	rest=${ref#*:}
	tag=${rest%%@*}
	pinned=${rest#*@}
	now=$(digest "$image" "$tag") || { echo "$image:$tag: registry not reachable, skipped"; continue; }
	case $now in sha256:*) ;; *) echo "$image:$tag: no digest, skipped"; continue ;; esac
	if [ "$now" != "$pinned" ]; then
		echo "$image:$tag: $pinned -> $now"
		grep -rl "$image:$tag@$pinned" .github/workflows | while read -r f; do
			sed -i "s#$image:$tag@$pinned#$image:$tag@$now#g" "$f"
		done
	else
		echo "$image:$tag: current"
	fi
	case $tag in
	latest | *[!0-9.]*) ;;
	*)
		latest=$(digest "$image" latest) || continue
		if [ "$latest" != "$now" ]; then
			echo "* $image: $image:latest is no longer $tag; a newer stable release is out. Test it in linux.yml and name it in scripts/install/templates." >> "$report"
		fi
		;;
	esac
done

# The BSD versions the pinned virtual machine action supports, from its
# readme on its main branch, against the versions ci.yml tests.
readme=$(curl -fsS https://raw.githubusercontent.com/cross-platform-actions/action/master/readme.md) || readme=""
for os in freebsd openbsd netbsd; do
	tested=$(grep -oE "os: $os, version: \"[0-9.]+\"" .github/workflows/ci.yml | grep -oE '[0-9.]+' | sort -V | tail -n 1)
	name=$(printf '%s' "$os" | sed 's/bsd$/BSD/; s/^free/Free/; s/^open/Open/; s/^net/Net/')
	newest=$(printf '%s\n' "$readme" | grep -oE "$name [0-9]+\.[0-9]+" | grep -oE '[0-9]+\.[0-9]+' | sort -V | tail -n 1)
	[ -n "$tested" ] && [ -n "$newest" ] || continue
	if [ "$(printf '%s\n%s\n' "$tested" "$newest" | sort -V | tail -n 1)" != "$tested" ]; then
		echo "* $name $newest is supported by cross-platform-actions; ci.yml tests $tested." >> "$report"
	fi
done
cat "$report"
