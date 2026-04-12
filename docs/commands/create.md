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

On first run, gopass writes two built-in templates (website login and PIN code) to this
folder. You can modify them or add your own alongside them.

To add a new template create a YAML file in `.gopass/create/` and commit it:

```bash
# open the store directory
cd "$(gopass config mounts.path)"
mkdir -p .gopass/create
$EDITOR .gopass/create/aws.yml
git add .gopass/create/aws.yml && git commit -m "Add AWS credential template"
```

## Template Structure

Each template file is a YAML document with the following top-level fields:

| Field        | Type     | Required | Description                                                                                                                                                          |
|--------------|----------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`       | string   | yes      | Human-readable name shown in the wizard's selection menu (e.g., `"Website login"`).                                                                                  |
| `priority`   | int      | no       | Sort order in the wizard menu. Lower numbers appear first. Default: `0`. The two built-in templates use priorities `0` (website) and `1` (PIN).                      |
| `prefix`     | string   | yes      | Directory inside the store where the new secret will be saved (e.g., `"websites"` → secret stored under `websites/<name>`).                                          |
| `name_from`  | []string | no       | List of attribute names whose values are joined to form the secret's file name. If empty, the user is prompted for a path. Values are sanitised with `CleanFilename`. |
| `welcome`    | string   | no       | Message printed at the start of the wizard for this template. Supports Unicode/emoji.                                                                                 |
| `attributes` | list     | yes      | Ordered list of attribute definitions (see [Attribute Fields](#attribute-fields) below).                                                                              |

Example skeleton:

```yaml
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
    type: "password"
    prompt: "Password"
    charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%&*"
    min: 10
    strict: true
  - name: "comment"
    type: "string"
    prompt: "Comments"
```

## Attribute Fields

Each entry in `attributes` supports the following fields:

| Field           | Type   | Applies to          | Description                                                                                                                                                                                                    |
|-----------------|--------|---------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`          | string | all                 | Key under which the value is stored in the secret's YAML body. Also used as the human-readable label when `prompt` is omitted.                                                                                  |
| `type`          | string | all                 | Controls the input behaviour. One of `string`, `hostname`, `password`, or `multiline` — see [Attribute Types](#attribute-types) below.                                                                          |
| `prompt`        | string | all                 | Override the text shown to the user. Defaults to the `name` field with the first letter upper-cased.                                                                                                            |
| `min`           | int    | string, hostname, password | Minimum acceptable length. Validation is skipped when `0` (default). For `password` with auto-generation the minimum is passed to the generator.                                                         |
| `max`           | int    | string, hostname, password | Maximum acceptable length. Validation is skipped when `0` (default).                                                                                                                                     |
| `charset`       | string | password            | Explicit character set for generated passwords. When omitted, the standard mixed-class generator is used. Ignored when the user opts out of generation.                                                         |
| `always_prompt` | bool   | password            | When `true`, skip the "Generate Password?" prompt and always ask the user to type one in. Default: `false`.                                                                                                     |
| `strict`        | bool   | password            | When `true` (and `charset` is set), every character class detected in `charset` (upper, lower, digit, symbol) must appear at least once in the generated password. Equivalent to `gopass generate --strict`. Default: `false`. |

## Attribute Types

### `string`

Prompts for a single-line text value. The value is stored as-is under the attribute's
`name` key in the secret's YAML body.

```yaml
- name: "username"
  type: "string"
  prompt: "Username"
  min: 1
  max: 64
```

### `hostname`

Like `string`, but additionally:

* Extracts the hostname component from the entered value (e.g., `https://example.com/login` → `example.com`).
* The extracted hostname is used as the `name_from` component if this attribute is listed there.
* Looks up password-change URLs via the built-in `pwrules` database and stores them as `password-change-url` if found.

```yaml
- name: "url"
  type: "hostname"
  prompt: "Website URL"
  min: 1
```

### `password`

Prompts the user `"Generate Password?"`. If yes, generates a password using the standard
gopass generator (respecting `charset` and `strict`). If no, asks the user to type one
(with confirmation) and applies `min`/`max` length validation.

The password is stored as the **first line** of the secret (the gopass password field),
not as a YAML key.

```yaml
- name: "password"
  type: "password"
  prompt: "Password"
  charset: "0123456789"   # digits only, e.g. for PIN codes
  min: 4
  max: 8
  always_prompt: true     # skip the "generate?" question
```

### `multiline`

Opens the user's `$EDITOR` (or the editor configured via `gopass config core.editor`)
with any existing gopass template for this attribute pre-filled. The full editor content
is written verbatim to the secret body.

Useful for SSH keys, certificates, or annotated notes.

```yaml
- name: "notes"
  type: "multiline"
  prompt: "Additional notes"
```

## File Naming Convention

Template files must end in `.yml` or `.yaml`. gopass ignores files with other extensions.
There is no enforced naming convention, but the built-in templates follow the pattern
`<priority>-<prefix>.yml` (e.g., `0-websites.yml`, `1-pin.yml`). Using the same convention
makes the order predictable in directory listings.

## Flags

| Flag      | Aliases | Description                                                      |
|-----------|---------|------------------------------------------------------------------|
| `--store` | `-s`    | Select the store to use. Will be used to look up user templates. |
| `--force` | `-f`    | For overwriting existing entries.                                |
| `--print` | `-p`    | Print the password to STDOUT.                                    |
