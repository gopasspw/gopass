# `cat` command

The `cat` command is used to pipe password in and out of STDIN and STDOUT
respectively. As it is intended to be used with binary data, it encodes the
data-stream to store it.

## Synopsis

```bash
$ echo "test" | gopass cat test/new
$ gopass cat test/new
```

## Modes of operation

* Create a new entry with data-stream from STDIN
* Change an existing entry to data-stream from STDIN
* Retrive encoded data from password-store and echo it to STDOUT

Cat is intended to work with binary data, so it accepts any kind of stream from
STDIN. It reads the binary-stream from STDIN and encodes it Base64 and saves it
in the password store encoded, with some metadata about the input-stream and the
used encoding (currently only Base64 supported).

### Example
```
$ echo "234" | gopass cat test/new
$ gopass show -f test/new
Secret: test/new


content-disposition: attachment; filename="STDIN"
content-transfer-encoding: Base64
MjM0Cg==
$ gopass cat test/new
234
```

### Differences to `insert`

In contrast to `insert` it handles any kind of data-stream from STDIN and
encodes it.
Drawback: you can not just simply read the password with `gopass show`.

## Flags

This command has currently no supported flags except the gopass globals.
