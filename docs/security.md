# Security, Known Limitations, and Caveats

This project aims to provide a secure and dependable credential store that can be used by individuals or teams.

We acknowledge that designing and implementing bullet-proof cryptography is very hard and try to leverage existing and proven technology instead of rolling our own implementations.

## Security Goals

* **Confidentially** - Ensure that only authorized parties can understand the data.
  * gopass attempts to protect the content of the secrets that it manages using [GNU Privacy Guard](#gnu-privacy-guard).
  * gopass does NOT protect the presence of the secrets OR the names of the secrets. Care must be taken not to disclose any confidential information through the
	name of the secrets.
* **Integrity** - Ensure that only authorized parties are allowed to modify data.
  * gopass makes no attempt at protecting the integrity of a store. However, we plan to do this in the future.
* **Availability** - Secrets must always be readable by exactly the specified recipients.
  * gopass provides fairly good availability due to its decentralized nature. For example, if your local password store is corrupted or destroyed, you can easily clone it from the Git server again. Conversely, if the Git server is offline or is destroyed, you (and your teammates) have a complete copy of all of the secrets on your local machine(s).
* **Non-repudiation** - Ensure that the involved parties actually transmitted and received messages.
  * gopass makes no attempt to ensure this.

### Additional Usability Goals

* Sensible Defaults - This project shall try to make the right things easy to do and make the wrong things hard to do.

## Threat Model

The threat model of gopass assumes there are no attackers on your local machine. Currently no attempts are taken to verify the integrity of the password store. We plan on using signed git commits for this. Anyone with access to the git repository can see which secrets are stored inside the store, but not their content.

## GNU Privacy Guard

gopass uses [GNU Privacy Guard](https://www.gnupg.org) (or GPG for short) to encrypt its secrets. This makes it easy to build a software we feel comfortable
trusting our credentials with. The first production release of GPG was on [September 9th, 1999](https://en.wikipedia.org/wiki/GNU_Privacy_Guard#History) and it is mature enough for most security experts to place a high degree of confidence in the software.

With that said, GPG isn't known for being the most user-friendly software. We try to work around some of the usability limitations of GPG but we always do so keeping security in mind. This means that, in some cases, the project carefully makes some security trade-offs in order to achieve better usability.

Since gopass uses GPG to encrypt data, GPG needs to be properly set up beforehand. (GPG installation is covered in the [gopass installation documentation](https://github.com/justwatchcom/gopass/blob/master/docs/setup.md).) However, basic knowledge of how [public-key cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography) and the [web of trust model](https://en.wikipedia.org/wiki/Web_of_trust) is assumed and necessary.

## Generating Passwords

Password generation uses the same approach as the popular tool `pwgen`. It reads uses the `crypto/rand` to select random characters from the selected character classes.

## git history and local files

Please keep in mind that by default, gopass stores its encrypted secrets in git. *This is a deviation from the behavior of `pass`, which does not force you to use `git`.* This has important implications.

First, it means that every user of gopass (and any attacker with access to your git repo) has a local copy with the full history. If we revoke access to a store from an user and re-encrypt the whole store this user won't be able to access any changed or added secrets but he'll be always able to access to
secrets by checking out old revisions from the repository.

**If you revoke access from a user you SHOULD change all secrets he had access to!**

## Private Keys Required

Please note that we try to make it hard to lock yourself out from your secrets. To ensure that a user is always able to decrypt his own secrets, gopass requires that the user has at least one private key available (that matches the current public keys on file for the password store).
