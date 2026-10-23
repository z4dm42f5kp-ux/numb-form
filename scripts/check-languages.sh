#!/bin/sh
# Daily language tracking (.github/workflows/updates.yml, job languages).
# Run from the repository root.
#
# 1. Go. The newest stable Go replaces the version named in the build
#    instructions of docs/install.md. When Go no longer fixes security
#    bugs in the minimum of go.mod or in CI's oldest tested Go, it is
#    reported.
# 2. Python. Every package pinned in the hash-locked requirements of the
#    workflows is checked against the OSV vulnerability database.
#
# Findings go to languages-report.txt. CHECK_GO_JSON and CHECK_OSV_JSON
# name files to use instead of the network, for trying it out.
set -eu
report=languages-report.txt
: > "$report"

if [ -n "${CHECK_GO_JSON:-}" ]; then
	gojson=$(cat "$CHECK_GO_JSON")
else
	gojson=$(curl -fsS 'https://go.dev/dl/?mode=json')
fi
# The list holds the supported releases (the two newest), newest first.
newest=$(printf '%s' "$gojson" | python3 -c 'import json,sys; print([r["version"] for r in json.load(sys.stdin) if r.get("stable")][0])')
oldest=$(printf '%s' "$gojson" | python3 -c '
import json, sys
v = [r["version"][2:].split(".") for r in json.load(sys.stdin) if r.get("stable")]
print(min((int(a), int(b)) for a, b, *_ in v).__str__().strip("()").replace(", ", "."))')
printf '%s' "$newest" | grep -Eq '^go1\.[0-9]+\.[0-9]+$' || { echo "unexpected Go version $newest" >&2; exit 1; }
echo "newest Go: $newest, oldest with security fixes: $oldest"

documented=$(grep -oE 'go1\.[0-9]+\.[0-9]+' docs/install.md | head -n 1)
if [ -n "$documented" ] && [ "$documented" != "$newest" ]; then
	echo "docs/install.md: $documented -> $newest"
	sed -i "s/$documented/$newest/g" docs/install.md
fi

older() { [ "$1" != "$2" ] && [ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | head -n 1)" = "$1" ]; }
gomod=$(sed -n 's/^go \([0-9]*\.[0-9]*\).*/\1/p' go.mod)
cimin=$(grep -oE 'go: "1\.[0-9]+\.x"' .github/workflows/ci.yml | grep -oE '1\.[0-9]+' | sort -V | head -n 1)
if older "$gomod" "$oldest"; then
	echo "* go.mod allows Go $gomod, but Go fixes security bugs only from $oldest on (newest: $newest). Raise it once the Linux distributions in docs/install.md ship $oldest or newer." >> "$report"
fi
if [ -n "$cimin" ] && older "$cimin" "$oldest"; then
	echo "* CI's oldest tested Go is $cimin, which no longer gets security fixes (oldest supported: $oldest)." >> "$report"
fi

# Python: name==version of every pinned package, then one OSV query.
reqs=$(ls .github/translate/requirements.txt .github/*requirements*.txt 2>/dev/null || true)
if [ -n "$reqs" ]; then
	# shellcheck disable=SC2086 # one argument per file
	query=$(python3 - $reqs <<'PY'
import json, re, sys
q = []
for f in sys.argv[1:]:
    for line in open(f, encoding="utf-8"):
        m = re.match(r"^([A-Za-z0-9_.-]+)(\[[^\]]*\])?==([^\s;\\]+)", line)
        if m:
            q.append({"package": {"name": m.group(1), "ecosystem": "PyPI"}, "version": m.group(3), "_file": f})
print(json.dumps(q))
PY
)
	body=$(printf '%s' "$query" | python3 -c 'import json,sys; print(json.dumps({"queries": [{k: v for k, v in x.items() if k != "_file"} for x in json.load(sys.stdin)]}))')
	if [ -n "${CHECK_OSV_JSON:-}" ]; then
		answer=$(cat "$CHECK_OSV_JSON")
	else
		answer=$(curl -fsS -X POST -H 'Content-Type: application/json' -d "$body" https://api.osv.dev/v1/querybatch)
	fi
	QUERY=$query python3 -c '
import json, os, sys
q = json.loads(os.environ["QUERY"])
for x, r in zip(q, json.load(sys.stdin).get("results", [])):
    ids = [v["id"] for v in r.get("vulns", [])]
    if ids:
        print("* %s %s (%s): %s" % (x["package"]["name"], x["version"], x["_file"], ", ".join(ids)))
' <<EOF2 >> "$report"
$answer
EOF2
	echo "Python: checked $(printf '%s' "$query" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))') pinned packages"
fi
cat "$report"
