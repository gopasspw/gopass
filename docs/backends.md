# Backends

gopass supports pluggable backends for Storage (`store`), Encryption (`crypto`) and Source-Control-Management (`sync`).

As of today the names and responsibilities of these backends are still unstable and will probably change.

By providing suiteable backends gopass can use differnt kinds of encryption (see XC below) or storage.
For example it is pretty straight forward to add mercurial or bazaar as an SCM backend or
implement a Vault storage.

All backends are in their own packages below `backend/`. They need to implement the
interfaces defined in the backend package and have their identification added to
the context handlers in the same package.

## Storage Backends (store)

### Filesystem (fs)

Right now there is only one storage backend implemented: Storing bytes on disk.

## SCM Backends (sync)

### CLI-based git (gitcli)

The CLI-based git backend requires a properly configured git binary. It has the
most features of all SCM backends and is pretty stable. One major drawback is that
it sometimes fails if commit signing is enabled and the interaction with GPG
fails.

### gogit.v4 (gogit)

This backend is based on the amazing work of [source{d}](https://sourced.tech/)
and implements a pure-Go SCM backend. It works pretty well but there is one major
show stopped: It only supports fast-forward merges. Unfortunately this makes
it unseable for most gopass usecases. However we still keep this backend around
in case upstream manages to implement proper merges. In that case this will
quickly become the default SCM backend.

### Git Mock

This is a no-op backend for testing SCM-less support.

## Crypto Backends (crypto)

### CLI-based GPG

This backend is based on calling the gpg binary. This is the recommended backend
since we believe that it's the most secure and one and it's compatible with
other implementations of the `password-store` layout. However GPG is notoriously
difficult to use, there are lot's of different versions being used and the
output is not very machine readable. We will continue to support this backend
in the future, but we'd like to to move to a different default backend if possible.

### GPG Mock

This is a no-op backend used for testing.

### openpgp pure-Go (openpgp)

We're planning to implement a pure-Go GPG backend based on the [openpgp package](https://godoc.org/golang.org/x/crypto/openpgp),
but unfortunately this packaged doesn't support recent versions of GPG.
If the openpgp package or a proper fork gains support for recent GPG versions
we'll try to move to this yet-to-be-written backend as the default backend.

### NaCl-based custom crypto backend (xc)

We implemented a pure-Go backend using a custom message format based on the excellent
[NaCl library](https://nacl.cr.yp.to/) [packages](https://godoc.org/golang.org/x/crypto/nacl).
The advantage of this backend that it's properly integrated into gopass, has a stable API,
stable error handling and only the feature we absolutely need. This makes it
very easy to setup, use and support. The big drawback is that it didn't receive
any of the scrunity and peer review that GPG got. And since it's very easy to
make dangerous mistakes when dealing with cryptography - even when it's only
using existing building blocks - we're a little wary to recommend it for broader use.

Also it requires it's own Keyring/Agent infrastructure as the keyformat is quite
different from what GPG is using.
