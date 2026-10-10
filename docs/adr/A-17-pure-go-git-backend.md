# A-17: Pure-Go (go-git) RCS backend as an opt-in alternative to the git CLI

**Status:** proposed  
**Source:** Maintainer request — a pure-Go git backend removes the runtime
dependency on the `git` binary for cross-compiled and single-binary
distributions, and for Windows users without Git for Windows on `PATH`. The
whole point is a backend that builds without cgo and without the CLI.

---

## Background

`gitfs` (`internal/backend/storage/gitfs/`) is gopass's default and primary
storage/RCS backend. It keeps the encrypted secrets on disk and shells out to
the external `git` binary for every version-control operation: `init`, `clone`,
`add`, `commit`, `push`, `pull`, `log`, `show`, `status`, `gc`, `remote`, and
signed-commit handling. It also reads `.git/config` through
`github.com/gopasspw/gitconfig`.

That design buys full fidelity with the user's git installation — credential
helpers, `gpg-agent`, smartcards, SSH signing, hooks, `core.autocrlf` — at the
cost of a hard runtime dependency on a `git` executable being present and on
`PATH`. This is the single remaining external process dependency for the
default configuration (the `gpg`/`age` crypto backends are separate concerns
and are out of scope here).

A pure-Go git implementation makes it possible to ship a static binary that can
still version and sync a store. The project already avoids cgo to keep
cross-compilation feasible (`AGENTS.md`, "Libraries and Frameworks"); `git2go`
/ libgit2 is therefore not an option. `github.com/go-git/go-git/v5` is a pure-Go
implementation, is the de-facto standard (Keybase, Gitea, Pulumi), and is
released under Apache-2.0, which is on the allow-list in `.license-lint.yml`.

### Verified codebase facts this record relies on

These are the exact seams the implementation must plug into.

* **The storage interface.** `internal/backend/storage.go` defines
  `backend.Storage`, which embeds the unexported `backend.rcs`
  (`internal/backend/rcs.go`):

  ```go
  // internal/backend/rcs.go
  type rcs interface {
      Add(ctx context.Context, args ...string) error
      Commit(ctx context.Context, msg string) error
      Push(ctx context.Context, remote, location string) error
      Pull(ctx context.Context, remote, location string) error

      TryAdd(ctx context.Context, args ...string) error
      TryCommit(ctx context.Context, msg string) error
      TryPush(ctx context.Context, remote, location string) error

      InitConfig(ctx context.Context, name, email string) error
      AddRemote(ctx context.Context, remote, location string) error
      RemoveRemote(ctx context.Context, remote string) error

      Revisions(ctx context.Context, name string) ([]Revision, error)
      GetRevision(ctx context.Context, name, revision string) ([]byte, error)

      Status(ctx context.Context) ([]byte, error)
      Compact(ctx context.Context) error
  }
  ```

  `backend.Storage` adds the file operations (`Get`, `Set`, `Delete`, `Exists`,
  `Move`, `List`, `IsDir`, `Prune`, `Link`), plus `Name()`, `Path()`,
  `Version(context.Context) semver.Version`, `Fsck(context.Context) error`, and
  `fmt.Stringer`. **This interface contract does not change.** The new backend
  implements it as-is.

* **`LinkTarget` is optional.** `store/leaf/link_target.go` discovers it by type
  assertion, not through the `Storage` interface. The new backend should
  implement it (delegating to `fs.Store`) so symlinked secrets keep working.

* **The registry.** `internal/backend/registry.go` exposes
  `backend.StorageRegistry`, a `Registry[StorageBackend, StorageLoader]`.
  A backend registers itself from an `init()` via
  `StorageRegistry.Register(id, name, &loader{})`. The loader must satisfy
  `backend.StorageLoader`:

  ```go
  type StorageLoader interface {
      fmt.Stringer
      Prioritized                 // Priority() int
      New(context.Context, string) (Storage, error)
      Init(context.Context, string) (Storage, error)
      Clone(context.Context, string, string) (Storage, error)
      Handles(context.Context, string) error
  }
  ```

  A loader may additionally implement `Commands(...)` to contribute CLI
  commands (see `internal/action/commands.go` `storeCommander`), and `Open`
  (used by `gitfs`).

* **Backend identifiers.** `StorageBackend` is an `int` enum in
  `internal/backend/storage.go`: `FS, GitFS, FossilFS, JJFS, CryptFS`. The
  `String()` name is resolved through the registry, so enum order is not
  persisted anywhere. A new constant is appended.

* **Backend registration happens through blank imports.** Each backend has a
  sibling file in the parent package, e.g.
  `internal/backend/storage/gitfs.go`:

  ```go
  package storage
  import _ "github.com/gopasspw/gopass/internal/backend/storage/gitfs" // register gitfs backend
  ```

  The whole set is pulled in by `pkg/gopass/api/api.go`
  (`_ "github.com/gopasspw/gopass/internal/backend/storage"`).

* **Selection precedence** (`backend.DetectStorage`, `internal/backend/storage.go`):

  1. A backend explicitly set in the context (`backend.WithStorageBackend`,
     e.g. from the `--storage` flag or `gopass init`). If `New` fails it falls
     back to `FS`.
  2. The config key `storage.backend` (`config.String(ctx, "storage.backend")`,
     documented in `docs/config.md`). If it fails, detection continues.
  3. Auto-detection: iterate `StorageRegistry.Prioritized()` (ascending
     `Priority()`) calling each loader's `Handles(ctx, path)`.
  4. Fallback `FS`.

  Existing priorities: `gitfs` = 11, `fs` = 50, `jjfs` = 15.

* **`Init`/`Clone` wiring.** `internal/store/leaf/store.go` `Init` and
  `internal/store/leaf/storage.go` `initStorageBackend` call
  `backend.InitStorage` / `backend.DetectStorage`. `internal/action/init.go`
  `initParseContext` defaults the context to `backend.GitFS` when no
  `--storage` flag was given. `internal/action/rcs.go` `RCSInit` does the same.

* **Config/environment plumbing.** Read config with
  `config.Bool/String/Int(ctx, key)` (`internal/config/utils.go`). The
  `docs/config.md` options table and `GOPASS_*` env table are policed by
  `TestConfigOptsInDocs` and `TestEnvVarsInDocs`
  (`internal/config/docs_test.go`): **any new `os.Getenv("GOPASS_…")` or
  `config.<T>(ctx, "x.y")` call that is not in `docs/config.md` fails the unit
  tests.** New env vars must be documented (or added to `ignoredEnvs`).

* **The git-CLI sub-command is contributed by the `gitfs` loader**
  (`internal/backend/storage/gitfs/commands.go`), and because the registry
  iterates *all* loaders, `gopass git` stays available regardless of which
  git-family backend is selected — provided a `.git` directory exists (which it
  does for both backends). No change required.

* **`doctor`** (`internal/action/doctor.go`) has `doctorCheckGit`, which fails
  when a store uses the git backend but no `git` binary is found. This check
  must become aware of the pure-Go backend.

* **Errors.** `internal/store/err.go` defines `ErrGitNotInit`,
  `ErrGitNoRemote`, `ErrGitNothingToCommit`. The new backend returns the same
  sentinels so `store/leaf` and `internal/action` error handling is unchanged.

* **The `fs` backend supplies the file layer.** `gitfs.Git` embeds `*fs.Store`
  for all file operations (`internal/backend/storage/gitfs/storage.go`). The new
  backend reuses `fs.Store` unchanged, so `Get`/`Set`/`Delete`/`List`/`Move`/
  symlink handling and the `safePath` traversal guard
  (`internal/backend/storage/fs/store.go`) are identical.

---

## Decision

Add a new, **additive** storage/RCS backend package
`internal/backend/storage/gogit` that implements `backend.Storage` using
`github.com/go-git/go-git/v5`. The existing `gitfs` (CLI) backend is kept
exactly as-is; the new backend is opt-in and becomes the automatic fallback
when no `git` binary is available.

Concretely:

1. **Implement behind the existing interface.** The new `GoGit` type embeds
   `*fs.Store` for file operations and adds a `*git.Repository` for RCS
   operations. `backend.Storage` and the unexported `rcs` interface are not
   touched.
2. **Register** the backend as `gogit`/`GoGit` with a priority *higher* (later)
   than `gitfs`, so the CLI backend remains the auto-detected default when
   `git` is present.
3. **Select** via the existing `storage.backend` config key (value `gogit`), a
   dedicated environment variable `GOPASS_GIT_BACKEND=gogit`, or the
   `--storage=gogit` flag. Precedence follows the existing scheme.
4. **Auto-fallback** to `gogit` whenever the `git` binary is not on `PATH`.
   This is implemented by making the `gitfs` loader fail `Handles`/`New` with a
   sentinel error when `exec.LookPath("git")` fails, plus a resolver used on the
   context-locked path (which currently falls back to `FS`). Result: on a
   machine without `git`, a `.git`-backed store transparently uses go-git
   instead of degrading to the unversioned `fs` backend.
5. **Implement the RCS surface natively** for `init`, `clone`, `add`, `commit`
   (unsigned), `status`, `log`, `fetch`, `push`, fast-forward `pull`, and
   `remote add/remove/list`.
6. **Never use `Repository.Merge`.** go-git v5 only implements
   `FastForwardMerge` (no three-way merge, no merge commit, no rebase). Instead
   perform a **file-level merge** and create a **two-parent commit** via
   `git.CommitOptions{Parents: …}`, preserving valid history for anyone later
   using the real git CLI.
7. **Never silently discard a secret.** On a true content conflict keep the
   local version and materialise the remote version as
   `<name>.conflict-<shortsha>`. Losing a password to automatic resolution is
   strictly worse than leaving two files to reconcile by hand.
8. **Shell out (or document as unsupported) for the operations go-git cannot
   replicate faithfully:** GPG/SSH-signed commits, git credential helpers, and
   git hooks. See "Feature parity matrix" below.
9. **Isolate every transport/auth call behind a single file** so a future bump
   to go-git `/v6` (currently alpha, rewritten transport layer) is a contained
   change.
10. **Do not make it the default** until the signing and credential-helper
    stories are solved. This record explicitly ships it opt-in first.

---

## Options considered

### Option A — `git2go` / libgit2 bindings

Rejected. Requires cgo, breaks static builds and cross-compilation, and pulls in
a C toolchain. This defeats the primary goal.

### Option B — Keep shelling out, only bundle a `git` binary

Rejected. Bundling git does not remove the process-dependency or the platform
matrix; it makes the release artifact enormous and licensing/attribution
awkward.

### Option C — `go-git/v6`

Rejected for now. `/v6` is alpha with a rewritten transport layer. Building on
`v5` (stable, `v5.19.x`) and isolating transport calls is the lower-risk path;
the design must make the later bump mechanical.

### Option D (chosen) — `go-git/v5`, additive, opt-in, CLI fallback

Chosen. Pure Go, no cgo, Apache-2.0, and a clean fit against the existing
`gitfs`/`fs` split. The CLI backend stays the default; the pure-Go backend
covers the no-git-binary case and can be selected explicitly.

---

## Detailed design

### Package layout

```text
internal/backend/storage/gogit/
  gogit.go     // GoGit struct, New/Init/Clone, open + Version + IsInitialized
  loader.go    // init() Registration, StorageLoader impl, Priority, Handles
  storage.go   // file ops delegation to *fs.Store (mirror gitfs/storage.go)
  rcs.go       // Add/TryAdd/Commit/TryCommit/Push/TryPush/Pull/Status/Compact
  revisions.go // Revisions, GetRevision (log/show equivalents)
  merge.go     // file-level merge + two-parent commit + conflict policy
  auth.go      // ALL transport/auth: SSH agent/keys, HTTPS creds, proxy. v6 seam.
  config.go    // local config reading/writing (user.name/email, remotes, fixConfig)
  signing.go   // optional GPG signer (shell-out) + gpg.format detection
internal/backend/storage/gogit.go   // blank import to register the backend
```

Add a blank-import aggregator file exactly like the existing ones:

```go
// internal/backend/storage/gogit.go
package storage

import _ "github.com/gopasspw/gopass/internal/backend/storage/gogit" // register gogit backend
```

### Backend identifier

Append a constant to the enum in `internal/backend/storage.go` (append at the
end to avoid perturbing existing `iota` values, even though only the registry
name is observed):

```go
const (
    FS StorageBackend = iota
    GitFS
    FossilFS
    JJFS
    CryptFS
    // GoGit is a filesystem-backed storage with Git using the pure-Go go-git
    // implementation instead of the git binary.
    GoGit
)
```

### Loader

```go
// internal/backend/storage/gogit/loader.go
package gogit

const name = "gogit"

func init() {
    backend.StorageRegistry.Register(backend.GoGit, name, &loader{})
}

type loader struct{}

func (l loader) String() string { return name }

// Priority must be *higher* (later) than gitfs (11) so that, when both the git
// binary and go-git are available, auto-detection still prefers the CLI
// backend. Auto-fallback to go-git is driven by gitfs.Handles failing when git
// is absent, not by priority alone.
func (l loader) Priority() int { return 12 }

func (l loader) New(ctx context.Context, path string) (backend.Storage, error) {
    return New(path)
}

func (l loader) Init(ctx context.Context, path string) (backend.Storage, error) {
    return Init(ctx, path, termio.DetectName(ctx, nil), termio.DetectEmail(ctx, nil))
}

func (l loader) Clone(ctx context.Context, repo, path string) (backend.Storage, error) {
    return Clone(ctx, repo, path, termio.DetectName(ctx, nil), termio.DetectEmail(ctx, nil))
}

// Handles returns nil iff path contains a .git directory. Unlike gitfs, it does
// NOT require a git binary.
func (l loader) Handles(_ context.Context, path string) error {
    path = fsutil.ExpandHomedir(path)
    if !fsutil.IsDir(filepath.Join(path, ".git")) {
        return fmt.Errorf("no .git at %s", path)
    }
    return nil
}
```

The gogit loader deliberately does **not** implement `storeCommander`; the
`gopass git` passthrough command continues to be contributed by the `gitfs`
loader and keeps working for any store with a `.git` directory.

### Core type

```go
// internal/backend/storage/gogit/gogit.go
package gogit

type GoGit struct {
    fs   *fs.Store
    repo *git.Repository
}

var _ backend.Storage = (*GoGit)(nil) // compile-time interface check

func New(path string) (*GoGit, error) {
    path = fsutil.ExpandHomedir(path)
    repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: false})
    if err != nil {
        return nil, fmt.Errorf("no git repository at %s: %w", path, err)
    }
    return &GoGit{fs: fs.New(path), repo: repo}, nil
}
```

`Init` mirrors `gitfs.Init`:

```go
func Init(ctx context.Context, path, userName, userEmail string) (*GoGit, error) {
    path = fsutil.ExpandHomedir(path)

    var repo *git.Repository
    if fsutil.IsDir(filepath.Join(path, ".git")) {
        r, err := git.PlainOpen(path)
        if err != nil {
            return nil, err
        }
        repo = r
    } else {
        r, err := git.PlainInitWithOptions(path, &git.PlainInitOptions{
            InitOptions: git.InitOptions{DefaultBranch: plumbing.NewBranchReferenceName("master")},
            Bare:        false,
        })
        if err != nil {
            return nil, fmt.Errorf("failed to init git repo: %w", err)
        }
        repo = r
        out.Printf(ctx, "git (go-git) initialized at %s", path)
    }

    g := &GoGit{fs: fs.New(path), repo: repo}

    if !ctxutil.IsGitInit(ctx) {
        return g, nil
    }
    if err := g.InitConfig(ctx, userName, userEmail); err != nil {
        return g, fmt.Errorf("failed to configure git: %w", err)
    }

    // stage and commit the current content
    if err := g.Add(ctx, path); err != nil {
        return g, fmt.Errorf("failed to add %q to git: %w", path, err)
    }
    if !g.HasStagedChanges(ctx) {
        return g, nil
    }
    if ctxutil.HasSetupRemote(ctx) {
        return g, nil
    }
    if err := g.Commit(ctx, "Add current content of password store"); err != nil {
        return g, fmt.Errorf("failed to commit: %w", err)
    }
    return g, nil
}
```

`Clone` uses `git.PlainCloneContext` with the auth resolved in `auth.go` (see
below) and then `InitConfig`.

`Version` returns the go-git module version via `debug.ModuleVersion` (the CLI
backend parses `git version`; there is no such string here). `String()` returns
`gogit(<version>,path:<path>)`. `Path()` and `Name()` return the store path and
`"gogit"` respectively. `IsInitialized` checks `.git/config` like `gitfs` for
behavioural parity with `gopass rcs init`.

### File operations

`storage.go` is a verbatim mirror of `gitfs/storage.go`: `Get`, `Set`,
`Delete`, `Exists`, `List`, `IsDir`, `Prune`, `Link`, `Move`, and `LinkTarget`
all delegate to `g.fs`. `Fsck` calls the `fs` Fsck and then commits any
untracked files (see `addUntrackedFiles` equivalent below).

> **Important:** file writes bypass go-git. They go straight to disk via
> `fs.Store`. This is intentional and matches `gitfs`. The worktree index is
> updated on the next `Add`, which reads the on-disk bytes. Do **not** route
> `Set`/`Delete` through `worktree` — the `fs` layer is the single source of
> truth for file content, symlink safety, and the `ErrMeaninglessWrite` guard.

### RCS operation mapping

| `backend.rcs` method | go-git implementation |
| --- | --- |
| `Add(ctx, files...)` | `wt.AddWithOptions(&git.AddOptions{All: true, Path: rel})` per path (or `All:true` with no path to stage everything). Normalise absolute paths to repo-relative first, exactly as `gitfs.Add` trims the store prefix. Return `store.ErrGitNotInit` if not initialised. |
| `TryAdd` | `Add`; swallow `ErrGitNotInit`. |
| `HasStagedChanges` | helper: `wt.Status()`; true if any entry has a non-`Unmodified` `Staging` code, or (`All` semantics) any change at all. |
| `Commit(ctx, msg)` | `wt.Commit(msg, &git.CommitOptions{Author: sig, Committer: sig, AllowEmptyCommits: false})`. Honours `ctxutil.GetCommitTimestamp` by setting `sig.When`. Returns `store.ErrGitNothingToCommit` on `git.ErrEmptyCommit`; `store.ErrGitNotInit` if uninitialised. If signing is configured see "Signing". |
| `TryCommit` | `Commit`; swallow `ErrGitNothingToCommit`, `ErrGitNotInit`. |
| `Push(ctx, remote, branch)` | Respect `ctxutil.IsNoNetwork` (return nil). Resolve remote/branch via `config.go` (`defaultRemote`, `defaultBranch`). `repo.PushContext(ctx, &git.PushOptions{RemoteName: remote, RefSpecs: …})`. Map "already up to date" (`git.NoErrAlreadyUpToDate`) to nil. Return `store.ErrGitNoRemote` when the remote/config is missing. |
| `TryPush` | `Push`; swallow `ErrGitNotInit`, `ErrGitNoRemote`. |
| `Pull(ctx, remote, branch)` | Respect `IsNoNetwork`. Fetch, then fast-forward if possible, else **file-level merge** (see below). |
| `AddRemote` | `repo.CreateRemote(&config.RemoteConfig{Name: remote, URLs: []string{url}})`. |
| `RemoveRemote` | `repo.DeleteRemote(remote)`. |
| `Revisions(ctx, name)` | `repo.Log(&git.LogOptions{From: head, Order: git.LogOrderCommitterTime, PathFilter: exactPath(name)})`, map each `*object.Commit` to `backend.Revision{Hash, AuthorName, AuthorEmail, Date: c.Author.When, Subject: firstLine, Body: rest}`. |
| `GetRevision(ctx, name, rev)` | `h, _ := repo.ResolveRevision(plumbing.Revision(rev))`; `c, _ := repo.CommitObject(*h)`; `f, _ := c.File(name)`; `f.Contents()`. Handle `"latest"`/`"HEAD"` by resolving `HEAD` (the `fs` backend treats these specially; keep that behaviour). |
| `Status(ctx)` | `wt.Status()` formatted deterministically (see "Status output"). |
| `Compact(ctx)` | `repo.RepackObjects(&git.RepackConfig{})`, then `repo.Prune(git.PruneOptions{})` with a handler that never errors. Best-effort: log and return nil on unsupported errors. |

`InitConfig(ctx, name, email)` writes `user.name` / `user.email` into the local
`.git/config` (via `repo.Config()` + `repo.SetConfig`, or `repo.ConfigScoped(config.LocalScope)`),
applies the same "sane config" defaults as `gitfs.fixConfig` where they map to
go-git (`push.default=matching`, `pull.rebase=false`, `diff.gpg.binary=true`,
`diff.gpg.textconv=gpg --no-tty --decrypt`), and writes the `.gitattributes`
file, then stages+commits it — reusing the existing
`gitattributesForBackend` logic. The `core.sshCommand` default is **not** set by
this backend: go-git does not read `core.sshCommand`, and setting it would only
matter if the user later switches back to the CLI backend. Leave it unset to
avoid surprising the CLI backend; document the difference.

`defaultRemote`/`defaultBranch`/`HasStagedChanges`/`ListUntrackedFiles`
helpers are reimplemented on top of go-git (not config-file string parsing):
`defaultBranch` from `repo.Head().Name()`, defaulting to `"master"` (go-git's
default and gopass's historical default) when `HEAD` is unborn; `defaultRemote`
from `branch.<name>.remote` in the local config, defaulting to `origin`.

### Selection wiring

**Config key (primary).** Reuse the existing generic key — no new option is
required:

```ini
storage.backend = gogit
```

`backend.DetectStorage` already calls `detectStorageByName(ctx, "gogit", path)`,
which resolves the name through `StorageRegistry.Backend(name)` and calls the
loader. **No change to `DetectStorage` is needed for this path.** Document
`gogit` in the valid-values list of the `storage.backend` row in
`docs/config.md`.

**Environment variable (convenience/override).** Add
`GOPASS_GIT_BACKEND` with values `gitfs` (alias `cli`, `cmd`) and `gogit`.
Precedence: explicit `--storage` flag > `GOPASS_GIT_BACKEND` > `storage.backend`
> auto-detect. Implement this in one place — a small helper in
`internal/action/init.go`/`rcs.go` and in `backend.DetectStorage`:

```go
// internal/backend/storage.go (new helper, used by DetectStorage and callers)
//
// applyGitBackendEnv returns the backend named by GOPASS_GIT_BACKEND, if any.
func gitBackendFromEnv() (StorageBackend, bool) {
    switch strings.ToLower(os.Getenv("GOPASS_GIT_BACKEND")) {
    case "gogit":
        return GoGit, true
    case "gitfs", "cli", "cmd":
        return GitFS, true
    default:
        return FS, false
    }
}
```

Because the env var is read with `os.Getenv("GOPASS_GIT_BACKEND")`, the
implementer **must** add a row to the `GOPASS_*` table in `docs/config.md`
(`GOPASS_GIT_BACKEND | string | Select the git backend: gitfs (default) or gogit …`)
or `TestEnvVarsInDocs` fails.

**Auto-fallback when no `git` binary is present.** Two mechanisms, both needed:

1. Make `gitfs` report itself unusable without git. In
   `internal/backend/storage/gitfs/loader.go`:

   ```go
   var ErrGitBinaryNotFound = errors.New("git binary not found in PATH")

   func (l loader) Handles(ctx context.Context, path string) error {
       path = fsutil.ExpandHomedir(path)
       if !fsutil.IsDir(filepath.Join(path, ".git")) {
           return fmt.Errorf("no .git at %s", path)
       }
       if _, err := exec.LookPath("git"); err != nil {
           return ErrGitBinaryNotFound
       }
       return nil
   }
   ```

   and return the same sentinel from `New`/`Init` so the config-locked path
   (`detectStorageByName`) also fails and lets detection continue.

2. Add a resolver for the **context-locked** path, which today falls back to
   `FS` on failure. In `internal/backend/storage.go`, before constructing a
   requested `GitFS` backend, downgrade to `GoGit` if git is absent:

   ```go
   func ResolveStorageBackend(ctx context.Context, requested StorageBackend) StorageBackend {
       if requested != GitFS {
           return requested
       }
       if _, err := exec.LookPath("git"); err != nil {
           if _, rerr := StorageRegistry.Get(GoGit); rerr == nil {
               debug.Log("git binary not found; falling back to pure-Go go-git backend")
               return GoGit
           }
       }
       return requested
   }
   ```

   Call it in `DetectStorage` (both the context-locked and the default branch)
   and in `internal/action/init.go` `initParseContext` where the default is set
   to `GitFS`. This keeps `gopass setup` on a git-less machine working with a
   versioned store rather than silently degrading to `FS`.

**Precedence with an explicit `--storage=gitfs`.** If the user *explicitly*
named `gitfs` and git is missing, the resolver still downgrades to `gogit`
(availability wins). If the user explicitly named `gogit`, it is used directly
and `gitfs` is never attempted. Emit a one-line `out.Warningf` when a fallback
happens so the user is not surprised.

### Merge — file-level, not content-level

go-git v5 offers only `MergeStrategy = FastForwardMerge` (and `repo.Merge`
errors with `ErrFastForwardMergeNotPossible` otherwise). There is no three-way
merge, no merge-commit construction helper, and no rebase. **Do not call
`repo.Merge`.** Implement the merge directly:

```go
// internal/backend/storage/gogit/merge.go
//
// mergeRemote performs a file-level merge of remoteHash into the current HEAD.
// It is correct for gopass because the files are encrypted blobs: a
// content-level 3-way merge is meaningless, so we resolve per path.
func (g *GoGit) mergeRemote(ctx context.Context, remoteHash plumbing.Hash) error {
    wt, err := g.repo.Worktree()
    if err != nil { return err }

    head, err := g.repo.Head()
    if err != nil { return err }
    localHash := head.Hash()

    localCommit, err := g.repo.CommitObject(localHash)
    if err != nil { return err }
    remoteCommit, err := g.repo.CommitObject(remoteHash)
    if err != nil { return err }

    // 1. best common ancestor
    bases, err := localCommit.MergeBase(remoteCommit)
    if err != nil { return err }
    // A missing base (unrelated histories) is treated as an empty tree.
    var baseTree *object.Tree
    if len(bases) > 0 {
        baseTree, err = bases[0].Tree()
        if err != nil { return err }
    }
    localTree, err := localCommit.Tree()
    if err != nil { return err }
    remoteTree, err := remoteCommit.Tree()
    if err != nil { return err }

    // 2. per-side changes vs base
    localChanges  := diffOrEmpty(baseTree, localTree)   // object.Changes
    remoteChanges := diffOrEmpty(baseTree, remoteTree)

    // 3. per-path resolution
    conflicts := []string{}
    for path := range unionPaths(localChanges, remoteChanges) {
        lc := changeFor(localChanges, path)   // may be nil
        rc := changeFor(remoteChanges, path)  // may be nil
        switch {
        case lc != nil && rc == nil:
            // changed only locally -> keep local (already on disk)
            // (if the local change is a delete, remove it from the index)
        case rc != nil && lc == nil:
            // changed only remotely -> apply remote content to worktree + index
            applyRemoteChange(ctx, wt, remoteTree, rc)
        case sameBlob(lc, rc):
            // changed both sides identically -> keep either; no-op
        case isDelete(lc) != isDelete(rc):
            // delete vs modify -> KEEP THE MODIFICATION (never drop a secret)
            keepModification(ctx, wt, remoteTree, lc, rc)
        default:
            // changed both sides, different content -> CONFLICT
            // keep local on disk, materialise remote as <name>.conflict-<shortsha>
            materializeConflict(ctx, wt, remoteTree, path, remoteHash)
            conflicts = append(conflicts, path)
        }
    }

    if len(conflicts) > 0 {
        out.Warningf(ctx, "Merge conflict in %d file(s); the remote versions were kept as *.conflict-<sha>:",
            len(conflicts))
        for _, p := range conflicts {
            out.Warningf(ctx, "  - %s", p)
        }
    }

    // 4. two-parent merge commit so real git users see valid history
    msg := fmt.Sprintf("Merge remote-tracking branch into %s (gopass go-git backend)", head.Name().Short())
    _, err = wt.Commit(msg, &git.CommitOptions{
        Author:    g.commitSignature(ctx),
        Committer: g.commitSignature(ctx),
        Parents:   []plumbing.Hash{localHash, remoteHash},
    })
    return err
}
```

Helper semantics:

* `diffOrEmpty(base, tree)` — if `base == nil`, use `object.DiffTree` against an
  empty tree (all files are additions on that side). `object.DiffTree(a, b)`
  returns `object.Changes` (each `Change` carries `From`/`To` `ChangeEntry`
  with `Name`, `Mode`, `Hash`).
* `applyRemoteChange` — for an add/modify, read the blob from `remoteTree`
  (`remoteTree.File(path)` → `Contents()`), write it into the worktree via
  `fs.Store.Set`, then `wt.Add(path)`. For a remote delete, `fs.Store.Delete`
  then `wt.Remove(path)`.
* `keepModification` — if local modified and remote deleted, leave local in
  place and `wt.Add`. If local deleted and remote modified, restore the remote
  content (the deletion must not win). **A secret is never dropped.**
* `sameBlob(lc, rc)` — compare `ChangeEntry.TreeEntry.Hash`.
* `materializeConflict` — leave the local file untouched; write the remote blob
  to `<dir>/<base>.conflict-<shortsha>` (shortsha = first 7 hex chars of
  `remoteHash`; use `path` with the extension preserved before the suffix, e.g.
  `secrets/api.gpg` → `secrets/api.conflict-1a2b3c4.gpg`), then `wt.Add` the new
  path. The conflict file *is* part of the merge commit tree, so the remote
  content is preserved in history even if the user deletes the file.

**Why two parents matter.** Later, a teammate may run the real `git` CLI. A
two-parent merge commit keeps the DAG connected: `git log`, `git merge-base`,
and future CLI pushes/pulls behave correctly, instead of the remote appearing as
an unexplained forced update. Tests must assert `NumParents() == 2`.

**Alternative conflict policy (non-default).** A `gogit.conflict=local|remote|newest`
(or `--prefer-newest`) option may be offered later, resolving true conflicts by
`Committer.When`. It **must not** be the default: "newest ciphertext" is not
meaningful for confidentiality, and silently discarding the other side can lose a
password. Document the default ("never drop a secret; keep both") loudly.

### Pull flow

```go
func (g *GoGit) Pull(ctx context.Context, remote, branch string) error {
    if ctxutil.IsNoNetwork(ctx) { return nil }
    remote, branch = g.resolveRemoteBranch(ctx, remote, branch)

    if err := g.repo.FetchContext(ctx, &git.FetchOptions{RemoteName: remote, Auth: g.auth(ctx, remote), Force: false}); err != nil &&
        !errors.Is(err, git.NoErrAlreadyUpToDate) {
        return err
    }

    remoteRef, err := g.repo.Reference(plumbing.NewRemoteReferenceName(remote, branch), true)
    if err != nil { return err }
    remoteHash := remoteRef.Hash()

    head, err := g.repo.Head()
    if err != nil { return err }
    if head.Hash() == remoteHash { return nil } // already up to date

    localCommit, _ := g.repo.CommitObject(head.Hash())
    remoteCommit, _ := g.repo.CommitObject(remoteHash)
    // fast-forward when local is an ancestor of remote
    if ff, err := localCommit.IsAncestor(remoteCommit); err == nil && ff {
        wt, _ := g.repo.Worktree()
        return wt.PullContext(ctx, &git.PullOptions{
            RemoteName: remote, ReferenceName: head.Name(), Auth: g.auth(ctx, remote),
        })
    }
    return g.mergeRemote(ctx, remoteHash)
}
```

### Push flow

Fetch first (to report "already up to date" and to surface divergence), warn on
untracked files exactly like `gitfs.PushPull`, then
`repo.PushContext(ctx, &git.PushOptions{RemoteName: remote, RefSpecs: …,
Auth: g.auth(ctx, remote)})`. A non-fast-forward push returns
`git.ErrNonFastForwardUpdate`; surface it verbatim so the user can run
`gopass sync` (pull/merge) first. Do **not** force-push.

### Transport and authentication (`auth.go` — the v6 seam)

All network/auth concerns live in exactly one file, `auth.go`. Every call to
`FetchOptions.Auth`, `PushOptions.Auth`, `PullOptions.Auth`, and `CloneOptions.Auth`
must go through `g.auth(ctx, remote)`. This is the file to rewrite for `/v6`.

```go
func (g *GoGit) auth(ctx context.Context, remote string) transport.AuthMethod {
    url := g.remoteURL(remote)
    switch {
    case strings.HasPrefix(url, "ssh://"), strings.Contains(url, "@"), strings.HasPrefix(url, "git@"):
        // SSH: native via golang.org/x/crypto/ssh + ssh-agent.
        if a, err := ssh.NewSSHAgentAuth(userFromURL(url)); err == nil {
            return a
        }
        // fall back to a key file from ~/.ssh (respect GIT_SSH_KEY / identity)
        if a, err := ssh.NewPublicKeysFromFile(userFromURL(url), sshKeyPath(), ""); err == nil {
            return a
        }
    case strings.HasPrefix(url, "https://"), strings.HasPrefix(url, "http://"):
        // HTTPS: go-git ignores credential.helper. See "Credentials".
        return httpsAuth(url)
    }
    return nil // local file paths need no auth
}
```

**SSH is fully supported** (`golang.org/x/crypto/ssh` is already a direct
dependency in `go.mod`): use `ssh.NewSSHAgentAuth` (ssh-agent, honouring
`SSH_AUTH_SOCK`), else `ssh.NewPublicKeysFromFile` with a key discovered from
`core.sshCommand`-like env or `~/.ssh/id_*`.

**Credential helpers are not supported by go-git.** `credential.helper`,
`osxkeychain`, `wincred`, and Git Credential Manager are ignored by go-git. Two
supported paths:

1. If a `git` binary is available, shell out to `git credential fill`
   (`printf 'protocol=…\nhost=…\n\n' | git credential fill`) with a timeout, parse
   `username=`/`password=`, and return `http.NewBasicAuth(user, pass)` (or
   `http.NewTokenAuth(pass)` when the password is an OAuth/token secret and
   username is empty). Never log the secret.
2. If no `git` is available, read the URL userinfo (`https://user:pass@host/…`)
   or `GIT_ASKPASS`/`.netrc`. Otherwise fail with a clear message:
   *"this store uses HTTPS with a credential helper; install git or configure
   the go-git backend's auth explicitly"*.

This asymmetry is a documented limitation and one of the reasons the backend is
not made default yet.

### Signing (`signing.go`)

go-git signs commits/tags with an in-process `*openpgp.Entity` (`SignKey`) or a
`Signer` (v5.12+). It does **not** use `gpg`, `gpg-agent`, pinentry, smartcards,
or `gpg.format=ssh`, and it cannot use a passphrase-protected agent key without
the user supplying the key material.

Policy:

* **Default: unsigned commits.** This matches go-git's native capability and is
  why the backend ships opt-in.
* If the local config has `commit.gpgsign=true`, attempt to honour it by shelling
  out to the git CLI for **that commit only** (hybrid mode): stage via go-git,
  then run `git commit -S -m <msg>` (honouring `gpg.program`, `user.signingkey`).
  This requires a `git` binary; if absent, emit a warning and create an unsigned
  commit rather than failing the user's write.
* Alternatively, implement `github.com/go-git/go-git/v5.Signer` that pipes the
  object payload to `gpg --status-fd=2 -bsau <key>`. Prefer the shell-out to
  `git commit` for correctness; a `Signer` shelling to `gpg` is acceptable but
  must handle `gpg.format=ssh` by delegating to the CLI backend or warning.
* `gpg.format=ssh` signing is **unsupported** by the pure-Go backend. Detect it
  and warn: *"SSH commit signing is not supported by the gogit backend; use the
  gitfs backend for signed commits."*

Document this prominently in `docs/backends/gogit.md`.

### Hooks

go-git never runs git hooks (`pre-commit`, `commit-msg`, `post-merge`,
`post-checkout`, …). gopass's own hook system (`core.pre-hook`/`core.post-hook`,
`internal/hook`) is independent of git and continues to work unchanged. Document
that git hooks are not executed by this backend. The credential/signing
shell-outs are the *only* places the backend may invoke an external binary, and
only when explicitly configured.

### CRLF, file mode, and index divergence

go-git stores worktree bytes as-is and does not apply `core.autocrlf` the way
the CLI does. For encrypted blobs (`.gpg`/`.age`) this is irrelevant, but text
files in a store (`.gitattributes`, `.public-keys/*`, config) could diverge.
Rules:

* Always write the `.gitattributes` file on init (same content as the CLI
  backend: `*.gpg diff=gpg\n` or `*.age binary\n`). If the CLI backend wrote
  it earlier, do not rewrite it.
* Prefer file mode `0o644` for worktree files (matches `fs.Store.Set`); do not
  fight git's executable-bit handling — gopass secrets are never executable.
* `Status` and object hashes are compared in CI rather than status text (see
  Testing).

### Status output

`gitfs.Status` prints raw `git status`. The pure-Go backend cannot replicate
that byte-for-byte. Implement a deterministic summary derived from
`wt.Status()`:

```text
# on branch <branch>
# changes to be committed:
#   modified:   secret/foo.gpg
# untracked files:
#   secret/bar.gpg
```

Keep it stable (sorted by path) so tests can assert on it. Note the divergence
in `docs/backends/gogit.md` and do not compare status output between backends in
CI — compare the object graph.

### doctor

Update `internal/action/doctor.go` `doctorCheckGit` to accept the pure-Go
backend: it should only fail when a store uses **`gitfs` specifically** and the
`git` binary is missing. If the store resolves to `gogit`, the check passes
(with an informational note that the pure-Go backend is in use). Import the
`gogit` package for the backend name/constant if needed.

### Error handling parity

Return the same sentinels the rest of the code base switches on, so no upstream
change is required:

* `store.ErrGitNotInit` — operations on an uninitialised repo.
* `store.ErrGitNothingToCommit` — empty commit (map from `git.ErrEmptyCommit`).
* `store.ErrGitNoRemote` — push/pull without a usable remote.

`store/leaf`'s `TryAdd`/`TryCommit`/`TryPush` call sites
(`internal/store/leaf/write.go`, `move.go`, `recipients.go`, `templates.go`,
`link.go`, `fsck.go`, `reencrypt.go`) already tolerate these sentinels.

---

## Feature parity matrix

| Capability | `gitfs` (CLI) | `gogit` (pure Go) |
| --- | --- | --- |
| init / clone | native | native |
| add / status / log / show | native | native |
| commit (unsigned) | native | native |
| fetch / push / fast-forward pull | native | native |
| remote add/remove/list | native | native |
| non-fast-forward pull | native (git merge) | **file-level merge + two-parent commit** (custom, see above) |
| GPG-signed commits | native (gpg/gpg-agent/pinentry/smartcard) | shell out to `git commit -S` if git present; else unsigned + warning |
| SSH-signed commits (`gpg.format=ssh`) | native | **unsupported** — warning |
| SSH remote auth | native (git) | native (x/crypto/ssh + ssh-agent) |
| HTTPS auth via credential.helper / GCM | native | shell out to `git credential` if git present; else URL/netrc/askpass |
| git hooks | native | ignored (documented) |
| `core.autocrlf`, filemode fidelity | native | approximate (blobs unaffected) |
| gc / compact | `git gc --aggressive` | `RepackObjects` + `Prune` (best effort) |
| `gopass git` passthrough | provided by gitfs loader | provided by gitfs loader (needs git binary at runtime) |
| cgo / external binary required | no cgo, **needs `git`** | **pure Go, no `git`** |

---

## Testing and rollout

### Unit tests (`internal/backend/storage/gogit/*_test.go`)

* `New`/`Init`/`Clone` against `t.TempDir()`, using the `plain` crypto backend
  for speed where the RCS layer is under test.
* `Add`/`Commit`/`Revisions`/`GetRevision` round trip; assert
  `backend.Revision` fields and that `GetRevision` bytes equal the file bytes.
* **Merge matrix** — table-driven over the resolution rules:
  add-one-side, modify-one-side, identical-modify-both, delete-vs-modify (both
  directions must keep the secret), true conflict (local kept, `.conflict-<sha>`
  written for remote). Assert `NumParents() == 2` and that no secret is lost.
* `Push`/`Pull`/`TryPush`/`TryPull` against a bare local repo (no network),
  covering fast-forward, "already up to date", and a divergence that triggers
  the file-level merge.
* Error parity: `ErrGitNotInit`, `ErrGitNothingToCommit`, `ErrGitNoRemote`.
* `Version`, `Name`, `String`, `Path`, `IsInitialized`, `LinkTarget`.
* `auth.go` unit-tested with a fake `git credential fill` on `PATH`.

### Cross-backend fixture tests (the required CI comparison)

Add a test that runs **both** backends against the **same fixture repo(s)** and
diffs the resulting object graphs (not status text):

* Create a repo with the CLI backend (git present) and with the pure-Go backend
  from identical starting content and an identical sequence of operations
  (add/commit/modify/commit/delete/commit).
* Compare, per commit: the root **tree hash**, each **blob hash**, and **file
  modes**; and compare `NumParents`. Normalise away commit hashes (which depend
  on author/committer time and encoding) — compare **tree** hashes by walking
  the tree, not commit hashes.
* Watch specifically for index/file-mode/CRLF (`core.autocrlf`) divergence. If a
  text file differs only by line endings, the test should assert the *bytes*
  stored are identical (since gopass stores ciphertext, they should be).
* Guard the test so it `t.Skip`s when the `git` binary is unavailable (the CLI
  side needs it), mirroring how `tests/` skips without `GOPASS_BINARY`.

### Integration tests (`tests/`)

Add a `gogit`-varianted integration test that initialises a store with
`GOPASS_GIT_BACKEND=gogit` / `--storage=gogit`, performs write/read/sync, and
asserts behaviour matches the existing git-backed tests. Keep the existing git
tests unchanged.

### Rollout

1. Ship opt-in only: `storage.backend=gogit` or `GOPASS_GIT_BACKEND=gogit`.
2. Enable automatic fallback when `git` is absent (the resolver above).
3. **Do not make it the default.** Only reconsider after the signing and
   credential-helper stories in `signing.go`/`auth.go` reach parity, or after
   the CLI backend is deliberately demoted in a follow-up ADR.

---

## Documentation deliverables

* **`docs/backends/gogit.md`** (new; name it per `docs/conventions.md` §5.3):
  what works natively, what shells out, and what is unsupported. Include the
  parity matrix above, the conflict policy, and how to select the backend.
* **`docs/backends.md`**: add `gogit` to the storage/RCS list with a one-line
  description and a link.
* **`docs/config.md`**:
  * `storage.backend` valid-values list gains `gogit`.
  * New `GOPASS_GIT_BACKEND` row in the env table (required by
    `TestEnvVarsInDocs`).
* **`docs/components.dot`**: optionally add a `gogit` node next to `gitfs`
  (pure-Go, no binary edge). Not strictly required, but keeps the diagram
  accurate if the backend lands.
* Regenerate completions and the man page (`make man`, `make *.completion`) —
  the `--storage` help text is built from `StorageRegistry.BackendNames()` and
  the checked-in completion/man files must reflect `gogit`.
* **`CHANGELOG.md`**: no hand-written entry needed (this is a `feat`, generated
  at release). If any `pkg/` symbol changes (none are planned), A-12 applies.

---

## Implementation checklist

Files to **add**:

* [ ] `internal/backend/storage/gogit/gogit.go` — `GoGit`, `New`, `Init`,
      `Clone`, `Version`, `Name`, `Path`, `String`, `IsInitialized`.
* [ ] `internal/backend/storage/gogit/loader.go` — registration, `Priority()`
      = 12, `Handles` (no git-binary requirement).
* [ ] `internal/backend/storage/gogit/storage.go` — delegate file ops +
      `LinkTarget` + `Fsck` to `fs.Store`.
* [ ] `internal/backend/storage/gogit/rcs.go` — `Add`/`TryAdd`/`Commit`/
      `TryCommit`/`Push`/`TryPush`/`Pull`/`Status`/`Compact`/`AddRemote`/
      `RemoveRemote`/`InitConfig`.
* [ ] `internal/backend/storage/gogit/revisions.go` — `Revisions`,
      `GetRevision`.
* [ ] `internal/backend/storage/gogit/merge.go` — file-level merge, two-parent
      commit, conflict policy.
* [ ] `internal/backend/storage/gogit/auth.go` — the single transport/auth seam.
* [ ] `internal/backend/storage/gogit/config.go` — local config read/write,
      `.gitattributes`, `defaultRemote`/`defaultBranch`.
* [ ] `internal/backend/storage/gogit/signing.go` — GPG signer / `gpg.format`
      detection.
* [ ] `internal/backend/storage/gogit.go` — blank import to register.
* [ ] `internal/backend/storage/gogit/*_test.go` — unit + merge-matrix tests.
* [ ] `tests/gogit_test.go` — integration variant (opt-in).
* [ ] `docs/backends/gogit.md`.

Files to **modify**:

* [ ] `internal/backend/storage.go` — add `GoGit` constant; add
      `ResolveStorageBackend`; apply it (and `GOPASS_GIT_BACKEND`) in
      `DetectStorage`.
* [ ] `internal/backend/storage/gitfs/loader.go` — return
      `ErrGitBinaryNotFound` from `Handles`/`New`/`Init` when git is missing.
* [ ] `internal/action/init.go` — apply `ResolveStorageBackend` where the
      default is `GitFS`; honour `GOPASS_GIT_BACKEND`.
* [ ] `internal/action/rcs.go` — same as `init.go`.
* [ ] `internal/action/doctor.go` — `doctorCheckGit` aware of `gogit`.
* [ ] `go.mod` / `go.sum` — add `github.com/go-git/go-git/v5` (Apache-2.0;
      already allowed by `.license-lint.yml`) and its pure-Go deps
      (`go-billy`, `go-git-fixtures` only as a test dep if used).
* [ ] `docs/backends.md`, `docs/config.md`, `docs/components.dot`.
* [ ] Regenerated `gopass.1`, `bash.completion`, `zsh.completion`,
      `fish.completion`.
* [ ] `docs/adr/README.md` — add this record to the index (same commit).

Files that must **not** change:

* `internal/backend/rcs.go` and the `Storage` interface contract in
  `internal/backend/storage.go` (the whole point is that the interface is
  stable).
* The `gitfs` behaviour and its tests (except the additive
  `ErrGitBinaryNotFound` on `Handles`/`New`).

---

## Future go-git v6 migration

`/v6` is currently alpha with a rewritten transport layer. To keep the bump
contained:

* Route **all** imports of `go-git/v5` through the `gogit` package; no other
  package imports go-git.
* Keep **all** transport/auth logic in `auth.go`. The v5 → v6 change is expected
  to be concentrated there (`transport.AuthMethod`, `ssh.*`, `http.*` packages
  were reworked).
* Keep object/plumbing usage in `rcs.go`, `revisions.go`, `merge.go`; the v6
  object model is broadly compatible, with renames documented per-file by the
  upstream release notes.
* A single `go.mod` major-version bump plus edits to those files should suffice.
  Do not leak `plumbing.Hash` or `object.Commit` types across the package
  boundary — the `backend.Storage` interface already uses gopass's own
  `Revision` type precisely so this stays internal.

---

## Conflict policy — rationale (restated)

The default is **never silently discard a secret**. On a true conflict:

* Local content stays in place.
* The remote content is written to `<name>.conflict-<shortsha>` and committed.
* The user is warned with the list of conflicted paths.

The alternative ("prefer newest by commit time") may be offered as an explicit
opt-in but must never be the default: for encrypted blobs there is no semantic
meaning to "newest" that distinguishes content, and an automatic pick can lose a
password that exists only on the losing side. Two files to reconcile by hand is
a recoverable annoyance; a lost password is not.

---

## Out of scope

* Changing the crypto backends (gpg/age) or the `Storage`/`rcs` interface
  (A-03 remains deferred and unrelated).
* Making `gogit` the default backend.
* Full byte-for-byte `git status` parity.
* Submodules, sparse checkout, partial/shallow clone, or reflog management.
* Repository migration tooling between the two backends (the `.git` directory is
  shared; switching selection is enough).

---

## Consequences

### Positive

* One static, cgo-free binary can version and sync a store without `git` on
  `PATH` — the main win for Windows and for single-binary distributions.
* Automatic fallback avoids silently degrading to the unversioned `fs` backend
  on git-less systems.
* Because both backends share `fs.Store` and the `.git` on-disk format, users can
  switch between them freely (config/env only).
* The interface contract is unchanged, so no churn in `store/leaf`,
  `internal/action`, or the public API.

### Negative and risks

* **Signed commits are degraded** (shell-out or unsigned) — must be documented.
* **Credential helpers are not honoured natively** for HTTPS — shell-out to
  `git credential` or explicit configuration required.
* **Git hooks do not run.**
* A new third-party dependency (`go-git/v5` + `go-billy`), accepted because it
  is pure Go and Apache-2.0.
* Merge semantics differ from git (file-level, two-parent). Correct for
  encrypted blobs, but users must understand the conflict files.
* Behavioural drift risk on CRLF/file mode for non-blob files; mitigated by
  choosing to store ciphertext and by the cross-backend object-graph test.
