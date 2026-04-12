# Security, Known Limitations, and Caveats

This project aims to provide a secure and dependable credential store that can be used by individuals or teams.

We acknowledge that designing and implementing bullet-proof cryptography is very hard and try to leverage existing and proven technology instead of rolling our own implementations.

## Security Goals

* **Confidentiality** - Ensure that only authorized parties can understand the data.
  * gopass attempts to protect the content of the secrets that it manages using [GNU Privacy Guard](#gnu-privacy-guard).
  * gopass does NOT protect the presence of the secrets OR the names of the secrets. Care must be taken not to disclose any confidential information through the
	name of the secrets.
* **Integrity** - Ensure that only authorized parties are allowed to modify data.
  * gopass makes no attempt at protecting the integrity of a store. However, we plan to do this in the future.
* **Availability** - Secrets must always be readable by exactly the specified recipients.
  * gopass provides fairly good availability due to its decentralized nature.
    For example, if your local password store is corrupted or destroyed, you can easily clone it from the Git server again.
    Conversely, if the Git server is offline or is destroyed, you (and your teammates) have a complete copy of all of the secrets on your local machine(s).
* **Non-repudiation** - Ensure that the involved parties actually transmitted and received messages.
  * gopass makes no attempt to ensure this.

### Additional Usability Goals

* Sensible Defaults - This project shall try to make the right things easy to do and make the wrong things hard to do.

## Threat Model

The threat model of gopass assumes there are no attackers on your local machine.
Currently no attempts are taken to verify the integrity of the password store.
We plan on using signed git commits for this.
Anyone with access to the git repository can see which secrets are stored inside the store, but not their content.

## GNU Privacy Guard

gopass uses [GNU Privacy Guard](https://www.gnupg.org) (or GPG for short) to encrypt its secrets.
This makes it easy to build software we feel comfortable trusting our credentials with.
The first production release of GPG was on [September 9th, 1999](https://en.wikipedia.org/wiki/GNU_Privacy_Guard#History) and by now it is mature enough for most security experts to place a high degree of confidence in the software.

With that said, GPG isn't known for being the most user-friendly software.
We try to work around some of the usability limitations of GPG but we always do so keeping security in mind.
This means that, in some cases, the project carefully makes some security trade-offs in order to achieve better usability.

Since gopass uses GPG to encrypt data, GPG needs to be properly set up beforehand.
(GPG installation is covered in the [gopass installation documentation](https://github.com/gopasspw/gopass/blob/master/docs/setup.md).)
However, basic knowledge of how [public-key cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography) and the [web of trust model](https://en.wikipedia.org/wiki/Web_of_trust) work is assumed and necessary.

## Generating Passwords

Password generation uses the same approach as the popular tool `pwgen`.
It uses `crypto/rand` to select random characters from the selected character classes.

## git history and local files

Please keep in mind that by default, gopass stores its encrypted secrets in git.
*This is a deviation from the behavior of `pass`, which does not force you to use `git`.* This has important implications.

First, it means that every user of gopass (and any attacker with access to your git repo) has a local copy with the full history.
If we revoke access to a store from a user and re-encrypt the whole store, this user won't be able to access any *changed* or *added* secrets -- but they'll always be able to access
secrets by checking out old revisions from the repository.

**If you revoke access from a user you SHOULD change all secrets they had access to!**

## Private Keys Required

Please note that we try to make it hard to lock yourself out from your secrets.
To ensure that a user is always able to decrypt his own secrets, gopass requires that the user has at least one private key available (that matches the current public keys on file for the password store).

---

## Security Engineering Highlights

The following design decisions and implementation choices contribute positively
to the security posture of gopass. They are documented here both as a record for
maintainers and as guidance for contributors who must preserve these properties.

### Binary Hardening

gopass is built with Position-Independent Executable (PIE) support, stripped
debug symbols, `-trimpath`, `CGO_ENABLED=0`, and the `netgo` build tag. This
eliminates CGo dependencies (and with them an entire class of memory-safety
issues) and makes cross-compilation straightforward.

### No Shell Injection

Every external command invocation — GPG, Git, and editors — uses
`exec.Command` with a string-slice argument list rather than constructing a
shell command string. This categorically prevents shell-injection attacks even
when user-supplied input reaches command arguments.

### Signed Update Verification

The built-in updater downloads a checksum file, verifies its GPG signature
against a hardcoded project public key, and enforces TLS 1.3 as a minimum
protocol version. An attacker would need both a fraudulent TLS certificate
**and** the project signing key to serve a malicious update.

### Age Agent Socket Security

Before connecting to the age-plugin agent socket, gopass checks that the
socket file has `0o600` permissions and that its owning UID matches the current
process UID. This prevents a local attacker from substituting a malicious
socket.

### Clipboard Auto-Clear

The clipboard helper spawns a detached `unclip` process that sleeps for a
configurable interval and then clears the clipboard. Before clearing, it
re-reads the clipboard and verifies an Argon2id checksum of the expected value
to ensure it only erases secrets it placed there (not content from an
unrelated application the user copied in the meantime).

### Debug Log Secret Protection

All types that hold secret material implement the `SafeStr()` interface, which
returns `"(elided)"` rather than the secret value. The `out` package respects
this interface everywhere. Full secret values are written to the debug log
**only** when `GOPASS_DEBUG_LOG_SECRETS=1` is explicitly set in the
environment.

### Temporary File Security

Temporary files created for editor sessions use `/dev/shm` on Linux, a private
ramdisk on macOS, and `0o600` permissions on all platforms. The files are
deleted in a deferred cleanup step so that even abnormal process exits
minimise the window during which plaintext is on disk.

### Recipient Validation

Before encrypting to a GPG key, gopass validates that the key is usable: it
checks for expiration, minimum trust level, and the presence of an encryption
sub-capability. Expired or untrusted keys are rejected, preventing silent
encryption to keys that can no longer decrypt.

### OpenBSD Pledge

On OpenBSD, gopass calls `protect.Pledge("stdio rpath wpath cpath tty proc
exec fattr")` to restrict the set of permitted syscalls to only those
actually needed. This limits the blast radius of any exploitation attempt
on that platform.
