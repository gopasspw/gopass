# `fsck` command

`gopass` can check integrity of it's password stores with the `fsck` command.
It will ensure proper file and directory permissions as well as proper
recipient coverage (on supported crypto backends, only).

## Synopsis

```
$ gopass fsck
```

## Modes of operation

* Check the entire password store, incl. all mounts
* Check only the specified mount

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--decrypt` | | Decrypt and reencrypt all secrets. WARNING: This will update all secrets to the native `gopass` MIME format. This might be incompatible with other password store clients.
