# Backends

gopass supports pluggable backends for Storage (`store`), Encryption (`crypto`) and Source-Control-Management (`sync`).

As of today the names and responsibilities of these backends are still unstable and will probably change.

By providing suiteable backends gopass can use differnt kinds of encryption (see XC below) or storage.
For example it is pretty straight forward to add mercurial or bazaar as an SCM backend or
implement a Vault storage.

All backends are in their own packages below `backend/`. They need to implement the
interfaces defined in the backend package and have their identification added to
the context handlers in the same package.

## Storage Backends (storage)

### Filesystem (fs)

Right now there is only one storage backend implemented: Storing bytes on disk.

### In Memory (inmem)

This is a volatile in-memory backend for tests.

WARNING: All data is lost when gopass stops!

### Consul (consul)

This is an experimental storage backend that stores data in Consul.
Make sure to either combine this with a crypto backend or make sure
the data in Consul is properly protected as this backend does no
encryption on it's own.

#### Usage

Until Consul support is fully integrated you need to manually setup a mount
using the Consul backend.

Add a new mount to your `config.yml` (usually at `.config/gopass/config.yml`):

```bash
cat <<EOF >> $HOME/.config/gopass/config.yml
mounts:
  consul:
    path: plain-noop-consul+https://consul:8500/some/prefix/?token=some-token&datacenter=your-dc
EOF
```

This will setup an unecrypted backend, i.e. your secrets in Consul will be only
protected by Consul's ACLs and anyone who can access your Consul K/V prefix
can read your secrets.

You probably want to use a crypto backend to protect your secrets like in the
following example:

```bash
gopass xc generate
KEY=$(gopass xc list-private-keys | tail -1 | cut -d' ' -f1)
gopass init --path='xc-noop-consul+https://consul:8500/foo/bar/?token=some-token&datacenter=you-dc' --store=consul --crypto=xc --sync=noop $KEY
gopass mounts
```

## RCS Backends (rcs)

These are revision control backends talking to difference source control
management systems.

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

### Noop (noop)

This is a no-op backend for testing SCM-less support.

## Crypto Backends (crypto)

### CLI-based GPG (gitcli)

This backend is based on calling the gpg binary. This is the recommended backend
since we believe that it's the most secure and one and it's compatible with
other implementations of the `password-store` layout. However GPG is notoriously
difficult to use, there are lot's of different versions being used and the
output is not very machine readable. We will continue to support this backend
in the future, but we'd like to to move to a different default backend if possible.

### Plaintext (plain)

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

Please see the backend [Readme](https://github.com/justwatchcom/gopass/blob/master/backend/crypto/xc/README.md) for more details. Proper documentation for this
backend still needs to written and will be added at a later point.
