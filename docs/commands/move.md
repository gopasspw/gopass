# `move` command

Note: The implementations for `copy` and `move` are exactly the same. The only difference is that `move` will remove the source after a successful copy.

The `move` command works like the Unix `mv` or `rsync` binaries. It allows moving either single entries or whole folders around. Moving across mounts is supported.

If the source is a directory, the source directory is re-created at the destination if no trailing slash is found. Otherwise the contained secrets are placed into the destination directory (similar to what `rsync` does).

Please note that `move` will always decrypt the source and re-encrypt at the destination.

Moving a secret onto itself is a no-op.

## Synopsis

```
# Overwrite new/leaf
$ gopass move path/to/leaf new/leaf
# Move the content of path/to/somedir to new/dir/somedir
$ gopass move path/to/somedirdir new/dir
# Does nothing
$ gopass move entry entry
```

## Modes of operation

* Move a single secret from source to destination
* Move a folder of secrets, possibly with sub folders, from source to destination

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--force` | `-f` | Overwrite existing destination without asking.

## Details

* To simplify the implementation and support multiple backends a `copy` or `move` operation will always decrypt and re-encrypt all affected secrets. Even if moving encrypted files around might be possible.
* You can move a secret to another secret, i.e. overwrite the destination. But `gopass` won't let you move a directory over a file. In that case you have to delete the destination first.

