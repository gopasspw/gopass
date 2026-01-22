# `fsmove` command

The `fsmove` command is used to move a file from your filesystem into your
password store, erasing it from your local filesystem after having stored it in your encrypted store.

## Synopsis

```bash
$ gopass fsmove ~/test/file data/test/file-entry
$ gopass fsmove data/test/file-entry ~/file
```

## Modes of operation

This command either reads a file from the filesystem and writes the
encoded and encrypted version in the store or it decrypts and decodes
a secret and writes the result to a file. Either source or destination
must be a file and the other one a secret. The source will be wiped
from disk or from the store after it has been copied successfully
and validated. If you don't want the source to be removed use
'gopass fscopy'.

`fsmove` is intended to work with raw files.

### Example
```
$ gopass fsmove ~/test/file data/test/file-entry
$ gopass cat data/test/file-entry
```

See also the docs for the [`cat` action](cat.md).

## Flags

This command has currently no supported flags except the gopass globals.
