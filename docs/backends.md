# Backends

gopass supports pluggable backends for Storage and Revision Control System (`storage`) and Encryption (`crypto`).

As of today, the names and responsibilities of these backends are still unstable and will probably change.

By providing suitable backends, gopass can use different kinds of encryption or storage.
For example, it is pretty straightforward to add mercurial or bazaar as an SCM backend.

All backends are in their own packages below `backend/`. They need to implement the
interfaces defined in the backend package and have their identification added to
the context handlers in the same package.

## Storage and RCS Backends (storage)

* [fs](backends/fs.md) - Filesystem storage without RCS support
* [gitfs](backends/gitfs.md) - Filesystem storage with Git RCS
* [ondisk](backends/ondisk.md) - EXPERIMENTAL Fully encrypted filesystem storage with custom RCS

## Crypto Backends (crypto)

* [gpgcli](backends/gpg.md) - depends on a working gpg installation
* plain -  A no-op backend used for testing. WARNING: DOES NOT ENCRYPT!
* [xc](../internal/backend/crypto/xc/README.md) - EXPERIMENTAL custom crypto backend. WARNING: DO NOT USE!
* [age](backends/age.md) -  This backend is based on [age](https://github.com/FiloSottile/age). It adds an encrypted keyring on top (using age in scrypt password mode). It also has (largely untested) support for specifying recipients as github users. This will use their ssh public keys for age encryption. This backend might very well become the new default backend.

