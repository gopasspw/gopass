# `convert` command

The `convert` command exists to migrate stores between different backend
implementations.

Note: This command exists to enable a possible migration path. If we agree
on a single set of backend implementations the multiple backend support
might go away and this command as well.

Warning: Converting between different RCS backends will loose part of the history. While we try to retain as much information as possible especially the commit timestamps will be set to the convert time.

## Synopsis

```
$ gopass convert --store=foo --move=true --storage=gitfs --crypto=age
$ gopass convert --store=bar --move=false --storage=fs --crypto=plain
```

## Flags

Flag | Description
---- | -----------
`--store` | Substore to convert.
`--move` | Remove backup after converting? (default: `false`)
`--storage` | Target storage backend.
`--crypto` | Target crypto backend.
