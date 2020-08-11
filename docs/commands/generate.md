# `generate` command

The `generate` command is used to generate a new password and store it into the password store.

Note: If you only want generate a password without storing it in the store, use the `pwgen` command.

## Synopsis

```
$ gopass generate entry [length]
$ gopass generate entry key [length]
```

## Modes of operation

* Generate a new entry with a new password, e.g. a new login. Setting the `Password` field, `gopass generate entry [chars]`
* Re-generating a new password and setting it in the `Password` field of an existing entry
* Generate a new password and setting it to a new key of an existing secret, e.g. `gopass generate entry key [chars]
* Re-generate a new password for an existing key in an existing entry

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--clip` | `-c` | Copy the generated password into the clipboard. Default: Value of `autoclip`
`--print` | `-p` | Print the generated password to the terminal. Default: false.
`--force` | `-f` | Force overwriting an existing entry.
`--edit` | `-e` | Generate a password and open the entry for editing in `$EDITOR`.
`--generator` | `-g` | Choose of of the available password generators, desribed below. Default: `cryptic`
`--symbols` | `-s` | Include symbols in the generated password (default: `false`)
`--strict` | | Ensure each requested character class is actually included. Without this option all requested classes can be included, but not necessarily are. (default: `false`)
`--sep` | | Word separator for multi-word generators.
`--lang`| | Language for word-based generators.

## Password Generators

Use `--generator` to select one of the available password generators:

Generator | Description
--------- | -----------
`cryptic` | The default generator yields cryptic passwords that should work with most sites. Use `--symbols` and `--strict` if the site has specific requirements. Please note that we auto-detect the correct rules for some sites. The length argument specifies the number of characters.
`xkcd` | Use an [XKCD#936](https://xkcd.com/936/) style password. Use `--lang` and `--sep` to refine it's behaviour. The length argument specifies the number of words.
`memorable` | Generate a memorable password. The length argument specifies the minimum lenght of characters. Please note that the password might be longer if not all necessary rules were satisfied by the minimum length solution.
`external` | Use the external generator from `$GOPASS_EXTERNAL_PWGEN`

## Relevant configuration options

* `autoclip` only applies to `generate`. If set the generated password is automatically copied to the clipboard - unless `--clip` is `false`
* `safecontent` will suppress printing of the password, unless `-p` is set. The password will not be copied, unless `-c` or the `autoclip` option are set.
