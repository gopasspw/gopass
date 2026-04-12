# `doctor` command

The `doctor` command checks your gopass installation for common configuration issues and reports the results.

## Synopsis

```
$ gopass doctor [--verbose]
```

## Description

Runs a series of diagnostic checks on the gopass installation. Exits with a non-zero status if any check fails, which makes it suitable for scripting.

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

## Flags

| Flag | Aliases | Description |
|---|---|---|
| `--verbose` | `-v` | Show passing checks in addition to warnings and errors |

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
```
