# `grep` command

The `grep` command works like the Unix `grep` tool. It decrypts all secrets
and performs a substring or regexp match on the given pattern.

## Synopsis

```
$ gopass grep foobar
```

## Modes of operation

* Search for the given pattern in all secrets

## Flags

| Flag       | Aliases | Description                                    |
|------------|---------|------------------------------------------------|
| `--regexp` | `-r`    | Parse the pattern as a RE2 regular expression. |
