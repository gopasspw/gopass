# `unclip` command

The `unclip` command is an internal helper that clears the clipboard after a
secret has been copied to it. It is not intended to be invoked by users
directly; [`show`](show.md) and [`otp`](otp.md) spawn it automatically when
copying a value.

Note: This command is hidden from the regular help output.

## Synopsis

```
$ gopass unclip --timeout 45
```

## Modes of operation

* Wait for the configured timeout and then clear the clipboard

## Flags

| Flag        | Description                                        |
|-------------|----------------------------------------------------|
| `--timeout` | Time to wait, in seconds.                          |
| `--force`   | Clear the clipboard even if the checksum mismatches.|

## Details

The expected clipboard name and checksum are passed to the helper via the
`GOPASS_UNCLIP_NAME` and `GOPASS_UNCLIP_CHECKSUM` environment variables.

Before clearing, the helper re-reads the clipboard and verifies a checksum of
the expected value. This ensures it only erases secrets gopass itself placed
there, and not content the user copied from another application in the meantime.
`--force` disables that check.

See [`show`](show.md) for the clipboard timeouts and the `core.cliptimeout`
configuration option.

## Exit codes

| Code | Meaning                              |
|-----:|--------------------------------------|
| 0    | Clipboard cleared successfully       |
| 18   | Clearing the clipboard failed        |

See [docs/exit-codes.md](../exit-codes.md) for the full table.
