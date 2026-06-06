# Use case: Team Workflows

## Summary

`gopass` is frequently used by teams to share secrets. A team is, from
gopass' point of view, a set of recipients (GPG or age keys) that share access
to one password store. Because gopass supports mounting any number of
substores, a single user can be a member of many teams at the same time, with
each team mapped to its own store.

This document describes the *supported* team workflows. It is the reference for
what gopass promises to do and the contract that the implementation and the
tests are expected to uphold. Anything not listed here is either unsupported or
best-effort.

Related issues that motivated this document:

* [#2762](https://github.com/gopasspw/gopass/issues/2762) – importing a
  recipient's public key via sync is broken (ID vs. filename mismatch).
* [#2620](https://github.com/gopasspw/gopass/issues/2620) – cloning a store as a
  new recipient removes all other recipients' public keys.
* [#1430](https://github.com/gopasspw/gopass/issues/1430) – no straightforward
  way to refresh an expired recipient public key.

For the detailed analysis and the implementation plan, see
[ADR A-14](../adr/A-14-team-workflows.md).

## Terminology

* **Team** – the set of recipients that can decrypt a given store. Implemented
  as the list of key IDs in the store's `.gpg-id` file.
* **Store / substore** – a password store, either the root store or a mounted
  substore. Each store has exactly one team at its root, with optional
  per-subdirectory recipient overrides.
* **Recipient** – a public key that is part of a team. Identified by a key ID.
  The canonical, unambiguous form of a recipient ID is the full GPG
  fingerprint (or the age recipient string for the age backend).
* **Member** – a person operating gopass who holds the *private* key matching
  one of the team's recipients.
* **Exported public keys** – the copies of all recipients' public keys that
  gopass stores inside the store under `.public-keys/<id>` so that new members
  can import them without an out-of-band key exchange.
* **Owner / maintainer** – a member who already has decryption access and is
  therefore able to add or remove recipients (re-encryption requires being able
  to decrypt).

## Core principles

1. **Recipient identity is canonical.** A recipient is always stored using its
   canonical key ID (full fingerprint for GPG). The `.gpg-id` entry and the
   `.public-keys/<id>` filename for the same recipient must always match. gopass
   normalizes any user-supplied identifier (email, short ID, fingerprint) to the
   canonical form before persisting it.
2. **Never silently drop a recipient.** Operations that re-encrypt a store must
   not remove a recipient just because the operating member cannot currently
   resolve that recipient's key locally. Dropping a recipient is always an
   explicit, confirmed action.
3. **Public keys travel with the store.** A member who can `git pull`/`clone` a
   store must be able to obtain every recipient's public key from the store
   itself, without an out-of-band exchange.
4. **Only owners change the team.** Adding or removing recipients re-encrypts
   the store and therefore requires decryption access. A read-only or
   not-yet-authorized member can only *request* access.
5. **Sync is safe.** `gopass sync` must never destroy recipient data on the
   remote. At worst it adds missing exported public keys.

---

## UC-1: Bootstrap a new store for a new team

**Actor:** The first team owner.

**Goal:** Create a brand-new store, become its first recipient, and publish it
to a remote so others can join.

**Preconditions:**

* The owner has a usable GPG (or age) key pair.
* A remote git repository (empty) is available, or will be created later.

**Main flow:**

```text
$ gopass init <owner-key-id>                      # root store, or
$ gopass init --store team-a --path ... <owner-key-id>   # as a substore

# publish to a remote
$ gopass git remote add --store team-a origin git@host:org/team-a.git
$ gopass sync --store team-a
```

**Postconditions:**

* `.gpg-id` contains the owner's canonical key ID.
* `.public-keys/<owner-id>` contains the owner's armored public key (when
  `core.exportkeys` is enabled, the default for the root store).
* The store is pushed to the remote.

**Acceptance criteria:**

* The recipient ID written to `.gpg-id` is the canonical fingerprint, identical
  to the `.public-keys/` filename (see UC core principle 1).
* A second person cloning the store immediately sees the owner's public key in
  `.public-keys/`.

---

## UC-2: Add a new store (team) for a user who is already in other teams

**Actor:** A user who already operates one or more gopass stores.

**Goal:** Mount an additional team store alongside the existing ones.

**Main flow (creating a new team store):** as UC-1 but always with
`--store <alias>` and `--path <path>`.

**Main flow (joining an existing team store):** see UC-4.

**Postconditions:**

* The new store is listed in `gopass mounts`.
* Operations on other stores are unaffected.

**Acceptance criteria:**

* gopass supports many concurrently mounted stores (see
  [multi-store](multi-store.md)); adding one more must not change the behavior
  of the others.
* Each store keeps its own team, its own crypto backend, and its own
  `core.exportkeys`/`core.autoimport` settings.

---

## UC-3: A new member requests access to an existing team

**Actor:** A prospective member who does **not** yet have decryption access.

**Goal:** Join a team by cloning the store and making their public key available
to an owner, who will grant access.

**Main flow:**

```text
# new member
$ gopass clone git@host:org/team-a.git team-a
# gopass detects the member cannot decrypt yet:
#   - it imports all recipients' public keys from .public-keys/
#   - it exports the member's own public key into .public-keys/
#   - it prints: "request access" and pushes the new public key
$ gopass sync --store team-a    # publishes the member's public key

# existing owner
$ gopass sync --store team-a            # pulls the new public key
$ gopass recipients add --store team-a <new-member-id>   # re-encrypts, grants access

# new member
$ gopass sync --store team-a    # pulls the re-encrypted secrets
$ gopass list team-a            # can now decrypt
```

**Postconditions:**

* The new member is in `.gpg-id` and `.public-keys/`.
* The store is re-encrypted for the extended team.

**Acceptance criteria (the bugs this must fix):**

* Cloning as a new member **must not** remove any existing recipient's public
  key from `.public-keys/` (regression test for
  [#2620](https://github.com/gopasspw/gopass/issues/2620)).
* The new member's public key is exported under their **canonical** ID, so the
  owner's `gopass recipients add` and the subsequent import on other members'
  machines resolve correctly (regression test for
  [#2762](https://github.com/gopasspw/gopass/issues/2762)).
* The clone flow and the `gopass setup --remote ...` flow must produce an
  identical store state (the two paths diverged in #2620).
* The member's root store location (`~/.password-store` vs.
  `~/.local/share/gopass/stores/root`) must not influence whether keys are
  preserved.

---

## UC-4: An owner adds a member to an existing team

**Actor:** An owner with decryption access.

**Goal:** Grant a new recipient access to the store and re-encrypt all secrets
for the new team.

**Main flow:**

```text
$ gopass sync --store team-a                       # get the member's public key
$ gopass recipients add --store team-a <member-id>
# gopass:
#   - normalizes <member-id> to its canonical fingerprint
#   - confirms the key identity with the owner
#   - adds it to .gpg-id and .public-keys/
#   - re-encrypts every secret for the new recipient set
$ gopass sync --store team-a                       # publish
```

**Postconditions:**

* All secrets are decryptable by the new member.
* `recipients.hash` is updated.

**Acceptance criteria:**

* The owner is warned and asked to confirm the identity of the key before it is
  added.
* If the member's public key is not in the local keyring but **is** present in
  `.public-keys/`, gopass offers to import it rather than failing.
* Adding a recipient whose key is already present updates/re-exports the key
  (useful when a member rotated or extended their key).

---

## UC-5: An owner removes a member from a team

**Actor:** An owner with decryption access.

**Goal:** Revoke a recipient's access and re-encrypt so the removed member can
no longer decrypt *new* secrets.

**Main flow:**

```text
$ gopass recipients remove --store team-a <member-id>
# gopass:
#   - resolves <member-id> (canonical ID, email, short ID, or .public-keys file)
#   - removes it from .gpg-id
#   - removes the corresponding .public-keys/<id> file
#   - re-encrypts every secret for the reduced recipient set
$ gopass sync --store team-a
```

**Postconditions:**

* The removed recipient is no longer in `.gpg-id` or `.public-keys/`.
* Secrets are re-encrypted; the removed member cannot decrypt anything written
  after removal.

**Acceptance criteria:**

* Removal must reliably match the recipient regardless of which form of the ID
  the operator typed (resolves the matching problem behind #2762).
* The removal cleans up the corresponding `.public-keys/` file (controlled,
  recipient-driven cleanup — distinct from the blanket `removeExtraKeys` that
  was disabled in #2620).
* gopass clearly states that revocation is **not retroactive**: anyone who had
  access to the git history can still decrypt old revisions. Rotation of the
  affected secrets is recommended and documented.
* gopass refuses to remove the last recipient (would make the store
  unreadable).

---

## UC-6: Refresh / rotate a recipient's public key (e.g. after expiry)

**Actor:** Any member (to refresh their own key) or an owner (to re-encrypt).

**Goal:** Replace an expired or rotated public key in the store and make sure
every member picks up the new key.

**Main flow:**

```text
# member whose key expired extends it locally, then:
$ gopass recipients update --store team-a            # (proposed command)
# gopass re-exports the member's current public key from the local keyring
# into .public-keys/<id>, overwriting the stale copy, and commits it.
$ gopass sync --store team-a

# other members, on next sync:
$ gopass sync --store team-a
# gopass detects that .public-keys/<id> is newer than the keyring copy and
# offers to import the refreshed key (subject to core.autoimport).
```

**Postconditions:**

* `.public-keys/<id>` holds the current key.
* Other members' keyrings are updated on sync (with confirmation).

**Acceptance criteria (the bug this must fix):**

* There is a first-class way to refresh a recipient's exported public key
  without re-adding the recipient or hand-editing files (resolves
  [#1430](https://github.com/gopasspw/gopass/issues/1430)).
* `core.autoimport` (and the interactive import prompt) must update an existing
  but *outdated/expired* key, not only import *missing* keys.
* gopass warns about recipients whose keys are expired or about to expire (see
  [ADR A-13](../adr/A-13-expired-gpg-key-handling.md)), pointing the user at
  this workflow.

---

## UC-7: Sync a team store safely

**Actor:** Any member.

**Goal:** Pull team changes and publish local changes without corrupting the
shared store.

**Main flow:**

```text
$ gopass sync --store team-a
# per store: pull -> import missing/updated public keys -> (if core.exportkeys)
#            export own/missing public keys -> push
```

**Acceptance criteria:**

* Sync never removes a recipient or a recipient's exported public key.
* Sync surfaces, but does not silently swallow, key-import failures.
* Sync is idempotent: running it twice with no remote changes produces no
  commits.

---

## UC-8: Per-subdirectory recipients within a team

**Actor:** An owner.

**Goal:** Restrict a subtree of a store to a subset of the team.

gopass offers *limited* support for additional `.gpg-id` files in
subdirectories. The preferred approach for strongly separated access remains a
separate substore (see [multi-store](multi-store.md)). This use case documents
the limited support so its boundaries are clear:

* A subdirectory `.gpg-id` overrides the root recipients for that subtree.
* All such recipients still need their public keys in the store's
  `.public-keys/`.
* Re-encryption respects the most specific `.gpg-id` for each secret.

---

## Non-goals

* gopass does **not** provide server-side access control. Repository read/write
  permissions are enforced by the git hosting platform (see
  [readonly-store](readonly-store.md)).
* gopass does **not** retroactively protect secrets from removed members; git
  history is accessible to anyone who ever had read access.
* gopass does **not** manage key trust or the web of trust; it relies on the
  member confirming key identities.
