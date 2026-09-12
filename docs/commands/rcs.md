# `rcs` commands

The `rcs` commands run version control operations inside a password store. They
are backend agnostic wrappers around the operations the storage backend
implements.

Note: This command is hidden from the regular help output. The `store`, `name`
and `email` flags are only meaningful for backends that need them (e.g. `gitfs`).

## Synopsis

```
$ gopass rcs init
$ gopass rcs status
```

## Subcommands

### `rcs init`

Create and initialize a new RCS repository in the store.

| Flag      | Aliases    | Description                       |
|-----------|------------|-----------------------------------|
| `--store` |            | Store to operate on.              |
| `--name`  | `--username` | Git author name.                |
| `--email` | `--useremail`| Git author email.               |
| `--storage` |          | Select the storage backend.       |

### `rcs status`

Show the current RCS status of the store.

| Flag      | Description             |
|-----------|-------------------------|
| `--store` | Store to operate on.    |

## Details

The equivalent per-store operations are also available as the hidden
[`pull`](pull.md) and [`push`](push.md) commands, and the
[`git`](git.md) command from the `gitfs` backend.

## Exit codes

See the `git`, `pull` and `push` entries in
[docs/exit-codes.md](../exit-codes.md) for the relevant codes and the full
table.
