# A-11: Context-Threading in pkg/ctxutil

**Status:** active — pattern is in use; reduction is planned but not scheduled  
**Source:** Architectural Review 2026-05-03 (filed as issue #3417)

---

## Background

`pkg/ctxutil/` defines more than 30 unexported context-key types used to pass
configuration values, terminal state, credentials, and backend selection flags
through the call stack. `ARCHITECTURE.md` describes this as "pragmatic
(read: non-idiomatic) approach to pass very specific configuration options
through multiple layers of abstraction."

The pattern was introduced to avoid bloating function signatures and interfaces
at a time when the codebase was evolving rapidly. It is effective in that
regard, but it carries ongoing costs.

---

## A-11-1: Silent failures on wrong key type

**Location:** `pkg/ctxutil/` (all key-lookup functions)

Context values in Go are retrieved by key type. A caller that passes the wrong
key type — or retrieves a value with the wrong expectation — gets the zero value
at runtime rather than a compile error. There is no tooling to enforce correct
key usage across packages.

**Assessment:** Low probability per call site, but the surface is large (30+
keys) and grows over time. New keys are easy to add; the incentive to add
another context key rather than thread an explicit parameter is strong.

**Recommendation:** Freeze — no new context keys should be introduced after
this ADR. New state that would previously have been a context key should be
passed as an explicit parameter or collected into a small config/state struct.

---

## A-11-2: Credentials in context

**Location:** `pkg/ctxutil/` — keys that carry passphrases or key identities

Credential-adjacent values passed through context carry a risk of accidental
exposure if a context is ever passed to a logging or serialisation call. The
current code does not log context contents, but the absence of a structural
barrier means this is a discipline constraint rather than an enforcement
constraint.

**Assessment:** No known current exposure. Risk is latent and grows with
codebase complexity.

**Recommendation:** Audit all ctxutil keys for credential sensitivity. For any
key that carries a passphrase, key ID, or token, replace context storage with
explicit passing to the narrowest function scope possible.

---

## A-11-3: Scope creep

**Location:** `pkg/ctxutil/` — file grows as new keys are added

The frictionless nature of adding a context key means the pattern expands
without review. Each new key increases the cognitive overhead of understanding
what state a given function implicitly depends on.

**Assessment:** Ongoing. Observed trend in git history.

**Recommendation:** Phase the reduction:

| Phase | Action | Effort |
|-------|--------|--------|
| 1 — Freeze | Add package-doc comment: "No new keys. Use explicit parameters." | Trivial |
| 2 — Extract display config | Collect terminal-display keys (width, color, verbose, auto-print) into a `DisplayConfig` struct passed explicitly through `Action` | Medium |
| 3 — Credential audit | Replace credential-adjacent keys with explicit scoped passing | Targeted |

Phase 1 alone is sufficient to stop the problem from growing. Phases 2 and 3
can be tackled incrementally without a flag day.
