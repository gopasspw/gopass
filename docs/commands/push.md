# `push` command

The `push` command pushes a store to its configured remote. It is a thin wrapper
around the underlying RCS backend (e.g. `git push`).

Note: This command is hidden from the regular help output. gopass already pushes
automatically after each write when `core.autopush` is enabled (the default).
Prefer [`sync`](sync.md) for normal operation.

## Synopsis

```
$ gopass push
$ gopass push --store foo origin master
```

## Modes of operation

* Push the selected store to its remote
* Push to a specific remote and branch

## Flags

| Flag      | Aliases | Description                       |
|-----------|---------|-----------------------------------|
| `--store` | `-s`    | Select the store to push.         |

## Details

The arguments are the optional remote and branch. If neither is given the
storage backend defaults are used.

If the store has no remote configured, gopass prints `No Git remote. Not
pushing` and exits successfully.

See also [`pull`](pull.md) and [`rcs`](rcs.md).

## Exit codes

| Code | Meaning                        |
|-----:|--------------------------------|
| 0    | Push completed successfully    |
| 7    | Push to the remote failed      |

See [docs/exit-codes.md](../exit-codes.md) for the full table.
