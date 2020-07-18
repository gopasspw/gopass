# Backends

gopass supports pluggable backends for Storage and Revision Control System (`storage`) and Encryption (`crypto`).

As of today, the names and responsibilities of these backends are still unstable and will probably change.

By providing suitable backends, gopass can use different kinds of encryption or storage.
For example, it is pretty straightforward to add mercurial or bazaar as an SCM backend.

All backends are in their own packages below `backend/`. They need to implement the
interfaces defined in the backend package and have their identification added to
the context handlers in the same package.

## Storage and RCS Backends (storage)

### Filesystem (fs)

This is the simplest storage backend. It stores the encrypted (see "Crypto Backends") data directly in the filesystem. It does not use version control and the relevant operations are no ops.

### Filesystem with Git (gitfs)

This is the default storage backend. It stores the encrypted data directly in the
filesystem. It uses an external git binary to provide history and remote sync
operations.

### On Disk (ondisk)

This is an experimental on disk K/V backend. It stores the encrypted data in the
filesystem in a content adressable manner. It is fully encrypted, including
metadata. Content can be encrypted using any of the supported encryption
backend but it's only being tested with age. Metadata is always encrypted with
age.

This might become the default storage and RCS backend in gopass 2.x.

**WARNING**: The disk format is still experimental and will change. **DO NOT USE** unless you want to help with the implementation.

This backend can be fully decrypted and parsed without gopass. The index is
age encrypted serialized JSON. It maps the keys (secret names) to content
addressable blobs on the filesystem. Those are usually encrypted with age.
The age keyring itself is also age encrypted serialized JSON.

TODO: Add commands to decrypt an ondisk/age store without gopass.

#### Background: How do access ondisk secrets without gopass

This section assumes `age` and `jq` are properly installed.

```
# Decrypt the gopass-age keyring
age -d -o /tmp/keyring ~/.config/gopass/age-keyring.age
# Extract the private key
cat /tmp/keyring | jq ".[1].identity" | cut -d'"' -f2 > /tmp/private-key
# Decrypt the index
# TODO
# Locate the latest secrets
# TODO
# Decrypt it
age -d -i /tmp/private-key -o /tmp/plaintext ~/.local/share/gopass/stores/root/foo.age
```

## Crypto Backends (crypto)

### CLI-based GPG (gpgcli)

This backend is based on calling the `gpg` binary. This is the recommended backend
since we believe that it's the most secure one and it is compatible with
other implementations of the `password-store` vault layout. However GPG is notoriously
difficult to use, there are lots of different versions being used, and the
output is not very machine readable. We will continue to support this backend
in the future, but we'd like to to move to a different default backend if possible.

### Plaintext (plain)

This is a no-op backend used for testing.

**WARNING**: Do not use unless you know what you are doing.

### NaCl-based custom crypto backend (xc)

**WARNING**: The future of this backend is unclear. If [age](https://github.com/FiloSottile/age) proves feasible this backend will be dropped. Do not use in production!

We implemented a pure-Go backend using a custom message format based on the excellent
[NaCl library](https://nacl.cr.yp.to/) [packages](https://godoc.org/golang.org/x/crypto/nacl).
The advantage of this backend is that it's properly integrated into gopass, has a stable API,
stable error handling and only the features we absolutely need. This makes it
very easy to setup, use and support. The big drawback is that it didn't receive
any of the scrutiny and peer review that GPG got. And since it's very easy to
make dangerous mistakes when dealing with cryptography - even when it's only
using existing building blocks - we're a little wary to recommend it for broader use.

Also it requires its own Keyring/Agent infrastructure, as the keyformat is quite
different from what GPG is using.

Please see the backend [Readme](https://github.com/gopasspw/gopass/blob/master/internal/backend/crypto/xc/README.md) for more details. Proper documentation for this
backend still needs to written and will be added at a later point.

### Age crypto backend (age)

This backend is based the [age](https://github.com/FiloSottile/age). It adds an
encrypted keyring on top (using age in scrypt password mode). It also has
(largely untested) support for specifying recipients as github users. This will
use their ssh public keys for age encryption.

This backend might very well become the new default backend.

