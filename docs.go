// Package numbform holds the documentation that is built into the
// programs, so that "numb-form man" works without the repository.
//
// In plain words: the manual pages are copied into the programs when
// they are built, so they can be read on a computer without internet
// access.
package numbform

import "embed"

// Docs holds README.md, SECURITY.md, CHANGES.md and docs/ (English,
// Spanish and German).
//
//go:embed README.md SECURITY.md CHANGES.md docs/*.md docs/es/*.md docs/de/*.md
var Docs embed.FS
