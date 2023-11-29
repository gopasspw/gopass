# Configuration

## Environment Variables

Some configuration options are only available through setting environment variables.

| **Option**                   | **Type** | **Description**                                                                                                                                                   |
|------------------------------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `CHECKPOINT_DISABLE`         | `bool`   | Set to any non-empty value to disable calling the GitHub API when running `gopass version`.                                                                       |
| `GOPASS_AUTOSYNC_INTERVAL`   | `int`    | Set this to the number of days between autosync runs.                                                                                                             |
| `GOPASS_CHARACTER_SET`       | `bool`   | Set to any non-empty value to restrict the characters used in generated passwords                                                                                 |
| `GOPASS_CLIPBOARD_CLEAR_CMD` | `string` | Use an external command to remove a password from the clipboard. See [GPaste](usecases/gpaste.md) for an example                                                  |
| `GOPASS_CLIPBOARD_COPY_CMD`  | `string` | Use an external command to copy a password to the clipboard. See [GPaste](usecases/gpaste.md) for an example                                                      |
| `GOPASS_CONFIG_NO_MIGRATE`   | `bool`   | Do not attempt to migrate old gopass configs and option names                                                                                                     |
| `GOPASS_CONFIG_NOSYSTEM`     | `bool`   | Do not read `/etc/gopass/config` (if it exists)                                                                                                                   |
| `GOPASS_CONFIG`              | `string` | Set this to the absolute path to the configuration file                                                                                                           |
| `GOPASS_CPU_PROFILE`         | `string` | Path to write a CPU Profile to. Use `go tool pprof` to visualize.                                                                                                 |
| `GOPASS_DEBUG_FILES`         | `string` | Comma separated filter for console debug output (files)                                                                                                           |
| `GOPASS_DEBUG_FUNCS`         | `string` | Comma separated filter for console debug output (functions)                                                                                                       |
| `GOPASS_DEBUG_LOG_SECRETS`   | `bool`   | Set to any non-empty value to enable logging of credentials                                                                                                       |
| `GOPASS_DEBUG_LOG`           | `string` | Set to a filename to enable debug logging (only set GOPASS_DEBUG to log to stderr)                                                                                |
| `GOPASS_DEBUG`               | `bool`   | Set to any non-empty value to enable verbose debug output, by default on stderr, unless GOPASS_DEBUG_LOG is set                                                   |
| `GOPASS_EXTERNAL_PWGEN`      | `string` | Use an external password generator. See [Features](features.md#using-custom-password-generators) for details                                                      |
| `GOPASS_FORCE_CHECK`         | `string` | (internal) Force the updater to check for updates. Used for testing.                                                                                              |
| `GOPASS_FORCE_UPDATE`        | `bool`   | Set to any non-empty value to force an update (if available)                                                                                                      |
| `GOPASS_GPG_BINARY`          | `string` | Set this to the absolute path to the GPG binary if you need to override the value returned by `gpgconf`, e.g. [QubesOS](https://www.qubes-os.org/doc/split-gpg/). |
| `GOPASS_GPG_OPTS`            | `string` | Add any extra arguments, e.g. `--armor` you want to pass to GPG on every invocation                                                                               |
| `GOPASS_HOMEDIR`             | `string` | Set this to the absolute path of the directory containing the `.config/` tree                                                                                     |
| `GOPASS_HOOK`                | `int`    | (internal) Set when invoking hook scripts.                                                                                                                        |
| `GOPASS_MEM_PROFILE`         | `string` | Path to write a Memory Profile to. Use `go tool pprof` to visualize.                                                                                              |
| `GOPASS_NO_AUTOSYNC`         | `bool`   | Set this to `true` to disable autosync. Deprecated. Please use `core.autosync`                                                                                    |
| `GOPASS_NO_NOTIFY`           | `bool`   | Set to any non-empty value to prevent notifications                                                                                                               |
| `GOPASS_NO_REMINDER`         | `bool`   | Set to any non-empty value to prevent reminders                                                                                                                   |
| `GOPASS_PW_DEFAULT_LENGTH`   | `int`    | Set to any integer value larger than zero to define a different default length in the `generate` command. By default the length is 24 characters.                 |
| `GOPASS_UMASK`               | `octal`  | Set to any valid umask to mask bits of files created by gopass                                                                                                    |
| `GOPASS_UNCLIP_CHECKSUM`     | `string` | (internal) Used between gopass and it's unclip helper.                                                                                                            |
| `GOPASS_UNCLIP_NAME`         | `string` | (internal) Used between gopass and it's unclip helper.                                                                                                            |
| `PWGEN_RULES_FILE`           | `string` | (internal) Used for testing the pwgen rules generator.                                                                                                            |

Variables not exclusively used by gopass:

| **Option**             | **Type** | **Description**                                                                                        |
|------------------------|----------|--------------------------------------------------------------------------------------------------------|
| `PASSWORD_STORE_DIR`   | `string` | absolute path containing the password store (a directory). Only supported during initialization!       |
| `PASSWORD_STORE_UMASK` | `string` | Set to any valid umask to mask bits of files created by gopass (GOPASS_UMASK has precedence over this) |
| `EDITOR`               | `string` | command name to execute for editing password entries                                                   |
| `PAGER`                | `string` | the pager program used for `gopass list`. See [Features](features.md#auto-pager) for details           |
| `GIT_AUTHOR_NAME`      | `string` | name of the author, used by the rcs backend to create a commit                                         |
| `GIT_AUTHOR_EMAIL`     | `string` | email of the author, used by the rcs backend to create a commit                                        |
| `NO_COLOR`             | `bool`   | disable color output. See [no-color.org](https://no-color.org) for more information.                   |

## Configuration Options

During start up, gopass will look for a configuration file at `$HOME/.config/gopass/config` on unix-like systems or at `%APPDATA%\gopass\config` on Windows. If one is not present, it will create one. If the config file already exists, it will attempt to parse it and load the settings. If this fails, the program will abort. Thus, if gopass is giving you trouble with a broken or incompatible configuration file, simply rename it or delete it.

All configuration options are also available for reading and writing through the sub-command `gopass config`.

* To display all values: `gopass config`
* To display a single value: `gopass config generate.autoclip`
* To update a single value: `gopass config generate.autoclip false`
* As many other sub-commands this command accepts a `--store` flag to operate on a given sub-store, provided the sub-store is a remote one.

### Configuration format

`gopass` uses a configuration format inspired by and mostly compatible with the configuration format used by git. We support
different configuration sources that take precedence over each other, just like [git](https://mirrors.edge.kernel.org/pub/software/scm/git/docs/git-config.html).

#### Configuration precedence

* Hard-coded presets apply if nothing else is set
* System-wide configuration file allows operators or package maintainers to supply system-wide defaults in `/etc/gopass/config`.
* User-wide (aka. global) configuration allows to set per-user settings. This is the closest equivalent to the old gopass configs. Located in `$HOME/.config/gopass/config`
* Per-store (aka. local) configuration allow to set per-store settings, e.g. read-only. Located in `<STORE_DIR>/config`.
* Per-store unversioned (aka `config.worktree`) configuration allows to override versioned per-store settings, e.g. disabling read-only. Located in `<STORE_DIR>/config.worktree`
* Environment variables (or command line flags) override all other values. Read from `GOPASS_CONFIG_KEY_n` and `GOPASS_CONFIG_VALUE_n` up to `GOPASS_CONFIG_COUNT`. Command line flags take precedence over environment variables.

### Configuration options

This is a list of available options:

| **Option**                      | **Type** | Description                                                                                                                                                                                                                        | *Default*                           |
|---------------------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------|
| `age.usekeychain`               | `bool`   | Use the OS keychain to cache age passphrases.                                                                                                                                                                                      | `false`                             |
| `audit.concurrency`             | `int`    | Number of concurrent audit workers.                                                                                                                                                                                                | ``                                  |
| `audit.hibp-dump-file`          | `string` | Specify to a HIBPv2 Dump file (sorted) if you want `audit` to check password hashes against this file.                                                                                                                             | `None`                              |
| `audit.hibp-use-api`            | `bool`   | Set to true if you want `gopass audit` to check your secrets against the public HIBPv2 API. Use with caution. This will leak a few bit of entropy.                                                                                 | `false`                             |
| `autosync.interval`             | `string` | AutoSync interval, for example `2d`, `4h`, `2m` (for days, hours, minutes). A plain number without suffix is taken as days.                                                                                                        | `3`                                 |
| `core.autoimport`               | `bool`   | Import missing keys stored in the pass repository without asking.                                                                                                                                                                  | `false`                             |
| `core.autopush`                 | `bool`   | Always do a `git push` after a commit to the store. Makes sure your local changes are always available on your git remote.                                                                                                         | `true`                              |
| `core.autosync`                 | `bool`   | Automatically sync (fetch & push) the git remote on an interval.                                                                                                                                                                   | `true`                              |
| `core.cliptimeout`              | `int`    | How many seconds the secret is stored when using `-c`. Setting this to `0` disables auto-clear.                                                                                                                                    | `45`                                |
| `core.exportkeys`               | `bool`   | Export public keys of all recipients to the store.                                                                                                                                                                                 | `true`                              |
| `core.nocolor`                  | `bool`   | Do not use color.                                                                                                                                                                                                                  | `false`                             |
| `core.nopager`                  | `bool`   | Do not invoke a pager to display long lists.                                                                                                                                                                                       | `false`                             |
| `core.noreminder`               | `bool`   | Set to true to disable periodic update reminder. Equivalent to setting `GOPASS_NO_REMINDER`.                                                                                                                                       | `false`                             |
| `core.notifications`            | `bool`   | Enable desktop notifications.                                                                                                                                                                                                      | `true`                              |
| `core.post-hook`                | `string` | This hook is executed after any command invocation.                                                                                                                                                                                | `None`                              |
| `core.pre-hook`                 | `string` | This hook is executed before any command invocation.                                                                                                                                                                               | `None`                              |
| `core.readonly`                 | `bool`   | Disable writing to a store. Note: This is just a convenience option to prevent accidental writes. Enforcement can only happen on a central server (if repos are set up around a central one).                                      | `false`                             |
| `create.default-username`       | `string` | The settings allows users to specify the default username for logins created with `gopass create`.                                                                                                                                 | `None`                              |
| `create.post-hook`              | `string` | This hook is executed right after the secret creation. If the hook exits with a non-zero exit value the generated secret is discarded.                                                                                             | `None`                              |
| `create.pre-hook`               | `string` | This hook is executed right before the secret creation during `gopass create`.                                                                                                                                                     | `None`                              |
| `delete.post-hook`              | `string` | This hook is run right after removing a record with `gopass rm`.                                                                                                                                                                   | `None`                              |
| `domain-alias.<from>.insteadOf` | `string` | Alias from domain to the string value of this entry. Currently not supported at the local config level.                                                                                                                            | ``                                  |
| `edit.auto-create`              | `bool`   | Automatically create new secrets when editing.                                                                                                                                                                                     | `false`                             |
| `edit.editor`                   | `string` | This setting controls which editor is used when opening a file with `gopass edit`. Currently not supported at the local config level. It takes precedence over the `$EDITOR` environment variable. This setting can contain flags. | `None`                              |
| `edit.post-hook`                | `string` | This hook is run right after editing a record with `gopass edit`.                                                                                                                                                                  |
| `edit.pre-hook`                 | `string` | This hook is run right before editing a record with `gopass edit`.                                                                                                                                                                 |
| `generate.autoclip`             | `bool`   | Always copy the password created with `gopass generate`.                                                                                                                                                                           | `false`                             |
| `generate.generator`            | `string` | Default password generator. `xkcd`, `memorable`, `external` or ``.                                                                                                                                                                 | ``                                  |
| `generate.length`               | `int`    | Default length for generated password.                                                                                                                                                                                             | `24`                                |
| `generate.strict`               | `bool`   | Use strict mode for generated password.                                                                                                                                                                                            | `false`                             |
| `generate.symbols`              | `bool`   | Include symbols in generated password.                                                                                                                                                                                             | `false`                             |
| `mounts.path`                   | `string` | Path to the root store.                                                                                                                                                                                                            | `$XDG_DATA_HOME/gopass/stores/root` |
| `recipients.check`              | `bool`   | Check recipients hash. The global config option takes precedence over local ones here for security reasons.                                                                                                                        | `false`                             |
| `recipients.hash`               | `string` | SHA256 hash of the recipients file. Used to notify the user when the recipients files change. Not set, nor read at the local level for security reasons.                                                                           | ``                                  |
| `recipients.remove-extra-keys`  | `bool`   | Remove extra recipients during key import. Not supported at the local level for security reasons.                                                                                                                                  | `false`                             |
| `show.autoclip`                 | `bool`   | Autoclip in `gopass show` by default.                                                                                                                                                                                              | `false`                             |
| `show.post-hook`                | `string` | This hook is run right after displaying a secret with `gopass show`.                                                                                                                                                               | `None`                              |
| `show.safecontent`              | `bool`   | Only output *safe content* (i.e. everything but the first line of a secret) to the terminal. Use *copy* (`-c`) to retrieve the password in the clipboard, or *force* (`-f`) to still print it.                                     | `false`                             |
| `updater.check`                 | `bool`   | Check for updates when running `gopass version`. Only supported as a global, system or env config option, not at the local level.                                                                                                  | `true`                              |
| `output.internal-pager`         | `bool`   | Use the internal pager `ov`.                                                                                                                                                                                                       | `false`                             |
| `pwgen.xkcd-sep`                | `string` | `xkcd` password generator separator.                                                                                                                                                                                               | ` `                                 |
| `pwgen.xkcd-lang`               | `string` | `xkcd` password generator language.                                                                                                                                                                                                | `en`                                |
| `pwgen.xkcd-capitalize`         | `bool`   | Capitalize the first character of each word. Default is `false`, except when the separator is empty.                                                                                                                               | `false`                             |
| `pwgen.xkcd-numbers`            | `bool`   | Add random numbers after each word.                                                                                                                                                                                                | `false`                             |
| `pwgen.xkcd-len`                | `bool`   | The number of words to generated.                                                                                                                                                                                                  | `4`                                 |

Furthermore, the following table list the legacy options (starting with v1.15.9) and their new names, their migration should be automatic
unless you've set them at the system level or using Env variables, in which case you'll need to migrate them manually:

| **Legacy option name** | **New option name** | **Version of migration** |
|------------------------|---------------------|--------------------------|
| `core.showsafecontent` | `show.safecontent`  | v1.15.9                  |
| `core.autoclip`        | `generate.autoclip` | v1.15.9                  |
| `core.showautoclip`    | `show.autoclip`     | v1.15.9                  |
