# `copy` command

The `copy` command copies a secret or a whole directory of secrets from one
location to another. It is the counterpart of [`move`](move.md): where `move`
removes the source after a successful copy, `copy` keeps it in place.

Copying across mounts is supported.

## Synopsis

```
# Copy new/leaf to other/leaf
$ gopass copy path/to/leaf new/leaf
# Copy the content of path/to/somedir into new/dir/somedir
$ gopass copy path/to/somedir new/dir
```

## Modes of operation

* Copy a single secret from source to destination
* Copy a folder of secrets, possibly with sub folders, from source to destination

## Flags

| Flag                   | Aliases | Description                                                  |
|------------------------|---------|--------------------------------------------------------------|
| `--force`              | `-f`    | Overwrite existing destination without asking.               |
| `--commit-message`     | `-m`    | Set the commit message.                                      |
| `--interactive-commit` | `-i`    | Open an editor for the commit message.                       |

## Details

If the source is a directory, the source directory is re-created at the destination if no trailing slash is found. Otherwise the contained secrets are placed into the destination directory (similar to what `rsync` does).

Note: The implementations for `copy` and `move` are exactly the same. The only difference is that `move` will remove the source after a successful copy.

To simplify the implementation and support multiple backends, a `copy` or `move` operation will always decrypt and re-encrypt all affected secrets, even if copying encrypted files around would be possible.
