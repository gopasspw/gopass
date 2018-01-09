Experimental Crypto Backend for gopass
======================================

This package contains an experimental crypto backend for gopass.
The goal is to provide an implementation that is feature complete
compared to the GPG backend but doesn't require any external binaries,
especially no GPG. Of course this would break compatilibity to existing
GPG deployments and users of different pass implementations, but
especially for closed teams with no existing GPG deployment this should
make little to no difference.

Motivation
----------

While GPG is believed to be very secure and it supports a wide range of
applications and devices, it's not really user friendly. Even passioned
[crypto experts](https://moxie.org/blog/gpg-and-me/) don't enjoy working with GPG and for
newcomers it's a major hurdle. For the gopass developers it's about the
most time consuming task to provide support and implement workaround for
GPG issues. This doesn't mean that GPG is bad, but security is hard and
complex and GPG adds a lot of flexiblity on top of that so the result
is complex and complicated.

WARNING
-------

We are no crypto experts. While this code uses professional implementations of
well known and rather easy to use crypto primitives there is still a lot of room
for making mistakes. This code so far has recieved no kind of security audit.
Please don't use it for anything critical unless you have reviewed and verified
it yourself and are willing to take any risk.

Status
------

Working, needs more of testing.

Design
------

* Hybrid encryption
    * Symmetric encryption for payload (secrets)
	* Using [Chacha20Poly1305](https://godoc.org/golang.org/x/crypto/chacha20poly1305) [AEAD](https://godoc.org/crypto/cipher#AEAD)
	* [Random session key](https://godoc.org/crypto/rand)
    * Asymmetric encryption per recipient
	* Using [Curve25519, XSalsa20, Poly1305 (NaCL Box)](https://godoc.org/golang.org/x/crypto/nacl/box)
	* (optional) Unencrypted Metadata
    * Disk format uses [protocol buffers version 3](https://developers.google.com/protocol-buffers/) encoding
* Keystore
    * Unencrypted public keys / metadata
    * Private Keys encrypted with [XSalsa20, Poly1305 (NaCL Secretbox)](https://godoc.org/golang.org/x/crypto/nacl/secretbox)
	* Using [Argon2 KDF](https://godoc.org/golang.org/x/crypto/argon2)
    * Key ID similar to GnuPG, using the low 64 bits of a SHA-3 / SHAKE-256 hash
    * Disk format uses [protocol buffers version 3](https://developers.google.com/protocol-buffers/) encoding
* Agent
    * (optional) Listens on Unix Socket
    * Invokes pinentry if necessary, caches passphrase

Attack Vectors
--------------

* Information Disclosure
    * Header
      * Session key is encrypted and authenticated using NaCl Box
      * Metadata is unencrypted, but unused right now
    * Body
      * Plaintext is encrypted and authenticated using Chacha20Poly1305 AEAD

Testing Notes
-------------

```bash
# Create two different homedirs
mkdir -p /tmp/gp1 /tmp/gp2

# Create a shared remote
mkdir -p /tmp/gpgit
cd /tmp/gpgit && git init --bare

# Generate two keypairs
GOPASS_HOMEDIR=/tmp/gp1 gopass xc generate
GOPASS_HOMEDIR=/tmp/gp2 gopass xc generate

# Get the key IDs (the init wizard doesn't support XC, yet)
GOPASS_HOMEDIR=/tmp/gp1 gopass xc list-private-keys
GOPASS_HOMEDIR=/tmp/gp2 gopass xc list-private-keys

# Init first password store
GOPASS_HOMEDIR=/tmp/gp1 gopass init --crypto=xc --sync=gitcli <key-id-1>

# add git remote
GOPASS_HOMEDIR=/tmp/gp1 ./gopass git remote add --remote origin --url /tmp/gpgit

# push to git remote (will produce a warning which can be ignored)
GOPASS_HOMEDIR=/tmp/gp1 ./gopass git push

# clone second store
GOPASS_HOMEDIR=/tmp/gp2 ./gopass clone --crypto=xc --sync=gitcli /tmp/gpgit

# Generate some secrets
GOPASS_HOMEDIR=/tmp/gp1 gopass generate foo/bar 24

# Sync stores
GOPASS_HOMEDIR=/tmp/gp1 gopass sync
GOPASS_HOMEDIR=/tmp/gp2 gopass sync

# Try to decrypt
GOPASS_HOMEDIR=/tmp/gp2 gopass show foo/bar # should fail

# Export recipient
GOPASS_HOMEDIR=/tmp/gp2 gopass xc export --id <key-id-2> --file /tmp/pub

# Import recipient
GOPASS_HOMEDIR=/tmp/gp1 gopass xc import --file /tmp/pub

# Add recipient
GOPASS_HOMEDIR=/tmp/gp1 gopass recipients add <key-id-2>

# Sync
GOPASS_HOMEDIR=/tmp/gp2 gopass sync

# Display secret
GOPASS_HOMEDIR=/tmp/gp2 gopass show foo/bar
```
