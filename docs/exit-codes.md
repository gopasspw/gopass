# gopass Exit Codes

gopass uses structured numeric exit codes so that shell scripts and other
programs can distinguish between different failure modes without parsing
human-readable output.

The numeric values are **stable** — they will never be renumbered between
releases. New codes may be added at the end of the list in future versions.

Run `gopass --exit-codes` to print this table at any time.

## Full Table

| Code | Name | Description |
|-----:|------|-------------|
| 0 | `OK` | Success — no error |
| 1 | `Unknown` | Unclassified or unexpected error |
| 2 | `Usage` | Bad invocation: wrong arguments or flags |
| 3 | `Aborted` | User deliberately aborted the operation |
| 4 | `Unsupported` | Operation is not supported |
| 5 | `AlreadyInitialized` | Store is already initialized |
| 6 | `NotInitialized` | Store is not initialized |
| 7 | `Git` | Git operation failed |
| 8 | `Mount` | Substore mount operation failed |
| 9 | `NoName` | No name provided for the entry |
| 10 | `NotFound` | Requested secret not found |
| 11 | `Decrypt` | Reading or decrypting a secret failed |
| 12 | `Encrypt` | Writing or encrypting a secret failed |
| 13 | `List` | Listing store contents failed |
| 14 | `Audit` | Audit found one or more issues |
| 15 | `Fsck` | Integrity check found errors |
| 16 | `Config` | Configuration error (reserved, not yet emitted) |
| 17 | `Recipients` | Recipient operation failed |
| 18 | `IO` | Miscellaneous I/O error |
| 19 | `GPG` | Miscellaneous GPG error (reserved, not yet emitted) |
| 20 | `Hook` | Hook execution failed |
| 21 | `Doctor` | Doctor found one or more failing checks |

## Per-Command Summary

### `show`

| Code | When |
|-----:|------|
| 0 | Secret displayed successfully |
| 10 | Secret not found |
| 11 | Secret could not be decrypted |

### `insert`

| Code | When |
|-----:|------|
| 0 | Secret inserted successfully |
| 11 | Existing secret could not be read for append/key-insert |
| 12 | Secret could not be encrypted and saved |

### `generate`

| Code | When |
|-----:|------|
| 0 | Password generated and stored successfully |
| 12 | Generated secret could not be encrypted and saved |

### `find`

| Code | When |
|-----:|------|
| 0 | Matches found (or single match displayed) |
| 10 | No matching secret found |

### `delete`

| Code | When |
|-----:|------|
| 0 | Secret deleted successfully |
| 10 | Secret not found |

### `audit`

| Code | When |
|-----:|------|
| 0 | No issues found |
| 14 | One or more weak passwords or issues detected |

### `fsck`

| Code | When |
|-----:|------|
| 0 | Store integrity OK |
| 15 | One or more integrity errors found |

### `doctor`

| Code | When |
|-----:|------|
| 0 | All checks passed |
| 21 | One or more checks failed |

## Scripting Example

```bash
gopass show myservice/password
case $? in
  0)  echo "OK" ;;
  10) echo "Secret not found" ;;
  11) echo "Decryption failed" ;;
  *)  echo "Unexpected error ($?)" ;;
esac
```
