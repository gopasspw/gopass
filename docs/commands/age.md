# `age` commands

The `age` commands are provided by the [age crypto backend](../backends/age.md).
They allow limited interactions with the gopass-specific age identities. They
are only available when a store uses the `age` backend.

Identities added here are automatically used as recipients when encrypting, but
they are **not** added to your recipients automatically. Make sure to keep your
recipients and identities in sync, see
[`recipients`](recipients.md).

## Synopsis

```
$ gopass age agent start
$ gopass age agent status
$ gopass age identities
$ gopass age identities add AGE-SECRET-KEY-1... age1...
$ gopass age identities sort
$ gopass age lock
```

## Subcommands

### `age agent`

Manage the background agent that caches age identities in memory. The agent is
optional but recommended if your identities require a password or are managed by
a plugin.

| Subcommand | Description                                                            |
|------------|------------------------------------------------------------------------|
| `start`    | Start the age agent.                                                   |
| `stop`     | Stop the age agent.                                                    |
| `status`   | Check if the age agent is running. Exits 0 when running, non-zero otherwise. |
| `unlock`   | Unlock the agent and reload identities (prompts for the PIN).          |
| `lock`     | Lock the agent and clear all cached identities.                        |

### `age identities`

List or manage the age identities used for decryption and encryption.

| Subcommand | Aliases | Description                                                              |
|------------|---------|--------------------------------------------------------------------------|
| `add`      |         | Add an existing age identity.                                            |
| `keygen`   |         | Generate a new age identity.                                             |
| `remove`   | `rm`    | Remove an identity.                                                      |
| `sort`     |         | Interactively configure the preferred order of identities for decryption. |

The preferred order is persisted in the `age.identities` configuration option
and used to deterministically sort identities before handing them to age. Only
public recipient strings are shown and stored, never secret key material.

### `age lock`

Alias for `age agent lock`, kept for convenience. It is hidden from the help
output.

## Flags

| Flag                 | Aliases | Description                                                  |
|----------------------|---------|--------------------------------------------------------------|
| `--age-sshkeys`      |         | Load SSH keys from the default SSH directory.                |
| `--age-ssh-key-path` |         | Custom path to an additional SSH key or directory of keys.   |

## Details

All age identities, including plugin ones, are supported. GitHub identities are
supported as well, but deprecated by age; gopass falls back to the corresponding
SSH identities and keeps a local cache of SSH keys for a given GitHub identity.

For automation, set `GOPASS_AGE_PASSWORD` to the passphrase for the identity
file, and `GOPASS_AGE_STDIN_PASSPHRASE` to force reading the passphrase from the
terminal instead of pinentry.

See [backends/age.md](../backends/age.md) for setup instructions, the agent
socket location and Yubikey usage.

## Exit codes

| Code | Meaning                                      |
|-----:|----------------------------------------------|
| 0    | Operation completed successfully             |
| 1    | Operation failed; or the agent is not running |

See [docs/exit-codes.md](../exit-codes.md) for the full table (the age backend
uses the generic `Unknown` code for failures).
