# `audit` command

The `audit` command will decrypt all secrets and scan for weak passwords or other common flaws.

## Synopsis

```
$ gopass audit
```

## Excludes

You can exclude certain secrets from the audit by adding a `.gopass-audit-exclude` file to the secret. The file should contain a list of RE2 patters to exclude, one per line. For example:

```
# Lines starting with # are ignored. Trailing comments are not supported.
# Exclude all secrets in the pin folder.
# Note: These are RE2, not Glob patterns!
pin/.*
# Literal matches are also valid RE2 patterns
test_folder/ignore_this
# Gopass internally uses forward slashes as path separators, even on Windows. So no need to escape backslashes.
```

## Password strength backends

| Backend                                         | Description                                                            |
|-------------------------------------------------|------------------------------------------------------------------------|
| [`crunchy`](https://github.com/muesli/crunchy)  | Crunchy password strength checker                                      |
| `name`                                          | Checks if password equals the name of the secret                       |
