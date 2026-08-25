# `export` command

The `export` command decrypts secrets and writes them to a portable JSON
document. It is meant to help migrate a store into another password manager:
the JSON keeps the password, the body and all key/value attributes so nothing
is lost, and you can transform it into whatever format the target expects.

By default the whole store is exported. Provide one or more folder prefixes to
export only those subfolders.

The output contains your secrets in **plaintext**, so treat the file (or the
terminal it is printed to) with the same care as the secrets themselves.

## Synopsis

```
$ gopass export
$ gopass export websites/
$ gopass export --out backup.json
```

## Flags

| Flag    | Aliases | Description                                     |
|---------|---------|-------------------------------------------------|
| `--out` | `-o`    | Write the export to this file instead of stdout. When writing to a file the mode is set to `0600`. |

## Output format

The document is a JSON array of objects:

```json
[
  {
    "name": "websites/example",
    "password": "correct horse battery staple",
    "body": "some notes",
    "attributes": {
      "url": ["https://example.com"],
      "user": ["alice"]
    }
  }
]
```

Empty fields are omitted. Attributes are encoded as lists because a key can
hold more than one value.
