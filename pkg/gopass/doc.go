// Package gopass contains the public gopass API.
//
// # Stability
//
// This package is best-effort stable. Additive changes (new exported symbols,
// new functional-option parameters) may appear in any release. Breaking changes
// (removal or signature change of an exported symbol, change of interface method
// sets or error semantics) require a [PKG-BREAK] entry in CHANGELOG.md and a
// minimum deprecation window of two minor releases or three months before the
// old symbol is removed. See docs/adr/A-12-pkg-api-stability.md for the full
// policy.
//
// Embedders that need to answer age identity passphrase prompts without using
// the gopass CLI can install a callback on the context with
// [github.com/gopasspw/gopass/pkg/ctxutil.WithPasswordCallback]. For
// non-interactive use cases with a fixed passphrase, use
// [github.com/gopasspw/gopass/pkg/ctxutil.WithAgePassphrase].
//
// Known consumers of this API:
//   - https://github.com/gopasspw/gopass-hibp
//   - https://github.com/gopasspw/gopass-jsonapi
//   - https://github.com/gopasspw/git-credential-gopass
//   - https://github.com/gopasspw/gopass-summon-provider
package gopass
