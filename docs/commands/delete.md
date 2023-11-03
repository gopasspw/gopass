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
* Delete a directoy of secrets

## Flags

| Flag          | Aliases | Description                           |
|---------------|---------|---------------------------------------|
| `--recursive` | `-r`    | Recursively delete files and folders. |
| `--force`     | `-f`    | Do not ask for confirmation.          |

## Details

* Removing a single key will need to decrypt the secret
