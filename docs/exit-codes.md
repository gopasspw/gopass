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

### `audit`

| Code | When |
|-----:|------|
| 0 | No issues found |
| 1 | Audit run itself failed |
| 13 | Store contents could not be listed |
| 14 | One or more weak passwords or issues detected |
| 18 | Report file could not be written |

### `cat`

| Code | When |
|-----:|------|
| 0 | Content read or written successfully |
| 9 | No secret name provided |
| 11 | Secret could not be decrypted for output |
| 18 | I/O error reading from stdin or writing content |

### `clone`

| Code | When |
|-----:|------|
| 0 | Store cloned successfully |
| 2 | No repository URL provided |
| 5 | Root store is already initialized (cannot clone over it) |
| 6 | Root store not initialized when trying to add a mount |
| 7 | Git clone operation failed |
| 8 | Adding the cloned store as a mount failed |
| 18 | Could not read repository URL or mount point interactively |

### `config`

| Code | When |
|-----:|------|
| 0 | Config value displayed or set successfully |
| 2 | Wrong number of arguments |
| 1 | Config value could not be set |

### `convert`

| Code | When |
|-----:|------|
| 0 | Store converted successfully |
| 2 | Unknown backend name given for `--storage` or `--crypto` |
| 10 | Named store not found |
| 1 | Conversion failed |

### `copy` / `cp`

| Code | When |
|-----:|------|
| 0 | Secret or directory copied successfully |
| 2 | Not enough arguments |
| 3 | Destination exists and user declined overwrite |
| 10 | Source path does not exist |
| 13 | Could not list source subtree |
| 18 | Copy operation failed |

### `create`

| Code | When |
|-----:|------|
| 0 | Secret created successfully |
| 1 | Interactive create wizard failed to initialize |
| 3 | User cancelled the create wizard |
| 18 | Generated password could not be copied to clipboard |

### `delete` / `rm`

| Code | When |
|-----:|------|
| 0 | Secret deleted successfully |
| 2 | No name provided; or multiple names with `-r`; or target is a directory without `-r` |
| 4 | `--key` value conflicts with an existing secret name |
| 10 | Secret not found |
| 18 | Delete or YAML-key removal failed |
| 20 | Post-delete hook execution failed |

### `doctor`

| Code | When |
|-----:|------|
| 0 | All checks passed |
| 21 | One or more checks failed |

### `edit`

| Code | When |
|-----:|------|
| 0 | Secret saved successfully |
| 2 | No name provided |
| 11 | Existing secret could not be decrypted before editing |
| 12 | Edited secret could not be encrypted and saved |
| 17 | Recipients for the secret are invalid |
| 20 | Pre-edit hook execution failed |

### `env`

| Code | When |
|-----:|------|
| 0 | Program executed successfully with secrets in environment |
| 2 | No program to execute; conflicting input-mode flags; non-secret path used with `--stdin` |
| 10 | Named secret not found |
| 13 | Store contents could not be listed |

### `find`

| Code | When |
|-----:|------|
| 0 | Matches found (or single match displayed) |
| 2 | No search pattern provided; or invalid regular expression |
| 3 | User aborted interactive selection |
| 10 | No matching secret found |
| 13 | Store contents could not be listed |

### `fsck`

| Code | When |
|-----:|------|
| 0 | Store integrity OK |
| 10 | Specified filter path not found |
| 15 | One or more integrity errors found |

### `generate`

| Code | When |
|-----:|------|
| 0 | Password generated and stored successfully |
| 2 | Length argument is not a valid positive integer |
| 3 | User declined to overwrite existing secret |
| 9 | No secret name provided |
| 12 | Generated secret could not be encrypted and saved |
| 18 | Generated password could not be copied to clipboard |

### `git`

| Code | When |
|-----:|------|
| 0 | Git operation completed successfully |
| 2 | Not enough arguments for `git remote add` or `git remote rm` |
| 7 | VCS init or remote push operation failed |

### `grep`

| Code | When |
|-----:|------|
| 0 | Search completed (results printed or nothing matched) |
| 2 | No search argument provided; or invalid regular expression |
| 13 | Store contents could not be listed |

### `history`

| Code | When |
|-----:|------|
| 0 | Revision history displayed successfully |
| 2 | No secret name provided |
| 10 | Secret does not exist |
| 1 | Revision list could not be retrieved |

### `init`

| Code | When |
|-----:|------|
| 0 | Store initialized successfully |
| 1 | Store initialization failed |
| 6 | Store is not initialized (checked via `IsInitialized`) |

### `insert`

| Code | When |
|-----:|------|
| 0 | Secret inserted successfully |
| 1 | Editor could not be launched for buffer-based insert |
| 2 | YAML key could not be parsed |
| 3 | Secret exists and user declined overwrite |
| 9 | No secret name provided |
| 11 | Existing secret could not be read for append/key-insert |
| 12 | Secret could not be encrypted and saved |
| 18 | I/O error reading from stdin or prompting for password |

### `link`

| Code | When |
|-----:|------|
| 0 | Link created successfully |
| 2 | Not enough arguments |

### `list` / `ls`

| Code | When |
|-----:|------|
| 0 | Store contents listed successfully |
| 10 | Specified filter path not found |
| 13 | Store tree could not be built |

### `merge`

| Code | When |
|-----:|------|
| 0 | Secrets merged successfully |
| 2 | Missing source or destination argument |
| 11 | Source secret could not be decrypted |
| 12 | Merged secret could not be encrypted and saved |

### `mounts`

| Code | When |
|-----:|------|
| 0 | Mount added or removed successfully |
| 2 | No alias provided for `mounts remove`; or wrong argument count for `mounts add` |
| 8 | Mount operation failed |

### `move` / `mv`

| Code | When |
|-----:|------|
| 0 | Secret or directory moved successfully |
| 2 | Not exactly two arguments provided |
| 3 | Destination exists and user declined overwrite |
| 1 | Move operation failed |

### `otp`

| Code | When |
|-----:|------|
| 0 | OTP token generated successfully |
| 2 | No secret name provided |
| 10 | Secret contains no OTP key |
| 1 | OTP URI not found or token calculation failed |
| 18 | Token could not be copied to clipboard |

### `process`

| Code | When |
|-----:|------|
| 0 | Template processed and output written successfully |
| 2 | No file argument provided |
| 18 | Template file could not be read or processed |

### `recipients`

| Code | When |
|-----:|------|
| 0 | Recipient operation completed successfully |
| 3 | User aborted interactive key selection |
| 13 | Recipient list could not be retrieved |
| 17 | Recipient could not be added or removed |

### `reorg`

| Code | When |
|-----:|------|
| 0 | Store reorganized successfully |
| 2 | Secret count changed in editor; or invalid move in editor |
| 3 | User aborted confirmation |
| 4 | Running in non-interactive mode |
| 7 | Git commit after reorganization failed |
| 13 | Store contents could not be listed |

### `show`

| Code | When |
|-----:|------|
| 0 | Secret displayed successfully |
| 1 | Revision list could not be retrieved; or QR encoding failed |
| 2 | No name provided |
| 10 | Secret not found; or requested YAML key, line, or password field not found |
| 11 | Secret could not be decrypted |

### `sync`

| Code | When |
|-----:|------|
| 0 | Synchronization completed (sync does not emit specific non-zero codes; errors are logged as warnings) |

### `templates`

| Code | When |
|-----:|------|
| 0 | Template operation completed successfully |
| 2 | No template name provided for `templates rm` |
| 10 | Template not found for `templates rm` |
| 13 | Template list could not be retrieved |
| 18 | Template could not be read or written |

### `update`

| Code | When |
|-----:|------|
| 0 | gopass is up to date or update applied successfully |
| 1 | Update check or download failed |

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
