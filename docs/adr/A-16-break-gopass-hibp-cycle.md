# A-16: Break the `gopass` ↔ `gopass-hibp` module dependency cycle

**Status:** proposed  
**Source:** Maintainer discussion — the `gopass` module depends on `gopass-hibp`, while the
`gopass-hibp` module depends on `gopass`.

---

## Background

`github.com/gopasspw/gopass` and `github.com/gopasspw/gopass-hibp` currently require each
other. Go tolerates this because the *package* import graph is still a DAG, but the *module*
graph contains a cycle. The cycle is small and well understood, but it makes releases and
dependency management awkward and it contradicts the component diagram in
[`docs/components.dot`](../components.dot), which already draws a standalone `pkg/hibp` node.

### Current facts

The cycle has two edges.

**Edge 1 — `gopass` consumes the HIBP client.** `internal/audit/audit.go` imports:

```go
"github.com/gopasspw/gopass-hibp/pkg/hibp/api"
"github.com/gopasspw/gopass-hibp/pkg/hibp/dump"
```

* `api.Lookup` is used by the online HIBPv2 validator (`audit.hibp-use-api`).
* `dump.New(...).LookupBatch` is used by the offline dump check (`audit.hibp-dump-file`).

These are the only two import sites in the module.

**Edge 2 — `gopass-hibp` consumes gopass.** The standalone CLI's `main.go`, `hibp.go` and
`pkg/hibp/*` import gopass packages:

| File in `gopass-hibp` | Imports from `gopass` |
| --- | --- |
| `pkg/hibp/api/client.go` | `pkg/debug` |
| `pkg/hibp/api/downloader.go` | `pkg/debug`, `pkg/ctxutil`, `pkg/fsutil`, `pkg/termio` |
| `pkg/hibp/dump/scanner.go` | `pkg/debug`, `pkg/fsutil` |
| `hibp.go`, `main.go` (repo root) | `pkg/gopass/api` |

### Constraints

* There must not be a module cycle between the two projects.
* The `gopass-hibp` binary must remain its own repository/binary; we do not want a second
  binary target inside the `gopass` repository.
* New external dependencies should be avoided (`AGENTS.md`, "Libraries and Frameworks").
* `pkg/gopass` and friends have a documented stability contract ([A-12](A-12-pkg-api-stability.md)).

### Why the cycle is a problem in practice

* **Release ordering is circular.** A new `pkg/hibp` feature needs a `gopass-hibp` release
  before `gopass` can consume it; a new `pkg/gopass` API needs a `gopass` release before
  `gopass-hibp` can build. Neither can be released first in isolation.
* **Tooling friction.** Dependency bots, `govulncheck`, `go mod tidy` diffs and licence
  scanning all report the cycle as a repeated, noisy pair of changes.
* **Architecture confusion.** It is unclear which module owns the HIBP client and which is the
  consumer.

---

## Options considered

### Option A — Extract a dependency-free `gopass-hibp-client` module

Move `pkg/hibp/api` and `pkg/hibp/dump` into a new repository/module
(`github.com/gopasspw/gopass-hibp-client`). Both `gopass` and `gopass-hibp` then depend on it,
and only `gopass-hibp` depends on `gopass`.

**Critical caveat:** this only breaks the cycle if the new client module has **no** gopass
dependencies. Today the client imports `pkg/debug`, `pkg/ctxutil`, `pkg/fsutil` and
`pkg/termio`. A straight copy would become `gopass → client → gopass` and the cycle would
remain. The client must therefore be decoupled:

* `debug.Log` → remove, or accept an injected `*slog.Logger`.
* `fsutil.IsFile` / `fsutil.IsDir` → `os.Stat`.
* `ctxutil.IsHidden(ctx)` → a plain `hidden bool` parameter (download only).
* `termio.NewProgressBar` → drop, or accept an injected progress callback.

That is roughly six call sites across three files — small, but it is a real API change.

#### Option A — Pros

* Cleanest layering: the HIBP domain code lives with the HIBP project, `gopass` stays a
  password manager.
* The client becomes reusable by any third party with zero gopass dependency.
* `gopass-hibp` remains a separate binary/repo (constraint satisfied).

#### Option A — Cons

* New repository, CI, goreleaser config, licence linting and release pipeline to maintain.
* Three-module release train (client → gopass, client → gopass-hibp) instead of two.
* Requires changing/removing behaviour that is currently gopass-specific (progress bar, hidden
  context handling), which slightly degrades the CLI UX unless a callback is threaded through.
* The client's public API now needs its own stability policy.

### Option B — Move the HIBP client library into the `gopass` module as `pkg/hibp`

Relocate `pkg/hibp/api` and `pkg/hibp/dump` into this repository under `pkg/hibp/`. `gopass`
then imports its own package; `gopass-hibp` imports `gopass/pkg/hibp/*` and
`gopass/pkg/gopass/api` as it already imports the latter. The `gopass-hibp` binary stays in its
own repository.

#### Option B — Pros

* Breaks the cycle with no new repository and no release-train expansion.
* No decoupling refactor: `pkg/debug`, `pkg/fsutil`, `pkg/ctxutil` and `pkg/termio` are in the
  same module, so the existing client code moves verbatim.
* Matches [`docs/components.dot`](../components.dot), which already draws `pkg/hibp`.
* `gopass-hibp` becomes a genuinely thin CLI; it already depends on `gopass` for `pkg/gopass/api`,
  so no new coupling is introduced.
* Only six files plus tests move; `internal/audit` only changes its import paths.

#### Option B — Cons

* `gopass` now hosts the HIBP client, which is arguably domain-specific to `gopass-hibp`.
* Changes to the client ship on the `gopass` release cadence; `gopass-hibp` must bump `gopass`
  to pick them up.
* The library must be added to the A-12 stability scope (or explicitly excluded).

### Option C — Keep the cycle

Make no change and document that mutual module requirements are required.

#### Option C — Pros

* Zero effort. Compiles and runs today via MVS.

#### Option C — Cons

* All the release-ordering and tooling problems above persist.
* Depends on Go continuing to tolerate cyclic module requirements indefinitely.

### Option D — Fold `gopass-hibp` into `gopass` as a subcommand

Move the CLI into `internal/action` (e.g. extend `gopass audit` with the `api`, `download` and
`dump` sub-actions) and retire the separate repository.

#### Option D — Pros

* Single repository, single binary, no cycle, no client library to publish.
* `gopass audit` already performs HIBP checks, so the functionality is a natural fit.

#### Option D — Cons

* Directly conflicts with the constraint of not maintaining multiple binaries in one repo —
  and with the historical decision to keep integrations standalone ([A-12](A-12-pkg-api-stability.md)).
* Loses the standalone `gopass-hibp` release cadence and its independent packaging.
* Removing the standalone binary is a CLI-visible break.

### Option E — Dependency inversion via interfaces

`gopass` defines a lookup/scanner interface and the implementation is injected at runtime from
`gopass-hibp`.

#### Option E — Pros

* Theoretically removes the compile-time dependency.

#### Option E — Cons

* Go links statically; there is no clean plugin boundary, and gopass cannot import an
  implementation it does not depend on. This adds indirection without removing the edge.
* Not recommended.

---

## Recommendation

**Adopt Option B**, with **Option A** kept as the documented fallback if the library is later
needed independently of gopass.

Rationale:

1. It breaks the cycle completely and permanently with no new repository or release train.
2. It is the least invasive change — the HIBP client code moves without any refactor because
   its helper dependencies are already in this module.
3. It aligns with the existing component diagram, which already shows `pkg/hibp` as a gopass
   package.
4. It satisfies the "no second binary in the gopass repository" constraint: only library code
   moves, the `gopass-hibp` binary stays where it is.

Option A remains attractive on layering grounds, but its non-trivial decoupling work and the
additional repository/release pipeline are not justified by the current maintenance capacity —
the same reasoning A-12 applied when deferring full module semver.

---

## Implementation plan (Option B)

1. **Move the library.** Relocate `pkg/hibp/api` and `pkg/hibp/dump` (including `_test.go`
   files) from `gopass-hibp` into `gopass/pkg/hibp/api` and `gopass/pkg/hibp/dump` without
   behavioural changes.
2. **Update consumers in `gopass`.** Change `internal/audit/audit.go` imports to
   `github.com/gopasspw/gopass/pkg/hibp/{api,dump}`. Add doc comments describing stability.
3. **Update `gopass-hibp`.** Point its imports at `gopass/pkg/hibp/*`; remove the now-empty
   `gopass-hibp/pkg` tree. `go.mod` keeps the single `gopass` requirement; delete the
   `gopass-hibp → gopass-hibp` self-reference that does not exist.
4. **Verify there is no residual cycle.** Run `go mod graph` in both repositories and confirm
   `gopass` no longer lists `gopass-hibp`.
5. **Documentation.** Update `docs/components.dot` if needed, `docs/hacking.md` (the API
   example pointer) and the A-12 stability table to include `pkg/hibp`.
6. **Release ordering.** Do a `gopass` minor release containing `pkg/hibp`, then release
   `gopass-hibp` with the bumped dependency. This is a one-off ordering cost; afterwards the
   train is linear.
7. **Validation.** `make test`, `make codequality`, `make test-integration` in both
   repositories; confirm the HIBP API and dump audit paths still work end to end.

### Migration checklist

* [ ] `gopass/pkg/hibp/api` and `gopass/pkg/hibp/dump` created
* [ ] `internal/audit/audit.go` imports updated
* [ ] `gopass-hibp` imports updated and `pkg/` removed
* [ ] `go mod graph` shows no `gopass → gopass-hibp` edge
* [ ] `docs/components.dot` and A-12 stability table updated
* [ ] both repositories pass `make test` and `make codequality`

---

## Consequences

* The module graph becomes acyclic, removing the release chicken-and-egg problem.
* `gopass-hibp` becomes a thin CLI over `pkg/gopass/api` and `pkg/hibp`.
* `pkg/hibp` joins the A-12 stability surface; changes there follow the same best-effort
  policy and `PKG-BREAK:` conventions as `pkg/gopass`.
* If the library later needs to be consumed outside gopass without pulling in the module,
  it can be extracted per Option A; the move in this record does not preclude that.

---

## Affected packages

`internal/audit`, new `pkg/hibp/api`, new `pkg/hibp/dump`; in `gopass-hibp`: `main.go`,
`hibp.go`, and removal of `pkg/hibp/`.
