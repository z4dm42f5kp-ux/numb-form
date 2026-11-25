#!/bin/sh
# Write the short description shown on a release page.
# usage: scripts/release-notes.sh stable vX.Y.Z
#        scripts/release-notes.sh edge <commit>
# stable: the version's section of CHANGES.md; if it has none (a rebuild
# with a newer Go), the pull requests merged since the previous release.
# edge: the pull requests merged into main since the newest stable
# release, newest first.
# Needs the full history (fetch-depth: 0). REPO (owner/name) makes the
# links of CHANGES.md absolute; it defaults to this repository.
set -eu
kind=${1:?usage: release-notes.sh stable|edge ref}
ref=${2:?usage: release-notes.sh stable|edge ref}
repo=${REPO:-z4dm42f5kp-ux/numb-form}
cd "$(dirname "$0")/.."

# merged FROM TO: one line per pull request merged into main, newest first.
merged() {
	range=$2
	[ -n "$1" ] && range="$1..$2"
	# shellcheck disable=SC2086 # range is one argument
	git log --first-parent --format='%s%x09%b' $range |
		awk -F '\t' '$1 ~ /^Merge pull request #[0-9]+ / {
			n = $1; sub(/^Merge pull request /, "", n); sub(/ .*/, "", n)
			t = $2; sub(/\n.*/, "", t)
			if (t == "") t = "(no title)"
			print "* " t " (" n ")"
		}'
}

verify="Verify the downloads as described in [docs/install.md](https://github.com/$repo/blob/main/docs/install.md#verifying-who-built-a-release)."

case $kind in
stable)
	prev=$(git describe --tags --abbrev=0 --match 'v[0-9]*' "$ref^" 2>/dev/null || true)
	section=$(awk -v v="## $ref" '$0 == v { on = 1; next } on && /^## / { exit } on' CHANGES.md |
		sed -e '/./,$!d' |
		sed -E "s#\]\(([^):\#]+)\)#](https://github.com/$repo/blob/$ref/\1)#g")
	if [ -n "$section" ]; then
		printf '%s\n' "$section"
	else
		list=$(merged "$prev" "$ref")
		if [ -n "$list" ]; then
			printf 'Changes since %s:\n\n%s\n' "${prev:-the first release}" "$list"
		else
			printf 'The programs of %s, built again with a newer Go.\n' "${prev:-the previous release}"
		fi
	fi
	;;
edge)
	stable=$(git describe --tags --abbrev=0 --match 'v[0-9]*' "$ref" 2>/dev/null || true)
	list=$(merged "$stable" "$ref" | head -n 30)
	printf 'Rolling build of %s, rebuilt after every change to main (`numb-form update rolling`). Not a stable release.\n\n' "$(git rev-parse --short "$ref")"
	if [ -n "$list" ]; then
		printf 'Changes since %s:\n\n%s\n' "${stable:-the first release}" "$list"
	else
		printf 'The same programs as %s.\n' "${stable:-the newest release}"
	fi
	;;
*)
	echo "release-notes.sh: unknown kind $kind" >&2
	exit 2
	;;
esac
printf '\n%s\n' "$verify"
