# `generate` command

The `generate` command is used to generate a new password and store it into the password store.

Note: If you only want generate a password without storing it in the store, use the `pwgen` command.

## Synopsis

```sh
gopass generate entry [length]
gopass generate entry key [length]
```

## Modes of operation

* Generate a new entry with a new password, e.g. a new login. Setting the `Password` field, `gopass generate entry [chars]`
* Re-generating a new password and setting it in the `Password` field of an existing entry
* Generate a new password and setting it to a new key of an existing secret, e.g. `gopass generate entry key [chars]`
* Re-generate a new password for an existing key in an existing entry

## Flags

| Flag          | Aliases | Description                                                                                                                                                        |
|---------------|---------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--clip`      | `-c`    | Copy the generated password into the clipboard. Default: Value of `generate.autoclip`                                                                             |
| `--print`     | `-p`    | Print the generated password to the terminal. Default: false.                                                                                                      |
| `--force`     | `-f`    | Force overwriting an existing entry.                                                                                                                               |
| `--edit`      | `-e`    | Generate a password and open the entry for editing in `$EDITOR`.                                                                                                   |
| `--generator` | `-g`    | Choose one of the available password generators, described below. Default: `cryptic`                                                                               |
| `--symbols`   | `-s`    | Include symbols in the generated password (default: `false`)                                                                                                       |
| `--strict`    |         | Ensure each requested character class is actually included. Without this option all requested classes can be included, but not necessarily are. (default: `false`) |
| `--xkcd-sep`  | `--sep`, `--xkcdsep` | Word separator for multi-word generators.                                                                                                               |
| `--xkcd-lang` | `--lang`, `--xkcdlang` | Language for word-based generators.                                                                                                                 |
| `--xkcd-capitalize` | `--xkcdcapitalize` | Capitalize the first letter of each word when using the `xkcd` generator. Equivalent to setting `pwgen.xkcd-capitalize = true` in config.  |
| `--xkcd-numbers` | `--xkcdnumbers` | Append a random number to each word when using the `xkcd` generator. Equivalent to setting `pwgen.xkcd-numbers = true` in config.              |

## Password Generators

Use `--generator` to select one of the available password generators:

| Generator   | Description                                                                                                                                                                                                                                                                      |
|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `cryptic`   | The default generator yields cryptic passwords that should work with most sites. Use `--symbols` and `--strict` if the site has specific requirements. Please note that we auto-detect the correct rules for some sites. The length argument specifies the number of characters. |
| `xkcd`      | Use an [XKCD#936](https://xkcd.com/936/) style password. Use `--xkcd-lang` and `--xkcd-sep` to refine its behaviour. The length argument specifies the number of words.                                                                                                                   |
| `memorable` | Generate a memorable password. The length argument specifies the minimum length of characters. Please note that the password might be longer if not all necessary rules were satisfied by the minimum length solution.                                                               |
| `external`  | Use the external generator from `$GOPASS_EXTERNAL_PWGEN`                                                                                                                                                                                                                         |

## Exit codes

| Code | Meaning |
|-----:|---------|
| 0 | Password generated and stored successfully |
| 2 | Length argument is not a valid positive integer |
| 3 | User declined to overwrite existing secret |
| 9 | No secret name provided |
| 12 | Generated secret could not be encrypted and saved |
| 18 | Generated password could not be copied to clipboard |

See [docs/exit-codes.md](../exit-codes.md) for the full table.

## Relevant configuration options

* `generate.autoclip` only applies to `generate`. If set the generated password is automatically copied to the clipboard - unless `--clip` is explicitly set to `--clip=false`
* `show.safecontent` will suppress printing of the password, unless `-p` is set. The password will not be copied, unless `-c` or the `generate.autoclip` option are set.

## Templates

When creating a new entry gopass will look for the most specific template
by going up in the secret path looking for a file called `.pass-template`.

If any such file is found it will be used to pre-populate the generated
secret.
