# Setup

You can either use a package manager, download a pre-built binary or install from source. If you have
a working Go development environment, we recommend building from source.

### Package manager

#### macOS

```bash
$ brew tap justwatchcom/gopass
$ brew install gopass
```

#### Debian and Ubuntu

```bash
$ wget https://www.justwatch.com/gopass/releases/1.2.0/gopass-1.2.0-linux-amd64.deb
$ sudo dpkg -i gopass-1.2.0-linux-amd64.deb
```

#### Windows

**WARNING**: Windows is not officially supported, yet. We try to support windows
in the future. These are steps are only meant to help you setup `gopass` on windows
so you can provide us with feedback about the current state of our windows support.

* Download a suiteable windows build, e.g. https://github.com/justwatchcom/gopass/releases/download/v1.4.0-rc1/gopass-1.4.0-rc1-windows-amd64.zip
* Install [GPG4Win](https://www.gpg4win.org/)
* Install [git](https://git-scm.com/download/win)

### Download

Please visit https://www.justwatch.com/gopass/releases/ for a list of binary releases.

### From Source

To get the latest version of gopass, run `go get`:

    go get -u github.com/justwatchcom/gopass

If `$GOPATH/bin` is in your `$PATH`, you can now run `gopass` from anywhere on your system and use this.

If you like you can link `$GOPATH/bin/gopass` to `pass` somewhere in your `$PATH` to use gopass as a drop-in
replacement of `pass`.

Assuming `$HOME/bin/` exists and is present in your `$PATH`:

```bash
$ ln -s $GOPATH/bin/gopass $HOME/bin/pass
```

### Autocompletion

Run one of the following commands for your shell and you should have
autocompletion for subcommands like `gopass show`, `gopass ls` and others.

    source <(gopass completion bash)
    source <(gopass completion zsh)

### fish completion

Experimental [fish](https://fishshell.com/) shell completion is available.
Copy the file `fish.completion` to `~/.config/fish/completions/gopass.fish`
and start a new shell.

Since writing fish completion scripts is not yet supported by the CLI library we
use, this completion script is missing a few features. Feel free to contribute
if you want to improve it.

### dmenu/rofi support

In earlier versions gopass supported [dmenu](http://tools.suckless.org/dmenu/).
We removed this and encourage you to call dmenu yourself now.

This also makes it easier to call gopass with e.g. [rofi](https://github.com/DaveDavenport/rofi).

```bash
# Simply copy the selected password to the clipboard
$ gopass ls --flat | dmenu | xargs --no-run-if-empty gopass show -c
# First pipe the selected name to gopass, encrypt it and type the password with xdotool.
$ gopass ls --flat | dmenu | xargs --no-run-if-empty gopass show | head -n 1 | xdotool type --clearmodifiers --file -
```

### Dependencies

`gopass` needs some external programs to work.

* `gpg`
* `git`

As well as some external editor (using `vim` by default).

On Debian-based Linux systems you should run this command:

```bash
$ apt-get install gnupg git
```

On macOS with [homebrew](http://brew.sh) the following will do:

```bash
$ brew install gnupg2 git
```

### Setup GPG

`gopass` depends on `gpg` for encryption and decryption. You **must** have a
suitable key pair.

```bash
$ gpg --gen-key
# Key Type: Choose either "RSA and RSA" or "DSA and ElGamal"
# Key Size: Choose at least 2048
# Validity: 5 to 10 years is a good default
# Enter your real name and primary email address, comment is not necessary
# Passphrase: Make sure to pick a very long passphrase, not just a simple password. Remember this should be stronger than any of the secrets you store in the password store. You can configure the GPG Agent later, to save you repititive typing.
```

There are a lot of good manuals to get started with GPG out there.

We recommend these ones:

* ["git + gpg, know thy commits" at coderwall](https://coderwall.com/p/d3uo3w/git-gpg-know-thy-commits)
* ["Generating a new GPG key" by GitHub](https://help.github.com/articles/generating-a-new-gpg-key/)

### Securing your editor

Various editors may store temporary files outside of the secure working directory
when editing secrets. It's advised to check and disable this behaviour for
your editor of choice.

For vim on linux this setting may be helpful:

```
au BufNewFile,BufRead /dev/shm/gopass.* setlocal noswapfile nobackup noundofile
```

### Data Organization

Your data in `gopass` loosely resembles an filesystem. You need to have at least one
root store but you can mount as many sub-stores (think of volumes) under the root volume.

The stores do not impose any specific layout for your data. Any `key` can contain any kind of data.

Please note that sensitive data **should not** be put into the name of a secret.

If you mainly use a store for website logins or plan to use
[browserpass](https://github.com/dannyvankooten/browserpass) you should follow
the following pattern for storing your credentials:

```
example.org/user
example.com/john@doe.com
```

#### Storing and Syncing your Password Store with Google Drive/Dropbox/...

Please be warned that using a cloud-based storage _drive_ may negatively impact
to confidentially of your store, but if you wish to use one of these services
you can do so.

For example, if using [Google Drive](https://drive.google.com):

```bash
cd
gopass init --nogit
mv .password-store/ "Google Drive/Password-Store"
gopass config path "~/Google Drive/Password-Store"
```

### Using other GUIs with `gopass`

Because `gopass` is fully *backwards* compatible with `pass` you can simply use other existing interfaces.
We use the [Android](https://github.com/zeapo/Android-Password-Store) &
[iOS](https://github.com/davidjb/pass-ios#readme) apps ourselves. But there are more integrations for
[Chrome, Firefox](https://github.com/dannyvankooten/browserpass),
[Windows](https://github.com/mbos/Pass4Win) and many more.

### Migrating to `gopass` from other password stores.

Since `gopass` is fully compatible to `pass` you can use any of the migration
tools available for [`pass`](https://www.passwordstore.org) to import from 1Password, LastPass and many more.

