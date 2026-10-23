#!/bin/sh
# Enforce the zero third-party dependency policy: go.mod may not require
# any module and go.sum must not exist. Adding a dependency is a security
# decision; see SECURITY.md before changing this script.
set -eu
cd "$(dirname "$0")/.."
if grep -Eq '^[[:space:]]*require([[:space:]]|\()' go.mod; then
	echo "check-deps: go.mod requires third-party modules:" >&2
	grep -En 'require' go.mod >&2
	exit 1
fi
if [ -e go.sum ]; then
	echo "check-deps: go.sum exists, so a dependency was added" >&2
	exit 1
fi
if [ -d vendor ]; then
	echo "check-deps: vendor/ must not exist" >&2
	exit 1
fi
# Every import must be from the standard library or this module.
nonstd=$(go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./... | grep -v '^github.com/z4dm42f5kp-ux/numb-form' || true)
if [ -n "$nonstd" ]; then
	echo "check-deps: non-standard imports found:" >&2
	echo "$nonstd" >&2
	exit 1
fi
# Packages that bypass Go's safety or its secure defaults
# (.github/SECURE-CODING.md): no cgo, no math/rand, and unsafe only for
# Windows system calls.
bad=$(grep -rlE --include='*.go' '^[[:space:]]*import "C"|"math/rand(/v2)?"' . || true)
if [ -n "$bad" ]; then
	echo "check-deps: cgo or math/rand used in:" >&2
	echo "$bad" >&2
	exit 1
fi
bad=$(grep -rl --include='*.go' '"unsafe"' . | grep -v '_windows\.go$' || true)
if [ -n "$bad" ]; then
	echo "check-deps: unsafe outside Windows system calls in:" >&2
	echo "$bad" >&2
	exit 1
fi
echo "check-deps: OK, standard library only"
