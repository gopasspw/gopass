# Configuration

## Environment Variables

Some configuration options are only available through setting environment variables.

| **Option**    | *Type*    | Description |
| ------------- | --------- | ----------- |
| `CHECKPOINT_DISABLE` | `bool`    | Set to any non-empty value to disable calling the GitHub API when running `gopass version`. |
| `GOPASS_DEBUG` | `bool` | Set to any non-empty value to enable verbose debug output |

## Configuration Options

`gopass` create a new configration file if not already present during startup.
If a config exists it will attempt to parse and load the settings. If this
fails program execution aborts. If `gopass` can not start because  of broken
or incompatible config file move it away (make a backup) and start fresh.

All configuration options are also available for reading and writing through
the subcommand `gopass config`.

* To display all values type: `gopass config`.
* To display a single value type: `gopass config autosync`
* To update a single value type: `gopass config autosync false`
* As many other subcommands this command accepts a `--store` flag to operate on a given sub-store.

This is a list of options available for `gopass`:

| **Option**    | *Type*    | Description |
| ------------- | --------- | ----------- |
| `askformore`  | `bool`    | If enabled - it will ask to add more data after use of `generate` command. |
| `autoimport`  | `bool`    | Import missing keys stored in the pass repo without asking. |
| `autosync`    | `bool`    | Always do a `git push` after a commit to the store. Makes sure your local changes are always available on your git remote. |
| `cliptimeout` | `int`     | How many seconds the secret is stored when using `-c`. |
| `noconfirm`   | `bool`    | Do not confirm recipient list when encrypting. |
| `path`        | `string`  | Path to the root store. |
| `safecontent` | `bool`    | Only output _safe content_ (i.e. everything but the first line of a secret) to the terminal. Use _copy_ (`-c`) to retrieve the password in the clipboard. |

