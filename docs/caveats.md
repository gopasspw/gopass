# Known Limitations and Caveats

## GnuPG

`gopass` uses [gpg](https://www.gnupg.org) to encrypt its secrets. This makes it easy to build a software we feel comfortable
trusting our credentials with, but `gpg` isn't known for being the most user-friendly software.

We try to work around some of the usability limitations of `gpg` but we always have to keep the security
goals in mind, so some features have to trade some usability against security and vice-versa.

## git history and local files

Please keep in mind that by default `gopass` stores its encrypted secrets in git. *This is a deviation
from the behavior of `pass`, which does not force you to use `git`.* Furthermore, the decision has some important
properties.

First it means that every user of `gopass` (and any attacker with access to your git repo) has a local
copy with the full history. If we revoke access to a store from an user and re-encrypt the whole store
this user won't be able to access any changed or added secrets but he'll be always able to access to
secrets by checking out old revisions from the repository.

**If you revoke access from a user you SHOULD change all secrets he had access to!**

## Private Keys required

Please note that we try to make it hard to lock yourself out from your secrets.
To ensure that a user is always able to decrypt his own secrets we require you
to have at least the public **and** private part of an recipient key available.
