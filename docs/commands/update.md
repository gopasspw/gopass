# `update` command

The `update` command will attempt to auto-update `gopass` by downloading the
latest release from GitHub. It performs several pre-flight checks in order to
determine if the binary can be updated or not (e.g. if managed by a package
manager).

## Synopsis

```
$ gopass update
$ gopass update --pre
```

## Flags

Flag | Description
---- | -----------
`--pre` | Include pre-releases such as release candidates (default: `false`).

## Exit codes

| Code | Meaning                                              |
|-----:|------------------------------------------------------|
| 0    | gopass is up to date or the update applied successfully |
| 1    | Update check or download failed                       |

See [docs/exit-codes.md](../exit-codes.md) for the full table.
