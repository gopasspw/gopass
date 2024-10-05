# `show` command

The `show` command is the most important and most frequently used command.
It allows displaying and copying the content of the secrets managed by gopass.

## Synopsis

```
$ gopass show entry
$ gopass show entry key
$ gopass show entry --qr
$ gopass show entry --password
```

## Modes of operation

* Show the whole entry: `gopass show entry`
* Show a specific key of the given entry: `gopass show entry key` (only works for key-value or YAML secrets)

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--clip` | `-c` | Copy the password value into the clipboard and don't show the content.
`--alsoclip` | `-C` | Copy the password value into the clipboard and show the content.
`--qr` | | Encode the password field as a QR code and print it. Note: When combining with `-c`/`-C` the unencoded password is copied. Not the QR code.
`--unsafe` | `-u` | Display unsafe content (e.g. the password) even when the `safecontent` option is set. No-op when `safecontent` is `false`.
`--password` | `-o` | Display only the password. For use in scripts. Takes precedence over other flags.
`--revision` | `-r` | Display a specific revision of the entry. Use an exact version identifier from `gopass history` or the special `-<N>` syntax. Does not work with native (e.g. git) refs.
`--noparsing` | `-n` | Do not parse the content, disable YAML and Key-Value functions.
`--chars` | | Display selected characters from the password.

## Details

This section describes the expected behaviour of the `show` command with respect to different combinations of flags and
config options.

Note: This section describes the expected behaviour, not necessarily the observed behaviour.
If you notice any discrepancies please file a bug and we will try to fix it.

TODO: We need to specify the expectations around new lines.

* When no flag is set the `show` command will display the full content of the secret and will parse it to support key-value lookup and YAML entries.
  If the `safecontent` option is set to `true` any secret fields (current default is only `password`) are replaced with a random number of '*' characters (length: 5-10).
  Using the `--unsafe` flag will reveal these fields even if `safecontent` is enabled. `--password` takes precedence of `safecontent=true` as well and displays only the password.
* The `--noparsing` flag will disable all parsing of the output, this can help debugging YAML secrets for example, where `key: 0123` actually parses into octal for 83.
* The `--clip` flag will copy the value of the `Password` field to the clipboard and doesn't display any part of the secret.
* The `--alsoclip` option will copy the value of the `Password` field but also display the secret content depending on the `safecontent` setting, i.e. obstructing the `Password` field if `safecontent` is `true` or just displaying it if not.
* The `--qr` flags operates complementary to other flags. It will *additionally* format the value of the `Password` entry as a QR code and display it. Other than that it will honor the other options, e.g. `gopass show --qr` will display the QR code *and* the whole secret content below. One special case is the `-o` flag, this flag doesn't make a lot of sense in combination, so if both `--qr` and `-o` are given only the QR code will be displayed.
* Since gopass plans to supports different RCS backends we do not support arbitrary git refs as arguments to the `--revision` flag. Using those might work, but this is explicitly not supported and bug reports will be closed as `wont-fix`. There are two issues with using arbitrary git refs is that (a) this doesn't work with non-git RCS backends and (b) git versions a whole repository, not single files. So the revision `HEAD^`
  might not have any changes for a given entry. Thus we only support specifc revisions obtained from `gopass history` or our custom syntax `-N` where N is an integer identifying a specific commit before `HEAD` (cf. `HEAD~N`).

## Parsing and secrets

Secrets are stored on disk as provided, but are parsed upon display to provide extra features such as the ability
to show the value of a key using:  `gopass show entry key`.

The secrets are split into 3 categories:
 - the plain type, which is just a plain secret without key-value capabilities
    ```
    this is a plain secret
    using multiple lines

    and that's it
    ```
    gets parsed to the same value


 - the key-value type, which allows to query the value of a specific key. This does not preserve ordering.
    ```
    this is a KV secret
    where: the first line is the password
    and: the keys are separated from their value by :

    and maybe we have a body text
    below it
    ```
    will be parsed into (with `safecontent` enabled):
   ```
    and: the keys are separated from their value by :
    where: the first line is the password


    and maybe we have a body text
    below it
    ```


 - the YAML type which implements YAML support, which means that secrets are parsed as per YAML standard.
    ```
    s3cret
    ---
    invoice: 0123
    date   : 2001-01-23
    bill-to: &id001
        given  : Bob
        family : Doe
    ship-to: *id001
    ```
   will be parsed into (with `safecontent` enabled):
   ```
    bill-to: map[family:Doe given:Bob]
    date: 2001-01-23 00:00:00 +0000 UTC
    invoice: 83
    ship-to: map[family:Doe given:Bob]
    ```
   Note how the `0123` is interpreted as octal for 83. If you want to store a string made of digits such as a numerical
   username, it should be enclosed in string delimiters: `username: "0123"` will always be parsed as the string `0123`
   and not as octal.

Both the key-value and the YAML format support so-called "unsafe-keys", which is a key-value that allows you to specify keys that should be hidden when using `gopass show` with `gopass config safecontent` set to true.
E.g:
```
supersecret
---
age: 27
secret: The rabbit outran the tortoise
name: John Smith
unsafe-keys: age,secret
```
will display (with safecontent enabled):
```
age: *****
name: John Smith
secret: *****
unsafe-keys: age,secret

```
unless it is called with `gopass show -n` that would disable parsing of the body, but still hide the password, or `gopass show -f` that would show everything that was hidden, including the password.

You can read more about secrets formats in its [documentation](docs/secrets.md).

Notice that if the option `parsing` is disabled in the config, then all secrets are handled as plain secrets.
