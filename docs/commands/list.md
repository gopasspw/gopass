# `list` command

The `list` command is used to list all the entries in the password store or at a given prefix.

## Synopsis

```bash
gopass ls
gopass ls path/to/entries
```

- List all the entries in the password store including the one in mounted stores: `gopass list`
- List all the entries in a given folder showing their relative path from the root: `gopass list path/to/entries`

Note: `list` will not change anything, nor encrypt or decrypt anything.

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--limit value` | `-l value`| Max tree depth (default: -1)
`--flat`      |`-f`      | Print a flat list of secrets (default: false)
`--folders`    | `-d`    |  Print a flat list of folders (default: false)
`--strip-prefix` | `-s`    |  Strip prefix from filtered entries (default: false)

The `--flat` and `--folders` flags provide a plaintext list of the entries located at
the given prefix (default prefix being the root `/`). They are notably used to produce the
completion results.
The `--flat` one will list all entries, one per line, using its full path.
The `--folders` one will display all the folders, one per line, recursively per level.
For instance an entry `folder/sub/entry` would cause it to list both:

```bash
$ gopass list --folders
folder
folder/sub
```

whereas `gopass list --flat` would have just displayed one line: `folder/sub/entry`.

The `--strip-prefix` flag is meant to be used along with `--flat` or `--folders`.
It will list the relative path from the current prefix, removing the said prefix,
instead of listing the relative paths from the root.
For instance on entry `folder/sub/entry`, running `gopass ls -f -s folder` would display
 only `sub/entry` instead of `folder/sub/entry`.

The `--limit` flag starts counting its depth from the root store, which means that
a depth of 0 only lists the items in the root gopass store:

```bash
$ gopass list -l 0
gopass
├── bar/
├── foo/
└── test (/home/user/.local/share/gopass/stores/substore1)
```

A value of 1 would list all the items in the root, plus their sub-items but no more:

```bash
$ gopass list -l 1
gopass
├── bar/
│   └── bar
├── foo/
│   ├── bar
│   └── foo
└── test (/home/user/.local/share/gopass/stores/substore1)
    └── foo
```

A negative value lists all the items without any depth limit.

```bash
$ gopass list -l -1
gopass
├── bar/
│   └── bar
├── foo/
│   ├── bar/
│   │   ├── bar/
│   │   │   └── bar
│   │   └── baz
│   └── foo
└── test (/home/user/.local/share/gopass/stores/substore1)
    └── foo
```

The flags can be used together: `gopass -l 1 -d` will list only the folders up to a depth of 1:

```bash
$ gopass list -l 1 -d
bar/
foo/
foo/bar/
test/
test/foo/
```

## Shadowing

It is possible to have a path that is both an entry and a folder. In that case the list command
will display the folder with a marker of `(shadowed)`, it can still be accessed using
`gopass show path/to/it`, while the content of the folder can be listed using `gopass list path/to/it`.

It should also be noted that the `mount` command can completely "shadow" an entry in a password store,
simply by having the same name and this entry and its subentries will not be visible
using `ls` anymore until the substore is unmounted.
The entries shadowed by a mount will not show up in a search and cannot be accessed at all without unmounting.

For instance in our example above, maybe there is an entry test/zaz in the root store,
but since the substore is mounted as `test/`, it only displays the content of the substore.
Unmounting it reveals its shadowed entries:

```bash
$ gopass list test
test/ 
└── foo
$ gopass mounts rm test
$ gopass list test
test/ 
└── zaz
```
