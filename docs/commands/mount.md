# `mount` commands

The `mount` commands allow managing mounted substores. This is one of the
distinctive core features of `gopass` and we aim making working with substores
as seamless as possible.

Instead of support for encrypting different parts of a store for different
recipients we instead encourage users to mount different stores - each
encrypted to a uniform set of recipients - into a semless virtual tree structure.

This feature is modeled after standard POSIX mount semantics.

## Synopsis

```
$ gopass mounts
$ gopass mounts add mount/point /path/to/store
$ gopass mounts remove mount/point
```

## Modes of operation

* Add a new mount
* List existing mounts
* Remove an existing mount
