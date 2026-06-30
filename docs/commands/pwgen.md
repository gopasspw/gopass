# `pwgen` command

The `pwgen` command implements a subset of the features of the Unix/Linux
`pwgen` command line tool. It aims to eventually support most of the `pwgen`
flags and mirror it's behaviour. It is mainly implemented as a curtosy for
Windows users.

## Modes of operation

* Generate a few dozen random passwords with the chosen length

## Usage

```bash
gopass pwgen [optional length]
```

## Synopsis

```bash
gopass pwgen
gopass pwgen 24
```

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--no-numerals` | `-0` | Do not include numerals in the generated passwords.
`--one-per-line` | `-1` | Print one password per line.
`--xkcd` | `-x` | Use multiple random english words combined to a password.
`--xkcd-sep` | `--sep`, `--xkcdsep` | Word separator for multi-word passwords.
`--xkcd-lang` | `--lang`, `--xkcdlang` | Language to generate password from. Currently only supports english (en, default).
`--xkcd-capitalize` | `--xkcdcapitalize` | Capitalize the first letter of each word in the generated xkcd password.
`--xkcd-numbers` | `--xkcdnumbers` | Add a random number to the end of the generated xkcd password.
`--memorable` | `-m` | Use the memorable (word-based) password generator. The length is a minimum (output may be longer). Incompatible with `--no-numerals`.
`--memorable-capitalize` | `--memorablecapitalize` | Capitalize (some) words in the generated memorable password. Implies `--memorable`.

## Notes

* With `--memorable`, the requested length is a **minimum** — the generated password is usually longer because whole words are concatenated.
* `--memorable` always includes digits (one per word), so it is incompatible with `--no-numerals`; the command errors out instead of silently ignoring it.
* `--memorable` ignores `--ambiguous`. `--memorable` and `--xkcd` are mutually exclusive (combining them is an error).
