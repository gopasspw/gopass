# `recipients` commands

The set of `recipients` commands allow managing public keys that are able to
decrypt a given password store. These commands are one of the more unique
`gopass` features and we aim to make working with teams as seamless as possible.

For the full team workflow reference, see
[docs/usecases/team-workflows.md](../usecases/team-workflows.md).

## Synopsis

```
$ gopass recipients
$ gopass recipients add [--store=<store>] <recipient-id>...
$ gopass recipients remove [--store=<store>] <recipient-id>...
$ gopass recipients update [--store=<store>] [<recipient-id>...]
$ gopass recipients canonicalize [--store=<store>]
$ gopass recipients ack [--store=<store>]
```

## Subcommands

### `recipients list` (default — no subcommand)

Lists all existing recipients for every mounted store.

### `recipients add` (aliases: `authorize`)

Adds one or more recipients to a store and re-encrypts all secrets so the new
recipients can decrypt them.

Recipient identifiers may be email addresses, short key IDs, or full
fingerprints. gopass normalizes every identifier to its canonical form (full
GPG fingerprint) before storing it in `.gpg-id`, ensuring the entry and the
`.public-keys/<id>` filename always match (see
[ADR A-14](../adr/A-14-team-workflows.md)).

**Example:**

```
$ gopass recipients add --store team-a alice@example.com
Resolved 'alice@example.com' to canonical key ID '0x1A2B3C4D5E6F'
√ Added 1 recipients
You need to run 'gopass sync' to push these changes
```

**Flags:** `--store` (store to operate on), `--force` (skip confirmation).

### `recipients remove` (aliases: `rm`, `deauthorize`)

Removes a recipient from a store and re-encrypts all secrets. After removal,
the removed recipient can no longer decrypt *new* changes — but they can still
decrypt old revisions from the git history. Always rotate secrets after
removing a recipient.

Removal also performs **recipient-scoped cleanup**: the removed recipient's
`.public-keys/<id>` and legacy `.gpg-keys/<id>` files are deleted. No other
recipient's files are affected.

**Flags:** `--store`, `--force`.

### `recipients update` (aliases: `refresh`)

Re-exports the named recipients' public keys from the local keyring into
`.public-keys/`, overwriting stale copies. Use this after extending an expired
key or adding new subkeys. If no IDs are given, your own key is updated.

**Example:**

```
$ gopass recipients update --store team-a
Refreshing public keys in store "team-a" ...
Updated public key for '0x1A2B3C4D5E6F'.
Done. You may want to run 'gopass sync' to push the updated keys.
```

**Flags:** `--store`.

### `recipients canonicalize` (aliases: `canon`)

Rewrites the `.gpg-id` file of a store so that every recipient ID is in its
canonical (full-fingerprint) form and renames the corresponding
`.public-keys/` files to match. Safe migration — no re-encryption required.

Run this once on existing stores that use non-canonical IDs (email addresses
or short key IDs). After running, use `gopass sync` to publish the changes.

**Flags:** `--store`.

### `recipients ack` (aliases: `acknowledge`)

Updates `recipients.hash` after manually validating changes to the recipients
list. This is part of the experimental recipients hashing feature (see below).

**Flags:** `--store`.

## Common flags

| Flag | Description |
|------|-------------|
| `--store` | Store to operate on. |
| `--force` | Skip confirmation prompts (supported on `add` and `remove`). |

## Important Remarks

WARNING: Removing a recipient can only ever work for new or changed secrets.
When a recipient is removed they will still be able to access anything that
they used to have access to. As a logical consequence one **should** change
all secrets when removing a recipient.

## Recipients hashing

This is an experimental feature that will hash the content of each mount's
recipients file (only the top most file) to display a warning when this is
changed by anyone else (local changes update it without warning). This
can happen either when a teammate modifies that file or when an attacker
tries to modify the recipients file in the central storage to get themselves
added to any newly modified secrets.

## Key refresh and expiry recovery

When a GPG key expires, other team members will see warnings during sync.
The key owner should extend the key locally and then run:

```
$ gopass recipients update --store <team-store>
$ gopass sync --store <team-store>
```

Other members will pick up the refreshed key on their next `gopass sync`.
You can check the health of your recipient keys at any time with:

```
$ gopass doctor --recipients
```

See also [ADR A-13](../adr/A-13-expired-gpg-key-handling.md) for details
on expired key handling.
