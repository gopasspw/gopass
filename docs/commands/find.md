# `find` command

The `find` command will attempt to do a simple substring match on the names of all secrets.
If there is a single match it will directly invoke `show` and display the result.
If there are multiple matches a selection will be shown.

Note: The find command will not fall back to a fuzzy search.

## Synopsis

```
$ gopass find entry
$ gopass find --regex '^foo/.*'
$ gopass find --json entry
```

## Flags

| Flag       | Aliases | Description                                                   |
|------------|---------|---------------------------------------------------------------|
| `--unsafe` | `-u`    | In the case of an exact match, display the password even if `safecontent` is enabled. |
| `--regex`  | `-r`    | Interpret the pattern as a regular expression instead of a plain substring match. |
| `--json`   | `-j`    | Output matches as a JSON array.                               |

## Exit codes

| Code | Meaning |
|-----:|---------|
| 0 | Matches found (or a single match displayed) |
| 2 | No search pattern provided; or invalid regular expression |
| 3 | User aborted interactive selection |
| 10 | No matching secret found |
| 13 | Store contents could not be listed |

See [docs/exit-codes.md](../exit-codes.md) for the full table.

