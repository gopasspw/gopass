# Configuration

## Environment Variables

Some configuration options are only available through setting environment variables.

| **Option**              | **Type** | **Description**                                                                                              |
|-------------------------|----------|--------------------------------------------------------------------------------------------------------------|
| `CHECKPOINT_DISABLE`    | `bool`   | Set to any non-empty value to disable calling the GitHub API when running `gopass version`.                  |
| `GOPASS_DEBUG`          | `bool`   | Set to any non-empty value to enable verbose debug output                                                    |
| `GOPASS_UMASK`          | `octal`  | Set to any valid umask to mask bits of files created by gopass                                               |
| `GOPASS_GPG_OPTS`       | `string` | Add any extra arguments, e.g. `--armor` you want to pass to GPG on every invocation                          |
| `GOPASS_EXTERNAL_PWGEN` | `string` | Use an external password generator. See [Features](features.md#using-custom-password-generators) for details |
| `GOPASS_NOCOLOR`        | `bool`   | Set to true to disable colored output                                                                        |
| `GOPASS_CHARACTER_SET`  | `bool`   | Set to any non-empty value to restrict the characters used in generated passwords                            |
| `GOPASS_CONFIG`         | `string` | Set this to the absolute path to the configuration file                                                     |
| `GOPASS_HOMEDIR`        | `string` | Set this to the absolute path of the directory containing the `.config/` tree                               |
| `GOPASS_FORCE_UPDATE`   | `bool`   | Set to any non-empty value to force an update (if available)                                                 |
| `GOPASS_NO_NOTIFY`      | `bool`   | Set to any non-empty value to prevent notifications                                                          |

Variables not exclusively used by gopass

| **Option**             | **Type** | **Description**                                                                                        |
|------------------------|----------|--------------------------------------------------------------------------------------------------------|
| `PASSWORD_STORE_DIR`   | `string` | absolute path containing the password store (a directory)                                              |
| `PASSWORD_STORE_UMASK` | `string` | Set to any valid umask to mask bits of files created by gopass (GOPASS_UMASK has precedence over this) |
| `EDITOR`               | `string` | command name to execute for editing password entries                                                  |
| `PAGER`                | `string` | the pager program used for `gopass list`. See [Features](features.md#auto-pager) for details          |
| `GIT_AUTHOR_NAME`      | `string` | name of the author, used by the rcs backend to create a commit                                         |
| `GIT_AUTHOR_EMAIL`     | `string` | email of the author, used by the rcs backend to create a commit                                        |

## Configuration Options

During start up, gopass will look for a configuration file at `$HOME/.config/gopass/config.yml`. If one is not present, it will create one. If the config file already exists, it will attempt to parse it and load the settings. If this fails, the program will abort. Thus, if gopass is giving you trouble with a broken or incompatible configuration file, simply rename it or delete it.

All configuration options are also available for reading and writing through the sub-command `gopass config`.

* To display all values: `gopass config`
* To display a single value: `gopass config autosync`
* To update a single value: `gopass config autosync false`
* As many other sub-commands this command accepts a `--store` flag to operate on a given sub-store.

This is a list of available options:

| **Option**       | **Type** | Description |
| ---------------- | -------- | ----------- |
| `askformore`     | `bool`   | If enabled - it will ask to add more data after use of `generate` command. |
| `autoclip`       | `bool`   | Always copy the password created by `pass generate`. |
| `autoprint`      | `bool`   | Always print the password created by `pass generate`. |
| `autoimport`     | `bool`   | Import missing keys stored in the pass repository without asking. |
| `autosync`       | `bool`   | Always do a `git push` after a commit to the store. Makes sure your local changes are always available on your git remote. |
| `concurrency`    | `int`    | Number of threads to use for batch operations (such as reencrypting). |
| `cliptimeout`    | `int`    | How many seconds the secret is stored when using `-c`. |
| `noconfirm`      | `bool`   | Do not confirm recipient list when encrypting. |
| `path`           | `string` | Path to the root store. |
| `editrecipients` | `bool`   | Modify recipients when creating and editing passwords. |
| `exportkeys`     | `bool`   | Export public keys of all recipients to the store. |
| `recipient_hash` | `map`    | Map of recipient ids to their hashes. |
| `safecontent`    | `bool`   | Only output _safe content_ (i.e. everything but the first line of a secret) to the terminal. Use _copy_ (`-c`) to retrieve the password in the clipboard. |
| `usesymbols`     | `bool`   | If enabled - it will use symbols when generating passwords. |
| `notifications`  | `bool`   | Enable desktop notifications. |
| `nocolor`        | `bool`   | Do not use color. |
| `nopager`        | `bool`   | Do not invoke a pager to display long lists. |
