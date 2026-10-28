#!/bin/sh
# Sign the checksum file of a release with PGP, when the owner has set up
# a PGP key (.github/RELEASING.md). Without a key it does nothing: the
# Sigstore signature already proves where SHA256SUMS comes from.
# Only SHA256SUMS is signed: it holds the SHA-256 of every archive, so a
# signature per archive would add nothing.
# usage: scripts/pgp-sign.sh dist-folder
# Environment: PGP_SIGNING_KEY, the armored secret key (repository
# secret; empty means no PGP signature), and PGP_PASSPHRASE, its
# passphrase (may be empty).
set -eu
dir=${1:?usage: pgp-sign.sh dist-folder}
if [ -z "${PGP_SIGNING_KEY:-}" ]; then
	echo "pgp-sign: no PGP key is set up; no SHA256SUMS.asc (Sigstore signs SHA256SUMS)"
	exit 0
fi
pub="$(cd "$(dirname "$0")/.." && pwd)/.github/release-key.asc"
if [ ! -s "$pub" ]; then
	echo "pgp-sign: PGP_SIGNING_KEY is set, but .github/release-key.asc, its public key, is missing" >&2
	exit 1
fi
sign=$(mktemp -d)
check=$(mktemp -d)
trap 'GNUPGHOME="$sign" gpgconf --kill all 2>/dev/null || true; GNUPGHOME="$check" gpgconf --kill all 2>/dev/null || true; rm -rf "$sign" "$check"' EXIT INT TERM
printf '%s\n' "$PGP_SIGNING_KEY" | GNUPGHOME="$sign" gpg --batch --quiet --import
printf '%s' "${PGP_PASSPHRASE:-}" | GNUPGHOME="$sign" gpg --batch --yes --pinentry-mode loopback \
	--passphrase-fd 0 --armor --detach-sign --output "$dir/SHA256SUMS.asc" "$dir/SHA256SUMS"
# The signature must verify with the published public key alone.
GNUPGHOME="$check" gpg --batch --quiet --import "$pub"
GNUPGHOME="$check" gpg --batch --verify "$dir/SHA256SUMS.asc" "$dir/SHA256SUMS"
echo "pgp-sign: made $dir/SHA256SUMS.asc"
