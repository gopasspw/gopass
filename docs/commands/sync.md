# `sync` command

The `sync` command is the preferred way to manually synchronize changes between
your local stores and any configured remotes.

You can also `cd` into a git-based store and manually perform git operations,
or use the `gopass git` command to automatically run a command in the correct
directory.

Note: `gopass sync` only supports one remote per store.

## Flags

Flag | Description
---- | -----------
`--store` | Only sync a specific sub store
