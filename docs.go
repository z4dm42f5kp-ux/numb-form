// Package numbform holds the documentation that is built into the
// programs, so that "numb-form man" works without the repository.
package numbform

import "embed"

// Docs holds README.md, SECURITY.md and docs/ (English, Spanish and
// German).
//
//go:embed README.md SECURITY.md docs/*.md docs/es/*.md docs/de/*.md
var Docs embed.FS
