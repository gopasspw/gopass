# Configuration

## Environment Variables

Some configuration options are only available through setting environment variables.

| **Option**           | **Type** | **Description** |
| -------------------- | ---------| --------------- |
| `CHECKPOINT_DISABLE` | `bool`   | Set to any non-empty value to disable calling the GitHub API when running `gopass version`. |
| `GOPASS_DEBUG`       | `bool`   | Set to any non-empty value to enable verbose debug output |

## Configuration Options

During startup, gopass will look for a configuration file at `$HOME/.config/gopass/config.yml`. If one is not present, it will create one. If the config file already exists, it will attempt to parse it and load the settings. If this fails, the program will abort. Thus, if gopass is giving you trouble with a broken or incompatible configuration file, simply rename it or delete it.

All configuration options are also available for reading and writing through the subcommand `gopass config`.

* To display all values: `gopass config`
* To display a single value: `gopass config autosync`
* To update a single value: `gopass config autosync false`
* As many other subcommands this command accepts a `--store` flag to operate on a given sub-store.

This is a list of options available:

| **Option**    | **Type** | Description |
| ------------- | -------- | ----------- |
| `askformore`  | `bool`   | If enabled - it will ask to add more data after use of `generate` command. |
| `autoimport`  | `bool`   | Import missing keys stored in the pass repo without asking. |
| `autosync`    | `bool`   | Always do a `git push` after a commit to the store. Makes sure your local changes are always available on your git remote. |
| `cliptimeout` | `int`    | How many seconds the secret is stored when using `-c`. |
| `noconfirm`   | `bool`   | Do not confirm recipient list when encrypting. |
| `path`        | `string` | Path to the root store. |
| `safecontent` | `bool`   | Only output _safe content_ (i.e. everything but the first line of a secret) to the terminal. Use _copy_ (`-c`) to retrieve the password in the clipboard. |
