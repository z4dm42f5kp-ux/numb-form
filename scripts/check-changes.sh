#!/bin/sh
# A pull request that changes the programs must say so in CHANGES.md
# (.github/DOCUMENTATION.md, rule 3) and change its English manual page
# (rule 6), unless every CHANGES.md line it adds starts with "Internal:".
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
if ! printf '%s\n' "$files" | grep -qx 'CHANGES.md'; then
	echo "check-changes: these files change the programs, but CHANGES.md does not say so:" >&2
	printf '%s\n' "$code" >&2
	echo "Add a line under the next version (see .github/DOCUMENTATION.md)." >&2
	exit 1
fi
if printf '%s\n' "$files" | grep -qE '^(docs/[^/]+\.md|README\.md|SECURITY\.md)$'; then
	echo "check-changes: OK, CHANGES.md and the manual describe the change"
	exit 0
fi
added=$(git diff "$base" "$head" -- CHANGES.md | grep -E '^\+\* ' || true)
if [ -n "$added" ] && ! printf '%s\n' "$added" | grep -vqE '^\+\* Internal:'; then
	echo "check-changes: OK, an internal change that no person can notice"
	exit 0
fi
echo "check-changes: these files change the programs, but no English manual page changed:" >&2
printf '%s\n' "$code" >&2
echo "Describe the change in its page under docs/, or, if no person can notice it, start its CHANGES.md lines with 'Internal:'." >&2
exit 1
