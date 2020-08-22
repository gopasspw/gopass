# `templates` commands

The template support is one of the more unique `gopass` features. It allows
password stores to define templates that will automatically apply to any new
secret create at or below the template path. For example this can be useful
to generate a new email password and the salted hash at the same time. Or a
PostgreSQL password with the custom salted hash. This is certainly a feature
that's not used very often, but if used corretly it can greatly reduce the
toil of some common operations.

## Synopsis

```
$ gopass templates
$ gopass templates show template
$ gopass templates edit template
$ gopass templates remove template
```

## Flags

None.

## Examples

### Compute the salted hash for the password

```
Password: {{ .Content }}
SSHA256: {{ .Content | ssha256 }}
```

### Compute the SQL statements to create a new PostgreSQL user

```
{{ .Content }}
---
sql:  |
  CREATE ROLE {{ .Name }} LOGIN PASSWORD '{{ .Content }}';
  GRANT {{ .Name }} TO {{ .Name }};
  ALTER USER {{ .Name }} SET search_path = '{{ .Name }}';
```

## Template functions

Function | Example | Description
-------- | ------- | -----------
`md5sum` | `{{ .Content \| md5sum }}` | Calculate the hex md5sum of the input.
`sha1sum` | `{{ .Content \| sha1sum }}` | Calculate the hex sha1sum of the input.
`md5crypt` | `{{ .Content \| md5crypt }}` | Calculate the md5crypt of the input.
`ssha` | `{{ .Content \| ssha }}` | Calculate the salted SHA-1 of the input.
`ssha256` | `{{ .Content \| ssha256 }}` | Calculate the salted SHA-256 of the input.
`ssha512` | `{{ .Content \| ssha512 }}` | Calculate the slated SHA-512 of the input.
`get` | `{{ get "foo/bar" }}` | Insert the full secret.
`getpw` | `{{ getpw "foo/bar" }}` | Insert the value of the password field from the given secret.
`getval` | `{{ getval "foo/bar" "baz" }}` | Insert the value of the named filed from the given secret.

## Template variables

Note: These examples assume being evaluated for the secret `foo/bar/baz` and
the generated password `VerySecure`.

Name | Example | Description
---- | ------- | -----------
`Dir` | `foo/bar` | The directory containing the secret.
`Path` | `foo/bar/baz` | The path or full name of the secret.
`Name` | `baz` | The last element of the path or short name of the secret.
`Content` | `VerySecure` | The generated password.
