# `edit` command

The `edit` command loads a new or existing secret into your `$EDITOR` (default: `vim`)
and saves the resulting content in the password store. It will attempt to create secure
temporary directory (depending on the OS) and will warn if insecure editor configuration
(currently only `vim`) is detected.

Native `gopass` MIME secrets are syntax checked and invalid encodings are rejected.
Any other type of secret is accepted as is.

`gopass` will honor templates when creating a new entry.

## Synopsis

```
$ gopass edit entry
$ gopass edit -e /bin/nano entry
$ EDITOR=/bin/nano gopass edit entry
```

## Modes of operation

* Create a new secret
* Edit an existing secret

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--editor` | `-e` | Specify the path to an editor. Must accept the filename as it's first argument.
`--create` | `-c` | Create a new secret. You can create a new secret with `edit` with or without `-c`, but `-c` will skip searching for existing matches.
