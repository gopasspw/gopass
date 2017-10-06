# Setup

## Table of Contents

1. [Pre-Installation Steps](#pre-installation-steps)
2. [Installation Steps](#installation-steps)
3. [Optional Post-Installation Steps](#optional-post-installation-steps)
4. [Using gopass](#using-gopass)

## Pre-Installation Steps

### Download and Install Dependencies

`gopass` needs some external programs to work:

* `gpg`
* `git`
* An external editor (using `vim` by default)

#### Ubuntu & Debian

```bash
apt-get install gnupg git
```

#### RHEL & CentOS

```bash
yum install gnupg2 git
```

#### macOS

If you haven't already, install [homebrew](http://brew.sh). And then:

```bash
brew install gnupg2 git
```

#### Windows

* Download and install [GPG4Win](https://www.gpg4win.org/).
* Download and install [the Windows git installer](https://git-scm.com/download/win).

### Setup a GPG key pair

`gopass` depends on `gpg` for encryption and decryption. You **must** have a
suitable key pair. To list your current keys, you can do:

```bash
gpg --list-keys
```

If there is no output, then you don't have any keys. To create a new key:

```bash
gpg --gen-key
```

You will be presented with several options:

* Key type: Choose either "RSA and RSA" or "DSA and ElGamal".
* Key size: Choose at least 2048.
* Validity: 5 to 10 years is a good default.
* Enter your real name and primary email address.
* A comment is not necessary.
* Passphrase: Make sure to pick a very long passphrase, not just a simple password. Remember this should be stronger than any of the secrets you store in the password store. You can configure the GPG Agent later to save you repetitive typing.

Now, you have created a public and private key pair. If you don't know what that means, of if you are not familiar with GPG, we highly recommend you do a little reading on the subject. Check out the following manuals:

* ["git + gpg, know thy commits" at coderwall](https://coderwall.com/p/d3uo3w/git-gpg-know-thy-commits)
* ["Generating a new GPG key" by GitHub](https://help.github.com/articles/generating-a-new-gpg-key/)

### Securing Your Editor

Various editors may store temporary files outside of the secure working directory
when editing secrets. It's advised to check and disable this behavior for
your editor of choice.

For `vim` on Linux, the following setting may be helpful:

```
au BufNewFile,BufRead /dev/shm/gopass.* setlocal noswapfile nobackup noundofile
```

## Installation Steps

Depending on your operating system, you can either use a package manager,
download a pre-built binary, or install from source. If you have a working Go
development environment, we recommend building from source.

#### macOS

If you haven't already, install [homebrew](http://brew.sh). And then:

```bash
brew tap justwatchcom/gopass
brew install gopass
```

Alternatively, you can grab the appropriate darwin release from the repository [releases page](https://github.com/justwatchcom/gopass/releases).

#### Ubuntu & Debian

First, find the latest .deb release from the repository [releases page](https://github.com/justwatchcom/gopass/releases). Then, download and install it:

```bash
wget [the URL of the latest .deb release]
sudo dpkg -i gopass-1.2.0-linux-amd64.deb
```

#### Windows

**WARNING**: Windows is not yet officially supported. We try to support it in
the future. These are steps are only meant to help you setup `gopass` on Windows
so you can provide us with feedback about the current state of our Windows support.

Download and install a suitable Windows build from the repository [releases page](https://github.com/justwatchcom/gopass/releases).

#### Installing from Source

If you have [Go](https://golang.org/) already installed, you can use `go get` to automatically download the latest version:

```
go get -u github.com/justwatchcom/gopass
```

If `$GOPATH/bin` is in your `$PATH`, you can now run `gopass` from anywhere on your system.

## Optional Post-Installation Steps

#### Migrating from `pass` to `gopass`

If you are migrating from `pass` to `gopass`, it may be helpful to link it to `pass` so that you can use it as a drop-in replacement for `pass`. For example, assuming `$HOME/bin/` exists and is present in your `$PATH`:

```bash
ln -s $GOPATH/bin/gopass $HOME/bin/pass
```

#### Enable Bash / Z Shell Autocompletion

Run one of the following commands for your shell and you should have
autocompletion for subcommands like `gopass show`, `gopass ls` and others.

```
source <(gopass completion bash)
source <(gopass completion zsh)
```

#### Enable fish completion

Experimental [fish](https://fishshell.com/) shell completion is available.
Copy the file `fish.completion` to `~/.config/fish/completions/gopass.fish`
and start a new shell.

Since writing fish completion scripts is not yet supported by the CLI library we
use, this completion script is missing a few features. Feel free to contribute
if you want to improve it.

#### dmenu / rofi support

In earlier versions `gopass` supported [dmenu](http://tools.suckless.org/dmenu/).
We removed this and encourage you to call dmenu yourself now.

This also makes it easier to call `gopass` with e.g. [rofi](https://github.com/DaveDavenport/rofi).

```bash
# Simply copy the selected password to the clipboard
gopass ls --flat | dmenu | xargs --no-run-if-empty gopass show -c
# First pipe the selected name to gopass, encrypt it and type the password with xdotool.
gopass ls --flat | dmenu | xargs --no-run-if-empty gopass show | head -n 1 | xdotool type --clearmodifiers --file -
```

#### Migrating to `gopass` from Other Password Stores

Because `gopass` is fully backwards compatible with `pass`, you can use any of the existing migration tools found under the "Migrating to pass" section of the [official pass website](https://www.passwordstore.org/).






### Data Organization

Your data in `gopass` loosely resembles an file system. You need to have at least one
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

#### Storing and Syncing your Password Store with Google Drive / Dropbox / etc.

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

Because `gopass` is fully backwards compatible with `pass`, you can use some existing interfaces:

* Android - [PwdStore](https://github.com/zeapo/Android-Password-Store)
* iOS - [Pass for iOS](https://github.com/davidjb/pass-ios#readme)
* Windows / macOS / Linux -  [QtPass](https://qtpass.org/)

Others can be found at the "Compatible Clients" section of the [official pass website](https://www.passwordstore.org/).

## Using gopass

Once you have installed gopass, check out the [features documentation](https://github.com/justwatchcom/gopass/blob/master/docs/usage.md) for some quick usage examples.
