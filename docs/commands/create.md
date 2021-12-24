# `create` command

The `create` command creates a new secret using a set of built-in or custom templates.
It implements a wizard that guides inexperienced users through the secret creating.

The main design goal of this command was to guide users through the creation of a secret
and asking for the necessary information to create a reasonable secret location.

## Synopsis

```
$ gopass create
$ gopass create --store=foo
```

## Modes of operation

* Create a new secret using a wizard

## Templates

`gopass create` will look for files ending in `.yml` in the folder `.gopass/create` inside
the selected store (by default the root store).

To add new templates to the wizard add templates to this folder.

Example:

```
$ cat $(gopass config path)/.gopass/create/aws.yml
---
priority: 5
name: "AWS"
prefix: "aws"
name_from:
  - "org"
  - "user"
welcome: "🧪 Creating AWS credentials"
attributes:
  org:
    type: "string"
    prompt: "Organization"
    min: 1
  user:
    type: "string"
    prompt: "User"
    min: 1
  password:
    type: "password"
    prompt: "Password"
  comment:
    type: "string"
    prompt: "Comments"
```

## Flags

Flag | Aliases | Description
---- | ------- | -----------
`--store` | `-s` | Select the store to use. Will be used to look up user templates.
`--force` | `-f` | For overwriting existing entries.
`--print` | `-p` | Print the password to STDOUT.
