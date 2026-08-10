# A-12: `pkg/gopass` API Stability Contract

**Status:** accepted  
**Source:** [GitHub Issue #3414](https://github.com/gopasspw/gopass/issues/3414)

---

## Background

`ARCHITECTURE.md` (lines 37–43) documents that gopass applies semantic versioning to the CLI
tool only, not the Go module. `pkg/gopass` is the documented integration point for external
consumers (gopass-hibp, gopass-jsonapi, git-credential-gopass, gopass-summon-provider, and
others). There is currently no documented contract for what constitutes a breaking change or
how consumers will be notified.

A breaking change to `pkg/gopass.Store`, `pkg/gopass.Secret`, or related types can affect
consumers silently — the module version does not increment, no changelog entry was mandated,
and `go get` would pull the break without warning.

---

## Decision

Implement **Options B and C**. They are complementary and together give consumers both
runtime/source-level signals (C) and a reliable notification channel (B).

Option A (full module semver / v2 path) is deferred — the current team size makes the
coordination overhead infeasible.

Option D (separate repository) is deferred for the same reasons.

---

## Option B — "Best-effort stable" policy with mandatory changelog tags

### Policy

`pkg/gopass` and its sub-packages are declared **best-effort stable**:

* Additive changes (new exported symbols, new optional parameters via functional options) may
  appear in any release without prior notice.
* **Breaking changes** (removal or signature change of an exported symbol, change of error
  semantics, change of an interface method set) require:
  1. A `PKG-BREAK:` footer on the commit, which produces a `[PKG-BREAK]`-prefixed entry in the
     `## [Unreleased]` section of `CHANGELOG.md` describing what changed and how consumers
     should migrate.
  2. A minimum deprecation window of **two minor releases or three months**, whichever is
     longer, between the first deprecation notice and removal. During this window the old
     symbol must remain available (possibly with a `// Deprecated:` GoDoc comment pointing to
     the replacement).

### Changelog tag convention

A breaking change to `pkg/` is declared with a `PKG-BREAK:` commit footer:

```
refactor(pkg/gopass): drop Store.GetRevision in favour of Store.History

Deprecated since 1.17.0; the two-minor / three-month window has elapsed.

PKG-BREAK: pkg/gopass: Remove Store.GetRevision — use Store.History instead
Signed-off-by: Your Name <your@example.com>
```

`helpers/commitmsg` reads that footer and `helpers/release` renders the entry with a
`[PKG-BREAK]` prefix inside the appropriate Keep a Changelog subsection, normally `Changed`.
A `Changelog-Section: Removed` footer moves it to `Removed` when the symbol is gone rather
than changed.

The footer is the machine-readable form of the tag. Writing `[PKG-BREAK]` by hand into
`## [Unreleased]` also works and is preserved by the release helper.

Note: `CHANGELOG.md` now follows Keep a Changelog 1.1.0, so `[SECURITY]`, `[BUGFIX]` and
`[FEATURE]` are no longer tags — they are the `Security`, `Fixed` and `Added` subsections.
`[PKG-BREAK]` remains a bullet prefix because it qualifies an entry rather than categorising
it. See [docs/conventions.md](../conventions.md).

### Enforcement

* Code reviewers must reject PRs that remove or change exported `pkg/` symbols without a
  corresponding `[PKG-BREAK]` changelog entry and a prior deprecation notice.
* The `golangci-lint` `godot` and `godox` rules already catch missing doc-comment periods and
  stray TODO/FIXME markers; no additional tooling is introduced.

---

## Option C — Explicit stability annotations in package doc comments

Each package under `pkg/gopass/` carries a doc comment that states its stability level.

| Package | Level | Rationale |
|---------|-------|-----------|
| `pkg/gopass` | **best-effort stable** | Core interfaces; multiple known consumers |
| `pkg/gopass/api` | **best-effort stable** | Primary API implementation |
| `pkg/gopass/secrets` | **best-effort stable** | Secret types consumed by integrations |
| `pkg/gopass/apimock` | **testing helper — no stability guarantee** | Internal test double; consumers should copy or vendor it |

The standard `// Deprecated:` GoDoc convention is used to signal pending removal of individual
symbols; `godoc` and `pkg.go.dev` render these prominently.

---

## Affected packages

`pkg/gopass/`, `pkg/gopass/api/`, `pkg/gopass/apimock/`, `pkg/gopass/secrets/`

---

## Consequences

* External integrators get a documented, reasonable stability promise without requiring a
  module-path change today.
* The changelog `[PKG-BREAK]` tag gives integrators a single place to scan for migration work
  when upgrading.
* The deprecation window gives integrators time to adapt before a symbol is removed.
* The policy can be upgraded to full module semver in the future if maintainer capacity grows,
  without breaking any existing convention.
