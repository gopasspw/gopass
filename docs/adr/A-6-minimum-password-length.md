# A-6: Minimum Password Length Enforcement

**Status:** deferred — user autonomy preserved; warning to be added  
**Source:** SECURITY_AUDIT_REPORT.md § M-4

---

## Background

The `gopass generate` command accepts a length argument that is ultimately
stored in the `generate.length` config key or read from the
`GOPASS_PW_DEFAULT_LENGTH` environment variable. Currently no lower bound is
enforced below the generator's character-class minimum (which can be as low as
1 character). A user can therefore generate a single-character "password".

---

## Impact

Accidentally generating very short passwords provides the appearance of
security while offering none. Users who script gopass (e.g. in CI pipelines)
may silently configure trivially weak credentials.

---

## Decision

Enforcing an absolute minimum is a **policy decision** that gopass deliberately
avoids making for users. Some legitimate use cases require short codes (e.g.
4-digit device PINs, legacy system constraints). A hard cutoff would break
these workflows.

The chosen approach is:

1. Display a **warning** when the requested length is below 12 characters,
   making the risk visible without blocking the operation.
2. Do **not** impose a hard minimum that silently rejects user input.

---

## Implementation

In `internal/action/generate.go` (or wherever length is read and validated
before the generator is called), add:

```go
const warnBelowLength = 12

if length < warnBelowLength {
    out.Warningf(ctx, "Generating a password of only %d characters. This may be too weak for most uses.", length)
}
```

The warning should be visible in non-interactive mode as well so that scripted
invocations are not silently insecure.

No config key changes are required. Users who want to suppress the warning can
do so by setting a length ≥ 12 in their config.
