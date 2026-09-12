# `reorg` command

The `reorg` command allows reorganizing a part of the password store by editing
a text file. It lists all secrets line by line, opens the list in your
`$EDITOR`, and moves the secrets according to your edits.

## Synopsis

```
$ gopass reorg
$ gopass reorg path/to/prefix
```

## Modes of operation

* Reorganize the whole store
* Reorganize only the secrets below a given prefix

## Flags

None.

## Details

The list of secrets is opened in your `$EDITOR`. Because only the names are
shown, this is a rename and move operation, not an edit of the secret contents.

The following restrictions apply:

* `reorg` is not supported in non-interactive mode.
* The number of lines must not change; adding or removing lines aborts the
  operation.
* Moving secrets across mounts is not supported.

After you save and exit, gopass prints the planned moves and asks for
confirmation. If you decline, nothing happens. Otherwise the moves are performed
and a single `Reorganize secrets` commit is created.

## Exit codes

| Code | Meaning                                                       |
|-----:|---------------------------------------------------------------|
| 0    | Store reorganized successfully (or no changes detected)       |
| 2    | Secret count changed in the editor; or invalid move           |
| 3    | User aborted confirmation                                     |
| 4    | Running in non-interactive mode                               |
| 7    | Git commit after reorganization failed                        |
| 13   | Store contents could not be listed                            |

See [docs/exit-codes.md](../exit-codes.md) for the full table.
