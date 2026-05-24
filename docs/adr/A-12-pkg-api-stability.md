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
  1. A `[PKG-BREAK]` entry in the `## Next` section of `CHANGELOG.md` describing what changed
     and how consumers should migrate.
  2. A minimum deprecation window of **two minor releases or three months**, whichever is
     longer, between the first deprecation notice and removal. During this window the old
     symbol must remain available (possibly with a `// Deprecated:` GoDoc comment pointing to
     the replacement).

### Changelog tag convention

Breaking changes to `pkg/` are tagged `[PKG-BREAK]` in `CHANGELOG.md`:

```
[PKG-BREAK] pkg/gopass: Remove Store.GetRevision — use Store.History instead (deprecated since 1.17.0)
```

The existing `helpers/changelog` generator reads `CHANGELOG.md` verbatim; no changes to that
tool are needed. The `[PKG-BREAK]` tag is a pure text convention, consistent with existing
tags such as `[SECURITY]`, `[BUGFIX]`, and `[FEATURE]`.

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
