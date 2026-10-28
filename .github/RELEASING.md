# Releasing

Releases are numbered `vX.Y.Z` ([semantic versioning](https://semver.org)):
a fix is a patch (`v1.4.1`), a new feature a minor release (`v1.5.0`).
Each release is one pull request titled `vX.Y.Z: what it changes`.

## What happens

1. The pull request is merged into `main`.
2. ci runs on `main`. When it passes, the **release approval** workflow
   checks the title, that `vX.Y.Z` is not taken and is newer than the
   newest release, and then waits. Before it asks you, it also waits for
   the Spanish and German translations of the merge (at most 40
   minutes): the commit that adds only them is the one released, so
   `numb-form man` in Spanish and German matches the release.
3. GitHub sends you a mail, *Review pending deployments*. Open it (or the
   run under Actions > release approval), tap **Review deployments**, tick
   **release** and tap **Approve and deploy**. **Reject** releases nothing.
4. The commit that ci tested is tagged `vX.Y.Z`, and the normal release
   build runs: archives, checksums, Sigstore signature, provenance,
   container image. The rolling build already contains it.
5. The **install-docs** workflow checks the new release's signature and
   opens a pull request *docs: install commands for vX.Y.Z*. Merge it so
   the documented install commands install the new release; merging it
   releases nothing.

A minor release (`vX.Y.0:`) also runs the **linux** workflow on its pull
request: the tests and the documented commands on every listed Linux
distribution. It takes up to an hour and a half; wait for it before
merging.

A pull request with another title (no `vX.Y.Z:` at the start) releases
nothing. To release the newest `main` without a pull request, run
Actions > release approval > Run workflow with the version; it waits for
your approval the same way.

## One-time setup

The approval comes from the `release` environment. Until it has you as
required reviewer, the workflow refuses to release (it never releases
without asking). On github.com (works in Safari on a phone; use the
desktop site if Settings is hidden):

1. Settings > Environments > New environment, name `release`,
   Configure environment.
2. Tick **Required reviewers** and add yourself. Leave
   **Prevent self-review** unticked, since you merge and approve.
3. Under **Deployment branches and tags**, choose Selected branches and
   tags and add the branch `main`.
4. Save protection rules.

## Optional PGP signature

Releases are signed with Sigstore without any key to look after.
A PGP signature of `SHA256SUMS` is added only when you set up a key.
The key is a long-lived secret: keep it offline, give it a passphrase,
and set an expiry date.
To set it up:

1. Make a signing key on your own computer, for example
   `gpg --quick-generate-key 'numb-form releases' ed25519 sign 2y`.
2. Commit its public key as `.github/release-key.asc`
   (`gpg --armor --export 'numb-form releases'`).
3. Settings > Secrets and variables > Actions > New repository secret:
   `PGP_SIGNING_KEY` with the output of
   `gpg --armor --export-secret-keys 'numb-form releases'`, and
   `PGP_PASSPHRASE` with its passphrase.

From the next release on, `SHA256SUMS.asc` is published with it, and the
release fails if the signature does not verify with
`.github/release-key.asc`.
To stop, delete the two secrets.

## Daily update checks

* Dependabot checks the actions and the translation tools every day
  (`.github/dependabot.yml`).
* The **updates** workflow checks every day for newer stable versions of
  the Go tools in the workflows and opens a pull request
  `ci: update Go tools: ...` (if workflows may not open pull requests in
  Settings > Actions > General, it opens an issue instead).
* It checks the operating systems the tests run on. When a pinned
  container image (Debian, Fedora, Arch, Gentoo, Void, openSUSE,
  Alpine, NixOS) has security updates, it opens a pull request
  `ci: security updates of the test systems`. When a distribution or
  FreeBSD, OpenBSD or NetBSD has a newer release, it opens or updates
  the issue *Newer operating system releases to test*.
* It also checks the binaries of the newest release and of the rolling
  build with govulncheck. When a newer stable Go fixes what it finds, it
  rebuilds the rolling build, and proposes the next patch version of the
  newest release, the same code built with that Go, which waits for your
  approval like any release. If you reject it, it asks again the next
  day; publishing any newer release ends it.
