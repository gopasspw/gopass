# `update` command

The `update` command will attempt to auto-update `gopass` by downloading the
latest release from GitHub. It performs several pre-flight checks in order to
determine if the binary can be updated or not (e.g. if managed by a package
manager).

## Synopsis

```
$ gopass update
$ gopass update --pre
```

## Flags

Flag | Description
---- | -----------
`--pre` | Update to pre-releases / release candidates (default: `false`).
