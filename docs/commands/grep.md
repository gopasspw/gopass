# `grep` command

The `grep` command works like the Unix `grep` tool. It decrypts all secrets
and performs a substring or regexp match on the given pattern.

## Synopsis

```
$ gopass grep foobar
```

## Modes of operations

* Search for the given pattern in all secrets

## Flags

None.
Flag | Aliases | Description
---- | ------- | -----------
`--regexp` | | Parse the pattern as a RE2 regular expression.
