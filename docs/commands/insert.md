# `insert` command

The `insert` command is used to manually set (insert, or change) a password in the store. It applies to either new or existing secrets.

## Synopsis

```
$ gopass insert entry
$ gopass insert entry key
```

## Modes of operation

* Create a new entry with a user-supplied password, e.g. a new site with a user-generated password or one picked from `gopass pwgen`: `gopass insert entry`
* Change an existing entry to a user-supplied password
* Create and change any field of a new or existing secret: `gopass insert entry key`
* Read data from STDIN and insert (or append) to a secret

Insert is similar in effect to `gopass edit` with the advantage of not displaying any content of the secret when changing a key.

Note: `insert` will not change anything but the `Password` field (using the `insert entry` invocation) or the specified key (using the `insert entry key` invocation).

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--echo` | `-e` | Display the secret while typing (default: `false`)
`--multiline` | `-m` | Insert using `$EDITOR` (default: `false`). This identical to running `gopass edit entry`. All other flags are ignored.
`--force` | `-f` | Overwrite any existing value and do not prompt. (default: `false`)
`--append` | `-a` | Append to any existing data. Only applies if reading from STDIN. (default: `false`)
