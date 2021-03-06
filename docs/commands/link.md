# `link` command

The `link` (or `ln`) command is used to create a symlink from one secret in a
store to a target in the same store.

Note: Symlinks across different stores / mounts are currently not supported!

Note: `audit` and `list` do not recognize symlinks, yet. They will treat
symlinks as regular (different) entries.

## Synopsis

```
$ gopass ln foo/bar bar/baz
$ gopass show foo/bar
$ gopass show bar/baz
```

## Modes of operations

* Create a symlink from an existing secret to a new name, the target must not exist, yet

Note: Use `gopass rm` to remove a symlink.

## Flags

None.

