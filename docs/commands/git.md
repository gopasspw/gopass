# `git` command

The `git` command runs an arbitrary `git` command inside a password store,
without the need to `cd` into the store directory first.

Note: This command is provided by the `gitfs` storage backend and is only
available when the selected store uses that backend.

## Synopsis

```
$ gopass git status
$ gopass git --store foo log --oneline
$ gopass git remote add origin git@example.com/store.git
```

## Modes of operation

* Run any `git` subcommand in the directory of the selected store

## Flags

| Flag      | Aliases | Description             |
|-----------|---------|-------------------------|
| `--store` |         | Store to operate on.    |

## Details

The command is hidden from the regular help output because it is intended as an
escape hatch. Prefer the dedicated subcommands (e.g. `gopass sync`) for common
operations.

The arguments after `gopass git` are passed through to `git` unchanged. `STDIN`,
`STDOUT` and `STDERR` are connected to the terminal.

See also [`rcs`](rcs.md), [`pull`](pull.md) and [`push`](push.md).

## Exit codes

The exit code is that of the invoked `git` process, with `0` indicating success.
See [docs/exit-codes.md](../exit-codes.md) for the `git` entry and the full
table.
