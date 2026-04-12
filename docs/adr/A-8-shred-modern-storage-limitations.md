# A-8: Shred Operation Is Ineffective on Modern Storage

**Status:** accepted — limitation documented; advisory notice to be added  
**Source:** SECURITY_AUDIT_REPORT.md § M-6

---

## Background

`pkg/fsutil/fsutil.go` provides a `Shred()` function that overwrites a file
with random bytes a configurable number of times before deleting it. The intent
is to prevent recovery of the plaintext after deletion.

---

## Why Shred Does Not Work on Modern Storage

The overwrite-before-delete approach relies on the assumption that writing to a
file path overwrites the same physical storage blocks each time. This assumption
has not held in practice for many years:

| Storage / FS type | Reason shred fails |
|-------------------|--------------------|
| SSD with wear levelling | Controller may write to different physical cells; old data remains until garbage-collected |
| ext4, NTFS, HFS+ (journaling) | Journal entries may contain copies of original data blocks |
| ZFS, Btrfs (copy-on-write) | Old snapshot blocks are never overwritten; new blocks are written alongside the old |
| APFS (macOS) | Copy-on-write; snapshots are created automatically by Time Machine |
| Network filesystems | Local overwrite does not affect server-side block allocation |

NIST SP 800-88 and academic literature (Gutmann, 1996; Wei et al., 2011)
confirm that software overwrite is unreliable on solid-state media.

---

## Current Mitigations Already in Place

gopass already uses the correct approach for the security-critical path:
plaintext is written to a **ramdisk** (macOS) or **`/dev/shm`** (Linux) for
editor sessions, so the sensitive data never reaches persistent storage where
shred would apply. The `Shred()` function is called on other files (e.g. old
encrypted files after re-encryption), where the data being removed is already
encrypted ciphertext. Shredding ciphertext provides minimal additional security
over simple deletion because the data is useless without the private key.

---

## Decision

Do not remove or deprecate `Shred()`. It provides a marginal additional layer
on rotating-disk (HDD) storage and its cost is low. However, it **must not be
presented to users as a guaranteed secure-deletion mechanism**.

Actions taken / to be taken:

1. Add a notice to the `Shred()` function's godoc comment explaining the
   limitation on SSDs and journaling/CoW filesystems.
2. If gopass ever presents a "securely deleted" user-facing message after
   calling `Shred()`, that message should be softened to "deleted" or include
   a caveat.
3. Sensitive plaintext must continue to be handled exclusively in ramdisk-backed
   temporary files (the existing behaviour) and never written to persistent
   disk in cleartext.

---

## References

- Gutmann, P. (1996). "Secure Deletion of Data from Magnetic and Solid-State Memory"
- Wei, M. et al. (2011). "Reliably Erasing Data from Flash-Based Solid State Drives" (FAST '11)
- NIST SP 800-88 Rev. 1: Guidelines for Media Sanitization
