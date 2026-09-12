# `delete` command

The `delete` command is used to remove a single secret or a whole subtree.

Note: Recursive operations crossing mount points are intentionally not supported.

## Synopsis

```
$ gopass delete entry
$ gopass rm -r path/to/folder
$ gopass rm -f entry
$ gopass delete entry key
```

## Modes of operation

* Delete a single secret
* Delete a single key from an existing secret
* Delete a directory of secrets

## Flags

| Flag          | Aliases | Description                           |
|---------------|---------|---------------------------------------|
| `--recursive` | `-r`    | Recursively delete files and folders. |
| `--force`     | `-f`    | Do not ask for confirmation.          |

## Exit codes

| Code | Meaning |
|-----:|---------|
| 0 | Secret deleted successfully |
| 2 | No name provided; or multiple names with `-r`; or target is a directory without `-r` |
| 4 | `--key` value conflicts with an existing secret name |
| 10 | Secret not found |
| 18 | Delete or YAML-key removal failed |
| 20 | Post-delete hook execution failed |

See [docs/exit-codes.md](../exit-codes.md) for the full table.

## Details

* Removing a single key will need to decrypt the secret
