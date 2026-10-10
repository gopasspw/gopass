# `gogit` storage backend

The `gogit` backend is a filesystem-backed storage with Git revision control
implemented in pure Go using [go-git](https://github.com/go-git/go-git). It is
an **opt-in** alternative to the default [`gitfs`](gitfs.md) backend and does
**not** require the `git` binary to be installed.

The main use cases are:

- Single-binary or cross-compiled distributions that cannot rely on a `git`
  executable being present.
- Windows users without Git for Windows on `PATH`.
- Systems where `git` is unavailable: gopass automatically falls back to
  `gogit` instead of degrading to the unversioned [`fs`](fs.md) backend.

Both backends share the same on-disk `.git` format and the same file layer, so
you can switch between them freely by changing the configuration.

## Selecting the backend

The backend can be selected in three ways, in decreasing precedence:

1. The `--storage=gogit` flag on `gopass init`, `gopass setup`, `gopass clone`
   and `gopass rcs init`.
2. The `GOPASS_GIT_BACKEND` environment variable (`gogit`, or `gitfs`/`cli`/`cmd`
   for the CLI backend).
3. The `storage.backend` configuration option:

   ```ini
   storage.backend = gogit
   ```

When no backend is selected explicitly, gopass auto-detects. If a `.git`
directory exists and the `git` binary is available, `gitfs` is preferred. If the
`git` binary is **not** available, gopass falls back to `gogit` automatically.

## Feature parity

| Capability | `gitfs` (CLI) | `gogit` (pure Go) |
| --- | --- | --- |
| init / clone | native | native |
| add / status / log / show | native | native |
| commit (unsigned) | native | native |
| fetch / push / fast-forward pull | native | native |
| remote add/remove/list | native | native |
| non-fast-forward pull | native (`git merge`) | file-level merge + two-parent commit |
| GPG-signed commits | native (gpg/gpg-agent/pinentry/smartcard) | shells out to `git commit -S` if git is present; otherwise unsigned + warning |
| SSH-signed commits (`gpg.format=ssh`) | native | **unsupported** — warning |
| SSH remote auth | native (git) | native (`x/crypto/ssh` + ssh-agent) |
| HTTPS auth via `credential.helper` / GCM | native | shells out to `git credential` if git is present; otherwise URL userinfo |
| git hooks | native | ignored |
| `core.autocrlf`, filemode fidelity | native | approximate (encrypted blobs unaffected) |
| gc / compact | `git gc --aggressive` | `RepackObjects` + `Prune` (best effort) |
| `gopass git` passthrough | provided by the `gitfs` loader | provided by the `gitfs` loader (needs the git binary at runtime) |
| cgo / external binary required | no cgo, **needs `git`** | **pure Go, no `git`** |

## Conflict policy

go-git v5 only implements fast-forward merges. When a pull cannot be
fast-forwarded, `gogit` performs a **file-level merge** and creates a
**two-parent merge commit** so that a later user of the real `git` CLI sees
valid, connected history.

Because gopass stores encrypted blobs, a content-level three-way merge is
meaningless. Changes are therefore resolved per path:

- Changed only locally → keep local.
- Changed only remotely → apply remote.
- Changed identically on both sides → no-op.
- Delete vs. modify → **keep the modification** (a secret is never dropped).
- Changed on both sides with different content → **conflict**: the local version
  is kept and the remote version is written to `<name>.conflict-<shortsha>`
  (e.g. `secrets/api.gpg` → `secrets/api.conflict-1a2b3c4.gpg`). The conflict
  file is part of the merge commit, so the remote content is preserved in
  history even if you delete the file.

The default is deliberately **never silently discard a secret**. Two files to
reconcile by hand is a recoverable annoyance; a lost password is not.

## Known limitations

- **Signed commits** are degraded. `commit.gpgsign=true` is honoured by shelling
  out to `git commit -S` when a git binary is available; otherwise an unsigned
  commit is created with a warning. `gpg.format=ssh` signing is unsupported.
- **Credential helpers** are not honoured natively for HTTPS. When a git binary
  is available, `gogit` shells out to `git credential fill`; otherwise only URL
  userinfo is used.
- **Git hooks** (`pre-commit`, `commit-msg`, `post-merge`, …) are not executed.
  gopass's own hook system (`core.pre-hook`/`core.post-hook`) is independent of
  git and continues to work.
- `core.sshCommand` is not set by this backend, since go-git does not read it.
- `git status` output is a deterministic summary, not byte-for-byte identical to
  the CLI.

## See also

- [ADR A-17](../adr/A-17-pure-go-git-backend.md) — the design record.
- [`gitfs`](gitfs.md) — the default git CLI backend.
