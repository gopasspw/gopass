# `fscopy` command

The `fscopy` command is used to copy a file from your filesystem into your
password store, while keeping it in clear in your local filesystem after
having stored it in your encrypted store.

## Synopsis

```bash
$ gopass fscopy ~/test/file data/test/file-entry
$ gopass fscopy data/test/file-entry ~/file
```

## Modes of operation

This command either reads a file from the filesystem and writes the
encoded and encrypted version in the store or it decrypts and decodes
a secret and writes the result to a file. Either source or destination
must be a file and the other one a secret.
If you want the source to be removed use 'gopass fsmove'.

`fscopy` is intended to work with raw files.

### Example
```
$ gopass fscopy ~/test/file data/test/file-entry
$ gopass cat data/test/file-entry
```

See also the docs for the [`cat` action](cat.md).

## Flags

This command has currently no supported flags except the gopass globals.
