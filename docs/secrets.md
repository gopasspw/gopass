# Secrets

`gopass` supports different secret formats. This page documents the different formats.

## Key-Value

The new [Key-Value implementation](../pkg/gopass/secrets/akv.go) fully maintains the secret format
when parsing but still does offer (limited) support for Key-Value operations, i.e. retrieving keys,
listing keys and writing (the first instance) keys. Some multi-value operations are not directly
supported. Use `gopass edit` for these.

Note: The parser will ensure that every parsed secret contains a terminating newline. Even if the
input didn't have one.

Format:

```text
Line | Description
   0 | Password
 1-n | Body
```

The parser uses the `: ` separator to identify potential Key Value pairs.
When updating existing pairs only the first value will be rewritten.
New pairs are always appended at the end.

## YAML

Note: Using YAML is discouraged as YAML can be troublesome for humans, e.g. parsing of unquoted numbers.

The [YAML Format](../pkg/gopass/secrets/yaml.go) is used if there is a YAML marker (`---`) after the body:

```text
YAML is a gopass secret that contains a parsed YAML data structure.
This is a legacy data type that is discouraged for new users as YAML
is neither trivial nor intuitive for users manually editing secrets (e.g.
unquoted phone numbers being parsed as octal and such).

Format
------
Line  | Description
    0 | Password
  1-n | Body
  n+1 | Separator ("---")
  n+2 | YAML content.
```

## Deprecated formats

`gopass` used to support different secret formats. These were deemed suboptimal and retired.
We still support parsing of these formats but don't write them anymore.

### MIME

`gopass` briefly had a custom secrets format based on multipart MIME. This did prove to be even more troublesome for humans than YAML so it was quickly deprecated.

These secrets are identified by a well known header.

```text
GOPASS-SECRET-1.0
Password: ...
[other headers]

[Body]
```

### Plain

The old KV implementation had some limitations so we did sometimes fall back to the old Plain format. With the new KV implementation this is not necessary anymore so this was removed.
