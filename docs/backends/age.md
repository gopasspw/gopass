# age crypto backend

The `age` backend is an experimental crypto backend based on [age](https://age-encryption.org). It adds an
encrypted keyring on top (using age in scrypt password mode). It also has
(largely untested) support for specifying recipients as github users. This will
use their ssh public keys for age encryption.
It is well positioned to eventually replace `gpg` as the default crypto backend.

## Getting started

WARNING: This backend is experimental and the on-disk format likely to change.

To start using the `age` backend initialize a new (sub) store with the `--crypto=age` flag:

```
$ gopass age identity add [AGE-... age1...]
<if you do not specify an age secret key, you'll be prompted for one>
$ gopass init --crypto age
```

or use the wizard that will help you create a new age key:
```
$ gopass setup --crypto age
```

This will automatically create a new age keypair and initialize the new store.

Existing stores can be migrated using `gopass convert --crypto age`.

N.B. for a fully scripted or **non-interactive setup**, you can use the `GOPASS_AGE_PASSWORD` env variable
to set your identity file secret passphrase, and specify the age identity and recipients
that should be used for encrypting/decrypting passwords as follows:
```
$ gopass age identity add <AGE-...> <age1...>
$  GOPASS_AGE_PASSWORD=mypassword gopass init --crypto age <age1...>
```
Notice the extra space in front of the command to skip most shell's history.
You'll need to set your name and username using `git` directly if you're using it as storage backend (the default one).

You can also specify the ssh directory by setting environment variable
```
$  GOPASS_SSH_DIR=/Downloads/new_ssh_dir gopass init --crypto age <age1...>
```

## Features

* Encryption using `age` library, can be decrypted using the `age` CLI
* Support for native age, ssh-ed25519 and ssh-rsa recipients
* Support for encrypted ssh private keys
* Support for using GitHub users' private keys, e.g. `github:user` as recipient
* Automatic downloading and caching of SSH keys from GitHub
* Encrypted keyring for age keypairs
* Support for age plugins

## Usage with a yubikey

To use with a Yubikey, `age` requires the usage of the [age-plugin-yubikey plugin](https://github.com/str4d/age-plugin-yubikey/).

Assuming you have Rust installed:
```bash
$ cargo install age-plugin-yubikey
$ age-plugin-yubikey -i
<should be empty>
$ age-plugin-yubikey
✨ Let's get your YubiKey set up for age! ✨
<follow instructions to setup a PIV slot>
$ age-plugin-yubikey -i
<should display your PIV slot information now>
$ gopass age identities add
Enter the age identity starting in AGE-:
<paste the `AGE-PLUGIN-YUBIKEY-...` identity from the previous command>
Provide the corresponding age recipient starting in age1:
<paste the `age1yubikey1...` recipient from the previous command>
```

If gopass tells you `waiting on yubikey plugin...` when decrypting secrets, it probably is waiting for you to touch
your Yubikey because you've set a Touch policy when setting up your PIV slot.

## Roadmap

The future of this backend largely depends on what is happening in the `age` project itself.

Assuming `age` is supporting this, we'd like to:

* Finalize GitHub recipient support
* Add Hardware token support
* Make age the default gopass backend
