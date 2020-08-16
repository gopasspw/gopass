# `sync` command

The `sync` command is the preferred way to manually synchronize changes between
your local stores and any configured remotes.

You can always `cd` into a git-based store and manually perform git operations,
but executing these through `gopass git` is deprecated and might be removed
at soe point.

Note: `gopass sync` only supports one remote per store.

## Flags

Flag | Description
---- | -----------
`--store` | Only sync a specific sub store


