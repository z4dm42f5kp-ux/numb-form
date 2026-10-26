#!/bin/sh
# A pull request that changes the programs must say so in CHANGES.md
# (.github/DOCUMENTATION.md, rule 3).
# usage: scripts/check-changes.sh base-commit head-commit
set -eu
base=${1:?usage: check-changes.sh base head}
head=${2:?usage: check-changes.sh base head}
cd "$(dirname "$0")/.."
files=$(git diff --name-only "$base" "$head")
code=$(printf '%s\n' "$files" | grep -E '^(cmd/|internal/|docs\.go$)' | grep -vE '(_test\.go$|^internal/testfix/)' || true)
if [ -z "$code" ]; then
	echo "check-changes: the programs are unchanged"
	exit 0
fi
if printf '%s\n' "$files" | grep -qx 'CHANGES.md'; then
	echo "check-changes: OK, CHANGES.md lists the change"
	exit 0
fi
echo "check-changes: these files change the programs, but CHANGES.md does not say so:" >&2
printf '%s\n' "$code" >&2
echo "Add a line under the next version (see .github/DOCUMENTATION.md)." >&2
exit 1
