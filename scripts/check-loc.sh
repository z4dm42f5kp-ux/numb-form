#!/bin/sh
# Enforce the size limits of AGENTS.md: every Go, shell and Python file
# at most 300 lines, every Go function at most 100 lines. A function is
# counted from its "func" line to its closing brace, as gofmt lays it out.
set -eu
cd "$(dirname "$0")/.."
maxfile=300
maxfunc=100
status=0
for f in $(git ls-files '*.go' '*.sh' '*.py'); do
	n=$(wc -l < "$f")
	if [ "$n" -gt "$maxfile" ]; then
		echo "::error file=$f::$n lines, the limit is $maxfile; split the file" >&2
		status=1
	fi
done
for f in $(git ls-files '*.go'); do
	awk -v max="$maxfunc" -v f="$f" '
		/^func / { start = NR; name = $0; sub(/\(.*$/, "", name); sub(/^func /, "", name)
			if ($0 ~ /^func \(/) { name = $0; sub(/^func \([^)]*\) /, "", name); sub(/\(.*$/, "", name) }
			if ($0 ~ /}$/) start = 0; next }
		start && /^}/ { n = NR - start + 1
			if (n > max) { printf "::error file=%s,line=%d::function %s has %d lines, the limit is %d; split it\n", f, start, name, n, max; bad = 1 }
			start = 0 }
		END { exit bad }' "$f" >&2 || status=1
done
exit "$status"
