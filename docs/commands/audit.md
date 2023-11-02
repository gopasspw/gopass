# `audit` command

The `audit` command will decrypt all secrets and scan for weak passwords or other common flaws.

## Synopsis

```
$ gopass audit
```

## Password strength backends

| Backend                                         | Description                                                            |
|-------------------------------------------------|------------------------------------------------------------------------|
| [`zxcvbn`](https://github.com/nbutton23/zxcvbn) | [zxcvbn](https://github.com/dropbox/zxcvbn) password strength checker. |
| [`crunchy`](https://github.com/muesli/crunchy)  | Crunchy password strength checker                                      |
| `name`                                          | Checks if password equals the name of the secret                       |


