## Security

This project aims to provide a secure and dependable credential store that can
be used by individuals or teams.

We acknowledge that designing and implementing bullet-proof cryptography is very
hard and try to leverage existing and proven technology instead of rolling
our own implementations.

### Ultimate Goals of Security

* Confidentially - Ensure that only authorized parties can understand the data.
	gopass does only try to protect the content of the secrets.
	Neither their presence nor their names. Care must be taken not to
	disclose any confidential information through the name of the secrets.
* Authentication - Ensure that whoever supplies some secret if an authorized party.
	gopass fully relies on GnuPG in this regard.
* Integrity - Ensure that only authorized parties are allowed to modify data.
	Currently gopass makes no attempt at protecting the integrity of a store.
	However we plan to do this in the future.
* Nonrepudiation - Ensure that the involved parties actually transmitted and
	received messages. gopass makes not attempt to ensure this.

### Additional Usability Goals

* Availability - Secrets must always be readable by exactly the specified recipients.
* Sensible Defaults - This project shall try to make the right things easy to do and make the wrong things hard to do.

### Password Store Initialization

gopass only uses GPG for encrypting data. GPG needs to be properly set up before using gopass.
The user is responsible for distributing and importing the necessary public keys. Knowledge
of the web of trust model of GPG is assumed and necessary.

### Generating Passwords

Password generation uses the same approach as the popular tool `pwgen`.
It reads uses the `crypto/rand` to select random characters from the selected
character classes.

### Threat model

The threat model of gopass assumes there are no attackers on your local machine. Currently
no attempts are taken to verify the integrity of the password store. We plan on using
signed git commits for this. Anyone with access to the git repository can see which
secrets are stored inside the store, but not their content.
