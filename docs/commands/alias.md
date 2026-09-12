# `alias` command

The `alias` command prints the domain aliases configured in the `pwrules`
database. These aliases map a known site to the domains its password generation
rules apply to.

## Synopsis

```
$ gopass alias
```

## Modes of operation

* Print all known domain aliases

## Flags

None.

## Details

The output lists each alias and the domains it expands to, for example
`- google.com -> gmail.com, youtube.com`. The list is sorted by alias name.

Note: This command is unrelated to `domain-alias.<from>.insteadOf`, which is a
configuration option for looking up secrets under several domain names. See
[Configuration](../config.md) for that option.

## Exit codes

| Code | Meaning                    |
|-----:|----------------------------|
| 0    | Aliases printed successfully |

See [docs/exit-codes.md](../exit-codes.md) for the full table.
