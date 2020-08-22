# `recipients` commands

The set of `recipients` commands allow managing public keys that are able to
decrypt a given password store.

These commands are one of the more unique `gopass` features and we aim to
make working with this as seamless as possible.

## Synopsis

```
$ gopass recipients
$ gopass recipients add
$ gopass recipients remove
```

## Modes of operation

* List all existing recipients, per mount: `gopass recipients`
* Add/Authorize a new public key to decrypt a store (mount): `gopass recipients add`
* Remove/Deuathorize an existing public key from a store (mount): `gopass recipients remove`

## Flags

Flag | Aliases | Description
`--store` | | Store to operate on.
`--force` | | Do not ask for confirmation.

## Important Remarks

WARNING: Removing a recipient can only ever work for new or changed secrets.
When a recipient is removed they will still be able to access anything that
they used to have access to. As a logical consequence one **should** change
all secrets when removing a recipient.
