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
apt-get install gnupg git rng-tools
```

_Note:_ installing in Ubuntu 16.04 will require you to install `gnupg2`.

#### RHEL & CentOS

```bash
yum install gnupg2 git rng-tools
```

#### MacOS

If you haven't already, install [homebrew](http://brew.sh). And then:

```bash
brew install gnupg2 git
```

#### Windows

* Download and install [GPG4Win](https://www.gpg4win.org/).
* Download and install [the Windows git installer](https://git-scm.com/download/win).

#### OpenBSD

For OpenBSD -current:
```
pkg_add gopass
```

For OpenBSD 6.2 and earlier, install via `go get`.

Please note that the OpenBSD builds uses `pledge(2)` to disable some syscalls,
so some features (e.g. version checks, auto-update) are unavailable.

### Set up a GPG key pair

gopass depends on the `gpg` program for encryption and decryption. You **must** have a
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
* Pass phrase: Make sure to pick a very long pass phrase, not just a simple password. Remember this should be stronger than any of the secrets you store in the password store. You can configure the GPG Agent later to save you repetitive typing.

Now, you have created a public and private key pair. If you don't know what that means, of if you are not familiar with GPG, we highly recommend you do a little reading on the subject. Check out the following manuals:

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

If the last step fails please investigate this before continuing.

## Installation Steps

Depending on your operating system, you can either use a package manager, download a pre-built binary, or install from source. If you have a working Go development environment, we recommend building from source.

### MacOS

If you haven't already, install [homebrew](http://brew.sh). And then:

```bash
brew install gopass
```

Alternatively, you can install gopass from the appropriate Darwin release from the repository [releases page](https://github.com/justwatchcom/gopass/releases).

### Ubuntu & Debian

When installing on Ubuntu or Debian you can either download the `deb` package and install manually or use our repository.

#### Using the gopass repository

First you need to add our archive signing key and add the package source.

```bash
wget -O- https://www.justwatch.com/gopass/releases/0x0C92225A97F6B666.pub | sudo apt-key add -
echo "deb https://www.justwatch.com/gopass/releases/binary-amd64 ./" | sudo tee /etc/apt/sources.list.d/gopass.list
```

Now you can update your package lists and install using `apt-get`:

```bash
sudo apt-get update
sudo apt-get install gopass
```

#### Manual download

First, find the latest .deb release from the repository [releases page](https://github.com/justwatchcom/gopass/releases). Then, download and install it:

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

### Windows

**WARNING**: Windows is not yet officially supported. We try to support it in the future. These are steps are only meant to help you setup gopass on Windows so you can provide us with feedback about the current state of our Windows support.

Download and install a suitable Windows build from the repository [releases page](https://github.com/justwatchcom/gopass/releases).

### Installing from Source

If you have [Go](https://golang.org/) already installed, you can use `go get` to automatically download the latest version:

```
go get -u github.com/justwatchcom/gopass
```

If `$GOPATH/bin` is in your `$PATH`, you can now run `gopass` from anywhere on your system.

## Optional Post-Installation Steps

### Securing Your Editor

Various editors may store temporary files outside of the secure working directory when editing secrets. We advise you to check and disable this behavior for your editor of choice.

For `vim` on Linux, the following setting may be helpful:

```
au BufNewFile,BufRead /dev/shm/gopass.* setlocal noswapfile nobackup noundofile
```

### Migrating from pass to gopass

If you are migrating from pass to gopass, you can simply use your existing password store and everything should just work. Furthermore, it may be helpful to link the gopass binary so that you can use it as a drop-in replacement. For example, assuming `$HOME/bin/` exists and is present in your `$PATH`:

```bash
ln -s $GOPATH/bin/gopass $HOME/bin/pass
```

### Migrating to gopass from Other Password Stores

Before migrating to gopass, you may have been using other password managers (such as [KeePass](https://keepass.info/), for example). If you were, you might want to import all of your existing passwords over. Because gopass is fully backwards compatible with pass, you can use any of the existing migration tools found under the "Migrating to pass" section of the [official pass website](https://www.passwordstore.org/).

### Enable Bash Auto completion

If you use Bash, you can run one of the following commands to enable auto completion for sub commands like `gopass show`, `gopass ls` and others.

```
source <(gopass completion bash)
```

**MacOS**: The version of bash shipped with MacOS may [require a workaround](https://stackoverflow.com/questions/32596123/why-source-command-doesnt-work-with-process-substitution-in-bash-3-2) to enable auto completion. If the instructions above do not work try the following one:

```
source /dev/stdin <<<"$(gopass completion bash)"
```

### Enable  Z Shell Auto completion

If you use zsh, `make install` or `make install-completion` should install the completion in the correct location.

### Enable fish completion

If you use the [fish](https://fishshell.com/) shell, you can enable experimental shell completion. Copy the file `fish.completion` to `~/.config/fish/completions/gopass.fish` and start a new shell.

Since writing fish completion scripts is not yet supported by the CLI library we use, this completion script is missing a few features. Feel free to contribute if you want to improve it.

### dmenu / rofi support

In earlier versions gopass supported [dmenu](http://tools.suckless.org/dmenu/). We removed this and encourage you to call dmenu yourself now.

This also makes it easier to call gopass with any drop-in replacement of dmenu, like [rofi](https://github.com/DaveDavenport/rofi), for example, since you would just need to replace the `dmenu` call below by `rofi -dmenu`.

```bash
# Simply copy the selected password to the clipboard
gopass ls --flat | dmenu | xargs --no-run-if-empty gopass show -c
# First pipe the selected name to gopass, encrypt it and type the password with xdotool.
gopass ls --flat | dmenu | xargs --no-run-if-empty gopass show -f | head -n 1 | xdotool type --clearmodifiers --file -
```

You can then bind these command lines to your preferred shortcuts in your window manager settings, typically under `System Settings > Keyboard > Shortcuts > Custom Shortcuts`. In some cases you may need to wrap it with `bash -c 'your command'` in order for it to work (tested and working in Ubuntu 18.04).

### Filling in passwords from browser

Gopass allows filling in passwords in browsers leveraging a browser plugin like [gopass bridge](https://github.com/martinhoefling/gopassbridge). 
The browser plugin communicates with gopass via JSON messages. To allow the plugin to start gopass, a [native messaging manifest](https://developer.mozilla.org/en-US/Add-ons/WebExtensions/Native_messaging) must be installed for each browser.
Chrome, chromium and Firefox are supported, currently. Further a wrapper must be installed to setup gpg-agent and execute `gopass jsonapi listen`.

```bash
# Asks all questions concerning browser and setup
gopass jsonapi configure

# Do not copy / install any files, just print their location and content
gopass jsonapi configure --print-only

# Specify browser and wrapper path
gopass jsonapi configure --browser chrome --path /home/user/.local/

```

The user name/login is determined from `login`, `username` and `user` yaml attributes (after the --- separator) like this:
```
<your password>
---
username: <your username>
```
As fallback, the last part of the path is used, e.g. `theuser1` for `Internet/github.com/theuser1` entry.
 
### Storing and Syncing your Password Store with Google Drive / Dropbox / etc.

Please be warned that using cloud-based storage may negatively impact to confidentially of your store. However, if you wish to use one of these services, you can do so.

For example, to use gopass with [Google Drive](https://drive.google.com):

```bash
gopass init --nogit
mv .password-store/ "Google Drive/Password-Store"
gopass config path "~/Google Drive/Password-Store"
```

### Download a GUI

Because gopass is fully backwards compatible with pass, you can use some existing graphical user interfaces / frontends:

* Android - [PwdStore](https://github.com/zeapo/Android-Password-Store)
* iOS - [Pass for iOS](https://github.com/davidjb/pass-ios#readme)
* Windows / MacOS / Linux -  [QtPass](https://qtpass.org/)

Others can be found at the "Compatible Clients" section of the [official pass website](https://www.passwordstore.org/).

## Using gopass

Once you have installed gopass, check out the [features documentation](https://github.com/justwatchcom/gopass/blob/master/docs/features.md) for some quick usage examples.

### Using the onboarding wizard

Running `gopass` with no existing store will start the onboarding wizard which
will guide you through the setup of gopass.

### Batch bootstrapping

In order to simplify the setup of gopass for your team members if can be run in a fully scripted bootstrap mode.

```bash
# First initialize a new shared store and push it to an empty remote
gopass --yes setup --remote github.com/example/pass.git --alias example --create --name "John Doe" --email "john.doe@example.com"

# For every other team member initialize a new store and clone the existing remote
gopass --yes setup --remote github.com/example/pass.git --alias example --name "Jane Doe" --email "jane.doe@example.com"
```
