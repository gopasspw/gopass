# `merge` command

The `merge` command implements a merge workflow to help deduplicate secrets. It
merges multiple existing secrets into a single destination entry and optionally
removes the sources afterwards.

## Synopsis

```
$ gopass merge to/entry from/entry-1 from/entry-2
```

## Modes of operation

* Merge two or more secrets into one, editing the result in `$EDITOR`
* Merge two or more secrets into one, unattended (`--force`)

## Flags

| Flag      | Aliases | Description                                     |
|-----------|---------|-------------------------------------------------|
| `--delete`| `-d`    | Remove merged entries (default: `true`).        |
| `--force` | `-f`    | Skip editor, merge entries unattended.          |

## Details

The command requires exactly one destination (which may already exist) and at
least one source (which must exist; multiple sources are allowed). The bodies of
all involved entries are concatenated, each preceded by a `# Secret: <name>`
comment, and loaded into your `$EDITOR`. Once you save and exit, the result is
written to the destination.

If the content is unchanged, nothing is written. With `--force` the editor is
skipped and the concatenated content is written directly.

Because the merged result is re-parsed, a merged entry that happens to contain a
`---` separator may be treated as YAML. See [secrets.md](../secrets.md) for the
parsing rules.

If `--delete` is enabled (the default), the source entries are removed after the
destination has been written. The deletion happens after the merge has been
committed.
