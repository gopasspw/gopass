# `version` command

The `version` command prints version and build time information for the `gopass`
binary.

## Synopsis

```
$ gopass version
```

## Modes of operation

* Print the version and build time
* Check whether a newer release is available (unless disabled)

## Flags

None.

## Details

By default `gopass version` additionally checks GitHub for a newer release and
prints a notice if your version is out of date. The check is skipped when:

* `CHECKPOINT_DISABLE` is set to any non-empty value;
* the `updater.check` configuration option is set to `false`;
* the running binary is a development build (`+HEAD`); or
* gopass runs under OpenBSD's `pledge(2)`.

The check times out after two seconds to avoid delaying the version output.

See [`update`](update.md) to install an available update.

## Exit codes

| Code | Meaning                     |
|-----:|-----------------------------|
| 0    | Version printed successfully |
| 3    | User aborted                 |

See [docs/exit-codes.md](../exit-codes.md) for the full table.
