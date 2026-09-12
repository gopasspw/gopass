# `init` command

The `init` command is used to initialize a new password store.
If no recipients are specified a usable existing private key is used.

The `init` command must be used to initialize new mounts. `gopass mounts add` only supports adding existing mounts.

Note: We do not support adding recipients using `init`. Please use `gopass recipients add` for that!

## Synopsis

```
$ gopass init
$ gopass init --crypto [age|gpgcli] --storage=[fs|gitfs]
```

## Flags

| Flag        | Aliases | Description                                                                                                 |
|-------------|---------|-------------------------------------------------------------------------------------------------------------|
| `--path`    | `-p`    | Initialize the (sub) store in this location.                                                                |
| `--store`   | `-s`    | Mount the newly initialized sub-store at this mount point                                                   |
| `--crypto`  |         | Select the crypto backend. Choose one of: `gpgcli`, `age` or `plain`. Default: `gpgcli` |
| `--storage` |         | Select the storage and RCS backend. Choose one of: `gitfs`, `fs`, `fossilfs`, `jjfs` or `cryptfs`. Default: `gitfs` |

Note: `plain` does not encrypt and is only intended for testing. `cryptfs`,
`fossilfs` and `jjfs` are experimental.

See [backends.md](../backends.md) for more information on the available backends.
