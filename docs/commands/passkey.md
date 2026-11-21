# `passkey` command

The `passkey` command manages and uses passkeys: the public key credentials
that web sites accept instead of a password through the
[Web Authentication API](https://www.w3.org/TR/webauthn-3/) (WebAuthn).

gopass acts as a software authenticator in the *synced passkey* model, like
the credential managers built into browsers and operating systems do. The
private key of a passkey lives in an encrypted secret and is available on
every device that has access to the password store.

Note: Passkeys are used by web sites through a browser. The command line
described here is the foundation for that and useful for scripting, testing
and maintenance. Using a passkey to sign in to a site needs a browser
extension that talks to gopass, e.g. through
[gopass-jsonapi](https://github.com/gopasspw/gopass-jsonapi).

## Synopsis

```
$ gopass passkey [list [rp-id]]
$ gopass passkey create <rp-id> <user>
$ gopass passkey assert <rp-id|secret> --challenge <challenge>
$ gopass passkey remove <rp-id|secret>
$ gopass passkey import <file>
$ gopass passkey export [rp-id]
```

## Terminology

* The *relying party* is the web site a passkey is used with. It is identified
  by its *RP ID*, a domain name like `example.com`.
* The *origin* is the address of the page that asks for a passkey, e.g.
  `https://login.example.com`. A page may only use a passkey if the RP ID is
  its own domain or a parent domain of it that is not a public suffix. This is
  what makes passkeys resistant to phishing. gopass enforces it.
* A *challenge* is a random value the relying party wants to have signed.
* The *user handle* is an identifier the relying party assigned to a user
  account. In contrast to the user name it never changes.

## How passkeys are stored

A passkey is a single line in a secret named `passkeys/<rp-id>/<user>`:

```
$ gopass show passkeys/example.com/alice

passkey: {"v":1,"rpId":"example.com","userName":"alice","userDisplayName":"Alice","userHandle":"...","credentialId":"...","key":"...","signCount":0,"coseAlg":-7,"backupEligible":true,"backupState":true,"creationDate":"2026-11-20T18:51:15Z"}
```

The line holds everything that is needed to use the passkey, including the
private key (`key`, a PKCS#8 structure). Binary values are base64url encoded.
The line is hidden by `gopass show` if `show.safecontent` is enabled. The
password of such a secret, its first line, is not used. Members of the line
that this version of gopass does not know, e.g. ones a later version added,
are kept when gopass changes the secret.

A secret holds all passkeys of a user account. It can have more than one
`passkey:` line, e.g. after importing passkeys for the same account from
different credential managers or after registering with a site twice. gopass
never removes a passkey on its own: whether a site accepted a new passkey is
not known to gopass, and a passkey that works must not be lost to one that
does not. Remove a passkey that is no longer needed with
`gopass passkey remove`.

The names of secrets are not encrypted. Storing passkeys in a folder of their
own means that the secrets holding the passkeys of a site can be found without
decrypting anything. It also means that the names of the sites and accounts
you have passkeys for are visible to anyone with access to the store, just
like the names of all other secrets.

The name of a secret is only used to find it. What a passkey is bound to is
decided by the `rpId` inside the encrypted line alone. A passkey that sits in
the folder of another relying party is ignored.

Passkeys are found in every folder named `passkeys`, also inside mounted
stores: `team/passkeys/example.com/alice` is a passkey you share with your
team. New passkeys are always written below the folder configured as
`passkey.prefix` (default: `passkeys`), never into any other folder. They are
encrypted for the recipients of that part of your store. If the same passkey
is found more than once, the one below the prefix is used.

### Signature counter

Authenticators can count how often a passkey was used, so that a relying party
can detect a cloned authenticator. A passkey that is synced between devices
can not maintain such a counter: a sync that is late or missing would make the
counter go backwards and the relying party suspect a clone. Passkeys are
therefore created with the counter fixed at zero, as recommended for synced
passkeys.

If you use your store on a single device you can opt in to a counter for new
passkeys with `gopass config passkey.signcount true`. The counter is then
written back to the secret with every use. The first value a site sees for
such a passkey is `1`.

### Syncing

Passkeys are synced like every other secret. If a new passkey was written to
the store but the store then fails, e.g. because the remote can not be
reached, the passkey is still stored. gopass completes the registration and
warns about it, instead of failing and leaving the site without the passkey
you just stored. Run `gopass sync` once the remote is available again.

## User presence and user verification

Every use of a passkey must be confirmed. This is the *test of user presence*.
The command asks on the terminal before it creates a passkey or signs with
one. Use the global `--yes` flag to skip the question in scripts:
`gopass --yes passkey assert ...`.

A relying party can also ask for the user to be *verified*, i.e. for proof
that it is you who is using the passkey, e.g. a PIN or a fingerprint. gopass
has no way to verify a user on its own. All it knows is that the store could
be decrypted, and it can not tell whether you just entered a passphrase or
whether an agent still had the key. A confirmation is therefore reported as
user presence only, and a site that *requires* user verification can not be
used.

If you protect your store well enough to consider every use of it as
verified, you can say so:

```
$ gopass config passkey.user-verification true
```

A ceremony you confirmed is then reported as user verified, which is how
many password managers treat an unlocked vault. Do not enable it if e.g.
your GPG key has no passphrase or your agent caches it for a long time on a
computer others have access to. Note that the confirmation can be skipped with
`--yes`: together with this option, a ceremony is then reported as user
present and verified without any interaction.

This option, like all `passkey.*` options, is only read from your own config
(`gopass config passkey.user-verification true`), the system config and the
environment, never from the config of a store. That is versioned with the
store, and anyone who can push to the store could otherwise claim user
verification for you, or have new passkeys written to a folder they can read.

## Commands

### `passkey list`

Lists the stored passkeys, optionally limited to those of one relying party.
The secrets holding them are decrypted to do so.

```
$ gopass passkey list
SECRET                      RP ID        USER           ALGORITHM  CREATED     CREDENTIAL ID
passkeys/example.com/alice  example.com  alice (Alice)  ES256      2026-11-20  3yZL-uu9GR2Vmm6_61WyIlXIfjNzKWZB8bfjMX3fWmI
```

### `passkey create`

Creates a passkey for a user account at a relying party, stores it in
`passkeys/<rp-id>/<user>` and prints the registration response as JSON. The
response has the format a browser returns from
[`PublicKeyCredential.toJSON()`](https://www.w3.org/TR/webauthn-3/#dictdef-registrationresponsejson)
and contains the public key of the passkey. If the output is a terminal, a
line saying where the passkey was stored precedes it.

```
$ gopass passkey create example.com alice --challenge cmVnaXN0ZXI
```

A passkey that already exists for the account is kept, see above. If no
`--user-handle` is given, the passkey is created for the account of the user
with that name if `passkeys/<rp-id>/<user>` holds one, and for a new account
otherwise. Only that secret is decrypted to find out.

| Flag             | Aliases | Description                                                                                            |
|------------------|---------|--------------------------------------------------------------------------------------------------------|
| `--algorithm`    | `-a`    | Signature algorithm: `ES256` (default), `EdDSA` or `RS256`.                                            |
| `--display-name` | `-n`    | Friendly name of the user account. Defaults to the user name.                                          |
| `--user-handle`  |         | User handle, base64url encoded. Defaults to the one of the user with that name, or a random one.       |
| `--challenge`    | `-c`    | Challenge of the relying party, base64url encoded. Defaults to a random one.                           |
| `--origin`       | `-o`    | Origin of the page that requests the passkey. Defaults to `https://<rp-id>`.                           |

### `passkey assert`

Signs the challenge of a relying party with a passkey and prints the
authentication response as JSON, in the format of
[`AuthenticationResponseJSON`](https://www.w3.org/TR/webauthn-3/#dictdef-authenticationresponsejson).
The passkey is selected by the RP ID or by the name of the secret holding it.

```
$ gopass passkey assert example.com --challenge c2lnbiBpbg
$ gopass passkey assert passkeys/example.com/alice --challenge c2lnbiBpbg --origin https://login.example.com
```

A secret that is named can be anywhere in the store, e.g.
`websites/example.com` next to the password, and only that secret is
decrypted. Given an RP ID, the secrets in `passkeys/<rp-id>/` (and in the
other folders named `passkeys`) are decrypted once each. A secret named like
the RP ID that holds no passkey, like the `example.com` of many stores, does
not get in the way.

If more than one passkey matches you are asked to pick one. The choices show
when a passkey was created and the start of its ID, so that two passkeys of
the same account can be told apart. In scripts use `--credential` to select
one.

| Flag                  | Aliases | Description                                                                              |
|-----------------------|---------|------------------------------------------------------------------------------------------|
| `--challenge`         | `-c`    | Challenge of the relying party, base64url encoded. Required.                             |
| `--origin`            | `-o`    | Origin of the page that requests the assertion. Defaults to `https://<rp-id>`.           |
| `--credential`        |         | ID of the passkey to use, base64url encoded.                                             |
| `--user-verification` |         | Requirement of the relying party: `required`, `preferred` (default) or `discouraged`.    |

### `passkey remove`

Removes a passkey from the secret holding it, after asking for confirmation.
The passkey is selected like for `assert`: by the RP ID or by the name of the
secret, with `--credential` or a choice if more than one matches. The secret
and everything else in it are kept.

```
$ gopass passkey remove example.com
$ gopass passkey remove passkeys/example.com/alice --credential 3yZL-uu9GR2Vmm6_61WyIlXIfjNzKWZB8bfjMX3fWmI
```

The site still lists the passkey. Remove it there as well.

| Flag           | Aliases | Description                                         |
|----------------|---------|-----------------------------------------------------|
| `--credential` |         | ID of the passkey to remove, base64url encoded.     |
| `--force`      | `-f`    | Remove the passkey without asking for confirmation. |

### `passkey import` and `passkey export`

Passkeys can be moved between credential managers with the FIDO Alliance
[Credential Exchange Format](https://fidoalliance.org/specifications-credential-exchange-specifications/)
(CXF).

```
$ gopass passkey import export.json
$ gopass passkey export --output passkeys.json
$ gopass passkey export example.com
```

`import` reads the passkeys of an unencrypted CXF JSON document, or of
standard input if the file is `-`. Other kinds of credentials in the document
are ignored. Importing a passkey that is already stored does not duplicate it.

`export` writes all passkeys, or those of one relying party, as a CXF JSON
document to standard output or to the file given with `--output`. An existing
file is not overwritten. The file is created readable for you only. On
Windows, where that is decided by the folder, keep it in a folder nobody else
has access to.

WARNING: The export is not encrypted and contains the private keys. Delete it
as soon as it is imported elsewhere.

The format has no signature counter. Passkeys with a counter are left out of
an export, as the format demands, and imported passkeys never count.

## Supported features

| Feature                              | Status                                                       |
|--------------------------------------|--------------------------------------------------------------|
| Algorithms                           | ES256 (ECDSA P-256), EdDSA (Ed25519), RS256 (RSA 2048)       |
| Discoverable credentials             | Yes, every passkey is discoverable                           |
| Attestation                          | None (`none` attestation statement format)                   |
| Extensions                           | `credProps`. Data of other extensions is kept on import.     |
| Related origin requests              | No                                                           |
| Hardware backed keys                 | No                                                           |

## Config options

See [config.md](../config.md): `passkey.prefix`, `passkey.signcount` and
`passkey.user-verification`. They are only read from your own config, the
system config and the environment, not from the config of a store.
