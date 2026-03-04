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
`--sep` | `--xs` | Word separator for multi-word passwords.
`--lang` | `--xl` | Language to generate password from. Currently only supports english (en, default).
