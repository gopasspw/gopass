# `find` command

The `find` command will attempt to do a simple substring match on the names of all secrets.
If there is a single match it will directly invoke `show` and display the result.
If there are multiple matches a selection will be shown.

Note: The find command will not fall back to a fuzzy search.

## Synopsis

```
$ gopass find entry
$ gopass find -f entry
$ gopass find -c entry
```

## Flags

| Flag       | Aliases | Description                                                   |
|------------|---------|---------------------------------------------------------------|
| `--clip`   | `-c`    | Copy the password into the clipboard.                         |
| `--unsafe` | `-u`    | Display any unsafe content, even if `safecontent` is enabled. |
| `--regex`  | `-r`    | Interpret the pattern as a regular expression instead of a plain substring match. |

## Exit codes

| Code | Meaning |
|-----:|---------|
| 0 | Matches found and displayed |
| 10 | No matching secret found |

See [docs/exit-codes.md](../exit-codes.md) for the full table.

