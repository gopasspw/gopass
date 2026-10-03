# age crypto backend

The `age` backend is an experimental crypto backend based on [age](https://age-encryption.org). It adds an
encrypted keyring on top (using age passphrase or recipient encryption). It also has
(largely untested) support for specifying recipients as github users. This will
use their ssh public keys for age encryption.
It is well positioned to eventually replace `gpg` as the default crypto backend.

## Getting started

WARNING: This backend is experimental and the on-disk format likely to change.

To start using the `age` backend initialize a new (sub) store with the `--crypto=age` flag:

```
$ gopass age identities add [AGE-... age1...]
<if you do not specify an age secret key, you'll be prompted for one>
$ gopass init --crypto age
```

or use the wizard that will help you create a new age key:
```
$ gopass setup --crypto age
```

This will automatically create a new age keypair and initialize the new store.

Existing stores can be migrated using `gopass convert --crypto age`.

N.B. for a fully scripted or **non-interactive setup**, you can use the `GOPASS_AGE_PASSWORD` env variable
to set your identity file secret passphrase, and specify the age identity and recipients
that should be used for encrypting/decrypting passwords as follows:
```
$ gopass age identities add <AGE-...> <age1...>
$  GOPASS_AGE_PASSWORD=mypassword gopass init --crypto age <age1...>
```
Notice the extra space in front of the command to skip most shell's history.
You'll need to set your name and username using `git` directly if you're using it as storage backend (the default one).

For test automation or other environments where `pinentry` is unavailable, set
`GOPASS_AGE_STDIN_PASSPHRASE` to force gopass to read the passphrase from the terminal instead.

You can also specify the ssh directory by setting environment variable
```
$  GOPASS_SSH_DIR=/Downloads/new_ssh_dir gopass init --crypto age <age1...>
```

## Features

* Encryption using `age` library, can be decrypted using the `age` CLI
* Support for native age, ssh-ed25519 and ssh-rsa recipients
* Support for encrypted ssh private keys
* Support for using GitHub users' private keys, e.g. `github:user` as recipient
* Automatic downloading and caching of SSH keys from GitHub
* Encrypted keyring for age keypairs
* Support for age plugins
* Caching of passphrases via an agent

## Identity preference

Run `gopass age identities sort` to configure the order in which identities are
tried for decryption. Save the selection to persist the public recipient order
in `age.identities`. An explicitly preferred plugin identity is tried before a
native age identity, even when both can decrypt the entry.

An identity that does not match the entry allows the next identity to be tried.
An authentication failure or cancellation reported as an error stops that
decryption attempt; preference ordering does not require every available
identity to authenticate.

The agent retains the identity order loaded when it was unlocked. After changing
the preference, run `gopass age agent lock` to clear the cached identities; the
next decryption loads the updated order.

## Agent

The age backend comes with an agent that can cache unlocked age identities.
The agent is started automatically by gopass if it's not already running.
You can disable the agent by setting `age.agent-enabled` to `false` in your gopass config.

The agent performs secret decryption without writing unlocked identities to disk. Native age
identities are loaded into the agent after their keyring is unlocked. For passphrase-protected
SSH identities, gopass prompts only when a matching key is first used, then sends the unlocked
identity—not its passphrase—to the agent over the protected local socket. Locking the agent or
letting its timeout expire clears both native age and SSH identities from agent memory.
The agent listens on a unix socket at `$XDG_RUNTIME_DIR/gopass/gopass-age-agent.sock`.

You can interact with the agent using the following commands:
- `gopass age agent`: starts the agent in the foreground.
- `gopass age lock`: locks the agent, clearing all cached identities.

## Hardware-backed keyring unlocking

A hardware age plugin can protect the local identity keyring rather than every
store entry. For example, `age-plugin-se` can require Touch ID to unlock ordinary
age identities, which the agent then caches for the session. Reading further
entries encrypted for those identities does not invoke the hardware plugin again.
This also works with other age plugins; gopass does not implement the hardware
authentication itself.

Keep the software identities used by your store in the gopass keyring. Put the
hardware plugin identity in a separate, plaintext age identity file, outside the
encrypted keyring. Protect that file's permissions and retain any backups required
by the plugin. The file must be readable before the keyring can be unlocked.

For an existing passphrase-protected keyring, configure its new protection and
migrate it explicitly. Substitute your bootstrap file and its public recipient:

```fish
gopass config age.keyring-identities ~/.config/gopass/age/keyring-unlock.txt
gopass config age.keyring-recipients age1...
gopass age identities reencrypt
gopass config age.agent-enabled true
gopass config age.agent-timeout 300
gopass age agent stop
gopass age agent unlock
```

Migration asks for the old keyring passphrase and verifies hardware decryption
of the new envelope before atomically replacing the keyring. Back up the original
encrypted keyring before migrating. This command does not change the recipients
or ciphertext of store entries. Entries encrypted only for a hardware recipient
still require hardware authentication for each decryption; session caching works
for entries encrypted for the software identities inside the keyring.

The first entry read can also unlock the session automatically. With the agent
disabled, each entry read unlocks the keyring directly. An authentication error
or cancellation fails the operation without falling back to a password. Do not
leave an ordinary recovery identity in the bootstrap file if hardware
authentication should be required on this device.

`gopass age agent lock` clears the cached identities. After the configured idle
timeout, the next read requires authentication again (`0` disables the timeout).
Changes to the keyring, bootstrap file, preferred identity order or session
configuration invalidate the cached session on the next read. A lock event while
authentication is pending prevents that attempt from loading a new session.
The agent does not detect screen locking itself; a desktop hook can invoke the
lock command. Local processes running as the same user can use an unlocked agent.

Adding, removing or re-encrypting identities clears the agent and can require
additional hardware authentication to verify the replacement keyring. Keep
`age.keyring-recipients` configured for these writes; gopass refuses to silently
downgrade an existing recipient-encrypted keyring to passphrase protection.
Restart an agent from an older gopass version before using this feature.

For recovery, include an additional public recipient in `age.keyring-recipients`
(comma-separated), and keep its private identity securely on another device or
offline. Hardware-bound identities may not be portable. To recover, set
`age.keyring-identities` to an independent file containing the recovery identity.
The encrypted identity keyring can also be decrypted with the standard `age` CLI.
The store itself still needs a separate backup.

## Usage with a yubikey

To use with a Yubikey, `age` requires the usage of the [age-plugin-yubikey plugin](https://github.com/str4d/age-plugin-yubikey/).

Assuming you have Rust installed:
```bash
$ cargo install age-plugin-yubikey
$ age-plugin-yubikey -i
<should be empty>
$ age-plugin-yubikey
✨ Let's get your YubiKey set up for age! ✨
<follow instructions to setup a PIV slot>
$ age-plugin-yubikey -i
<should display your PIV slot information now>
$ gopass age identities add
Enter the age identity starting in AGE-:
<paste the `AGE-PLUGIN-YUBIKEY-...` identity from the previous command>
Provide the corresponding age recipient starting in age1:
<paste the `age1yubikey1...` recipient from the previous command>
```

If gopass tells you `waiting on yubikey plugin...` when decrypting secrets, it probably is waiting for you to touch
your Yubikey because you've set a Touch policy when setting up your PIV slot.

## Roadmap

The future of this backend largely depends on what is happening in the `age` project itself.

Assuming `age` is supporting this, we'd like to:

* Finalize GitHub recipient support
* Add Hardware token support
* Make age the default gopass backend
