# `sum` command

The `sum` command decodes a Base64 encoded secret and computes the SHA256
checksum over the decoded data. This is useful to verify the integrity of a
secret inserted with [`cat`](cat.md), [`fscopy`](fscopy.md) or
[`fsmove`](fsmove.md).

## Synopsis

```
$ gopass sum my/private.key
```

The command is also available as `gopass sha` and `gopass sha256`.

## Modes of operation

* Compute the SHA256 checksum of the decoded content of a secret

## Flags

None.

## Exit codes

| Code | Meaning                                     |
|-----:|---------------------------------------------|
| 0    | Checksum computed and printed successfully  |
| 2    | No secret name provided                     |
| 11   | Secret could not be read or decrypted       |

See [docs/exit-codes.md](../exit-codes.md) for the full table.
