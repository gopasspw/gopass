# Setup

## Table of Contents

1. [Pre-Installation Steps](#pre-installation-steps)
2. [Installation Steps](#installation-steps)
3. [Optional Post-Installation Steps](#optional-post-installation-steps)
4. [Using gopass](#using-gopass)

## Pre-Installation Steps

### Download and Install Dependencies

gopass needs some external programs to work:

* `gpg` - [GnuPG](https://www.gnupg.org/), preferably in Version 2 or later
* `git` - [Git SCM](https://git-scm.com/), any Version should be OK

It is recommended to have either `rng-tools` or `haveged` installed to speed up
key generation if these are available for your platform.

#### Ubuntu & Debian

```bash
apt-get update
apt-get install git gnupg rng-tools
```

#### RHEL & CentOS

```bash
yum install gnupg2 git rng-tools
```

#### Arch Linux

```bash
pacman -S gnupg2 git rng-tools
```

#### MacOS

If you haven't already, install [homebrew](http://brew.sh). And then:

```bash
brew install gnupg2 git
```

#### Windows

* Download and install [GPG4Win](https://www.gpg4win.org/).
* Download and install [the Windows git installer](https://git-scm.com/download/win).

Alternatively, it can be installed via [chocolatey](https://chocolatey.org/packages/gopass)

* `choco install gopass` (requires admin privileges)

#### OpenBSD

For OpenBSD -current:
```
pkg_add gopass
```

For OpenBSD 6.2 and earlier, install via `go install`.

Please note that the OpenBSD builds uses `pledge(2)` to disable some syscalls,
so some features (e.g. version checks, auto-update) are unavailable.

#### FreeBSD

For FreeBSD 11 and newer:

```
pkg install gopass
```

### Set up a GPG key pair

gopass depends on the `gpg` program for encryption and decryption. You **must** have a
suitable key pair. To list your current keys, you can do:

```bash
gpg --list-secret-keys
```

If there is no output, then you don't have any keys. To create a new key:

```bash
gpg --full-generate-key
```

You will be presented with several options:

* Key type: Choose either "RSA and RSA" or "DSA and ElGamal".
* Key size: Choose at least 2048.
* Validity: 5 to 10 years is a good default.
* Enter your real name and primary email address.
* A comment is not necessary.
* Pass phrase: Make sure to pick a very long pass phrase, not just a simple password. Remember this should be stronger than any of the secrets you store in the password store. You can configure the GPG Agent later to save you repetitive typing.

Now, you have created a public and private key pair. If you don't know what that means, or if you are not familiar with GPG, we highly recommend you do a little reading on the subject. Check out the following manuals:

* ["git + gpg, know thy commits" at coderwall](https://coderwall.com/p/d3uo3w/git-gpg-know-thy-commits)
* ["Generating a new GPG key" by GitHub](https://help.github.com/articles/generating-a-new-gpg-key/)

### Git and GPG

gopass will configure git to sign commits by default, so you should make sure that git can
interface with GPG.

```bash
mkdir some-dir
cd some-dir
git init
touch foo
git add foo
git commit -S -m "test"
```
Here the `-S` flag will sign your commit using GPG, allowing you to test your GPG setup with git.
If you get an error like: "gpg failed to sign the data" try to see if creating a clear text signature works:

```bash
 echo "test" | gpg2 --clearsign
```

If this fails with an error: "Inappropriate ioctl for device" run the following command:

```bash
 export GPG_TTY=$(tty)
```

If you are using CSH or TCSH:
```
setenv GPG_TTY `tty`
```

Now you should be able to create a clear text signature and the commit should work flawlessly.

If you are presented with a different error please investigate this before continuing. If that works
set it in your `.zprofile`, `.bashrc` or simliar.

Also if you have both `gnupg` and `gnupg2` installed, make sure to use the latter in git:
```bash
git config --global gpg.program gpg2
```

## Installation Steps

Depending on your operating system, you can either use a package manager, download a pre-built binary, or install from source. If you have a working Go development environment, we recommend building from source.

### MacOS

If you haven't already, install [homebrew](http://brew.sh). And then:

```bash
brew install gopass
```

Alternatively, you can install gopass from the appropriate Darwin release from the repository [releases page](https://github.com/gopasspw/gopass/releases).

If you're using a password on your GPG key, you also have to install `pinentry-mac` from brew and configure it in your `~/gpg/gpg-agent.conf`:

```bash
brew install pinentry-mac
PINENTRY=$(which pinentry-mac)
echo "pinentry-program ${PINENTRY}" >>~/.gnupg/gpg-agent.conf
defaults write org.gpgtools.common UseKeychain NO
```

The last step is important if you want to stop `pinentry-mac` from caching your passphrase in the MacOS Keychain by default (indefinitely).

### Ubuntu, Debian, Deepin, Devuan, Kali Linux, Pardus, Parrot, Raspbian

**WARNING**: The official Debian repositories (and derived distributions) contain
a package named `gopass` that is not related to this project in any way.
It's a similar tool with a completely independent implementation and feature set.
We are aware of this issue but can not do anything about it.

When installing on Ubuntu or Debian you can either download the `deb` package,
[install manually or build from source](#installing-from-source) or use our APT repository.

```bash
$ curl http://packages.gopass.pw/repos/gopass/gopass-archive-keyring.gpg | sudo tee /usr/share/keyrings/gopass-archive-keyring.gpg
$ cat << EOF | sudo tee /etc/apt/sources.list.d/gopass.sources
Types: deb
URIs: https://packages.gopass.pw/repos/gopass
Suites: stable
Architectures: all amd64 arm64 armhf
Components: main
Signed-By: /usr/share/keyrings/gopass-archive-keyring.gpg
EOF
$ sudo apt update
$ sudo apt install gopass gopass-archive-keyring
```

#### Manual download

First, find the latest .deb release from the repository [releases page](https://github.com/gopasspw/gopass/releases). Then, download and install it:

```bash
wget [the URL of the latest .deb release]
sudo dpkg -i gopass-1.2.0-linux-amd64.deb
```

### Gentoo

There is an overlay that includes gopass. Run these commands to install gopass through `emerge`.

```bash
layman -a go-overlay
emerge -av gopass
```

### Fedora

```bash
dnf install gopass
```

### Red Hat / CentOS

There is an unofficial RPM build maintained by a contributor.

```bash
# if you're using dnf (needs dnf-plugins-core)
dnf copr enable daftaupe/gopass
dnf install gopass
# of if you're using an older distribution (needs yum-plugin-copr)
yum copr enable daftaupe/gopass
yum install gopass
```

### Arch Linux

```bash
pacman -S gopass
```

### Windows

**WARNING**: Windows is not yet officially supported. We try to support it in the future. These are steps are only meant to help you setup gopass on Windows so you can provide us with feedback about the current state of our Windows support.

You can install `gopass` by [Chocolatey](https://chocolatey.org/):

```bash
choco install gopass
```

Or by [Scoop](https://scoop.sh/):

```bash
scoop install gopass
```

Alternatively, download and install a suitable Windows build from the repository [releases page](https://github.com/gopasspw/gopass/releases).

### Installing from Source

If you have [Go](https://golang.org/) already installed, you can use `go install` to automatically download the latest version:

```bash
GO111MODULE=on go install -u github.com/gopasspw/gopass@latest
```

WARNING: `latest` is not a stable release. It is recommended to use a specific version.

If `$GOPATH/bin` is in your `$PATH`, you can now run `gopass` from anywhere on your system.

## Optional Post-Installation Steps

### Securing Your Editor

Various editors may store temporary files outside of the secure working directory when editing secrets. We advise you to check and disable this behavior for your editor of choice.

For `vim` on Linux, the following setting may be helpful:

```
au BufNewFile,BufRead /dev/shm/gopass.* setlocal noswapfile nobackup noundofile
```

For MacOS consider this setting:

```
au BufNewFile,BufRead /private/**/gopass** setlocal noswapfile nobackup noundofile
```

### Migrating from pass to gopass

If you are migrating from pass to gopass, you can simply use your existing password store and everything should just work. Furthermore, it may be helpful to link the gopass binary so that you can use it as a drop-in replacement. For example, assuming `$HOME/bin/` exists and is present in your `$PATH`:

```bash
ln -s $GOPATH/bin/gopass $HOME/bin/pass
```

### Migrating to gopass from Other Password Stores

Before migrating to gopass, you may have been using other password managers (such as [KeePass](https://keepass.info/), for example). If you were, you might want to import all of your existing passwords over. Because gopass is fully backwards compatible with pass, you can use any of the existing migration tools found under the "Migrating to pass" section of the [official pass website](https://www.passwordstore.org/).

### Enable Bash Auto completion

If you use Bash, you can run one of the following commands to enable auto completion for sub-commands like `gopass show`, `gopass ls` and others.

```bash
source <(gopass completion bash)
```

**MacOS**: The version of bash shipped with MacOS may [require a workaround](https://stackoverflow.com/questions/32596123/why-source-command-doesnt-work-with-process-substitution-in-bash-3-2) to enable auto completion. If the instructions above do not work try the following one:

```bash
source /dev/stdin <<<"$(gopass completion bash)"
```

### Enable Z Shell Auto completion

If you use zsh, `make install` or `make install-completion` should install the completion in the correct location.

If zsh autocompletion is still not functional, or if you want to install it manually, you can run the following commands:

```bash
$ gopass completion zsh > ~/_gopass 
$ sudo mv ~/_gopass /usr/share/zsh/site-functions/_gopass
$ rm -i ${ZDOTDIR:-${HOME:?No ZDOTDIR or HOME}}/.zcompdump && compinit

```
Then exit and re-run zsh if the last command failed.

Notice that it is important to directly redirect Gopass' output to a file,
using pipes or echo mess up the output. Also notice that the generated `_gopass` file is
a completion file that is supposed to be handled by zsh and to be installed in the zsh
completions directory, as defined by either the standard `/usr/share/zsh/site-functions/` path,
or by a user-specified `fpath` folder. It is not meant to used with `source`.

If zsh completion is still not working, you might want to add the following to your `.zshrc` file:
```bash
autoload -U compinit && compinit
```
if you don't have it already.

### Enable fish completion

If you use the [fish](https://fishshell.com/) shell, you can enable experimental shell completion by the following command:
```fish
$ mkdir -p ~/.config/fish/completions and; gopass completion fish > ~/.config/fish/completions/gopass.fish
```
and start a new shell afterwards.

Since writing fish completion scripts is not yet supported by the CLI library we use, this completion script is missing a few features. Feel free to contribute if you want to improve it.

### dmenu / rofi support

In earlier versions gopass supported [dmenu](http://tools.suckless.org/dmenu/). We removed this and encourage you to call dmenu yourself now.

This also makes it easier to call gopass with any drop-in replacement of dmenu, like [rofi](https://github.com/DaveDavenport/rofi), for example, since you would just need to replace the `dmenu` call below by `rofi -dmenu`.

```bash
# Simply copy the selected password to the clipboard
gopass ls --flat | dmenu | xargs --no-run-if-empty gopass show -c
# First pipe the selected name to gopass, decrypt it and type the password with xdotool.
gopass ls --flat | dmenu | xargs --no-run-if-empty gopass show -o | xdotool type --clearmodifiers --file -
# First pipe the selected name to gopass, and type the value from the key "username" with xdotool.
gopass ls --flat | dmenu | xargs --no-run-if-empty -- bash -c 'gopass show -f $0 username' | head -n 1 | xdotool type --clearmodifiers --file -
# Oterwise type the name of the entry using xdotool, in case you are not including a username key in your entries
gopass ls --flat | dmenu | sed 's!.*/!!' | tr -d '\n' | xdotool type --clearmodifiers --file -
```

You can then bind these command lines to your preferred shortcuts in your window manager settings, typically under `System Settings > Keyboard > Shortcuts > Custom Shortcuts`. In some cases you may need to wrap it with `bash -c 'your command'` in order for it to work (tested and working in Ubuntu 18.04).

### Filling in passwords from browser

Gopass allows filling in passwords in browsers leveraging a browser plugin like [gopass bridge](https://github.com/gopasspw/gopassbridge).
The browser plugin communicates with gopass-jsonapi via JSON messages.
To allow the plugin to start gopass-jsonapi, a [native messaging manifest](https://developer.mozilla.org/en-US/Add-ons/WebExtensions/Native_messaging) must be installed for each browser.
Chrome, Chromium and Firefox are supported, currently.

**Upgrade to gopass v1.10 / v1.11**:
`gopass-jsonapi` is now its own binary file, which you need to install separately.

The binary for v1.10 and v1.11 can be downloaded and unpacked from
[archive files on Github Releases](https://github.com/gopasspw/gopass/releases/tag/v1.11.0).

You need to run `gopass-jsonapi configure` after the upgrade to configure your browser for the new command.

**Upgrade to gopass v1.12**
The new binary can be downloaded from the latest
[Github Release on gopass-jsonapi](https://github.com/gopasspw/gopass-jsonapi/releases).

For more detailed instructions, please read: [gopass-jsonapi/README](https://github.com/gopasspw/gopass-jsonapi/blob/main/README.md).

### Storing and Syncing your Password Store with git

This is the recommended way to use `gopass`.

NOTE: We do recommend to use a private Git repository. A public one will keep
your credentials secure but it will leak metadata.

To use `gopass` with `git` either create a new git repository or clone an existing
password store.

#### New password store with git

Create a new repository, either locally or on a server, then specify this
repository during the `gopass setup`.

```bash
$ gopass setup --crypto gpg --storage gitfs # used by default
[...]
# provide an existing, empty git remote, e.g. git@gitlab.example.org:john/passwords.git
```

#### Existing password store with git

If you have created a password store with `git`, `gopass` can easily clone it.

```bash
$ gopass clone git@gitlab.example.org:john/passwords.git
```

### Storing and Syncing your Password Store with Google Drive / Dropbox / Syncthing / etc.

The recommended way to use Gopass is to sync your store with a git repository, preferably a private one, since the name and path of your secrets might reveal information that you'd prefer to keep private.
However, shall you prefer to, you might also use the `noop` storage backend that is meant to store data on a cloud provider instead of a git server.

Please be warned that using cloud-based storage may negatively impact the confidentiality of your store. However, if you wish to use one of these services, you can do so.

For example, to use gopass with [Google Drive](https://drive.google.com):

```bash
gopass setup --storage fs
mv .password-store/ "Google Drive/Password-Store"
gopass config path "~/Google Drive/Password-Store"
```

### Download a GUI

Because gopass is fully backwards compatible with pass, you can use some existing graphical user interfaces / frontends:

* Android - [PwdStore](https://github.com/zeapo/Android-Password-Store)
* iOS - [Pass for iOS](https://github.com/davidjb/pass-ios#readme)
* Windows / MacOS / Linux -  [QtPass](https://qtpass.org/)

There is also [Gopass UI](https://github.com/codecentric/gopass-ui) which was exclusively implemented for gopass and is available for MacOS, Linux and Windows.

Others can be found at the "Compatible Clients" section of the [official pass website](https://www.passwordstore.org/).

## Using gopass

Once you have installed gopass, check out the [features documentation](https://github.com/gopasspw/gopass/blob/master/docs/features.md) for some quick usage examples.

### Using the onboarding wizard

Running `gopass` with no existing store will start the onboarding wizard which
will guide you through the setup of gopass.

### Batch bootstrapping

In order to simplify the setup of gopass for your team members it can be run in a fully scripted bootstrap mode.

```bash
# First initialize a new shared store and push it to an empty remote
gopass --yes setup --remote github.com/example/pass.git --alias example --create --name "John Doe" --email "john.doe@example.com"

# For every other team member initialize a new store and clone the existing remote
gopass --yes setup --remote github.com/example/pass.git --alias example --name "Jane Doe" --email "jane.doe@example.com"
```

The first command will create a new mount named `example` and push it to an empty (`--create`) remote.
It will fail if the remote at `github.com/example/pass.git` is not empty.

The second command will clone the existing (no `--create` flag) remote `github.com/example/pass.git`
and mount it as the mount point `example`.

