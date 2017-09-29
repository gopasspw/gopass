# Features

## Standard Features

### Setup of a Store

If you don't have an existing password store or your store is completely empty you have to initialize it.

Please note: This document uses the term *password store* to refer to a directory (usually `$HOME/.password-store`) which is managed by either `gopass` or `pass`. This is entirely different from any OS-level credential store, your GPG Keyring or your SSH Keys.

Choose one of:
```bash
$ gopass init gopher@golang.org
$ gopass init A3683834
$ gopass init 1E52C1335AC1F4F4FE02F62AB5B44266A3683834    # preferred
```

This will encrypt any secret which is added to the store for the recipient.

#### Clone an existing store

If you already have a _password-store_ that you would clone to the system you can take one short cut:

```bash
$ gopass clone git@example.com/pass.git
$ gopass clone git@example.com/pass-work.git work # clone as mount called: work
```

This runs `git clone` in the background and also sets up the config file if necessary.

A second parameter tells gopass to clone and mount it to the store.
In the example above the repository would have been cloned to `$HOME/.password-store-work`.
Afterwards the directory would have been mounted as `work`.

Please note that the repository must contain an already initialized password
store. You can initialize a new store with `gopass init --path /path/to/store`.

### Adding secrets

Let's say you want to create an account.

| Website    | User   |
| ---------- | ------ |
| golang.org | gopher |


#### Type in a new secret

```bash
$ gopass insert golang.org/gopher
Enter secret for golang.org/gopher:       # hidden
Retype secret for golang.org/gopher:      # hidden
gopass: Encrypting golang.org/gopher for these recipients:
 - 0xB5B44266A3683834 - Gopher <gopher@golang.org>

Do you want to continue? [yn]: y
```

#### Generate a new secret

```bash
$ gopass generate golang.org/gopher
How long should the secret be? [20]:
gopass: Encrypting golang.org/gopher for these recipients:
 - 0xB5B44266A3683834 - Gopher <gopher@golang.org>

Do you want to continue? [yn]: y
The generated secret for golang.org/gopher is:
Eech4ahRoy2oowi0ohl
```

```bash
$ gopass generate golang.org/gopher 16    # length as paramenter
gopass: Encrypting golang.org/gopher for these recipients:
 - 0xB5B44266A3683834 - Gopher <gopher@golang.org>

Do you want to continue? [yn]: y
The generated password for golang.org/gopher is:
Eech4ahRoy2oowi0ohl

```

The `generate` command will ask for any missing arguments, like name of the secret or the length. If you don't want the password to be displayed use
the `-c` flag to copy it to your clipboard.

### Edit a secret

```bash
$ gopass edit golang.org/gopher
```

The `edit` command uses the `$EDITOR` environment variable to start your preferred editor where
you can easily edit multi-line content. `vim` will be the default if `$EDITOR` is not set.

### Listing existing secrets

You can list all entries of the store:

```bash
$ gopass
gopass
├── golang.org
│   └── gopher
└── emails
    ├── user@example.com
    └── user@justwatch.com
```

If your terminal supports colors the output will use ANSI color codes to highlight directories
and mounted sub stores. Mounted sub stores include the mount point and source directory. See
below for more details on mounts and sub stores.

### Show a secret

```bash
$ gopass golang.org/gopher

Eech4ahRoy2oowi0ohl
```

The default action of `gopass` is show. It also accepts the `-c` flag to copy the content of
the secret directly to the clipboard.

Since it may be dangerous to always display the password on `gopass` calls, the `safecontent`
setting may be set to `true` to allow one to display only the rest of the password entries by
default and display the whole entry, with password, only when the `-f` flag is used.

#### Copy secret to clipboard

```bash
$ gopass -c golang.org/gopher

Copied golang.org/gopher to clipboard. Will clear in 45 seconds.
```

### Removing secret

```bash
$ gopass rm golang.org/gopher
```

`rm` will remove a secret from the store. Use `-r` to delete a whole folder.
Please note that you **can not** remove a folder containing a mounted sub store.
You have to unmount any mounted sub stores first.

### Moving secrets

```bash
$ gopass mv emails/example.com emails/user@example.com
```

*Moving also works across different sub-stores.*

### Copying secrets

```bash
$ gopass cp emails/example.com emails/user@example.com
```

*Copying also works across different sub-stores.*

## Advanced Features

### Auto-Pager

Like other popular open-source projects `gopass` automatically pipe the output
to `$PAGER` if it's longer than one terminal page. You can disable this behaviour
by unsetting `$PAGER` or `gopass config nopager true`.

### git auto-push and auto-pull

If you want gopass to always push changes in git to your default remote (origin)
enable autosync:

```bash
$ gopass config autosync true
```

### Check Passwords for Common Flaws

gopass can check your passwords for common flaws, like being too short or coming
from a dictionary.

```bash
$ gopass audit
Detected weak secret for 'golang.org/gopher': Password is too short
```

### Check Passwords against leaked passwords

gopass can assist you in checking your passwords against those included in recent
data breaches. Right now this you still need to download and unpack those dumps
yourself, but gopass can take care of the rest.

First go to [haveibeenpwned.com/Passwords](https://haveibeenpwned.com/Passwords) and download
the dumps. Then unpack the 7-zip archives somewhere. Note that full path to those
files and provide it to gopass in the environment variable `HIBP_DUMPS`.

```bash
$ HIBP_DUMPS=/tmp/pwned-passwords-1.0.txt gopass audit hibp
```

### Support for Binary Content

gopass provides secure and easy support for working with binary files through the
`gopass binary` family of subcommands. One can copy or move secret from or to
the store. gopass will attempt to securely overwrite and remove any secret moved
to the store.

```bash
# copy file "/some/file.jpg" to "some/secret.b64" in the store
$ gopass binary cp /some/file.jpg some/secret
# move file "/home/user/private.key" to "my/private.key.b64", removing the file on disk
# after the file has been encoded, stored and verified to be intact (SHA256)
$ gopass binary mv /home/user/private.key my/private.key
# Calculate the checksum of some asset
$ gopass binary sha256 my/private.key
```

### Multiple Stores

gopass supports multi-stores that can be mounted over each other like filesystems
on Linux/UNIX systems.

To add an mount point to an existing store add an entry to the `mounts` object
of the store.

gopass tries to read its configuration from `$HOME/.config/gopass/config.yml` if present.
You can override this location by setting `GOPASS_CONFIG` to another location.

Mounting new stores can be done through gopass:

```bash
# Mount a new store
$ gopass mounts add test /tmp/password-store-test
# Show mounted stores
$ gopass mounts
# Umount a store
$ gopass mounts remove test
```

You can initialize a new store using `gopass init --store mount-point --path /path/to/store`.

Where possible sub stores are supported transparently through the path to the
secret. When specifying the name of a secret it's matched against any mounted
sub stores and the given action is executed on this store.

Commands that don't accept an secret name, e.g. `gopass recipients add` or
`gopass init` usually accept a `--store` parameter. Please check the help output
of each command for more information, e.g. `gopass help init` or
`gopass recipients help add`.

Commands that support the `--store` flag:

| **Command** | *Example* | Description |
| ----------- | --------- | ----------- |
| `gopass git` | `gopass git --store=foo push origin master` | Push all changes in the sub store *foo* to master
| `gopass git init` | `gopass git init --store=foo` | Initialize git in the sub store *foo*
| `gopass init` | `gopass init --store=foo` | Initialize and mount the new sub store *foo*
| `gopass recipients add`| `gopass recipients add --store=foo GPGxID` | Add the new recipient *GPGxID* to the store *foo*
| `gopass recipients remove` | `gopass recipients remove --store=foo GPGxID` | Remove the existing recipients *GPGxID* from the store *foo*
| `gopass config` | `gopass config --store=foo autosync false` | Set the config flag `autosync` to `false` for the store *foo*

### Directly edit structured secrets aka. YAML support

`gopass` supports directly editing structured secrets (only simple key-value maps so far).

```bash
$ gopass generate -n foo/bar 12
The generated password for foo/bar is:
7fXGKeaZgzty
$ gopass insert foo/bar baz
Enter password for foo/bar/baz:
Retype password for foo/bar/baz:
$ gopass foo/bar baz
zab
$ gopass foo/bar
7fXGKeaZgzty
baz: zab
```

### Edit the Config

`gopass` allows editing the config from the commandline. This is similar to how `git` handles `config`
changes through the commandline. Any change will be written to the configured `gopass` config file.

```bash
$ gopass config
alwaystrust: false
askformore: false
autoimport: false
autopull: false
autopush: true
cliptimeout: 10
loadkeys: false
noconfirm: false
path: /home/user/.password-store

$ gopass config cliptimeout 60
$ gopass config cliptimeout
```

### Managing Recipients

You can list, add and remove recipients from the commandline.

```bash
$ gopass recipients
gopass
└── 0xB5B44266A3683834 - Gopher <gopher@golang.org>

$ gopass recipients add 1ABB2C1A

$ gopass recipients
gopass
├── 0xB1C7DF661ABB2C1A - Someone <someone@example.com>
└── 0xB5B44266A3683834 - Gopher <gopher@golang.org>

$ gopass recipients remove 0xB5B44266A3683834

$ gopass recipients
gopass
└── 0xB1C7DF661ABB2C1A - Someone <someone@example.com>
```

Running `gopass recipients` will also try to load and save any missing GPG keys
from and to the store.

The commands manipulating recipients, i.e. `gopass recipients add` and
`gopass recpients remove` accept a `--store` flag that expects the
*name of a mount point* to operate on this mounted sub store.

### Debugging

To debug `gopass`, set the environment variable `GOPASS_DEBUG` to `true`.

### Restricting the characters in generated passwords

To restrict the characters used in generated passwords set `GOPASS_CHARACTER_SET` to
any non-empty string. Please keep in mind that this can considerably weaken the
strength of generated passwords.

### In-place updates to existing passwords

Running `gopass [generate|insert] foo/bar` on an existing entry `foo/bar` will only update
the first line of the secret, leaving any trailing data in place.

### Disabling Colors

Disabling colors is as simple as setting `GOPASS_NOCOLOR` to `true`.

### Password Templates

With gopass you can create templates which are searched when executing `gopass edit` on a new secret. If the folder, or any parent folder, contains a file called `.pass-template` it's parsed as a Go template, executed with the name of the new secret and an auto-generated password and loaded into your `$EDITOR`.

This makes it easy to e.g. generate database passwords or use templates for certain kind of secrets.

### JSON API

`gopass jsonapi` enables communication with gopass via json messages. This is in particular useful for browser plugins like [gopassbridge](https://github.com/martinhoefling/gopassbridge) running gopass as native app. More details can be found in [docs/jsonapi.md](docs/jsonapi.md)

## Roadmap

- [x] Be 100% pass 1.4 compatible
- [x] Storing binary files in gopass (almost done)
- [x] Storing structured files and templates (credit cards, DBs, websites...)
- [ ] UX improvements and more wizards
- [ ] Tackle the information disclosure issue
- [ ] Build a great workflow for requesting and granting access
- [ ] Better and more fine grained ACL
- [ ] Be nicely usable by semi- and non-technical users

*Note: Being 100% pass compatible was a milestone, not a promise for the future. We will eventually diverge from pass to support more advanced features. This will break compatibility.*

