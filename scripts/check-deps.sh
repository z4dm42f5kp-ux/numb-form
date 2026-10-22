#!/bin/sh
# Enforce the dependency policy: the standard library, plus exactly one
# module, golang.org/x/net (for its HTML parser, golang.org/x/net/html),
# at the version pinned in go.mod and checked against go.sum, which Go
# verifies against the Go checksum database. Adding anything else is a
# security decision; see SECURITY.md before changing this script.
set -eu
cd "$(dirname "$0")/.."
allowed='golang.org/x/net'
reqs=$(awk '/^require \(/{b=1;next} b&&/^\)/{b=0} b{print $1} /^require [^(]/{print $2}' go.mod)
for m in $reqs; do
	if [ "$m" != "$allowed" ]; then
		echo "check-deps: go.mod requires a module that is not allowed: $m" >&2
		exit 1
	fi
done
if grep -q '// indirect' go.mod; then
	echo "check-deps: go.mod has indirect requirements" >&2
	exit 1
fi
if [ -d vendor ]; then
	echo "check-deps: vendor/ must not exist" >&2
	exit 1
fi
if [ -n "$reqs" ] && [ ! -s go.sum ]; then
	echo "check-deps: go.sum is missing" >&2
	exit 1
fi
# Every import must be from the standard library, this module, or the
# HTML parser of golang.org/x/net.
nonstd=$(go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./... |
	grep -v -e '^github.com/z4dm42f5kp-ux/numb-form' -e '^golang.org/x/net/html$' -e '^golang.org/x/net/html/atom$' || true)
if [ -n "$nonstd" ]; then
	echo "check-deps: imports outside the standard library and golang.org/x/net/html:" >&2
	echo "$nonstd" >&2
	exit 1
fi
# The downloaded module must match go.sum (and so the checksum database).
go mod verify >/dev/null
echo "check-deps: OK, standard library and golang.org/x/net/html only"
