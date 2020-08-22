# `gopass` command

Calling `gopass` without any command argument is a common entry point and
has two different modes.

## Synopsis

```
$ gopass
$ gopass entry
```

## Modes of operation

* Invoked without any arguments `gopass` will start an interactive REPL shell. This includes zero-setup command completion and passphrase caching (for non-GPG backends).
* Invoked with one argument it will perform a (fuzzy) search and display a list of matches or the secret directly (if exactly one match).
* Invoked with two arguments it will do search and if there is a match display the named key.

## Flags

Note: `gopass` intentionally does not support any flags. If you need to use any flag consider using `gopass show` instead.
