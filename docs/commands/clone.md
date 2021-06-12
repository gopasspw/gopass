# `clone` command

The `clone` command allows cloning and setting up a new password store
from a remote location, e.g. a remote git repo.

## Synopsis

```
$ gopass clone git@example.com/store.git
$ gopass clone git@example.com/store.git sub/store
```

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--path` | | The path to clone the repo to.
`--crypto` | | Override the crypto backend to use if the auto-detection fails.
