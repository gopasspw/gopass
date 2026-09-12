# `pull` command

The `pull` command pulls a store from its configured remote. It is a thin
wrapper around the underlying RCS backend (e.g. `git pull`).

Note: This command is hidden from the regular help output. Prefer
[`sync`](sync.md) for normal operation, which pulls and pushes all stores and
also handles the associated GPG keys.

## Synopsis

```
$ gopass pull
$ gopass pull --store foo origin master
```

## Modes of operation

* Pull the selected store from its remote
* Pull a specific remote and branch

## Flags

| Flag      | Aliases | Description                       |
|-----------|---------|-----------------------------------|
| `--store` | `-s`    | Select the store to pull.         |

## Details

The arguments are the optional remote and branch. If neither is given the
storage backend defaults are used.

See also [`push`](push.md) and [`rcs`](rcs.md).
