# A-13: Expired GPG Key Handling and Recipient Validity Warnings

**Status:** partially implemented — core silent-drop warning shipped; remaining work tracked below  
**Source:** [GitHub Issue #2885](https://github.com/gopasspw/gopass/issues/2885)

---

## Background

When a recipient's GPG key expires, gopass silently drops that recipient from
the encryption target list. Secrets written after the key expires can no
longer be decrypted by that recipient. Neither the writing user nor the
affected recipient receives any notification that this happened.

The encryption path is:

```
Set() → useableKeys() → FindRecipients() → [expired key silently filtered]
      → Encrypt(filtered_list)            [warning in Encrypt() never fires]
```

`FindRecipients()` calls `KeyList.UseableKeys()`, which returns only keys
whose `ExpirationDate` is either zero or in the future. The difference between
the original recipient list and the returned key list was never surfaced to the
user.

An existing `CheckRecipients()` function in `internal/store/leaf/recipients.go`
already performs the correct per-recipient check and is called before
`RecipientsAdd`, but was not called on the write path.

---

## Implemented fix (this branch)

`useableKeys()` in `internal/store/leaf/store.go` now iterates over the
original recipient list and, for each recipient that `FindRecipients` returns
no useable key for, emits an `out.Warningf` message naming that recipient.
The return value (the filtered key list used for encryption) is unchanged.

`Encrypt()` in `internal/backend/crypto/gpg/cli/encrypt.go` was also updated
to use `out.Warningf` instead of `out.Printf` for its own per-recipient check,
so the severity is correct if that path is ever reached.

A regression test (`TestSetWarnsAboutInvalidRecipient`) was added to
`internal/store/leaf/write_test.go`.

---

## Remaining work

### R-1: Add recipient key-expiry check to `gopass audit`

**Problem:** `gopass audit` checks only password strength (crunchy, HIBP). It
does not check whether any recipient's key is expired or about to expire.
A team could run `gopass audit` regularly and still have no indication that
a re-encryption silently excluded a recipient.

**Recommendation:** Add a recipient-validity check to `internal/audit/`. The
`Auditor` type already has a `secretGetter` interface; a separate
`RecipientAuditor` (or an additional pass in the existing `Batch` method) could:

1. Collect the union of all recipient IDs across all secrets (or from the
   `.gpg-id` files) without decrypting anything.
2. Call `crypto.FindRecipients(ctx, id)` for each one.
3. Report any ID with no useable key as an error and any ID whose
   `ExpirationDate` is within a configurable window (default 60 days) as a
   warning.

This requires plumbing a `Crypto` backend reference into the audit path, which
the current `secretGetter` interface does not expose. A separate interface or
an `Auditor` constructor parameter is the cleanest extension point.

**Configuration key (proposed):** `audit.recipient-expiry-warning-days`
(default `60`; `0` disables the check).

---

### R-2: Proactive expiry warning on sync and fsck

**Problem:** Users learn that a key has expired only when they attempt to
write a secret. There is no early warning before expiry, and no warning to
the affected recipient when they pull a store that now contains secrets they
cannot decrypt.

**Recommendation:**

- In `gopass sync` (`internal/action/sync.go`) and `gopass fsck`
  (`internal/store/leaf/fsck.go`), call a lightweight
  `CheckRecipientExpiry(ctx, warningDays int)` helper (to be added to
  `internal/store/leaf/recipients.go`) that:
  1. Loads all recipient IDs from the store's `.gpg-id` files.
  2. Looks up each key via `FindRecipients`.
  3. For keys that are already expired: `out.Warningf`.
  4. For keys expiring within `warningDays`: `out.Warningf` including the
     expiry date and the recovery instructions (see R-3).

- The check must be cheap: it reads only the local keyring and `.gpg-id`
  files; no network calls or decryption.

**Configuration key (proposed):** `core.recipient-expiry-warning-days`
(default `60`; `0` disables). Distinct from the audit key so that the two
features can be tuned independently.

---

### R-3: Document the key-refresh recovery flow

**Problem:** The recovery path when a key has expired is not obvious. The
issue reporter required a multi-step manual process. A simpler path already
exists but is undocumented:

1. Recipient L extends the key expiry locally: `gpg --edit-key <L>` → `expire`.
2. L exports the updated key: `gpg -a --export <L> > L.pub.asc`.
3. L replaces `.public-keys/<L>` in the store with the new export, commits,
   and pushes.
4. Any other recipient A runs `gopass recipients add <L>` (the existing
   `AddRecipient` handler already asks for confirmation to re-encrypt when the
   key is already in the store).

Step 4 is the re-encryption trigger. Gopass already supports this workflow;
it just needs to be documented.

**Recommendation:**

- Add a section "Recipient key expiry" to `docs/commands/recipients.md`
  describing the flow above.
- Update the warning message emitted by R-1/R-2 to include a short hint, e.g.:
  `"Run 'gopass recipients add <id>' after the key is refreshed to re-encrypt."`.

---

### R-4: Detect that the committed public key differs from the local keyring

**Problem:** A recipient may extend their key locally and not update the copy
in `.public-keys/`. Conversely, another recipient may import an updated key
from `.public-keys/` while the local keyring still has the old (expired)
version. In both cases gopass has no way to tell the user that the two copies
are out of sync.

**Recommendation:** In `gopass fsck` or `gopass recipients`, compare the
`ExpirationDate` of the key stored in `.public-keys/<id>` against the key
in the local GPG keyring. If they differ by more than a negligible delta
(suggest: one day), emit a warning.

This requires parsing the armored public key from `.public-keys/` without
importing it, which can be done with `golang.org/x/crypto/openpgp` or
`github.com/ProtonMail/go-crypto/openpgp` (already an indirect dependency via
the age backend). Care must be taken not to introduce a new direct dependency
on a CGo package; `go-crypto` is pure Go.

---

## Rejected alternatives

**Return an error from `Set()` when a recipient is dropped:** This would be a
breaking behaviour change. Existing stores that happen to have stale recipient
entries (e.g. a team member who has left) would become unwritable. A warning
is the correct signal; the operator must decide whether to remove the recipient
or refresh the key.

**Block writes entirely when any recipient has no useable key:** Same objection
as above. A `--strict` flag could be added later if there is demand.

**Check every recipient on every read (Get):** Unnecessary overhead; expiry
affects the write path only.
