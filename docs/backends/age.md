# age backend

The `age` backend is an experimental crypto backend based on [age](https://age-encryption.org). It adds an
encrypted keyring on top (using age in scrypt password mode). It also has
(largely untested) support for specifying recipients as github users. This will
use their ssh public keys for age encryption.
It is well positioned to eventually replace `gpg` as the default crypto backend.

## Getting started

WARNING: This backend is experimental and the on-disk format likely to change.

To start using the `age` backend initialize a new (sub) store with the `--crypto=age` flag:

```
gopass init --crypto age
gopass recipients add github:user
```

This will automatically create a new age keypair and initilize the new store.

Existing stores can be migrated using `gopass convert --crypto age`.

## Features

* Encryption using `age` library, can be decrypted using the `age` CLI
* Support for native age, ssh-ed25519 and ssh-rsa recipients
* Support for encrypted ssh private keys
* Support for using GitHub users' private keys, e.g. `github:user` as recipient
* Automatic downloading and caching of SSH keys from GitHub
* Encrypted keyring for age keypairs

## Roadmap

The future of this backend largely depends on what is happening in the `age` project itself.

Assuming `age` is supporting this, we'd like to:

* Finalize GitHub recipient support
* Add Hardware token support
* Make age the default gopass backend

