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
$ gopass recipients ack
```

## Modes of operation

* List all existing recipients, per mount: `gopass recipients`
* Add/Authorize a new public key to decrypt a store (mount): `gopass recipients add`
* Remove/Deuathorize an existing public key from a store (mount): `gopass recipients remove`
* Acknowledge changes in the `recipients.hash`

## Flags

Flag | Aliases | Description
`--store` | | Store to operate on.
`--force` | | Do not ask for confirmation.

## Important Remarks

WARNING: Removing a recipient can only ever work for new or changed secrets.
When a recipient is removed they will still be able to access anything that
they used to have access to. As a logical consequence one **should** change
all secrets when removing a recipient.

## Recipients hashing

This is an experimental feature that will hash the content of each mounts
recipients file (only the top most file) to display a warning when this is
changed by anyone else (local changes update it without warning). This
can happen either when a teammate modifies that file or when an attacker
tries to modify the recipients file in the central storage to get themselves
added to any newly modified secrets.
