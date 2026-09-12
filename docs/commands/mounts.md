# `mounts` commands

The `mounts` commands allow managing mounted substores. This is one of the
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
$ gopass mounts versions
```

## Modes of operation

* Add a new mount
* List existing mounts
* Remove an existing mount
* Display the versions of the external tools used by the mounts

## Subcommands

| Subcommand | Aliases                        | Description                                                            |
|------------|--------------------------------|------------------------------------------------------------------------|
| `add`      | `mount`                        | Mount an existing or new password store. Use `--create` to create one. |
| `remove`   | `rm`, `unmount`, `umount`      | Unmount a store. This only updates the configuration, it does not delete the store. |
| `versions` | `version`                      | Display version information of external commands used by the mounts.   |

## Creating new mounts

You can also create new mounts using `init` even if your store is already initialized:

```
gopass init --store mynewsubstore pgpkeyidentitfier
```

(You can also specify a specific local path using `--path`, just make sure to keep your PGP key identifier, e.g. its email or fingerprint, as the last argument.)
