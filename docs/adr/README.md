# Architecture Decision Records

Each record states one architectural decision, the context it was made in, the
options considered, and the outcome. Amend a record only to update its status.
Never reuse a number: supersede a decision by writing a new record.

Name records `A-NN-<kebab-slug>.md`, with `NN` zero-padded to two digits. Write
the H1 as `# A-NN: <Title>`, matching the file name. Open the record with a
`**Status:**` line — `proposed`, `accepted`, `deferred`, `implemented`,
`partially implemented`, or `superseded by A-NN` — and a `**Source:**` line.
Update this index in the same commit that adds or supersedes a record.

## Index

| ADR | Title | Status | Added |
|---|---|---|---|
| [A-03](A-03-separate-storage-rcs.md) | Separate `Storage` and `RCS` interfaces | deferred | 2026-04-06 |
| [A-04](A-04-grep-match-error-counters.md) | Fix `grep` match and error counters | open | 2026-04-06 |
| [A-05](A-05-template-engine-text-vs-html.md) | Template engine uses `text/template` instead of `html/template` | deferred | 2026-04-06 |
| [A-06](A-06-minimum-password-length.md) | Minimum password length enforcement | deferred | 2026-04-06 |
| [A-07](A-07-hook-system-dead-code.md) | Hook system dead code and CVE-2023-24055 | deferred | 2026-04-06 |
| [A-08](A-08-shred-modern-storage-limitations.md) | Shred operation is ineffective on modern storage | accepted | 2026-04-06 |
| [A-09](A-09-low-severity-informational-findings.md) | Low severity and informational findings | accepted | 2026-04-06 |
| [A-10](A-10-code-quality-findings.md) | Code quality findings | open | 2026-04-06 |
| [A-11](A-11-secret-service.md) | Integrate `org.freedesktop.secrets` D-Bus service | proposed | 2026-05-24 |
| [A-12](A-12-pkg-api-stability.md) | `pkg/gopass` API stability contract | accepted | 2026-05-24 |
| [A-13](A-13-expired-gpg-key-handling.md) | Expired GPG key handling and recipient validity warnings | partially implemented | 2026-05-25 |
| [A-14](A-14-team-workflows.md) | Effortless team workflows | implemented | 2026-06-06 |
| [A-15](A-15-screenshot-build-tag.md) | `noscreenshot` build tag for OTP screen-capture feature | accepted | 2026-05-25 |

Status values are taken from each record's `**Status:**` line. Dates are the
authoring commit dates reported by `git log --diff-filter=A --follow`.

## Reserved and renumbered records

**A-01 and A-02 are reserved and have no records.** Both numbers are cited from
the `CHANGELOG.md` unreleased section: "Split Action handler into focused
handler types (A-1)" and "Replace context-key config system with typed structs
(A-2)". No file was written for either. Do not reuse these numbers; the
changelog citations would then point at unrelated decisions.

**A-15 was renumbered from A-13.** Two records carried the number A-13.
`A-13-expired-gpg-key-handling.md` keeps it: it is cited from
`docs/commands/recipients.md`, `docs/usecases/team-workflows.md`, and
`docs/adr/A-14-team-workflows.md`. The screenshot record had no inbound
citations and was renumbered to A-15.

**A-03 through A-09 were zero-padded.** Their file names previously used a
single digit, which sorts them after A-10 in any lexical listing.

## Unavailable sources

`SECURITY_AUDIT_REPORT.md` and `CODE_QUALITY_REPORT.md` are not present in the
working tree. Records A-03 through A-10 cite them in their `**Source:**` lines.
Both files were removed in commit `77894053` ("Clean up") and are recoverable
only from git history. The citations identify which finding each record answers,
for example "§ M-4", and are therefore retained.
