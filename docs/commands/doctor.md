# `doctor` command

The `doctor` command checks your gopass installation for common configuration
issues and reports the results. With the `--recipients` flag, it performs a
detailed recipient consistency diagnostic across all stores.

## Synopsis

```
$ gopass doctor [--verbose]
$ gopass doctor --recipients [--verbose]
```

## Description

Runs a series of diagnostic checks on the gopass installation. Exits with a
non-zero status if any check fails, which makes it suitable for scripting.

Checks performed:

| Check | Description |
|---|---|
| GPG binary | Verifies `gpg` is in `PATH` when a store uses GPG encryption |
| age binary | Verifies `age` is in `PATH` when a store uses age encryption |
| git binary | Verifies `git` is in `PATH` when a store uses the gitfs backend |
| git identity | Checks that `user.name` and `user.email` are set in git config for each git-backed store |
| store permissions | Checks that each store directory exists and is not world-writable |
| recipient keys | Checks that all recipient keys are valid and not expired |
| git remote | Warns (but does not fail) if a git-backed store has no remote configured |

### `--recipients` mode

When `--recipients` is passed, `gopass doctor` performs a detailed,
per-store, per-recipient consistency diagnostic:

- **Non-canonical IDs** — warns when a recipient is stored with an ambiguous
  identifier (e.g. email address) instead of its full fingerprint. Suggests
  running `gopass recipients canonicalize`.
- **Missing keys** — errors when a recipient is neither in the local keyring
  nor in `.public-keys/`.
- **`.public-keys/` only** — warns when a recipient's key is only available in
  the store's `.public-keys/` directory (not in the local keyring). Suggests
  running `gopass sync`.
- **Expired / unusable keys** — detects when a key is present in the keyring
  by fingerprint but not usable by direct ID lookup (expired). Recommends
  `gopass recipients update` followed by `gopass sync`.

The diagnostic is read-only (no decryption needed) and safe to run at any time.

## Flags

| Flag | Aliases | Description |
|---|---|---|
| `--verbose` | `-v` | Show passing checks in addition to warnings and errors |
| `--recipients` | | Run a detailed recipient consistency diagnostic across all stores |

## Exit codes

| Code | Meaning |
|---|---|
| 0 | All checks passed |
| non-zero | One or more checks failed |

## Examples

```
# Run all checks, show only failures
$ gopass doctor

# Run all checks, show every result
$ gopass doctor --verbose

# Run detailed recipient diagnostic
$ gopass doctor --recipients

# Run recipient diagnostic with full detail
$ gopass doctor --recipients --verbose
```
