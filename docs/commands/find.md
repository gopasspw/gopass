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

