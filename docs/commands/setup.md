# `setup` command

The `setup` command initializes a new password store and is also the onboarding
wizard that is automatically invoked when `gopass` is started without any
existing password store.

It exists so users can be provided with simple one-command setup instructions.

## Synopsis

```
$ gopass setup
$ gopass setup --crypto age --storage gitfs
$ gopass setup --remote git@example.com/store.git
$ gopass setup --team team-a --remote git@example.com/team-a.git
```

## Modes of operation

* Initialize a new local store interactively
* Join an existing team by cloning its remote (`--remote`)
* Create a new team and publish it to a remote (`--create-team`)

## Flags

| Flag            | Aliases   | Description                                                                                 |
|-----------------|-----------|---------------------------------------------------------------------------------------------|
| `--remote`      |           | URL to a git remote, will attempt to join this team.                                        |
| `--team`        | `--alias` | Name of the team to create or join (may contain slashes). Also used as the local mount point. |
| `--create-team` | `--create`| Create a new team instead of joining an existing one.                                        |
| `--name`        |           | Firstname and Lastname for unattended GPG key generation.                                   |
| `--email`       |           | EMail for unattended GPG key generation.                                                    |
| `--crypto`      |           | Select the crypto backend.                                                                  |
| `--storage`     |           | Select the storage backend.                                                                 |

`--alias` and `--create` are deprecated aliases kept for backward compatibility.

## Details

If the store is already initialized the wizard aborts to avoid overwriting
existing data.

If no usable private key is found during setup, gopass generates a new key pair.
For the `gpgcli` backend this may take up to a few minutes.

For non-interactive use supply `--name`, `--email` and (when joining or creating
a team) `--team` and `--remote`. Creating a team without a team name fails fast
in non-interactive mode.

See [docs/setup.md](../setup.md) for the full installation and setup guide and
[backends.md](../backends.md) for the available backends.
