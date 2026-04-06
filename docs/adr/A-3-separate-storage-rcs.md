# A-3: Separate `Storage` and `RCS` interfaces

**Status:** deferred — current implementation intentionally keeps them merged  
**Source:** CODE_QUALITY_REPORT.md § A-3

---

## Background

`backend.Storage` (defined in `internal/backend/storage.go`) currently embeds
the unexported `rcs` interface (defined in `internal/backend/rcs.go`). This
means every `Storage` implementation must satisfy ~15 VCS methods: `Add`,
`Commit`, `Push`, `Pull`, `TryAdd`, `TryCommit`, `TryPush`, `InitConfig`,
`AddRemote`, `RemoveRemote`, `Revisions`, `GetRevision`, `Status`, `Compact`.

The `fs` backend (`internal/backend/storage/fs/`) has no VCS support and
implements all of these as stubs that return `store.ErrGitNotInit` or
`backend.ErrNotSupported`.

This was a deliberate architectural choice: an earlier version of the codebase
had separate `Storage` and `RCS` interfaces, but merging them simplified the
leaf store significantly by eliminating a type-assertion at every VCS call site.
Since the overwhelming majority of storage backends (`gitfs`, `fossilfs`,
`jjfs`) are also RCS backends, the merged interface is **correct in practice**
even if it is not strictly correct in theory.

---

## Trade-offs of the current merged design

### Pros
- Single interface, no type-assertion boilerplate in `leaf.Store` or callers.
- Adding a new backend that supports VCS is straightforward — implement one
  interface and you're done.
- No layering confusion: `leaf.Store.storage` holds everything the store needs.

### Cons
- Pure-storage backends (`fs`) must implement ~15 no-op stubs.
- The `Storage` interface is larger than necessary, making it harder to mock in
  tests.
- Conceptually misleading: the interface promises VCS capabilities that some
  backends cannot deliver.

---

## Implementation plan (if splitting is chosen in the future)

### Phase 1 — Export `rcs` as a standalone `RCS` interface

**`internal/backend/rcs.go`**

- Rename the unexported `rcs` interface to exported `RCS`.
- Add a concrete `NopRCS` type in the same file that provides the same no-op
  behaviour currently in `fs/rcs.go`:
  - `TryAdd`, `TryCommit`, `TryPush`, `InitConfig`, `Compact` → `nil`
  - `Add`, `Commit`, `Push`, `Pull` → `store.ErrGitNotInit`
  - `Revisions` → single `{Hash:"latest", Date:time.Now()}` + `ErrNotSupported`
  - `GetRevision("HEAD"|"latest")` → delegates to a supplied getter func; other
    revisions → `ErrNotSupported`
  - `Status`, `AddRemote`, `RemoveRemote` → `ErrNotSupported`

**`internal/backend/storage.go`**

- Remove the `rcs` embed from `Storage` so it contains only file-operation
  methods: `Get`, `Set`, `Delete`, `Exists`, `Move`, `List`, `IsDir`, `Prune`,
  `Link`, `Name`, `Path`, `Version`, `Fsck`, `fmt.Stringer`.

### Phase 2 — Add an `rcs` field to `leaf.Store`

**`internal/store/leaf/store.go`**

- Add `rcs backend.RCS` field to the `Store` struct.
- In `leaf.Init()`, after the storage backend is created, derive `rcs` via a
  type assertion:

  ```go
  if r, ok := st.(backend.RCS); ok {
      s.rcs = r
  } else {
      s.rcs = backend.NopRCS{}
  }
  ```

- Update `leaf.GitInit()` in `internal/store/leaf/rcs.go` to re-derive `s.rcs`
  after reassigning `s.storage`.

**Leaf files that call RCS methods on `s.storage`** — change `s.storage.X` to
`s.rcs.X` for all VCS operations:

| File | Methods affected |
|---|---|
| `internal/store/leaf/write.go` | `TryAdd`, `TryCommit`, `TryPush` |
| `internal/store/leaf/move.go` | `TryAdd`, `TryCommit`, `TryPush` |
| `internal/store/leaf/link.go` | `Add` |
| `internal/store/leaf/fsck.go` | `Compact`, `Push`, `TryCommit` |
| `internal/store/leaf/reencrypt.go` | `TryAdd`, `TryCommit`, `TryPush` |
| `internal/store/leaf/recipients.go` | `TryAdd`, `TryCommit`, `Push`, `Add` |
| `internal/store/leaf/templates.go` | `TryAdd`, `Add` |
| `internal/store/leaf/rcs.go` | `Revisions`, `GetRevision`, `Status` |

### Phase 3 — Delete `fs/rcs.go`

`internal/backend/storage/fs/rcs.go` can be deleted in its entirety. The
`fs.Store` type no longer needs to implement `RCS`.

### Phase 4 — Verify other consumers

These locations hold a `backend.Storage` variable but call only file-op methods
and require no changes beyond a successful compile check:

- `internal/action/reorg.go` — `var storage backend.Storage`
- `internal/create/wizard.go` — `backend.Storage` parameter

The `StorageLoader` interface in `internal/backend/registry.go` returns
`Storage` — no change needed.

### Phase 5 — Update and add tests

- Delete `internal/backend/storage/fs/rcs_test.go` (tests for the removed
  stubs).
- Any mock `Storage` type in leaf or other test packages can stop implementing
  RCS methods; embed or compose `backend.NopRCS` instead.
- Add a small unit test for `NopRCS` in `internal/backend/rcs_test.go`.

---

## Files changed at a glance

| File | Action |
|---|---|
| `internal/backend/rcs.go` | Export `RCS`; add `NopRCS` |
| `internal/backend/storage.go` | Remove `rcs` embed from `Storage` |
| `internal/store/leaf/store.go` | Add `rcs backend.RCS` field + init logic |
| `internal/store/leaf/rcs.go` | Use `s.rcs`; re-derive in `GitInit` |
| `internal/store/leaf/{write,move,link,fsck,reencrypt,recipients,templates}.go` | `s.storage.X` → `s.rcs.X` for VCS calls |
| `internal/backend/storage/fs/rcs.go` | **Delete** |
| `internal/backend/storage/fs/rcs_test.go` | **Delete** |
| `internal/backend/rcs_test.go` | Add `NopRCS` tests |

`gitfs`, `fossilfs`, and `jjfs` require no changes — they already implement
the full `rcs` interface and will satisfy `backend.RCS` via the type assertion
with zero modification.
