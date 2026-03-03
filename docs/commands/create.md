# `create` command

The `create` command creates a new secret using a set of built-in or custom templates.
It implements a wizard that guides inexperienced users through the secret creating.

The main design goal of this command was to guide users through the creation of a secret
and asking for the necessary information to create a reasonable secret location.

## Synopsis

```bash
gopass create
gopass create --store=foo
```

## Modes of operation

* Create a new secret using a wizard

## Templates

`gopass create` will look for files ending in `.yml` in the folder `.gopass/create` inside
the selected store (by default the root store).

To add new templates to the wizard add templates to this folder.

Example:

```bash
$ cat $(gopass config mounts.path)/.gopass/create/aws.yml
---
priority: 5
name: "AWS"
prefix: "aws"
name_from:
  - "org"
  - "user"
welcome: "🧪 Creating AWS credentials"
attributes:
  - name: "org"
    type: "string"
    prompt: "Organization"
    min: 1
  - name: "user"
    type: "string"
    prompt: "User"
    min: 1
  - name: "password"
    type: "password" # hide input
    prompt: "Password"
    charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%&*"
    min: 10
    strict: true # ensure at least one char from each detected class (upper, lower, digit, symbol)
  - name: "comment"
    type: "string"
    prompt: "Comments"
```

## Template Attributes

Template attributes support the following fields:

| Field           | Type   | Description                                                                                                                                                                                                                                                                  |
|-----------------|--------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`          | string | The name of the attribute. This will be used as the key in the secret's YAML data.                                                                                                                                                                                          |
| `type`          | string | The type of attribute. Supported values: `string`, `hostname`, `password`.                                                                                                                                                                                                  |
| `prompt`        | string | The prompt text to display to the user.                                                                                                                                                                                                                                     |
| `charset`       | string | For password type: Custom character set to use when generating the password. If not specified, standard character classes will be used.                                                                                                                                      |
| `min`           | int    | Minimum length validation for the attribute value.                                                                                                                                                                                                                           |
| `max`           | int    | Maximum length validation for the attribute value.                                                                                                                                                                                                                           |
| `always_prompt` | bool   | For password type: Always prompt for the password instead of offering password generation. Default: `false`.                                                                                                                                                                 |
| `strict`        | bool   | For password type with `charset`: Enforce that all detected character classes (uppercase, lowercase, digits, symbols) present in the charset are represented in the generated password. Similar to `--strict` in `gopass generate`. Default: `false`.                        |

## Flags

| Flag      | Aliases | Description                                                      |
|-----------|---------|------------------------------------------------------------------|
| `--store` | `-s`    | Select the store to use. Will be used to look up user templates. |
| `--force` | `-f`    | For overwriting existing entries.                                |
| `--print` | `-p`    | Print the password to STDOUT.                                    |
